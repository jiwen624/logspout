package output

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMarshal(t *testing.T) {
	o := console
	b, err := json.Marshal(o)
	assert.Nil(t, err)
	assert.Equal(t, []byte("\"console\""), b)

	// receiver & method
	o = file
	b, err = o.MarshalJSON()
	assert.Nil(t, err)
	assert.Equal(t, []byte("\"file\""), b)

	// invalid type
	o = Type(upperbound + 1)
	b, err = json.Marshal(o)
	assert.NotNil(t, err)
}

func TestUnMarshal(t *testing.T) {
	s := []byte("\"console\"")
	typ := unspecified
	err := json.Unmarshal(s, &typ)
	assert.Nil(t, err)
	assert.Equal(t, console, typ)

	// receiver & method
	s = []byte("\"kafka\"")
	typ = unspecified
	err = typ.UnmarshalJSON(s)
	assert.Nil(t, err)
	assert.Equal(t, kafka, typ)

	// invalid type
	s = []byte("\"invalid\"")
	typ = unspecified
	err = json.Unmarshal(s, &typ)
	assert.NotNil(t, err)
}
