package replacer

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRandomStr(t *testing.T) {
	seed := "abcde12345"

	for i := 0; i < 100; i++ {
		s := RandomStr(seed, i)
		assert.Equal(t, i, len(s), fmt.Sprintf("Got: %s", s))
		for _, c := range s {
			assert.True(t, strings.Contains(seed, string(c)), fmt.Sprintf("Got: %s", s))
		}
	}
}
