package pattern

type Pattern interface {
	// Matches parses the input string and returns the slice of matches sub-strings
	Matches(string) []string

	// Names returns the names of capture groups
	Names() []string
}
