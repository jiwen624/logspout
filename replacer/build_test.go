package replacer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuild(t *testing.T) {
	rawMsg := []byte(`{
    "timestamp": {
      "type": "timestamp",
      "attrs": {
        "format": "MMM dd, yyyy hh:mm:ss.SSS a z"
      }
    },
    "severity": {
      "type": "fixedList",
      "attrs": {
        "method": "random",
        "listFile": "../examples/severity.sample"
      }
    },
    "subsystem": {
      "type": "float",
      "attrs": {
        "min": 100,
        "max": 10000,
        "precision": 20
      }
    },
    "ipaddress": {
      "type": "looksReal",
      "attrs": {
        "method": "ipv4China"
      }
    },
    "phone": {
      "type": "looksReal",
      "attrs": {
        "method": "cellphoneChina"
      }
    },
    "users": {
      "type": "fixedList",
      "attrs": {
        "method": "random",
        "list": [
          "GuoJing",
          "HuangRong",
          "ZhangSanfeng"
        ]
      }
    },
    "thread": {
      "type": "integer",
      "attrs": {
        "method": "next",
        "min": 1,
        "max": 100
      }
    },
    "transaction": {
      "type": "string",
      "attrs": {
        "chars": "abcde12345",
        "min": 10,
        "max": 20
      }
    },
    "msgid": {
      "type": "integer",
      "attrs": {
        "method": "random",
        "min": 0,
        "max": 2000000
      }
    }
  }`)

	r, err := Build(rawMsg)
	assert.Nil(t, err)
	assert.NotNil(t, r)

	tr, ok := r["timestamp"].(*TimeStampReplacer)
	assert.True(t, ok)
	assert.Equal(t, "MMM dd, yyyy hh:mm:ss.SSS a z", tr.format)

	fl, ok := r["severity"].(*FixedListReplacer)
	assert.True(t, ok)
	assert.Equal(t, RANDOM, fl.method)
	assert.Equal(t, []string{"Info", "Warning", "Error", "Debug"}, fl.valRange)

	fr, ok := r["subsystem"].(*FloatReplacer)
	assert.True(t, ok)
	assert.Equal(t, float64(100), fr.min)
	assert.Equal(t, float64(10000), fr.max)
	assert.Equal(t, int64(20), fr.precision)

	lr, ok := r["ipaddress"].(*LooksReal)
	assert.True(t, ok)
	assert.Equal(t, IPV4CHINA, lr.method)

	lr, ok = r["phone"].(*LooksReal)
	assert.True(t, ok)
	assert.Equal(t, CELLPHONECHINA, lr.method)

	fl, ok = r["users"].(*FixedListReplacer)
	assert.True(t, ok)
	assert.Equal(t, RANDOM, fl.method)
	assert.Equal(t, []string{"GuoJing", "HuangRong", "ZhangSanfeng"}, fl.valRange)

	ir, ok := r["thread"].(*IntegerReplacer)
	assert.True(t, ok)
	assert.Equal(t, NEXT, ir.method)
	assert.Equal(t, int64(1), ir.min)
	assert.Equal(t, int64(100), ir.max)

	sr, ok := r["transaction"].(*StringReplacer)
	assert.True(t, ok)
	assert.Equal(t, "abcde12345", sr.chars)
	assert.Equal(t, int64(10), sr.min)
	assert.Equal(t, int64(20), sr.max)

	ir, ok = r["msgid"].(*IntegerReplacer)
	assert.True(t, ok)
	assert.Equal(t, RANDOM, ir.method)
	assert.Equal(t, int64(0), ir.min)
	assert.Equal(t, int64(2000000), ir.max)

	// errors
	r, err = Build(nil)
	assert.NotNil(t, err)
	assert.Nil(t, r)
}
