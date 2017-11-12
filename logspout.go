package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"github.com/buger/jsonparser"
	"github.com/vjeantet/jodaTime"
	"io/ioutil"
	"math/rand"
	"os"
	"regexp"
	"strconv"
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
	METHOD      = "method"
	LIST        = "list"
	MIN         = "min"
	MAX         = "max"
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

		switch t {
		case "fixed-list":
			c, err := jsonparser.GetString(value, METHOD)
			if err != nil {
				return errors.New(fmt.Sprintf("No method found in %s", string(key)))
			}
			var vr []string = make([]string, 0)
			jsonparser.ArrayEach(value, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
				vr = append(vr, string(value))
			}, "list")
			replacerMap[k] = newFixedListReplacer(c, vr, 0)

		case "timestamp":
			if tsFmt, err := jsonparser.GetString(value, "format"); err == nil {
				replacerMap[k] = newTimeStampReplacer(tsFmt)
			} else {
				LevelLog(WARNING, err)
			}

		case "integer":
			c, err := jsonparser.GetString(value, METHOD)
			if err != nil {
				return errors.New(fmt.Sprintf("No %s found in %s", METHOD, string(key)))
			}
			min, err := jsonparser.GetInt(value, MIN)
			if err != nil {
				return errors.New(fmt.Sprintf("No %s found in %s", MIN, string(key)))
			}
			max, err := jsonparser.GetInt(value, MAX)
			if err != nil {
				return errors.New(fmt.Sprintf("No %s found in %s", MAX, string(key)))
			}
			replacerMap[k] = newIntegerReplacer(c, min, max, min)
		}
		return err
	}

	if err := jsonparser.ObjectEach(replace, handler); err != nil {
		LevelLog(ERROR, err)
		return
	}

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

type FixedListReplacer struct {
	method   string
	valRange []string
	currIdx  int
}

func newFixedListReplacer(c string, v []string, ci int) Replacer {
	return &FixedListReplacer{
		method:   c,
		valRange: v,
		currIdx:  ci,
	}
}

func (fl *FixedListReplacer) ReplacedValue() (string, error) {
	var newVal string

	switch fl.method {
	case "random":
		fl.currIdx = rand.Intn(len(fl.valRange))
	case "inorder":
		fl.currIdx = (fl.currIdx + 1) / len(fl.valRange)
	}
	newVal = fl.valRange[fl.currIdx]
	return newVal, nil
}

type TimeStampReplacer struct {
	format string
}

func newTimeStampReplacer(f string) Replacer {
	return &TimeStampReplacer{
		format: f,
	}
}

func (ts *TimeStampReplacer) ReplacedValue() (string, error) {
	return jodaTime.Format(ts.format, time.Now()), nil
}

type IntegerReplacer struct {
	method  string
	min     int64
	max     int64
	currVal int64
}

func newIntegerReplacer(c string, minV int64, maxV int64, cv int64) Replacer {
	return &IntegerReplacer{
		method:  c,
		min:     minV,
		max:     maxV,
		currVal: cv,
	}
}

func (i *IntegerReplacer) ReplacedValue() (string, error) {
	switch i.method {
	case "increase":
		i.currVal += 1
		if i.currVal > i.max {
			i.currVal = i.min
		}
	case "decrease":
		i.currVal -= 1
		if i.currVal < i.min {
			i.currVal = i.max
		}
	case "random":
	default:
		i.currVal = rand.Int63n(i.max-i.min) + i.min

	}
	return strconv.FormatInt(i.currVal, 10), nil
}
