package output

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKafka(t *testing.T) {
	ko := &Kafka{}
	assert.Equal(t, kafka, ko.Type())
	assert.Equal(t, "Kafka", ko.String())
	assert.Equal(t, ID(""), ko.ID())
	n, e := ko.Write(nil)
	assert.Equal(t, 0, n)
	assert.Nil(t, e)

	assert.Nil(t, ko.Activate())
	assert.Nil(t, ko.Deactivate())
}
