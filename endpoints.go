package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// Counter stores the counter values returned to the client
type Counter struct {
	Workers []uint64 `json:"Workers"`
	Total   uint64   `json:"TotalEPS"`
	Conf    string   `json:"ConfigFile"`
}

func fetchCounter(w http.ResponseWriter, r *http.Request) {
	details := r.URL.Query().Get("details")

	counter := Counter{
		Workers: make([]uint64, 0),
		Total:   0,
		Conf:    "",
	}

	wgCounter.Add(concurrency)

	cCounter.L.Lock()
	reqCounter = true
	cCounter.Broadcast()
	cCounter.L.Unlock()

	wgCounter.Wait()
	// Change this flag to false only after all the counter goroutines are done.
	reqCounter = false

	var total uint64
	var num = concurrency
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
	counter.Total = total * uint64(duplicate)
	counter.Conf = *confPath

	var retStr string
	if b, err := json.Marshal(&counter); err != nil {
		retStr = err.Error()
	} else {
		retStr = string(b)
	}

	fmt.Fprintln(w, retStr)
}

func config(w http.ResponseWriter, r *http.Request) {
	details := r.URL.Query().Get("details")
	if details == "true" {
		fmt.Fprintln(w, string(conf))
	} else {
		fmt.Fprintln(w, *confPath)
	}
}

func console() {
	http.HandleFunc("/counter", fetchCounter)
	http.HandleFunc("/config", config)

	err := http.ListenAndServe(":"+consolePort, nil)
	if err != nil {
		log.Fatal("listen and serve: ", err)
	}
}
