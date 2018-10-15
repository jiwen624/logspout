package config

import (
	"fmt"
	"os"
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

	// permission denied
	fn := "/tmp/test.file1030"
	os.Create(fn)
	os.Chmod(fn, 0200)
	b, e = readFile(fn, maxConfFileSize)
	assert.Nil(t, b)
	assert.NotNil(t, e)
	os.Remove(fn)
}

func TestFromFile(t *testing.T) {
	// not supported file type
	ext := ".not-supported"
	fn := fmt.Sprintf("/tmp/test%s", ext)
	sc, err := FromFile(fn)
	assert.Nil(t, sc)
	assert.Contains(t, err.Error(), errUnsupportedFileType.Error())
	assert.Contains(t, err.Error(), ext)

	// yaml file: not supported now
	fn = "/tmp/non-exist.yml"
	sc, err = FromFile(fn)
	assert.Nil(t, sc)
	assert.Contains(t, err.Error(), errUnsupportedFileType.Error())
	assert.Contains(t, err.Error(), "Yaml")

	// json file
	fn = "../examples/logspout.json"
	sc, err = FromFile(fn)
	assert.Nil(t, err)
	assert.NotNil(t, sc)

	fn = "/tmp/non-exist.json"
	sc, err = FromFile(fn)
	assert.Nil(t, sc)
	assert.NotNil(t, err)
}
