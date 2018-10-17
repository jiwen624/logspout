package replacer

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFixedListReplacer(t *testing.T) {
	// test constructor
	f := NewFixedListReplacer("invalid", nil, 100)
	assert.NotNil(t, f)
	assert.Equal(t, []string{""}, f.valRange)
	assert.Equal(t, 0, f.currIdx)

	valRange := []string{"one", "two"}
	f = NewFixedListReplacer("next", valRange, 0)
	assert.NotNil(t, f)

	// test copy
	fc := f.Copy()
	assert.Equal(t, fc, f)

	fp := fmt.Sprintf("%p", &f.valRange)

	fcp := fmt.Sprintf("%p", &fc.(*FixedListReplacer).valRange)
	assert.NotEqual(t, fp, fcp)

	// test replace with policy `next`
	f = NewFixedListReplacer("next", valRange, 0)
	r, err := f.ReplacedValue(NewTruncatedGaussian(0.5, 0.2))
	assert.Nil(t, err)
	assert.Equal(t, valRange[1], r)

	r, err = f.ReplacedValue(NewTruncatedGaussian(0.5, 0.2))
	assert.Nil(t, err)
	assert.Equal(t, valRange[0], r)

	// test replace with policy `random`
	f = NewFixedListReplacer("random", valRange, 0)
	r, err = f.ReplacedValue(NewTruncatedGaussian(0.5, 0.2))
	assert.Nil(t, err)
	assert.Contains(t, valRange, r)

	r, err = f.ReplacedValue(NewTruncatedGaussian(0.5, 0.2))
	assert.Nil(t, err)
	assert.Contains(t, valRange, r)
}
