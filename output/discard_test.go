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

func TestDiscard_Activate(t *testing.T) {
	d := Discard{}
	assert.Nil(t, d.Activate())
}

func TestDiscard_Deactivate(t *testing.T) {
	d := Discard{}
	assert.Nil(t, d.Deactivate())
}

func TestDiscard_Type(t *testing.T) {
	d := Discard{}
	assert.Equal(t, discard, d.Type())
}

func TestDiscard_String(t *testing.T) {
	d := Discard{}
	assert.Equal(t, "discard", d.String())
}

func TestDiscard_ID(t *testing.T) {
	d := Discard{}
	assert.Equal(t, id("discard"), d.ID())
}
