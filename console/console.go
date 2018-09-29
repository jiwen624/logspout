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
		log.Fatal("listen and serve: ", err)
	}
}
