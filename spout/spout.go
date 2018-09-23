package spout

import (
	"encoding/json"

	"github.com/jiwen624/logspout/config"
	"github.com/jiwen624/logspout/output"
)

type Spout struct {
	// BurstMode defines if the logspout runs under burst mode, aka. it generates
	// logs without any `thinking` time.
	//
	// BurstMode will be deprecated in future, use MinInterval=MaxInterval=0
	// instead to achieve the same outcome.
	BurstMode bool

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
	Replacers json.RawMessage
}

// Build reads the config from a SoutConfig object and build a Spout object.
func Build(cfg *config.SpoutConfig) *Spout {
	s := &Spout{}

	s.BurstMode = cfg.BurstMode
	s.Duration = cfg.Duration
	s.MaxEvents = cfg.MaxEvents
	s.ConsolePort = cfg.ConsolePort
	s.Concurrency = cfg.Concurrency
	s.MinInterval = cfg.MinInterval
	s.MaxInterval = cfg.MaxInterval

	s.LogType = cfg.LogType
	s.SampleFilePath = cfg.SampleFilePath

	s.TransactionIDs = cfg.TransactionIDs
	s.MaxIntraTransactionLatency = cfg.MaxIntraTransactionLatency

	s.Output = output.RegistryFromConf(cfg.Output)

	// TODO: pattern, replacers
	s.Pattern = cfg.Pattern
	s.Replacers = cfg.Replacement
	return s
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
	return s.StartAllOutput()
}

// Stop stops the spout
func (s *Spout) Stop() error {
	return s.StopAllOutputs()
}
