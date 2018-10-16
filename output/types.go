//go:generate jsonenums -type=Type
package output

type Type int

const (
	// the default type is unspecified
	unspecified Type = iota

	// To the console, which may be stdout or stderr
	console

	// To a regular file
	file

	// To a syslog receiver
	syslog

	// To a kafka topic
	kafka

	// To ElasticSearch // TODO: not implemented
	// es

	// To /dev/null
	discard

	// the upper bound of the types enumeration
	upperbound
)

// Types returns all output types
func Types() []Type {
	return []Type{console, file, syslog, kafka, discard}
}
