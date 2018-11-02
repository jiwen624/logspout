// Package console contains the methods and functions of the management console.
// The metrics data can be exposed through the management console.
package console

import (
	"net/http"

	"github.com/jiwen624/logspout/log"
)

// Start kicks of the HTTP console
func Start(host string) {
	log.Infof("Starting up the console on %s.", host)
	go startListener(host)
}

func startListener(host string) {
	err := http.ListenAndServe(host, nil)
	if err != nil {
		log.Error("listen and serve: ", err)
	}
}
