package spout

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/jiwen624/logspout/metrics"

	"github.com/jiwen624/logspout/log"
	"github.com/jiwen624/logspout/replacer"
	"github.com/jiwen624/logspout/utils"

	"github.com/leesper/go_rng"
)

// startWorker generates new logs with the replacement policies, in a infinite loop.
func (s *Spout) startWorker(m [][]string, names [][]string, workerID int) {
	workerName := fmt.Sprintf("worker%d", workerID)

	log.Infof("%s spawned", workerName)
	defer log.Infof("%s is exiting.", workerName)

	// the expected number of events assigned to this worker
	expectedEvents := int(s.MaxEvents / s.Concurrency)

	var newLog string
	defer s.Done()

	// Gaussian distribution
	grng := rng.NewGaussianGenerator(time.Now().UnixNano())

	matches := utils.StrSlice2DCopy(m)

	// the index of current log event in the transaction (which contains multiple
	// log events
	var evtIdxInTrans int
	// the transaction per second
	var tps int64
	// the total count of log events
	var generatedEvents int

	cTicker := time.NewTicker(time.Second * 1).C
	for {
		// The first message of a transaction
		for k, v := range s.Replacers {
			idx := utils.StrIndex(names[evtIdxInTrans], k)
			if idx == -1 {
				continue
			} else if evtIdxInTrans == 0 || utils.StrIndex(s.TransactionID, k) == -1 {
				// TODO: data race due to multiple workers using/manipulating the same
				// TODO: replacer object
				if s, err := v.ReplacedValue(grng); err == nil {
					matches[evtIdxInTrans][idx] = s
				}
			} else {
				matches[evtIdxInTrans][idx] = matches[0][idx]
			}
		}

		newLog = strings.Join(matches[evtIdxInTrans], "")
		// Print to logger streams, you may redirect it to anywhere else you want

		s.Spray(newLog)

		tps++
		// Exits after it exceeds the predefined maximum events.
		generatedEvents++
		if generatedEvents >= expectedEvents {
			return
		}

		// It never sleeps in hightide mode.
		if len(s.TransactionID) != 0 && s.BurstMode == false {
			time.Sleep(time.Millisecond * time.Duration(replacer.SimpleGaussian(grng, s.MaxIntraTransLat)))
		}

		evtIdxInTrans++
		if evtIdxInTrans >= len(s.seedLogs) {
			evtIdxInTrans = 0
			// think for a while between transactions
			s.think(grng)
		}

		select {
		case <-s.close:
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
func (s *Spout) think(grng *rng.GaussianGenerator) {
	if s.BurstMode {
		return
	}

	// Sleep for a short while.
	var sleepMsec = s.MinInterval
	if s.MaxInterval == s.MinInterval {
		sleepMsec = s.MinInterval
	} else {
		if s.UniformLoad == true {
			sleepMsec = s.MinInterval + replacer.SimpleGaussian(grng, int(s.MaxInterval-s.MinInterval))
		} else { // There should be a better algorithm here.
			x := float64((time.Now().Unix() % 86400) / 13751)
			y := (math.Pow(math.Sin(x), 2) + math.Pow(math.Sin(x/2), 2) + 0.2) / 1.7619
			sleepMsec = int(float64(s.MinInterval) / y)
			if sleepMsec > s.MaxInterval {
				sleepMsec = s.MaxInterval
			}
		}
	}
	time.Sleep(time.Millisecond * time.Duration(int(sleepMsec)))
}
