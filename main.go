// Package logspout implements a tiny tool to generate machine logs in accordance with
// the format of a sample log and the replacement policies defined in the config file.
// It can be used to generate logs in very high speed as well to do stress tests.
package main

import (
	"github.com/jiwen624/logspout/config"
	"github.com/jiwen624/logspout/flag"
	"github.com/jiwen624/logspout/log"
	"github.com/jiwen624/logspout/utils"

	"github.com/jiwen624/logspout/spout"
)

func main() {
	utils.CheckErr(log.SetLevel(flag.LogLevel))

	conf, err := config.FromJsonFile(flag.ConfigPath)
	if err != nil {
		log.Errorf("Error loading config: %s", err)
		return
	}

	spt, err := spout.Build(conf)
	if err != nil {
		log.Errorf("Failed to create logspout: %s", err.Error())
		return
	}

	if err := spt.Start(); err != nil {
		log.Errorf("Failed to start spout: %v", err)
		return
	}

	// TODO: Register stop to sig handler and timeout handler
	// spt.Stop()
}
