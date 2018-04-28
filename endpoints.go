package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

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

func console() {
	http.HandleFunc("/counter", fetchCounter)
	err := http.ListenAndServe(":"+consolePort, nil)
	if err != nil {
		sugar.Fatal("listen and serve: ", err)
	}
}
