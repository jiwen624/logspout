// Package pattern defines the log match and replacement patterns. It can be either
// regular expressions or rules of fixed-length strings.
package pattern

type Pattern interface {
	// Matches parses the input string and returns the slice of matches sub-strings
	Matches(string) []string

	// Names returns the names of capture groups
	Names() []string
}
