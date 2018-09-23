// Package logspout implements a tiny tool to generate machine logs in accordance with
// the format of a sample log and the replacement policies defined in the config file.
// It can be used to generate logs in very high speed as well to do stress tests.
package main

import (
	"bytes"
	"fmt"
	"io"
	l "log"
	"math"
	"os"
	"regexp"

	"github.com/jiwen624/logspout/config"
	"github.com/jiwen624/logspout/flag"
	"github.com/jiwen624/logspout/log"

	"github.com/jiwen624/logspout/spout"
	. "github.com/jiwen624/logspout/utils"
)

// Control the speed of log bursts, in milliseconds.
var minInterval = 1000.0
var maxInterval = 1000.0
var duration = 0
var maxEvents uint64 = math.MaxUint64
var concurrency = 1
var duplicate = 1
var highTide = false
var uniform = true
var trans = false
var transIds = make([]string, 0)
var rawMsgs = make([]string, 0)
var intraTransLat = 10

// termChans stores the channels for close requests
var termChans = make([]chan struct{}, 0, concurrency)

// The default log event output stream: stdout
var logger = l.New(os.Stdout, "", 0)

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

	log.Debug(summary(conf))

	spt, err := spout.Build(conf)
	if err != nil {
		log.Errorf("Failed to create logspout: %s", err.Error())
		return
	}

	var dests []io.Writer
	// TODO: remove
	log.Debugf("===>sput: %+v", spt.Output)
	// for _, value := range spt.Output {
	// 	dests = append(dests, value)
	// }
	logger.SetOutput(io.MultiWriter(dests...))

	// TODO: define a pattern struct and move it to that struct
	// TODO: support both Perl and PCRE
	log.Debugf("Original patterns:\n%v\n", spt.Pattern)
	var ptns []string
	for _, ptn := range spt.Pattern {
		ptns = append(ptns, ReConvert(ptn))
	}
	spt.Pattern = ptns
	log.Debugf("Converted patterns:\n%v\n", spt.Pattern)

	var matches = make([][]string, 0)
	var names = make([][]string, 0)

	for idx, ptn := range spt.Pattern {
		re := regexp.MustCompile(ptn)
		matches = append(matches, re.FindStringSubmatch(rawMsgs[idx]))
		names = append(names, re.SubexpNames())

		if len(matches[idx]) == 0 {
			log.Errorf("the re pattern doesn't match the sample log in #%d", idx)
			return
		}

		// Remove the first one as it is the whole string.
		matches[idx] = matches[idx][1:]
		names[idx] = names[idx][1:]
	}

	for idx, match := range matches {
		log.Debugf("   pattern #%d", idx)
		for i, group := range match {
			log.Debugf("       - %s: %s", names[idx][i], group)
		}
	}

	log.Debug("check above matches and change patterns if something is wrong")

	// TODO: change minInterval to int
	minInterval = float64(spt.MinInterval)
	maxInterval = float64(spt.MaxInterval)
	duration = spt.Duration // I suppose you won't set a large number that makes an int overflow.
	if spt.MaxEvents != 0 {
		maxEvents = uint64(spt.MaxEvents)
	}

	if minInterval > maxInterval {
		log.Error("minInterval should be less than maxInterval")
		return
	}

	if err := spt.Start(); err != nil {
		log.Errorf("Failed to start spout: %v", err)
		return
	}

	// TODO: Register stop to sig handler and timeout handler
	defer spt.Stop()

}
