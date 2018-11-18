// Package metrics contains the methods and functions of recording metrics, e.g.,
// TPS (transaction per second). The data recorded in this package can be offered
// to the management console which then exposes it through a particular HTTP
// endpoint.
package metrics

import (
	"encoding/json"
	"expvar"
	"fmt"
	"net/http"
)

const (
	TPS      = "tps"
	TotalTPS = "Total"
	EndPoint = "/metrics/tps"
)

var (
	tps *expvar.Map
)

func init() {
	initCounters()
	registerHandlers()
}

func initCounters() {
	// Don't publish it to debug/vars
	tps = &expvar.Map{}
	tps.Init()
}

func registerHandlers() {
	registerHandler("/metrics/tps", tpsHandler)
}

func registerHandler(url string, handler http.HandlerFunc) {
	http.HandleFunc(url, handler)
}

// SetTPS sets the transaction per second data of a particular worker
func SetTPS(worker string, val int64) {
	setTPS(tps, worker, val)
}

func setTPS(tps *expvar.Map, worker string, val int64) {
	v := &expvar.Int{}
	v.Set(val)
	tps.Set(worker, v)
}

// tpsHandler is an HTTP handler to expose the metrics. It also aggregates
// the metrics based on predefined rules.
func tpsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	var s string
	if b, err := json.Marshal(tpsSnapshot(tps)); err != nil {
		s = err.Error()
	} else {
		s = string(b)
	}
	fmt.Fprintf(w, s)
}

// tpsSnapshot takes a consistent snapshot of the current TPS metrics.
func tpsSnapshot(tps *expvar.Map) map[string]int64 {
	if tps == nil {
		return nil
	}

	tpsMap := make(map[string]int64)

	tps.Do(func(kv expvar.KeyValue) {
		tpsMap[kv.Key] = kv.Value.(*expvar.Int).Value()
	})

	var totalTPS int64
	for _, v := range tpsMap {
		totalTPS += v
	}
	tpsMap[TotalTPS] = totalTPS
	return tpsMap
}
