package output

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFileWrite(t *testing.T) {
	f := &File{}
	n, err := f.Write(nil)
	assert.Equal(t, 0, n)
	assert.NotNil(t, err)

	f = &File{loggers: []ClosableWriter{&alwaysSuccessfulWriter{}}}
	n, err = f.Write([]byte{'a'})
	assert.Equal(t, 1, n)
	assert.Nil(t, err)

	f = &File{loggers: []ClosableWriter{&alwaysFailedWriter{}}}
	n, err = f.Write([]byte{'a'})
	assert.Equal(t, 0, n)
	assert.NotNil(t, err)

	f = &File{loggers: []ClosableWriter{&alwaysSuccessfulWriter{}, &alwaysFailedWriter{}}}
	n, err = f.Write([]byte{'a'})
	assert.Equal(t, 0, n)
	assert.NotNil(t, err)
}

func TestFileString(t *testing.T) {
	f := &File{}
	assert.NotEmpty(t, f.String())
}

func TestFileID(t *testing.T) {
	f := &File{}
	assert.NotEmpty(t, f.ID(), f.ID)
}

func TestFileType(t *testing.T) {
	f := &File{}
	assert.Equal(t, file, f.Type())
}

func TestFileActivate(t *testing.T) {
	f := &File{}
	f.Activate()
	assert.Equal(t, defaultFileName, f.FileName)
	assert.Equal(t, defaultDir, f.Directory)
	assert.Equal(t, defaultMaxAge, f.MaxAge)
	assert.Equal(t, defaultMaxBackups, f.MaxBackups)
	assert.Equal(t, defaultMaxSize, f.MaxSize)
	assert.NotNil(t, f.loggers)

	fName := "nondefault_name"
	f.FileName = fName
	f.Activate()
	assert.Equal(t, fName, f.FileName)
}

func TestFileDeactivate(t *testing.T) {
	f := &File{}
	f.Activate()
	assert.Nil(t, f.Deactivate())
	assert.NotNil(t, f.Deactivate())
}

func TestBuildFile(t *testing.T) {
	f := &File{Duplicate: 3}
	f.Activate()
	assert.Equal(t, 3, len(f.loggers))
}
