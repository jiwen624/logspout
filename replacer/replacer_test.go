package replacer

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

type DumbReplacer struct {
	val string
}

func (d DumbReplacer) ReplacedValue(RandomGenerator) (string, error) {
	return d.val, nil
}

func (d DumbReplacer) Copy() Replacer {
	return &DumbReplacer{d.val}
}

func TestReplacers_Copy(t *testing.T) {
	m := make(map[string]Replacer)

	for i := 0; i < 10; i++ {
		m[strconv.Itoa(i)] = &DumbReplacer{strconv.Itoa(i)}
	}
	r := Replacers(m)
	rCopy := r.Copy()

	assert.NotNil(t, rCopy)
	assert.Len(t, rCopy, 10)
}
