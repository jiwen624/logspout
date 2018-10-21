// Program logspout implements a tiny tool to generate machine logs in accordance with
// the format of a sample log and the replacement policies defined in the config file.
// It can be used to generate logs in very high speed as well to do stress tests.
package main

import (
	"github.com/jiwen624/logspout/config"
	"github.com/jiwen624/logspout/flag"
	"github.com/jiwen624/logspout/log"
	"github.com/jiwen624/logspout/spout"
	"github.com/jiwen624/logspout/utils"
)

const (
	errFailedInMain = "Failed in main"
)

func main() {
	utils.ExitOnErr(errFailedInMain, log.SetLevel(flag.LogLevel))

	log.Infof("Starting up Logspout %s+%s", spout.Version(), spout.Commit())

	conf, err := config.FromFile(flag.ConfigPath)
	utils.ExitOnErr(errFailedInMain, err)

	spt, err := spout.Build(conf)
	utils.ExitOnErr(errFailedInMain, err)

	// It will block here if no error happens
	utils.ExitOnErr(errFailedInMain, spt.Start())

	spt.Stop()
	log.Info("Logspout is stopped. Bye.")
}
