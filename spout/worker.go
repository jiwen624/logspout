package spout

import (
	"math"
	"strings"
	"time"

	"github.com/jiwen624/logspout/metrics"

	"github.com/jiwen624/logspout/log"
	"github.com/jiwen624/logspout/replacer"
	"github.com/jiwen624/logspout/utils"

	"github.com/leesper/go_rng"
)

// PopNewLogs generates new logs with the replacement policies, in a infinite loop.
func (s *Spout) popNewLogs(m [][]string, names [][]string, workerID int) {
	log.Debugf("spawned worker #%d", workerID)
	defer log.Infof("worker #%d is exiting.", workerID)

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
	var totalCnt int

	cTicker := time.NewTicker(time.Second * 1).C
	for {
		// The first message of a transaction
		for k, v := range s.Replacers {
			idx := utils.StrIndex(names[evtIdxInTrans], k)
			if idx == -1 {
				continue
			} else if evtIdxInTrans == 0 || utils.StrIndex(s.TransactionID, k) == -1 {
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
		totalCnt++
		if totalCnt >= int(s.MaxEvents/s.Concurrency) {
			return
		}

		// It never sleeps in hightide mode.
		if len(s.TransactionID) != 0 && s.BurstMode == false {
			time.Sleep(time.Millisecond * time.Duration(replacer.SimpleGaussian(grng, s.MaxIntraTransLat)))
		}

		evtIdxInTrans++
		if evtIdxInTrans >= len(s.seedLogs) {
			evtIdxInTrans = 0

			// We will populate events as fast as possible in high tide mode. (Watch out your CPU!)
			if s.BurstMode == false {
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
		}

		select {
		case <-s.close:
			return
		case <-cTicker:
			metrics.SetTPS(workerID, tps)
			tps = 0
		default:
		}
	}
}

// Spray sprays the generated logs into the predefined destinations.
func (s *Spout) Spray(log string) error {
	return s.Output.Write(log)
}
