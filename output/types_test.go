package output

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTypesNum(t *testing.T) {
	l := len(Types())
	assert.Equal(t, int(upperbound)-int(unspecified)-1, l)
}
