package output

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSyslog_String(t *testing.T) {
	sl := Syslog{Protocol: "udp", Host: "test.com", Tag: "logspout"}
	s := sl.String()
	assert.Contains(t, s, "udp")
	assert.Contains(t, s, "test.com")
	assert.Contains(t, s, "logspout")
}

func TestSyslog_Write(t *testing.T) {
	sl := Syslog{Protocol: "udp", Host: "test.com", Tag: "logspout"}
	n, err := sl.Write([]byte("hello"))
	assert.Equal(t, 0, n)
	assert.Contains(t, err.Error(), errOutputNull.Error())

	sl.logger = &DumbClosableWriter{}
	n, err = sl.Write([]byte("hello"))
	assert.Equal(t, len("hello"), n)
	assert.Nil(t, err)
	assert.Nil(t, sl.Deactivate())
}

func TestSyslog_ID(t *testing.T) {
	sl := Syslog{Protocol: "udp", Host: "test.com", Tag: "logspout"}
	assert.Equal(t, id(sl.String()), sl.ID())
}

func TestSyslog_Type(t *testing.T) {
	sl := Syslog{Protocol: "udp", Host: "test.com", Tag: "logspout"}
	assert.Equal(t, syslog, sl.Type())
}

func TestSyslog_ActivateDeactivate(t *testing.T) {
	sl := Syslog{Protocol: "udp", Host: "localhost:10308", Tag: "logspout"}
	err := sl.Activate()
	assert.Nil(t, err)
	assert.Nil(t, sl.Deactivate())

	// bad host
	sl = Syslog{Protocol: "udp", Host: "localhost:88888", Tag: "logspout"}
	err = sl.Activate()
	assert.NotNil(t, err)
	assert.NotNil(t, sl.Deactivate())
}
