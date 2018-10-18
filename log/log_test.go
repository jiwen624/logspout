package log

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetLevel(t *testing.T) {
	assert.NotNil(t, SetLevel("invalid"))
	assert.Nil(t, SetLevel("info"))
	assert.Equal(t, INFO, GetLevel())
	assert.Equal(t, "info", ToString(INFO))
	assert.False(t, DEBUG.Printable())
}

func TestToString(t *testing.T) {
	assert.Equal(t, "debug", ToString(DEBUG))
	assert.Equal(t, "info", ToString(INFO))
	assert.Equal(t, "warn", ToString(WARN))
	assert.Equal(t, "error", ToString(ERROR))
	assert.Equal(t, "fatal", ToString(FATAL))
	assert.Equal(t, "invalid", ToString(Level(100)))
}

func TestToLevel(t *testing.T) {
	l, err := ToLevel("debug")
	assert.Nil(t, err)
	assert.Equal(t, DEBUG, l)

	l, err = ToLevel("INFO")
	assert.Nil(t, err)
	assert.Equal(t, INFO, l)

	l, err = ToLevel("waRN")
	assert.Nil(t, err)
	assert.Equal(t, WARN, l)

	l, err = ToLevel("hello")
	assert.NotNil(t, err)
	assert.Equal(t, DEBUG, l)
}
