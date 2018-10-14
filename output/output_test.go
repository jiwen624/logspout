package output

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

type alwaysSuccessfulWriter struct{}

func (s *alwaysSuccessfulWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func (s *alwaysSuccessfulWriter) Close() error {
	return nil
}

type alwaysFailedWriter struct{}

func (f *alwaysFailedWriter) Write(p []byte) (n int, err error) {
	return 0, errors.New("failed to write bytes to")
}

func (f *alwaysFailedWriter) Close() error {
	return nil
}

func TestInitializer(t *testing.T) {
	typ := Type(10000)
	initializer := func() Output { return nil }
	RegisterType(typ, initializer)
	i, ok := GetInitializer(typ)
	assert.True(t, ok)
	assert.NotNil(t, i)

	UnregisterType(typ)
	i, ok = GetInitializer(typ)
	assert.False(t, ok)
	assert.Nil(t, i)
}

func TestID(t *testing.T) {
	a := "hello"
	b := "world"
	c := "World"
	assert.NotEqual(t, "", id(a))
	assert.NotEqual(t, id(a), id(b))
	assert.NotEqual(t, id(b), id(c))
}
