package output

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeName(t *testing.T) {
	cases := map[string]string{
		"stdout":  "stdout",
		"stderr":  "stderr",
		"Stdout":  "stdout",
		"stdErr":  "stderr",
		" stdout": "stdout",
		"stderr ": "stderr",
		`stderr
`: "stderr",
		"stdin ": "stdin ",
		"HELLO":  "HELLO",
		"":       "",
		" ":      " ",
		"@#$@#%": "@#$@#%",
	}

	for in, out := range cases {
		assert.Equal(t, out, normalizeName(in), fmt.Sprintf("in: %s, out: %s", in, out))
	}
}

func TestDeactivate(t *testing.T) {
	c := &Console{}
	assert.NotNil(t, c.Deactivate())

	c = &Console{FileName: "stderr"}
	assert.Nil(t, c.Activate())
	assert.Nil(t, c.Deactivate())
	assert.Nil(t, c.logger)
}

func TestActivate(t *testing.T) {
	c := &Console{FileName: "stdout"}
	assert.Nil(t, c.Activate())
	assert.Equal(t, os.Stdout, c.logger)

	c = &Console{FileName: "stderr"}
	assert.Nil(t, c.Activate())
	assert.Equal(t, os.Stderr, c.logger)

	c = &Console{FileName: "invalid"}
	assert.NotNil(t, c.Activate())
	assert.Equal(t, nil, c.logger)
}

func TestId(t *testing.T) {
	c := &Console{FileName: "stdout"}
	assert.NotNil(t, c.ID())
}

func TestString(t *testing.T) {
	c := &Console{FileName: "stderr"}
	assert.NotNil(t, c.String())
}

func TestWrite(t *testing.T) {
	c := &Console{FileName: "fake"}
	n, err := c.Write([]byte{})
	assert.Equal(t, 0, n)
	assert.NotNil(t, err)

	c.logger = &alwaysSuccessfulWriter{}
	n, err = c.Write([]byte{'a'})
	assert.Equal(t, 1, n)
	assert.Nil(t, err)
}

func TestType(t *testing.T) {
	c := &Console{}
	assert.Equal(t, console, c.Type())
}
