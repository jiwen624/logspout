package pattern

import (
	"regexp"

	"github.com/jiwen624/logspout/utils"
)

type Re struct {
	re *regexp.Regexp
}

func New(pattern string) *Re {
	// TODO: support pcre with auto detection/fallback
	// TODO: support fixed arrays
	ptn := utils.ReConvert(pattern)
	return &Re{re: regexp.MustCompile(ptn)}
}

func (r *Re) Matches(s string) []string {
	return r.re.FindStringSubmatch(s)
}

func (r *Re) Names() []string {
	return r.re.SubexpNames()
}
