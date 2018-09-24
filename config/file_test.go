package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadFile(t *testing.T) {
	// test file not exist
	b, e := readFile("/not-exist.log", 1024)
	assert.Nil(t, b, "non-exist")
	assert.NotNil(t, e, e.Error())

	// test file too big
	b, e = readFile("./file_test.go", 10)
	assert.Nil(t, b, "too-big")
	assert.NotNil(t, e, e.Error())

	// test normal
	b, e = readFile("./file_test.go", maxConfFileSize)
	assert.NotNil(t, b, "normal")
	assert.Nil(t, e, "normal")
}
