package spout

import (
	"fmt"

	"github.com/jiwen624/logspout/utils"

	"github.com/jiwen624/logspout/config"
	"github.com/jiwen624/logspout/output"
	"github.com/jiwen624/logspout/pattern"
	"github.com/jiwen624/logspout/replacer"
	"github.com/pkg/errors"
)

type Spout struct {
	// BurstMode defines if the logspout runs under burst mode, aka. it generates
	// logs without any `thinking` time.
	//
	// BurstMode will be deprecated in future, use MinInterval=MaxInterval=0
	// instead to achieve the same outcome.
	BurstMode bool

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

	// TransactionIDs defines the transaction IDs for transaction mode. Under
	// transaction mode, if a certain number of logs have the same value in these
	// keys, they form a transaction.
	TransactionIDs []string

	// MaxIntraTransactionLatency defines the maximum latency of a transaction (
	// I need a better name for it).
	MaxIntraTransactionLatency int

	// Output defines the output destinations of the logs, which may be the console,
	// files or some message queues.
	// The output stored here may be active or inactive, and may be changed
	// on-the-fly.
	// TODO: currently the outputs are stored in a global registry
	Output map[string]output.Output

	// Pattern is a list of regular patterns that define the fields to be repalced
	// by policies defined in Replacement.
	Pattern []pattern.Pattern

	// Replacement defines the replacement policies for the fields extracted by
	// patterns defined in Pattern
	Replacers map[string]replacer.Replacer
}

// Build reads the config from a SoutConfig object and build a Spout object.
func Build(cfg *config.SpoutConfig) *Spout {
	s := &Spout{}

	s.BurstMode = cfg.BurstMode
	s.Concurrency = cfg.Concurrency
	s.MinInterval = cfg.MinInterval
	s.MaxInterval = cfg.MaxInterval

	s.LogType = cfg.LogType
	s.SampleFilePath = cfg.SampleFilePath

	s.TransactionIDs = cfg.TransactionIDs
	s.MaxIntraTransactionLatency = cfg.MaxIntraTransactionLatency

	s.Output = output.BuildOutputMap(cfg.Output)
	// TODO: pattern, replacers
	return s
}

func (s *Spout) touchOuput(name string, apply output.Apply) error {
	o, ok := s.Output[name]
	if !ok {
		return errors.Wrap(output.ErrNotFound,
			fmt.Sprintf("touchOutput:%s", name))
	}
	apply(o)
	return nil
}

func (s *Spout) touchAllOutput(apply output.Apply) []error {
	var errs []error

	for name := range s.Output {
		if err := s.touchOuput(name, apply); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

// StartOutput starts an output by its name
func (s *Spout) StartOutput(name string) error {
	return s.touchOuput(name, output.Register)
}

// StopOutput stops an output by its name
func (s *Spout) StopOutput(name string) error {
	return s.touchOuput(name, output.Unregister)
}

// StartAllOutput starts all the outputs
func (s *Spout) StartAllOutput() []error {
	return s.touchAllOutput(output.Register)
}

// StopAllOutput stops all the outputs
func (s *Spout) StopAllOutputs() []error {
	return s.touchAllOutput(output.Unregister)
}

// Start kicks off the spout
func (s *Spout) Start() error {
	return utils.CombineErrs(s.StartAllOutput())
}

// Stop stops the spout
func (s *Spout) Stop() error {
	return utils.CombineErrs(s.StopAllOutputs())
}
