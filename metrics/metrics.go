package metrics

import (
	"expvar"
	"strconv"
)

var (
	tps *expvar.Map
)

func init() {
	initCounters()
}

func initCounters() {
	tps = expvar.NewMap("tps")
}

// SetTPS sets the transaction per second data of a particular worker
func SetTPS(worker int, val int64) {
	tps.Set(strconv.Itoa(worker), &expvar.Int{val})
}
