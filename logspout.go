package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/buger/jsonparser"
	"io/ioutil"
	"os"
	"regexp"
)

const (
	LogType     = "logtype"
	File        = "file"
	Pattern     = "pattern"
	Replacement = "replacement"
	Type        = "type"
	Choose      = "choose"
	Range       = "range"
	RePattern   = `####<(?P<timestamp>.*?)>\s*<(?P<severity>.*?)>\s*<(?P<subsystem>.*?)>\s*<(?P<machine>.*?)>\s*<(?P<server>.*?)>\s*<(?P<thread>.*?)>\s*<(?P<user>.*?)>\s*<(?P<transaction>.*?)>\s*<(?P<diagcontext>.*?)>\s*<(?P<rawtime>.*?)>\s*<(?P<msgid>.*?)>\s*<(?P<msgtext>.*?)>`
)

// The silly big all-in-one main function. Yes I will refactor it when I have some time. :-P
func main() {
	confPath := flag.String("f", "logspout.json", "specify the config file in json format.")
	flag.Parse()

	conf, err := ioutil.ReadFile(*confPath)
	if err != nil {
		LevelLog(ERROR, err)
		return
	}

	var logType, sampleFile, ptn string
	if logType, err = jsonparser.GetString(conf, LogType); err != nil {
		LevelLog(ERROR, err)
		return
	}

	if sampleFile, err = jsonparser.GetString(conf, File); err != nil {
		LevelLog(ERROR, err)
		return
	}

	if ptn, err = jsonparser.GetString(conf, Pattern); err != nil {
		LevelLog(ERROR, err)
		return
	}

	LevelLog(INFO, fmt.Sprintf("Loaded configurations from %s\n", *confPath))

	LevelLog(DEBUG, fmt.Sprintf("  - logtype = %s\n", logType))
	LevelLog(DEBUG, fmt.Sprintf("  - file = %s\n", sampleFile))
	LevelLog(DEBUG, fmt.Sprintf("  - pattern = %s", ptn))

	file, err := os.Open(sampleFile)
	if err != nil {
		LevelLog(ERROR, err)
		return
	}
	defer file.Close()

	re := regexp.MustCompile(ptn)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fmt.Printf("Raw: %s\n\n", scanner.Text())

		matches := re.FindStringSubmatch(scanner.Text())
		names := re.SubexpNames()

		for idx, match := range matches {
			fmt.Printf("  - %s: %s\n", names[idx], match)
		}
	}
	LevelLog(INFO, "Started. Check above matches and change patterns if something is wrong.\n")

	replace, _, _, err := jsonparser.Get(conf, Replacement)
	if err != nil {
		LevelLog(ERROR, err)
	}

	handler = func(key []byte, value []byte, dataType jsonparser.ValueType, offset int) error {
		var err error = nil
		return err
	}

	jsonparser.ObjectEach(replace, handler)

}
