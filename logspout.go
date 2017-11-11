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
	flag.Parse()

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

	re := regexp.MustCompile(ptn)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		LevelLog(DEBUG, "---------------------------------------------------\n")
		LevelLog(DEBUG, fmt.Sprintf("**Raw**: %s\n\n", scanner.Text()))

		matches := re.FindStringSubmatch(scanner.Text())
		names := re.SubexpNames()

		for idx, match := range matches {
			if idx == 0 {
				continue
			}
			LevelLog(DEBUG, fmt.Sprintf("  - %s: %s\n", names[idx], match))
		}
	}
	LevelLog(INFO, "Started. Check above matches and change patterns if something is wrong.\n")

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

		var vr []string
		// TODO: use jsonparser.ArrayEach to build the value range

		switch t {
		case "fixed-list":
			replacerMap[k] = newFixedList(c, vr, 0)
		}

		return err
	}

	jsonparser.ObjectEach(replace, handler)

	// goroutine for future use, not necessary for now.
	go PopNewLogs()
}

func PopNewLogs() {
	// TODO: replace the matched capture group with the replacerMap values.

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

var replacerMap map[string]Replacer = make(map[string]Replacer)
