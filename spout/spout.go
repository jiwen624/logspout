package spout

import (
	"bufio"
	"bytes"
	"fmt"
	"math"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/jiwen624/logspout/utils"

	"github.com/jiwen624/logspout/gen"

	"github.com/pkg/errors"

	"github.com/jiwen624/logspout/config"
	"github.com/jiwen624/logspout/log"
	"github.com/jiwen624/logspout/output"
)

type Spout struct {
	// BurstMode defines if the logspout runs under burst mode, aka. it generates
	// logs without any `thinking` time.
	//
	// BurstMode will be deprecated in future, use MinInterval=MaxInterval=0
	// instead to achieve the same outcome.
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

	// MaxIntraTransactionLatency defines the maximum latency of a transaction (
	// I need a better name for it).
	MaxIntraTransactionLatency int

	// rawMsgs are the sample logs to be manipulated
	rawMsgs []string

	// Output defines the output destinations of the logs, which may be the console,
	// files or some message queues.
	// The output stored here may be active or inactive, and may be changed
	// on-the-fly.
	// TODO: currently the outputs are stored in a global registry
	Output *output.Registry

	// Pattern is a list of regular patterns that define the fields to be repalced
	// by policies defined in Replacement.
	// TODO:
	// Pattern []pattern.Pattern
	Pattern []string

	// Replacement defines the replacement policies for the fields extracted by
	// patterns defined in Pattern
	// TODO:
	// Replacers map[string]replacer.Replacer
	Replacers map[string]gen.Replacer

	// close is the indicator to close the spout
	close     chan struct{}
	closeOnce sync.Once
}

// Build reads the config from a SoutConfig object and build a Spout object.
func Build(cfg *config.SpoutConfig) (*Spout, error) {
	s := &Spout{close: make(chan struct{})}

	s.BurstMode = cfg.BurstMode
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
	if err := s.loadRawMessage(); err != nil {
		return nil, errors.Wrap(err, "build spout")
	}
	s.TransactionID = cfg.TransactionID
	s.MaxIntraTransactionLatency = cfg.MaxIntraTransactionLatency

	s.Output = output.RegistryFromConf(cfg.Output)

	// TODO: pattern, replacers
	// TODO: define a pattern struct and move it to that struct
	// TODO: support both Perl and PCRE
	var ptns []string
	for _, ptn := range cfg.Pattern {
		ptns = append(ptns, utils.ReConvert(ptn))
	}
	s.Pattern = ptns

	if len(s.rawMsgs) != len(s.Pattern) {
		return nil, fmt.Errorf("%d sample event(s) but %d pattern(s) found", len(s.rawMsgs), len(s.Pattern))
	}

	r, err := buildReplacerMap(cfg.Replacement)
	if err != nil {
		return nil, errors.Wrap(err, "build replacer")
	}
	s.Replacers = r

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

// Start kicks off the spout
func (s *Spout) Start() error {
	if err := s.StartAllOutput(); err != nil {
		return errors.Wrap(err, "logspout start")
	}

	go s.console()

	c := make(chan os.Signal, 10)
	signal.Notify(c,
		os.Interrupt,    // Ctrl-C
		syscall.SIGTERM, // kill
	)

	go s.sigHandler(c)

	log.Infof("LogSpout started with %d workers.", s.Concurrency)
	s.ProduceLogs()

	return nil
}

// Stop stops the spout
func (s *Spout) Stop() error {
	if err := s.StopAllOutputs(); err != nil {
		return errors.Wrap(err, "logspout stop")
	}

	log.Info("LogSpout ended")

	s.closeOnce.Do(func() {
		close(s.close)
	})
	return nil
}

func (s *Spout) sigHandler(c chan os.Signal) {
	for sig := range c {
		switch sig {
		case os.Interrupt:
			fallthrough
		case syscall.SIGTERM:
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
		// Use blank line as the delimiter of a log event.
		if vs = scanner.Text(); vs == "" {
			s.rawMsgs = append(s.rawMsgs, strings.TrimRight(buffer.String(), "\n"))
			buffer.Reset()
			continue
		}
		buffer.WriteString(scanner.Text())
		buffer.WriteString("\n") // Multi-line log support
	}

	if buffer.Len() != 0 {
		s.rawMsgs = append(s.rawMsgs, strings.TrimRight(buffer.String(), "\n"))
	}

	return nil
}

func (s *Spout) ProduceLogs() {
	// goroutine for future use, not necessary for now.
	var wg sync.WaitGroup

	wg.Add(s.Concurrency) // Add it before you start the goroutine.

	var matches = make([][]string, 0)
	var names = make([][]string, 0)

	for idx, ptn := range s.Pattern {
		re := regexp.MustCompile(ptn)
		matches = append(matches, re.FindStringSubmatch(s.rawMsgs[idx]))
		names = append(names, re.SubexpNames())

		if len(matches[idx]) == 0 {
			log.Errorf("the re pattern doesn't match the sample log in #%d", idx)
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
		go s.popNewLogs(matches, names, &wg, cCounter, resChan, i)
	}

	if s.Duration != 0 {
		select {
		case <-time.After(time.Second * time.Duration(s.Duration)):
			log.Debugf("Stopping logspout after: %v sec", s.Duration)
			// TODO: make sure it's closed only once
			close(s.close)
		}
	}

	wg.Wait()

}
