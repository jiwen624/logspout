package spout

import (
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/jiwen624/logspout/metrics"

	"github.com/jiwen624/logspout/log"
	"github.com/jiwen624/logspout/replacer"
	"github.com/jiwen624/logspout/utils"
)

type worker struct {
	// The name of the worker, which is mainly used for logging purpose.
	name string
	// The maximum number of events expected, the worker will quit after reaching
	// this number.
	// It's calculated via spout's maximum events / concurrency
	maxEvents int
	// The life cycle of this worker.
	duration time.Duration
	// The replacers used by the worker to do string substitutions.
	replacers replacer.Replacers
	// The transaction ID
	transIDs []string
	// The logs to be used for substitutions.
	seedLogs []string
	// The minimum interval in milliseconds between logs in the same transaction.
	minInterval int
	// The maximum interval in milliseconds between logs in the same transaction.
	maxInterval int
	// The maximum interval in milliseconds between two adjacent transactions.
	maxIntraTransLat int
	// Is the workload (aka TPS) is uniformed or with some jitter
	uniformLoad bool
	// The function to be called to write logs to the output destinations
	writeTo func(string) error
	// The callback function after the worker is finished.
	doneCallback func()
	// The channel that indicates the worker should exit when it's closed.
	closeChan chan struct{}
	// The random number generator.
	rand replacer.RandomGenerator
	// The flag indicates if the workload is in burst mode, where no think time exists.
	burstMode bool
}

// See the corresponding comments in fields of the struct worker.
type workerConfig struct {
	Index            int
	MaxEvents        int
	Seconds          int
	Replacers        replacer.Replacers
	TransIDs         []string
	SeedLogs         []string
	MinInterval      int
	MaxInterval      int
	UniformLoad      bool
	MaxIntraTransLat int
	WriteTo          func(string) error
	DoneCallback     func()
	CloseChan        chan struct{}
	BurstMode        bool
}

func NewWorker(c workerConfig) *worker {
	w := &worker{
		name:             fmt.Sprintf("worker%d", c.Index),
		maxEvents:        c.MaxEvents,
		duration:         time.Second * time.Duration(c.Seconds),
		replacers:        c.Replacers,
		transIDs:         c.TransIDs,
		minInterval:      c.MinInterval,
		maxInterval:      c.MaxInterval,
		uniformLoad:      c.UniformLoad,
		maxIntraTransLat: c.MaxIntraTransLat,
		seedLogs:         c.SeedLogs,
		writeTo:          c.WriteTo,
		doneCallback:     c.DoneCallback,
		closeChan:        c.CloseChan,
		rand:             replacer.NewTruncatedGaussian(0.5, 0.2),
		burstMode:        c.BurstMode,
	}
	return w
}

// startWorker generates new logs with the replacement policies, in a infinite loop.
func (w *worker) start(m [][]string, names [][]string, workerID int) {
	workerName := fmt.Sprintf("worker%d", workerID)

	log.Infof("%s spawned", workerName)
	defer log.Infof("%s is exiting.", workerName)

	defer w.doneCallback()

	// Make a deep copy of m, as it's used by multiple workers.
	matches := utils.StrSlice2DCopy(m)

	// the index of current log event in the transaction (which contains multiple
	// log events
	var evtIdx int
	// the transaction per second
	var tps int64
	// the total count of log events
	var generatedNum int

	// Does it `think` between two adjacent transactions.
	sleepIntraTrans := len(w.transIDs) != 0 && !w.burstMode
	// Does it `think` between two adjacent logs in the same transaction.
	sleepInterTrans := !w.burstMode && w.maxInterval > 0
	// The ticker defines the worker metrics flushing interval.
	cTicker := time.NewTicker(time.Second * 1).C

	for {
		// The first message of a transaction
		for k, v := range w.replacers {
			idx := utils.StrIndex(names[evtIdx], k)
			if idx == -1 {
				continue
			} else if evtIdx == 0 || utils.StrIndex(w.transIDs, k) == -1 {
				if s, err := v.ReplacedValue(w.rand); err == nil {
					matches[evtIdx][idx] = s
				}
			} else {
				matches[evtIdx][idx] = matches[0][idx]
			}
		}

		// Print to logger streams, you may redirect it to anywhere else you want
		if err := w.writeTo(strings.Join(matches[evtIdx], "")); err != nil {
			log.Warn(errors.Wrap(err, "err writing logs to output"))
		}

		tps++
		// Exits after it exceeds the predefined maximum events.
		generatedNum++
		if generatedNum >= w.maxEvents {
			return
		}

		// It never sleeps in burst mode.
		if sleepIntraTrans {
			sleepTime := time.Millisecond * time.Duration(w.rand.Next(w.maxIntraTransLat))
			time.Sleep(sleepTime)
		}

		evtIdx++
		if evtIdx >= len(w.seedLogs) {
			evtIdx = 0
			// think for a while between transactions
			if sleepInterTrans {
				w.think()
			}
		}

		select {
		case <-w.closeChan:
			return
		case <-cTicker:
			metrics.SetTPS(workerName, tps)
			tps = 0
		default:
		}
	}
}

// think calculates the think time and sleep for certain period if time
func (w *worker) think() {
	// Sleep for a short while.
	time.Sleep(w.calculateThinkTime())
}

func (w *worker) calculateThinkTime() time.Duration {
	if w.maxInterval <= w.minInterval {
		return time.Duration(w.minInterval) * time.Millisecond
	}

	if w.uniformLoad {
		d := w.minInterval + w.rand.Next(w.maxInterval-w.minInterval)
		return time.Duration(d) * time.Millisecond
	}
	// TODO: change to stochastic workload
	d := w.minInterval + w.rand.Next(w.maxInterval-w.minInterval)
	return time.Duration(d) * time.Millisecond
}
