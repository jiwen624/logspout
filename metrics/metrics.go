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
	// TODO: not neccessary to use expvar.Int as :1. it uses atomic to access
	// TODO: 2. a new Int object is created each time.
	v := &expvar.Int{}
	v.Set(val)
	tps.Set(strconv.Itoa(worker), v)
	// TODO: calculate the total TPS
}
