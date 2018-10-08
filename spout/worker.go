package spout

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/jiwen624/logspout/metrics"

	"github.com/jiwen624/logspout/log"
	"github.com/jiwen624/logspout/replacer"
	"github.com/jiwen624/logspout/utils"
)

type worker struct {
	name             string
	maxEvents        int
	duration         time.Duration
	replacers        replacer.Replacers
	transIDs         []string
	seedLogs         []string
	minInterval      int
	maxInterval      int
	maxIntraTransLat int
	uniformLoad      bool
	writeTo          func(string) error
	doneCallback     func()
	closeChan        chan struct{}
	rand             replacer.RandomGenerator
	burstMode        bool
}

type workerConfig struct {
	Index int
	// the maximum events of this particular worker
	// should be spout's maximum events / concurrency
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

	matches := utils.StrSlice2DCopy(m)

	// the index of current log event in the transaction (which contains multiple
	// log events
	var evtIdxInTrans int
	// the transaction per second
	var tps int64
	// the total count of log events
	var generatedEvents int

	sleepIntraTrans := len(w.transIDs) != 0 && !w.burstMode
	sleepInterTrans := !w.burstMode && w.maxInterval > 0
	cTicker := time.NewTicker(time.Second * 1).C

	for {
		// The first message of a transaction
		for k, v := range w.replacers {
			idx := utils.StrIndex(names[evtIdxInTrans], k)
			if idx == -1 {
				continue
			} else if evtIdxInTrans == 0 || utils.StrIndex(w.transIDs, k) == -1 {
				if s, err := v.ReplacedValue(w.rand); err == nil {
					matches[evtIdxInTrans][idx] = s
				}
			} else {
				matches[evtIdxInTrans][idx] = matches[0][idx]
			}
		}

		// Print to logger streams, you may redirect it to anywhere else you want
		if err := w.writeTo(strings.Join(matches[evtIdxInTrans], "")); err != nil {
			log.Warn(errors.Wrap(err, "err writing logs to output"))
		}

		tps++
		// Exits after it exceeds the predefined maximum events.
		generatedEvents++
		if generatedEvents >= w.maxEvents {
			return
		}

		// It never sleeps in burst mode.
		if sleepIntraTrans {
			time.Sleep(time.Millisecond * time.Duration(w.rand.Next(w.maxIntraTransLat)))
		}

		evtIdxInTrans++
		if evtIdxInTrans >= len(w.seedLogs) {
			evtIdxInTrans = 0
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

// TODO: use Jitter object as the only parameter
// think calculates the think time and sleep for certain period if time
func (w *worker) think() {
	// Sleep for a short while.
	var sleepMsec = w.minInterval
	if w.maxInterval <= w.minInterval {
		sleepMsec = w.minInterval
	} else {
		if w.uniformLoad == true {
			sleepMsec = w.minInterval + w.rand.Next(w.maxInterval-w.minInterval)
		} else { // There should be a better algorithm here.
			x := float64((time.Now().Unix() % 86400) / 13751)
			y := (math.Pow(math.Sin(x), 2) + math.Pow(math.Sin(x/2), 2) + 0.2) / 1.7619
			sleepMsec = int(float64(w.minInterval) / y)
			if sleepMsec > w.maxInterval {
				sleepMsec = w.maxInterval
			}
		}
	}
	time.Sleep(time.Millisecond * time.Duration(int(sleepMsec)))
}
