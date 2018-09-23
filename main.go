// Package logspout implements a tiny tool to generate machine logs in accordance with
// the format of a sample log and the replacement policies defined in the config file.
// It can be used to generate logs in very high speed as well to do stress tests.
package main

import (
	"bytes"
	"fmt"

	"github.com/jiwen624/logspout/config"
	"github.com/jiwen624/logspout/flag"
	"github.com/jiwen624/logspout/log"

	"github.com/jiwen624/logspout/spout"
)

func summary(conf *config.SpoutConfig) string {
	var b bytes.Buffer
	b.WriteString(fmt.Sprintf("loaded configurations from %s\n", flag.ConfigPath))
	b.WriteString(fmt.Sprintf("  - logtype = %s\n", conf.LogType))
	b.WriteString(fmt.Sprintf("  - file = %s\n", conf.SampleFilePath))
	for idx, ptn := range conf.Pattern {
		b.WriteString(fmt.Sprintf("  - pattern #%d = %s\n", idx, ptn))
	}
	return b.String()
}

func main() {
	log.SetLevel(flag.LogLevel)

	conf, err := config.FromJsonFile(flag.ConfigPath)
	if err != nil {
		log.Errorf("Error loading config: %s", err)
		return
	}

	// log.Debug(summary(conf))

	spt, err := spout.Build(conf)
	if err != nil {
		log.Errorf("Failed to create logspout: %s", err.Error())
		return
	}

	// TODO: remove
	log.Debugf("===>sput: %+v", spt.Output)

	if err := spt.Start(); err != nil {
		log.Errorf("Failed to start spout: %v", err)
		return
	}

	// TODO: Register stop to sig handler and timeout handler
	// spt.Stop()
}
