package spout

import (
	"math"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jiwen624/logspout/replacer"

	"github.com/jiwen624/logspout/log"

	"github.com/jiwen624/logspout/utils"

	"github.com/leesper/go_rng"
)

// PopNewLogs generates new logs with the replacement policies, in a infinite loop.
func (s *Spout) popNewLogs(m [][]string, names [][]string, wg *sync.WaitGroup, cCounter *sync.Cond,
	resChan chan uint64, idx int) {
	log.Debugf("spawned worker #%d", idx)
	defer log.Infof("worker #%d is exiting.", idx)

	var newLog string
	defer wg.Done()

	// Gaussian distribution
	grng := rng.NewGaussianGenerator(time.Now().UnixNano())

	matches := utils.StrSlice2DCopy(m)

	var currMsg int
	var counter uint64
	var totalCnt int

	var c uint64

	// This goroutine waits for the request from client to fetch the current counter value.
	go func(res chan uint64) {
		for {
			cCounter.L.Lock()
			for reqCounter == false {
				cCounter.Wait()
			}
			cCounter.L.Unlock()
			wgCounter.Done()

			res <- atomic.LoadUint64(&c)
		}
	}(resChan)

	cTicker := time.NewTicker(time.Second * 1).C
	for {
		// The first message of a transaction
		for k, v := range s.Replacers {
			idx := utils.StrIndex(names[currMsg], k)
			if idx == -1 {
				continue
			} else if currMsg == 0 || utils.StrIndex(s.TransactionID, k) == -1 {
				if s, err := v.ReplacedValue(grng); err == nil {
					matches[currMsg][idx] = s
				}
			} else {
				matches[currMsg][idx] = matches[0][idx]
			}
		}

		newLog = strings.Join(matches[currMsg], "")
		// Print to logger streams, you may redirect it to anywhere else you want
		s.Output.Write(newLog)
		counter++
		// Exits after it exceeds the predefined maximum events.
		totalCnt++
		if totalCnt >= int(s.MaxEvents/s.Concurrency) {
			return
		}

		// It never sleeps in hightide mode.
		if len(s.TransactionID) != 0 && s.BurstMode == false {
			time.Sleep(time.Millisecond * time.Duration(replacer.SimpleGaussian(grng, s.MaxIntraTransactionLatency)))
		}

		currMsg++
		if currMsg >= len(s.rawMsgs) {
			currMsg = 0

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
			atomic.StoreUint64(&c, counter)
			counter = 0
		default:
		}
	}
}
