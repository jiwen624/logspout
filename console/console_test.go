package console

import (
	"net/http"
	"testing"
	"time"

	"github.com/jiwen624/logspout/metrics"

	"github.com/stretchr/testify/assert"
)

func TestStart(t *testing.T) {
	// Test valid host
	url := "localhost:10300"
	Start(url)
	time.Sleep(time.Millisecond * 10)

	nc := &http.Client{
		Timeout: time.Second * 2,
	}
	rsp, err := nc.Get("http://" + url + metrics.EndPoint)
	assert.Nil(t, err)
	assert.Equal(t, 200, rsp.StatusCode)

	// bad url
	badUrl := "localhost:100"
	Start(badUrl)
	time.Sleep(time.Millisecond * 10)

	rsp, err = nc.Get("http://" + badUrl + metrics.EndPoint)
	assert.NotNil(t, err)
}
