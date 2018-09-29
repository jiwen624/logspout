package spout

import (
	"bufio"
	"bytes"
	"fmt"
	"math"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/pkg/errors"

	"github.com/jiwen624/logspout/config"
	"github.com/jiwen624/logspout/log"
	"github.com/jiwen624/logspout/output"
	"github.com/jiwen624/logspout/pattern"
	"github.com/jiwen624/logspout/replacer"
	"github.com/jiwen624/logspout/utils"
)

type Spout struct {
	// BurstMode defines if the logspout runs under burst mode, aka. it generates
	// logs without any `thinking` time.
	//
	// Deprecated: BurstMode will be deprecated in a future version, use
	// MinInterval=MaxInterval=0 instead to achieve the same outcome.
	BurstMode bool

	// UniformLoad means the workload is uniform
	UniformLoad bool

	// Duration means how long the logspout program will run for (in seconds)
	Duration int

	// MaxEvents means the maximum number of events logspout will generate
	MaxEvents int

	// ConsolePort specifies the port for management console. The console is
	// disabled if the port is 0 (the default value)
	ConsolePort int

	// Concurrency defines the number of workers to generate logs concurrently.
	Concurrency int

	// MinInterval is the minimum interval between two log entries.
	MinInterval int

	// MaxInterval is the maximum interval between two log entries.
	MaxInterval int

	// LogType defines the type of the logs, e.g., the application name.
	LogType string

	// SampleFilePath is the file name and path where Logspout fetch sample logs
	// from.
	SampleFilePath string

	// TransactionID defines the transaction IDs for transaction mode. Under
	// transaction mode, if a certain number of logs have the same value in these
	// keys, they form a transaction.
	TransactionID []string

	// MaxIntraTransLat defines the maximum latency of a transaction (
	// I need a better name for it).
	MaxIntraTransLat int

	// seedLogs are the sample logs to be manipulated
	seedLogs []string

	// Output defines the output destinations of the logs, which may be the console,
	// files or some message queues.
	// The output stored here may be active or inactive, and may be changed
	// on-the-fly.
	Output *output.Registry

	// Patterns is a list of regular patterns that define the fields to be repalced
	// by policies defined in Replacement.
	Patterns []pattern.Pattern

	// Replacement defines the replacement policies for the fields extracted by
	// patterns defined in Patterns
	Replacers map[string]replacer.Replacer

	// close is the indicator to close the spout
	close     chan struct{}
	closeOnce sync.Once

	// Used to coordinate between the main goroutine and the workers
	sync.WaitGroup
}

func NewDefault() *Spout {
	return &Spout{close: make(chan struct{})}
}

// Build reads the config from a SoutConfig object and build a Spout object.
func Build(cfg *config.SpoutConfig) (*Spout, error) {
	s := NewDefault()

	s.BurstMode = cfg.BurstMode
	s.UniformLoad = cfg.UniformLoad
	s.Duration = cfg.Duration

	s.MaxEvents = cfg.MaxEvents
	if s.MaxEvents == 0 {
		s.MaxEvents = int(math.MaxInt32)
	}

	s.ConsolePort = cfg.ConsolePort
	s.Concurrency = cfg.Concurrency
	s.MinInterval = cfg.MinInterval
	s.MaxInterval = cfg.MaxInterval
	s.LogType = cfg.LogType
	s.SampleFilePath = cfg.SampleFilePath
	s.TransactionID = cfg.TransactionID
	s.MaxIntraTransLat = cfg.MaxIntraTransactionLatency

	if err := s.loadRawMessage(); err != nil {
		return nil, errors.Wrap(err, "build spout")
	}

	op, err := output.RegistryFromConf(cfg.Output)
	if err != nil {
		log.Warn(errors.Wrap(err, "build logspout"))
		if op.Size() == 0 {
			return nil, errors.Wrap(err, "no valid output")
		}
	}
	s.Output = op

	for _, ptn := range cfg.Pattern {
		s.Patterns = append(s.Patterns, pattern.New(ptn))
	}

	r, err := replacer.Build(cfg.Replacement)
	if err != nil {
		return nil, errors.Wrap(err, "build replacer")
	}
	s.Replacers = r

	return s.SanityCheck()
}

// TODO: add more checks
func (s *Spout) SanityCheck() (*Spout, error) {
	var errs []error
	rl := len(s.seedLogs)
	pl := len(s.Patterns)
	if rl != pl {
		e := fmt.Errorf("%d seed log(s) but %d pattern(s) found", rl, pl)
		errs = append(errs, e)
		return nil, utils.CombineErrs(errs)
	}
	return s, nil
}

// StartAllOutput starts all the outputs
func (s *Spout) StartAllOutput() error {
	return s.Output.ForAll(func(o output.Output) error {
		return o.Activate()
	})
}

// StopAllOutput stops all the outputs
func (s *Spout) StopAllOutputs() error {
	return s.Output.ForAll(func(o output.Output) error {
		return o.Deactivate()
	})
}

// SigMon registers the signal handler and run the handler in a standalone goroutine.
func (s *Spout) SigMon() {
	c := make(chan os.Signal, 10)
	signal.Notify(c,
		os.Interrupt,    // Ctrl-C
		syscall.SIGTERM, // kill
	)

	go s.sigHandler(c)
}

// StartConsole starts the management console in a standalone goroutine.
func (s *Spout) StartConsole() {
	go s.console()
}

// Start kicks off the spout
func (s *Spout) Start() error {
	s.StartConsole()
	s.SigMon()
	s.ProduceLogs()

	return nil
}

// Stop stops the spout
func (s *Spout) Stop() {
	s.closeOnce.Do(func() {
		log.Info("LogSpout is closing.")

		if err := s.StopAllOutputs(); err != nil {
			log.Error(errors.Wrap(err, "logspout stop"))
		}
		close(s.close)
	})
}

func (s *Spout) sigHandler(c chan os.Signal) {
	for sig := range c {
		switch sig {
		case os.Interrupt:
			fallthrough
		case syscall.SIGTERM:
			// Don't call os.Exit(), wait until all workers are closed.
			s.Stop()
		default:
			err := fmt.Errorf("unhandled signal: %v", sig)
			fmt.Fprintln(os.Stderr, err)
		}
	}
}

func (s *Spout) loadRawMessage() error {
	file, err := os.Open(s.SampleFilePath)
	if err != nil {
		return errors.Wrap(err, "loadRawMessage")
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	var buffer bytes.Buffer
	var vs string
	for scanner.Scan() {
		vs = scanner.Text()
		// Use blank line as the delimiter of a log event.
		if vs == "" {
			if buffer.Len() == 0 {
				continue
			}
			s.seedLogs = append(s.seedLogs, buffer.String())
			buffer.Reset()
			continue
		}
		buffer.WriteString(vs)
		buffer.WriteString("\n") // Multi-line log support
	}

	if buffer.Len() != 0 {
		s.seedLogs = append(s.seedLogs, buffer.String())
	}

	return nil
}

func (s *Spout) ProduceLogs() {
	s.Add(s.Concurrency) // Add it before you start the goroutine.

	var matches = make([][]string, 0)
	var names = make([][]string, 0)

	for idx, ptn := range s.Patterns {
		matches = append(matches, ptn.Matches(s.seedLogs[idx]))
		names = append(names, ptn.Names())

		if len(matches[idx]) == 0 {
			log.Errorf("pattern doesn't match sample log in #%d", idx)
			return
		}

		// Remove the first one as it is the whole string.
		matches[idx] = matches[idx][1:]
		names[idx] = names[idx][1:]
	}

	for idx, match := range matches {
		log.Debugf("   pattern #%d", idx)
		for i, group := range match {
			log.Debugf("       - %s: %s", names[idx][i], group)
		}
	}

	for i := 0; i < s.Concurrency; i++ {
		go s.popNewLogs(matches, names, cCounter, resChan, i)
	}

	log.Infof("LogSpout started with %d workers.", s.Concurrency)

	if s.Duration != 0 {
		select {
		case <-time.After(time.Second * time.Duration(s.Duration)):
			log.Debugf("Stopping logspout after: %v sec", s.Duration)
			s.Stop()
		case <-s.close:
		}
	}

	s.Wait()

}
