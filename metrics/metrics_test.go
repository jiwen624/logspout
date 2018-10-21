package metrics

import (
	"encoding/json"
	"expvar"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/jiwen624/logspout/console"

	"github.com/stretchr/testify/assert"
)

func TestInitCounters(t *testing.T) {
	initCounters()
	assert.NotNil(t, tps)

	var i int
	tps.Do(func(kv expvar.KeyValue) {
		i++
	})
	assert.Equal(t, 0, i)
}

func TestRegisterHandler(t *testing.T) {
	// already registered in init() method
	rawUrl := "/metrics/tps"
	u, _ := url.Parse(rawUrl)

	handler, path := http.DefaultServeMux.Handler(&http.Request{Method: "GET", URL: u})
	// function in go is not addressable or comparable, so just make sure the handler
	// is not nil here.
	assert.NotNil(t, handler)
	assert.Equal(t, path, rawUrl)
}

func TestSetGetTPS(t *testing.T) {
	host := "localhost:12345"
	console.Start(host)

	SetTPS("worker1", 1)
	SetTPS("worker2", 2)
	SetTPS("worker3", 3)

	client := &http.Client{
		Timeout: time.Second * 10,
	}
	rsp, err := client.Get("http://" + host + "/metrics/tps")

	assert.Nil(t, err)
	defer rsp.Body.Close()

	body, err := ioutil.ReadAll(rsp.Body)
	assert.Nil(t, err)

	tps := struct {
		Total   int `json:"Total"`
		Worker1 int `json:"worker1"`
		Worker2 int `json:"worker2"`
		Worker3 int `json:"worker3"`
	}{}
	json.Unmarshal(body, &tps)

	assert.Equal(t, 1, tps.Worker1)
	assert.Equal(t, 2, tps.Worker2)
	assert.Equal(t, 3, tps.Worker3)
	assert.Equal(t, 6, tps.Total)
}

func TestTpsSnapshot(t *testing.T) {
	assert.Nil(t, tpsSnapshot(nil))
}
