package main

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"github.com/Pallinder/go-randomdata"
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
	MININTERVAL = "min-interval"
	MAXINTERVAL = "max-interval"
	LISTFILE    = "list-file"
	FIXEDLIST   = "fixed-list"
	TIMESTAMP   = "timestamp"
	INTEGER     = "integer"
	LOOKSREAL   = "looks-real"
	NEXT        = "next"
	PREV        = "prev"
	RANDOM      = "random"
	FORMAT      = "format"
	CONCURRENY  = "concurrency"
	HIGHTIDE    = "hightide"
)

// LooksReal data methods
const (
	IPV4      = "ipv4"
	IPV4CHINA = "ipv4china"
	IPV6      = "ipv6"
	UA        = "user-agent"
	COUNTRY   = "country"
	EMAIL     = "email"
	NAME      = "name"
)

// Control the speed of log bursts, in milliseconds.
var minInterval = 1000
var maxInterval = 1000
var concurrency = 1
var highTide = false

// The silly big all-in-one main function. Yes I will refactor it when I have some time. :-P
func main() {
	confPath := flag.String("f", "logspout.json", "specify the config file in json format.")
	level := flag.String("v", "warning", "Print level: debug, info, warning, error.")
	flag.Parse()

	if val, ok := levelsDbg[*level]; ok {
		globalLevel = DebugLevel(val)
	} else {
		globalLevel = INFO
	}

	conf, err := ioutil.ReadFile(*confPath)
	if err != nil {
		LevelLog(ERROR, err)
		return
	}

	if h, err := jsonparser.GetBoolean(conf, HIGHTIDE); err == nil {
		highTide = h
	}

	if c, err := jsonparser.GetInt(conf, CONCURRENY); err == nil {
		concurrency = int(c)
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

	var replacerMap = make(map[string]Replacer)
	var matches, names []string

	re := regexp.MustCompile(ptn)
	scanner := bufio.NewScanner(file)

	var buffer bytes.Buffer
	for scanner.Scan() {
		buffer.WriteString(scanner.Text())
	}
	rawMsg := buffer.String()

	LevelLog(DEBUG, "---------------------------------------------------\n")
	LevelLog(DEBUG, fmt.Sprintf("**Raw**: %s\n\n", rawMsg))

	matches = re.FindStringSubmatch(rawMsg)
	names = re.SubexpNames()

	if len(matches) == 0 {
		LevelLog(ERROR, "The re pattern doesn't match the sample log.")
		return
	} else {
		// Remove the first one as it is the whole string.
		matches = matches[1:]
		names = names[1:]
	}

	for idx, match := range matches {
		LevelLog(DEBUG, fmt.Sprintf("  - %s: %s\n", names[idx], match))
	}

	LevelLog(DEBUG, "Check above matches and change patterns if something is wrong.\n")

	replace, _, _, err := jsonparser.Get(conf, REPLACEMENT)
	if err != nil {
		LevelLog(ERROR, err)
	}

	if minI, err := jsonparser.GetInt(conf, MININTERVAL); err == nil {
		minInterval = int(minI)
	}
	if maxI, err := jsonparser.GetInt(conf, MAXINTERVAL); err == nil {
		maxInterval = int(maxI)
	}

	// Build the replacer map
	handler := func(key []byte, value []byte, dataType jsonparser.ValueType, offset int) error {
		var err error = nil
		k := string(key)
		t, err := jsonparser.GetString(value, TYPE)
		if err != nil {
			return errors.New(fmt.Sprintf("No type found in %s", string(key)))
		}

		switch t {
		case FIXEDLIST:
			c, err := jsonparser.GetString(value, METHOD)
			if err != nil {
				return errors.New(fmt.Sprintf("No method found in %s", string(key)))
			}
			var vr []string = make([]string, 0)
			_, err = jsonparser.ArrayEach(value, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
				vr = append(vr, string(value))
			}, LIST)
			// No list found
			if err != nil {
				if f, err := jsonparser.GetString(value, LISTFILE); err != nil {
					return err
				} else { //Open sample file and fill into vr
					fp, err := os.Open(f)
					if err != nil {
						return err
					}
					defer fp.Close()
					s := bufio.NewScanner(fp)
					for s.Scan() {
						vr = append(vr, s.Text())
					}
				}

			}
			replacerMap[k] = newFixedListReplacer(c, vr, 0)

		case TIMESTAMP:
			if tsFmt, err := jsonparser.GetString(value, FORMAT); err == nil {
				replacerMap[k] = newTimeStampReplacer(tsFmt)
			} else {
				return err
			}

		case INTEGER:
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

		case LOOKSREAL:
			// TODO: An empty string will be returned by the ReplacedValue() if the method is invalid.
			c, err := jsonparser.GetString(value, METHOD)
			if err != nil {
				return errors.New(fmt.Sprintf("No %s found in %s", METHOD, string(key)))
			}
			replacerMap[k] = newLooksReal(c)
		}
		return err
	}

	if err := jsonparser.ObjectEach(replace, handler); err != nil {
		LevelLog(ERROR, err)
		return
	}

	// goroutine for future use, not necessary for now.
	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		wg.Add(1) // Add it before you start the goroutine.
		LevelLog(DEBUG, fmt.Sprintf("Spawned worker #%d\n", i))
		go PopNewLogs(replacerMap, matches, names, wg)
	}
	LevelLog(INFO, "Started.\n")
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
		var sleepMsec int = 1000
		if maxInterval == minInterval {
			sleepMsec = minInterval
		} else {
			sleepMsec = minInterval + rand.Intn(maxInterval-minInterval)

		}
		// We will populate events as fast as possible in high tide mode. (Watch out your CPU!)
		if highTide == false {
			time.Sleep(time.Millisecond * time.Duration(sleepMsec))
		}
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
	case NEXT:
		fl.currIdx = (fl.currIdx + 1) % len(fl.valRange)

	case RANDOM:
		fallthrough
	default:
		fl.currIdx = rand.Intn(len(fl.valRange))
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
	case NEXT:
		i.currVal += 1
		if i.currVal > i.max {
			i.currVal = i.min
		}
	case PREV:
		i.currVal -= 1
		if i.currVal < i.min {
			i.currVal = i.max
		}
	case RANDOM:
		fallthrough
	default: // Use random by default
		i.currVal = rand.Int63n(i.max-i.min) + i.min
	}
	return strconv.FormatInt(i.currVal, 10), nil
}

type LooksReal struct {
	method string
}

func newLooksReal(m string) Replacer {
	return &LooksReal{
		method: m,
	}
}

func (ia *LooksReal) ReplacedValue() (data string, err error) {
	switch ia.method {
	case IPV4:
		data = randomdata.IpV4Address()
	case IPV4CHINA:
		data = GetRandomChinaIP()
	case IPV6:
		data = randomdata.IpV6Address()
	case UA:
		data = randomdata.UserAgentString()
	case COUNTRY:
		data = randomdata.Country(randomdata.FullCountry)
	case EMAIL:
		data = randomdata.Email()
	case NAME:
		data = randomdata.SillyName()
	}
	return data, nil
}
