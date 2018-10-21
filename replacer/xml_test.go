package replacer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInvalidXMLStr(t *testing.T) {
	s, err := XMLStr(0, 0, []string{"hello"})
	assert.NotNil(t, err)
	assert.Equal(t, "", s)
}

func TestNilDocXmlStr(t *testing.T) {
	assert.Equal(t, 0, xmlStr(nil, 1, 1, 1, nil, nil))
}

func TestRandomTag(t *testing.T) {
	assert.NotEqual(t, "", randomTag([]string{}))
}

func TestRandomAttrK(t *testing.T) {
	assert.NotEqual(t, "", randomAttrK())
}

func TestRandomAttrV(t *testing.T) {
	assert.NotEqual(t, "", randomAttrV())
}

func TestRandomComment(t *testing.T) {
	assert.NotEqual(t, "", randomComment())
}

func TestNeedComment(t *testing.T) {
	b := needComment()
	assert.True(t, b == true || b == false)
}

func TestRandomData(t *testing.T) {
	assert.NotEqual(t, "", randomData())
}
