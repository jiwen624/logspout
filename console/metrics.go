package console

import "expvar"

var (
	tps *expvar.Map
)

func init() {
	initCounters()
}

func initCounters() {
	tps = expvar.NewMap("tps")
}
