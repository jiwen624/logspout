package config

import (
	"testing"

	"io/ioutil"

	"github.com/stretchr/testify/assert"
)

func TestLoadJson(t *testing.T) {
	var data []byte
	sc, err := LoadJson(data)
	assert.Nil(t, sc)
	assert.NotNil(t, err)

	data = []byte("")
	sc, err = LoadJson(data)
	assert.Nil(t, sc)
	assert.NotNil(t, err)

	data = []byte("{}")
	sc, err = LoadJson(data)
	assert.NotNil(t, sc)
	assert.Nil(t, err)

	data = []byte(`{"burstMode": 123}`)
	sc, err = LoadJson(data)
	assert.Nil(t, sc)
	assert.NotNil(t, err)

	data = []byte(`{"burstMode": false, "concurrency": 1, "minInterval":1}`)
	sc, err = LoadJson(data)
	assert.NotNil(t, sc)
	assert.Nil(t, err)

	data, err = ioutil.ReadFile("../logspout-docker.json")
	assert.Nil(t, err)
	sc, err = LoadJson(data)
	assert.NotNil(t, sc)
	assert.Nil(t, err)

	data, err = ioutil.ReadFile("../logspout.json")
	assert.Nil(t, err)
	sc, err = LoadJson(data)
	assert.NotNil(t, sc)
	assert.Nil(t, err)
}
