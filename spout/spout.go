package spout

import (
	"bufio"
	"bytes"
	"fmt"
	"math"
	"os"
	"sync"
	"time"

	"github.com/jiwen624/logspout/console"

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

	// ConsolePort specifies the port for management console.
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

	// Patterns is a list of regular patterns that define the fields to be replaced
	// by policies defined in Replacement.
	Patterns []pattern.Pattern

	// Replacement defines the replacement policies for the fields extracted by
	// patterns defined in Patterns.
	//
	// This field is not concurrent-safe. Each worker should obtain its own copy
	// of the spout replacers.
	Replacers replacer.Replacers

	// close is the indicator to close the spout
	close     chan struct{}
	closeOnce sync.Once

	// Used to coordinate between the main goroutine and the workers
	sync.WaitGroup
}

const (
	defaultConsolePort = 10306
)

func NewDefault() *Spout {
	return &Spout{close: make(chan struct{})}
}

// setOrFallback returns val if val doesn't equal to init (the initial value),
// otherwise it returns the fallback value
func setOrFallback(val int, init int, fallback int) int {
	if val == init {
		return fallback
	}
	return val
}

// Build reads the config from a SoutConfig object and build a Spout object.
func Build(cfg *config.SpoutConfig) (*Spout, error) {
	s := NewDefault()

	s.BurstMode = cfg.BurstMode
	s.UniformLoad = cfg.UniformLoad
	s.Duration = cfg.Duration

	s.MaxEvents = setOrFallback(cfg.MaxEvents, 0, int(math.MaxInt32))
	s.ConsolePort = setOrFallback(cfg.ConsolePort, 0, defaultConsolePort)

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

// StartAllOutputs starts all the outputs
func (s *Spout) StartAllOutputs() error {
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

// StartConsole starts the management console in a standalone goroutine.
func (s *Spout) StartConsole() {
	host := fmt.Sprintf("localhost:%d", s.ConsolePort)
	console.Start(host)
}

// Start kicks off the spout
func (s *Spout) Start() error {
	if err := s.StartAllOutputs(); err != nil {
		return errors.Wrap(err, "start failed")
	}

	s.StartConsole()
	s.SigMon()

	return s.ProduceLogs()
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

// ProduceLogs generate logs based on the replacers configuration. It will block
// until the maximum number of events is reached, or timeout, or the program is
// stopped by Ctrl-C
func (s *Spout) ProduceLogs() error {

	matches, names, err := s.GenerateTokens()
	if err != nil {
		return errors.Wrap(err, "ProduceLogs")
	}

	s.StartWorkers(matches, names)
	s.WaitForWorkers()

	return nil
}

// Spray sprays the generated logs into the predefined destinations.
func (s *Spout) Spray(log string) error {
	return s.Output.Write(log)
}

// GenerateTokens matches the seed logs with the patterns and generate
// the tokens.
func (s *Spout) GenerateTokens() ([][]string, [][]string, error) {
	var matches = make([][]string, 0)
	var names = make([][]string, 0)

	for idx, ptn := range s.Patterns {
		matches = append(matches, ptn.Matches(s.seedLogs[idx]))
		names = append(names, ptn.Names())

		if len(matches[idx]) == 0 {
			return nil, nil, fmt.Errorf("#%d: unmatched pattern and logs", idx)
		}

		// Remove the first one as it is the whole string.
		matches[idx] = matches[idx][1:]
		names[idx] = names[idx][1:]
	}

	debugPrintPatterns(matches, names)
	return matches, names, nil
}

// StartWorkers creates and starts all workers.
func (s *Spout) StartWorkers(matches [][]string, names [][]string) {
	s.Add(s.Concurrency) // Add them before you start the goroutines.

	for i := 0; i < s.Concurrency; i++ {
		w := NewWorker(workerConfig{
			Index:            i,
			MaxEvents:        int(s.MaxEvents / s.Concurrency),
			Seconds:          s.Duration,
			Replacers:        s.Replacers.Copy(),
			TransIDs:         s.TransactionID,
			SeedLogs:         s.seedLogs,
			MinInterval:      s.MinInterval,
			MaxInterval:      s.MaxInterval,
			UniformLoad:      s.UniformLoad,
			MaxIntraTransLat: s.MaxIntraTransLat,
			WriteTo:          s.Spray,
			DoneCallback:     s.Done,
			CloseChan:        s.close,
			BurstMode:        s.BurstMode,
		})
		go w.start(matches, names, i)
	}

	log.Infof("LogSpout started with %d workers.", s.Concurrency)
}

// WaitForWorkers monitors the timer, stop all workers then and wait for them
// before exiting.
func (s *Spout) WaitForWorkers() {
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

func debugPrintPatterns(matches, names [][]string) {
	if !log.GetLevel().Printable() {
		return
	}

	for idx, match := range matches {
		log.Debugf("   pattern #%d", idx)
		for i, group := range match {
			log.Debugf("       - %s: %s", names[idx][i], group)
		}
	}
}
