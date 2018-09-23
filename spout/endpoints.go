package spout

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"sync"

	"github.com/jiwen624/logspout/flag"
	"github.com/jiwen624/logspout/log"
)

// Counter stores the counter values returned to the client
type Counter struct {
	Workers []uint64 `json:"Workers"`
	Total   uint64   `json:"TotalEPS"`
	Conf    string   `json:"ConfigFile"`
}

var (
	reqCounter = false

	// For fetching the counter values
	wgCounter sync.WaitGroup

	cCounter = sync.NewCond(&sync.Mutex{})

	resChan = make(chan uint64)
)

func (s *Spout) fetchCounter(w http.ResponseWriter, r *http.Request) {
	details := r.URL.Query().Get("details")

	counter := Counter{
		Workers: make([]uint64, 0),
		Total:   0,
		Conf:    "",
	}

	wgCounter.Add(s.Concurrency)

	cCounter.L.Lock()
	reqCounter = true
	cCounter.Broadcast()
	cCounter.L.Unlock()

	wgCounter.Wait()
	// Change this flag to false only after all the counter goroutines are done.
	reqCounter = false

	var total uint64
	var num = s.Concurrency
	for c := range resChan {
		if details == "true" {
			counter.Workers = append(counter.Workers, c)
		}
		total += c
		num--
		if num <= 0 {
			break
		}
	}
	counter.Total = total
	counter.Conf = flag.ConfigPath

	var retStr string
	if b, err := json.Marshal(&counter); err != nil {
		retStr = err.Error()
	} else {
		retStr = string(b)
	}

	fmt.Fprintln(w, retStr)
}

// TODO: print the spout obj in the memory
func currConfig(w http.ResponseWriter, r *http.Request) {
	details := r.URL.Query().Get("details")
	if details == "true" {
		var cfg string
		if b, err := ioutil.ReadFile(flag.ConfigPath); err != nil {
			cfg = err.Error()
		} else {
			cfg = string(b)
		}
		fmt.Fprintln(w, cfg)
	} else {
		fmt.Fprintln(w, flag.ConfigPath)
	}
}

func (s *Spout) console() {
	if s.ConsolePort == 0 {
		log.Infof("Management console is disabled with port=%d", s.ConsolePort)
		return
	}

	http.HandleFunc("/counter", s.fetchCounter)
	http.HandleFunc("/config", currConfig)

	err := http.ListenAndServe(":"+strconv.Itoa(s.ConsolePort), nil)
	if err != nil {
		log.Fatal("listen and serve: ", err)
	}
}
