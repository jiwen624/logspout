package replacer

import (
	"encoding/json"
	"fmt"
	"net"
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

func TestLooksReal(t *testing.T) {
	rng := NewTruncatedGaussian(0.5, 0.2)

	// IPv4 address
	lr := NewLooksReal(IPV4, nil)
	assert.NotNil(t, lr)
	ip, err := lr.ReplacedValue(rng)
	assert.Nil(t, err)
	assert.NotNil(t, net.ParseIP(ip))

	// unknown format
	lr = NewLooksReal("Unknown", nil)
	assert.NotNil(t, lr)
	ip, err = lr.ReplacedValue(rng)
	assert.NotNil(t, err)
	assert.Equal(t, "", ip)

	// IPv4 China address
	lr = NewLooksReal(IPV4CHINA, nil)
	assert.NotNil(t, lr)
	ip, err = lr.ReplacedValue(rng)
	assert.Nil(t, err)
	assert.NotNil(t, net.ParseIP(ip))

	// IPv6 address
	lr = NewLooksReal(IPV6, nil)
	assert.NotNil(t, lr)
	ip, err = lr.ReplacedValue(rng)
	assert.Nil(t, err)
	ipv6 := net.ParseIP(ip)
	assert.NotNil(t, ipv6)
	assert.Nil(t, ipv6.To4())

	// User Agent
	lr = NewLooksReal(UA, nil)
	assert.NotNil(t, lr)
	ua, err := lr.ReplacedValue(rng)
	assert.Nil(t, err)
	assert.Contains(t, ua, "Mozilla")

	// Country
	lr = NewLooksReal(COUNTRY, nil)
	assert.NotNil(t, lr)
	c, err := lr.ReplacedValue(rng)
	assert.Nil(t, err)
	assert.NotEmpty(t, c)

	// Email
	lr = NewLooksReal(EMAIL, nil)
	assert.NotNil(t, lr)
	e, err := lr.ReplacedValue(rng)
	assert.Nil(t, err)
	assert.Contains(t, e, "@")

	// Name
	lr = NewLooksReal(COUNTRY, nil)
	assert.NotNil(t, lr)
	n, err := lr.ReplacedValue(rng)
	assert.Nil(t, err)
	assert.NotEmpty(t, n)

	// Chinese name
	lr = NewLooksReal(CHINESENAME, nil)
	assert.NotNil(t, lr)
	n, err = lr.ReplacedValue(rng)
	assert.Nil(t, err)
	assert.NotEmpty(t, n)

	// Name
	lr = NewLooksReal(NAME, nil)
	assert.NotNil(t, lr)
	n, err = lr.ReplacedValue(rng)
	assert.Nil(t, err)
	assert.NotEmpty(t, n)

	// China cellphone
	lr = NewLooksReal(CELLPHONECHINA, nil)
	assert.NotNil(t, lr)
	cell, err := lr.ReplacedValue(rng)
	assert.Nil(t, err)
	assert.NotEmpty(t, cell)

	// Mac address
	lr = NewLooksReal(MAC, nil)
	assert.NotNil(t, lr)
	mac, err := lr.ReplacedValue(rng)
	assert.Nil(t, err)
	hw, err := net.ParseMAC(mac)
	assert.NotNil(t, hw)
	assert.Nil(t, err)

	// UUID
	lr = NewLooksReal(UUID, nil)
	assert.NotNil(t, lr)
	id, err := lr.ReplacedValue(rng)
	assert.Nil(t, err)
	assert.NotEmpty(t, id)

	// JSON
	lr = NewLooksReal(JSON, map[string]interface{}{
		MAXDEPTH:    1,
		MAXELEMENTS: 1,
		TAGSEED:     []string{"hello", "world"},
	})
	assert.NotNil(t, lr)
	json, err := lr.ReplacedValue(rng)
	assert.Nil(t, err)
	assert.True(t, isJSON(json))

	// JSON with invalid parameter
	// TODO: disable this case temporarily as the xml/json needs to be refactored.
	// lr = NewLooksReal(JSON, map[string]interface{}{
	// 	MAXDEPTH:    -1,
	// 	MAXELEMENTS: 1,
	// 	TAGSEED:     []string{"hello", "world"},
	// })
	// assert.NotNil(t, lr)
	// json, err = lr.ReplacedValue(rng)
	// assert.Nil(t, err)
	// assert.Equal(t, "", json)

	// XML
	lr = NewLooksReal(XML, map[string]interface{}{
		MAXDEPTH:    1,
		MAXELEMENTS: 1,
		TAGSEED:     []string{"hello", "world"},
	})
	assert.NotNil(t, lr)
	xml, err := lr.ReplacedValue(rng)
	assert.Nil(t, err)
	assert.True(t, isXML(xml))

	lrCopy := lr.Copy()
	assert.Equal(t, lrCopy, lr)

	// XML: missing mandatory parameter
	lr = NewLooksReal(XML, map[string]interface{}{
		MAXDEPTH:    1,
		MAXELEMENTS: 1,
	})
	xml, err = lr.ReplacedValue(rng)
	assert.NotNil(t, err)
	assert.Empty(t, xml)

	// XML: invalid mandatory parameter
	// TODO: disable it temporarily
	// lr = NewLooksReal(XML, map[string]interface{}{
	// 	MAXDEPTH:    1,
	// 	MAXELEMENTS: -1,
	// 	TAGSEED:     []string{"hello", "world"},
	// })
	// xml, err = lr.ReplacedValue(rng)
	// assert.NotNil(t, err)
	// assert.Empty(t, xml)
}

func isJSON(s string) bool {
	j := json.RawMessage{}
	return json.Unmarshal([]byte(s), &j) == nil
}

func isXML(s string) bool {
	// TODO
	return s != ""
}
