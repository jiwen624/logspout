package replacer

import (
	"fmt"
	"strconv"
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

func TestStringReplacer(t *testing.T) {
	seed := "abcde"
	s := NewStringReplacer(seed, 1, 1)
	rng := NewTruncatedGaussian(0.5, 0.2)
	assert.NotNil(t, s)

	sCopy := s.Copy()
	assert.NotNil(t, sCopy)
	assert.Equal(t, sCopy, s)

	str, err := s.ReplacedValue(rng)
	assert.Len(t, str, 1)
	assert.Nil(t, err)
	assert.Contains(t, seed, str)

	s = NewStringReplacer(seed, 2, 4)
	for i := 0; i < 100; i++ {
		str, err := s.ReplacedValue(rng)
		assert.Nil(t, err)
		assert.Condition(t, func() bool {
			return (2 <= len(str)) && (len(str) <= 4)
		})
	}
}

func TestFloatReplacer(t *testing.T) {
	rng := NewTruncatedGaussian(0.5, 0.2)
	f := NewFloatReplacer(1, 2, 2)
	assert.NotNil(t, f)

	fCopy := f.Copy()
	assert.NotNil(t, fCopy)
	assert.Equal(t, fCopy, f)

	for i := 0; i < 100; i++ {
		flt, err := f.ReplacedValue(rng)
		assert.Nil(t, err)
		assert.Len(t, flt, 4)
		assert.Condition(t, func() bool {
			n, e := strconv.ParseFloat(flt, 32)
			assert.Nil(t, e)
			return (1 <= n) && (n <= 2)
		})
	}
}

func TestIntegerReplacer(t *testing.T) {
	rng := NewTruncatedGaussian(0.5, 0.2)

	ir := NewIntegerReplacer("random", 0, 100, 0)
	assert.NotNil(t, ir)
	irCopy := ir.Copy()
	assert.Equal(t, irCopy, ir)

	for i := 0; i < 100; i++ {
		vs, err := ir.ReplacedValue(rng)
		assert.Nil(t, err)
		v, e := strconv.Atoi(vs)
		assert.Nil(t, e)
		assert.Condition(t, func() bool {
			return v >= 0 && v <= 100
		}, fmt.Sprintf("Case: %d", i))
	}

	ir = NewIntegerReplacer("next", 0, 100, 0)
	assert.NotNil(t, ir)
	irCopy = ir.Copy()
	assert.Equal(t, irCopy, ir)

	for i := 0; i < 150; i++ {
		vs, err := ir.ReplacedValue(rng)
		assert.Nil(t, err)
		v, e := strconv.Atoi(vs)
		assert.Nil(t, e)
		assert.Equal(t, i%101, v, fmt.Sprintf("Case: %d", i))
	}

	ir = NewIntegerReplacer("prev", 0, 100, 100)
	assert.NotNil(t, ir)
	irCopy = ir.Copy()
	assert.Equal(t, irCopy, ir)

	for i := 100; i >= 0; i-- {
		vs, err := ir.ReplacedValue(rng)
		assert.Nil(t, err)
		v, e := strconv.Atoi(vs)
		assert.Nil(t, e)
		assert.Equal(t, i%101, v, fmt.Sprintf("Case: %d", i))
	}

}
