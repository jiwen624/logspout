package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/pkg/errors"
)

func TestCombineErrs(t *testing.T) {
	e1 := errors.New("hello")
	e2 := errors.New("world")

	var e3 error
	var e4 error

	assert.Nil(t, CombineErrs(nil))
	assert.Nil(t, CombineErrs([]error{e3, e4}))
	assert.Equal(t, CombineErrs([]error{e1, e3}).Error(), "hello")
	assert.Equal(t, CombineErrs([]error{e1, e2}).Error(), "hello\nworld")
}
