package output

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDiscard_Write(t *testing.T) {
	d := Discard{}
	n, err := d.Write(nil)
	assert.Equal(t, 0, n)
	assert.Nil(t, err)

	n, err = d.Write([]byte{'a'})
	assert.Equal(t, 1, n)
	assert.Nil(t, err)
}
