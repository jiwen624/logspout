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
)

var (
	tps *expvar.Map
)

func init() {
	initCounters()
	registerHandler()
}

func initCounters() {
	// Don't publish it to debug/vars
	tps = &expvar.Map{}
	tps.Init()
}

func registerHandler() {
	http.HandleFunc("/metrics/tps", tpsHandler)
}

// SetTPS sets the transaction per second data of a particular worker
func SetTPS(worker string, val int64) {
	v := &expvar.Int{}
	v.Set(val)
	tps.Set(worker, v)
}

// tpsHandler is an HTTP handler to expose the metrics. It also aggregates
// the metrics based on predefined rules.
func tpsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	var s string
	if b, err := json.Marshal(tpsSnapshot()); err != nil {
		s = err.Error()
	} else {
		s = string(b)
	}
	fmt.Fprintf(w, s)
}

// tpsSnapshot takes a consistent snapshot of the current TPS metrics.
func tpsSnapshot() map[string]int64 {
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
