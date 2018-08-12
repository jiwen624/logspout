// Package config defines the basic configuration struct. The struct has explicit
// fields for top-level configuration items, but keep the inner layers as raw
// json strings, which will be consumed by the Logspout initializer.
package config

// SpoutConfig is the basic struct of Logspout configuration. It handles the
// marshalling/unmarshalling of the configuration file.
type SpoutConfig struct {
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
	// files or some message queues
	Output map[string]interface{}

	// Pattern is a list of regular patterns that define the fields to be repalced
	// by policies defined in Replacement.
	Pattern []string

	// Replacement defines the replacement policies for the fields extracted by
	// patterns defined in Pattern
	Replacement map[string]interface{}
}
