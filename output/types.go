//go:generate jsonenums -type=Type
package output

type Type int

const (
	// To the console, which may be stdout or stderr
	console Type = iota

	// To a regular file
	file

	// To a syslog receiver
	syslog

	// To a kafka topic
	kafka

	// To ElasticSearch // TODO: not implemented
	es
)
