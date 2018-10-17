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

	valRange := []string{"one", "two", "three"}
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
	assert.Equal(t, valRange[2], r)

	// test replace with policy `random`
	f = NewFixedListReplacer("random", valRange, 0)
	r, err = f.ReplacedValue(NewTruncatedGaussian(0.5, 0.2))
	assert.Nil(t, err)
	assert.Contains(t, valRange, r)

	r, err = f.ReplacedValue(NewTruncatedGaussian(0.5, 0.2))
	assert.Nil(t, err)
	assert.Contains(t, valRange, r)

	// test replace with policy `prev`
	f = NewFixedListReplacer("prev", valRange, 0)
	r, err = f.ReplacedValue(NewTruncatedGaussian(0.5, 0.2))
	assert.Nil(t, err)
	assert.Equal(t, valRange[2], r)

	r, err = f.ReplacedValue(NewTruncatedGaussian(0.5, 0.2))
	assert.Nil(t, err)
	assert.Equal(t, valRange[1], r)
}

func TestTimeStampReplacer(t *testing.T) {
	ts := NewTimeStampReplacer("MMM dd, yyyy hh:mm:ss.SSS a z")
	assert.NotNil(t, ts)

	tsc := ts.Copy().(*TimeStampReplacer)
	assert.NotNil(t, tsc)
	assert.Equal(t, tsc.format, ts.format)

	time, err := ts.ReplacedValue(NewTruncatedGaussian(0.5, 0.2))
	assert.NotEmpty(t, time)
	assert.Nil(t, err)
}
