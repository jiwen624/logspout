package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStrIndex(t *testing.T) {
	strs := []string{"hello", "world", "hi"}
	assert.Equal(t, 0, StrIndex(strs, "hello"))
	assert.Equal(t, 1, StrIndex(strs, "world"))
	assert.Equal(t, 2, StrIndex(strs, "hi"))
	assert.Equal(t, -1, StrIndex(strs, "cannotseeme"))

	assert.Equal(t, -1, StrIndex(nil, "hi"))
	assert.Equal(t, -1, StrIndex([]string{}, "hi"))
	assert.Equal(t, -1, StrIndex([]string{"ho"}, "hi"))
}

func TestStrSlice2DCopy(t *testing.T) {
	src := [][]string{
		{"hello", "world"},
		{"HELLO", "WORLD"},
	}
	assert.Equal(t, src, StrSlice2DCopy(src))

	src = [][]string(nil)
	assert.Equal(t, src, StrSlice2DCopy(src))

	src = [][]string{
		{"hello", "world"},
		nil,
	}
	assert.Equal(t, src, StrSlice2DCopy(src))

	src = [][]string{
		{"hello", ""},
		nil,
	}
	assert.Equal(t, src, StrSlice2DCopy(src))

	src = [][]string{
		nil,
		nil,
	}
	assert.Equal(t, src, StrSlice2DCopy(src))
}
