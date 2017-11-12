package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"github.com/buger/jsonparser"
	"io/ioutil"
	"math/rand"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

const (
	LOGTYPE     = "logtype"
	FILE        = "file"
	PATTERN     = "pattern"
	REPLACEMENT = "replacement"
	TYPE        = "type"
	CHOOSE      = "choose"
	RANGE       = "range"
)

// The silly big all-in-one main function. Yes I will refactor it when I have some time. :-P
func main() {
	confPath := flag.String("f", "logspout.json", "specify the config file in json format.")
	// TODO: not a good idea.
	level := flag.Int("v", 1, "Print level: 0->debug, 1->info, 2->warn, 3->error.")
	flag.Parse()

	globalLevel = DebugLevel(*level)

	conf, err := ioutil.ReadFile(*confPath)
	if err != nil {
		LevelLog(ERROR, err)
		return
	}

	var logType, sampleFile, ptn string
	if logType, err = jsonparser.GetString(conf, LOGTYPE); err != nil {
		LevelLog(ERROR, err)
		return
	}

	if sampleFile, err = jsonparser.GetString(conf, FILE); err != nil {
		LevelLog(ERROR, err)
		return
	}

	if ptn, err = jsonparser.GetString(conf, PATTERN); err != nil {
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

	var replacerMap map[string]Replacer = make(map[string]Replacer)
	var matches, names []string

	re := regexp.MustCompile(ptn)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		LevelLog(DEBUG, "---------------------------------------------------\n")
		LevelLog(DEBUG, fmt.Sprintf("**Raw**: %s\n\n", scanner.Text()))

		matches = re.FindStringSubmatch(scanner.Text())
		names = re.SubexpNames()

		// Remove the first one as it is the whole string.
		matches = matches[1:]
		names = names[1:]

		for idx, match := range matches {
			if idx == 0 {
				continue
			}
			LevelLog(DEBUG, fmt.Sprintf("  - %s: %s\n", names[idx], match))
		}
		continue // Currently it supports only one line in a file.
	}
	LevelLog(DEBUG, "Check above matches and change patterns if something is wrong.\n")
	LevelLog(INFO, "Started.\n")

	replace, _, _, err := jsonparser.Get(conf, REPLACEMENT)
	if err != nil {
		LevelLog(ERROR, err)
	}

	// Build the replacer map
	handler := func(key []byte, value []byte, dataType jsonparser.ValueType, offset int) error {
		var err error = nil
		k := string(key)
		t, err := jsonparser.GetString(value, "type")
		if err != nil {
			return errors.New(fmt.Sprintf("No type found in %s", string(key)))
		}

		c, err := jsonparser.GetString(value, "choose")
		if err != nil {
			return errors.New(fmt.Sprintf("No choose found in %s", string(key)))
		}

		var vr []string = make([]string, 0)

		jsonparser.ArrayEach(value, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
			vr = append(vr, string(value))
		}, "range")

		switch t {
		case "fixed-list":
			replacerMap[k] = newFixedList(c, vr, 0)
		case "timestamp":
			// TODO
		case "integer":
			// TODO
		}

		return err
	}

	jsonparser.ObjectEach(replace, handler)

	// goroutine for future use, not necessary for now.

	var wg sync.WaitGroup
	wg.Add(1) // Add it before you start the goroutine.
	go PopNewLogs(replacerMap, matches, names, wg)
	wg.Wait()
}

// PopNewLogs generates new logs with the replacement policies, in a infinite loop.
func PopNewLogs(replacers map[string]Replacer, matches []string, names []string, wg sync.WaitGroup) {
	var newLog string
	defer wg.Done()
	for {
		for k, v := range replacers {
			idx := StrIndex(names, k)
			if idx == -1 {
				continue
			}
			if s, err := v.ReplacedValue(); err == nil {
				matches[idx] = s
			}
		}

		newLog = strings.Join(matches, "")
		// Print to stdout, you may redirect it to anywhere else you want
		fmt.Fprintln(os.Stdout, newLog)
		time.Sleep(time.Second * 1) // TODO: configurable
	}
	// I never quit...
}

type Replacer interface {
	// ReplacedValue returns the new replaced value.
	ReplacedValue() (string, error)
}

type FixedList struct {
	choose   string
	valRange []string
	currIdx  int
}

func newFixedList(c string, v []string, ci int) Replacer {
	return &FixedList{
		choose:   c,
		valRange: v,
		currIdx:  ci,
	}
}

func (fl *FixedList) ReplacedValue() (string, error) {
	var newVal string

	switch fl.choose {
	case "random":
		fl.currIdx = rand.Intn(len(fl.valRange))
	case "inorder":
		fl.currIdx = (fl.currIdx + 1) / len(fl.valRange)
	}
	newVal = fl.valRange[fl.currIdx]
	return newVal, nil
}
