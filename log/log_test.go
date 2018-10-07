package log

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetLevel(t *testing.T) {
	assert.NotNil(t, SetLevel("invalid"))
	assert.Nil(t, SetLevel("info"))
	assert.Equal(t, INFO, GetLevel())
	assert.Equal(t, "info", ToString(INFO))
}
