package pattern

import (
	"regexp"
	"strings"
)

// Re is a struct which wraps the Regexp and implements Pattern interface
type Re struct {
	re *regexp.Regexp
}

// New returns a new Re object
func New(pattern string) *Re {
	// TODO: support pcre with auto detection/fallback
	// TODO: support fixed arrays
	ptn := reConvert(pattern)
	return &Re{re: regexp.MustCompile(ptn)}
}

// Matches parses the input string and returns the slice of matches sub-strings
func (r *Re) Matches(s string) []string {
	return r.re.FindStringSubmatch(s)
}

// Names returns the names of capture groups
func (r *Re) Names() []string {
	return r.re.SubexpNames()
}

// reConvert does the pre-process of the regular expression literal.
// It does the following things:
// 1. Remove P before captured group names
// 2. Add parenthesises to the other parts of the log event string.
func reConvert(ptn string) string {
	s := strings.Split(ptn, "")
	for i := 1; i < len(s)-1; i++ {
		if s[i] == "?" && s[i-1] == "(" && s[i+1] == "<" {
			// Replace "?" with "?P", it has a bug but works for 99% of the cases.
			// TODO: I'll keep it before I have time to write a better one.
			s[i] = "?P"
		}
	}
	return strings.Join(s, "")
}
