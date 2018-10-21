package output

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegistry(t *testing.T) {
	r := NewRegistry()
	assert.Equal(t, 0, r.Size())

	assert.Equal(t, ErrRegisterNilOutput, r.Register(nil))
	assert.Equal(t, ErrUnRegisterNilOutput, r.Unregister(nil))
	assert.Equal(t, ErrEmptyRegistry, r.ForEach(func(Output) error {
		return nil
	}, func(Output) bool { return true }))

	do := &Discard{}
	assert.Nil(t, r.Register(do))
	assert.Equal(t, 1, r.Size())

	fo := &Console{FileName: "stdout"}
	assert.Nil(t, r.Register(fo))

	// Register it again
	assert.Equal(t, ErrDuplicate, r.Register(fo))

	assert.Equal(t, 2, r.Size())

	assert.Nil(t, r.ForAll(func(o Output) error {
		return o.Activate()
	}))

	foGet, err := r.Get(fo.ID())
	assert.Nil(t, err)
	assert.Equal(t, fo, foGet)

	assert.Nil(t, r.ForAll(func(o Output) error {
		assert.True(t, o.Type() == console || o.Type() == discard)
		return nil
	}))

	assert.Nil(t, r.Write("hello"))

	assert.NotEqual(t, "", r.String())

	assert.NotNil(t, r.Unregister(&Console{FileName: "invalid"}))

	assert.Nil(t, r.Unregister(fo))
	assert.Equal(t, 1, r.Size())

	assert.Nil(t, r.Unregister(do))
	assert.Equal(t, 0, r.Size())
}

func TestRegistryFromConf(t *testing.T) {
	ow := map[string]Wrapper{
		"syslog": {
			T: syslog,
			Raw: []byte(`{
        "protocol": "udp",
        "host": "localhost:516",
        "tag": "logspout"
      }`),
		},
	}

	r, err := RegistryFromConf(ow)
	assert.Nil(t, err)
	assert.NotNil(t, r)
	assert.Equal(t, 1, r.Size())

	apply := func(o Output) error {
		assert.Equal(t, syslog, o.Type())

		so, ok := o.(*Syslog)
		assert.True(t, ok)

		assert.Equal(t, "udp", so.Protocol)
		assert.Equal(t, "localhost:516", so.Host)
		assert.Equal(t, "logspout", so.Tag)
		return nil
	}

	predicate := func(o Output) bool {
		return o.Type() == syslog
	}

	assert.Nil(t, r.ForEach(apply, predicate))
}
