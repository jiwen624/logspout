// Package logspout implements a tiny tool to generate machine logs in accordance with
// the format of a sample log and the replacement policies defined in the config file.
// It can be used to generate logs in very high speed as well to do stress tests.
package main

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"github.com/buger/jsonparser"
	"github.com/jiwen624/logspout/gen"
	. "github.com/jiwen624/logspout/utils"
	"github.com/leesper/go_rng"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

// Options in the configure file.
const (
	LOGTYPE     = "logtype"
	SAMPLEFILE  = "sample-file"
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
	FLOAT       = "float"
	PRECISION   = "precision"
	STRING      = "string"
	CHARS       = "chars"
	LOOKSREAL   = "looks-real"
	FORMAT      = "format"
	CONCURRENY  = "concurrency"
	HIGHTIDE    = "hightide"
	RECONVERT   = "re-convert"
)

// Control the speed of log bursts, in milliseconds.
var minInterval = 1000
var maxInterval = 1000
var concurrency = 1
var highTide = false
var reconvert = true

func main() {
	confPath := flag.String("f", "logspout.json", "specify the config file in json format.")
	level := flag.String("v", "warning", "Print level: debug, info, warning, error.")
	flag.Parse()

	if val, ok := LevelsDbg[*level]; ok {
		SetGlobalDebugLevel(DebugLevel(val))
	} else {
		SetGlobalDebugLevel(INFO)
	}

	conf, err := ioutil.ReadFile(*confPath)
	if err != nil {
		LevelLog(ERROR, err)
		return
	}

	if h, err := jsonparser.GetBoolean(conf, RECONVERT); err == nil {
		reconvert = h
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

	if sampleFile, err = jsonparser.GetString(conf, SAMPLEFILE); err != nil {
		LevelLog(ERROR, err)
		return
	}

	if ptn, err = jsonparser.GetUnsafeString(conf, PATTERN); err != nil {
		LevelLog(ERROR, err)
		return
	}

	if reconvert == true {
		ptn = ReConvert(ptn)
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

	var matches, names []string
	re := regexp.MustCompile(ptn)
	scanner := bufio.NewScanner(file)

	var buffer bytes.Buffer
	for scanner.Scan() {
		buffer.WriteString(scanner.Text())
		buffer.WriteString("\n") //Multi-line log support
	}

	rawMsg := strings.TrimRight(buffer.String(), "\n")

	LevelLog(DEBUG, fmt.Sprintf("**Raw message**: %s\n\n", rawMsg))

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

	var replacerMap map[string]gen.Replacer
	if replacerMap, err = BuildReplacerMap(replace); err != nil {
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

// BuildReplacerMap builds and returns an string-Replacer map for future use.
func BuildReplacerMap(replace []byte) (map[string]gen.Replacer, error) {
	var replacerMap = make(map[string]gen.Replacer)

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
			var vr = make([]string, 0)
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
			replacerMap[k] = gen.NewFixedListReplacer(c, vr, 0)

		case TIMESTAMP:
			if tsFmt, err := jsonparser.GetString(value, FORMAT); err == nil {
				replacerMap[k] = gen.NewTimeStampReplacer(tsFmt)
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
			replacerMap[k] = gen.NewIntegerReplacer(c, min, max, min)

		case FLOAT:
			min, err := jsonparser.GetInt(value, MIN)
			if err != nil {
				return errors.New(fmt.Sprintf("No %s found in %s", MIN, string(key)))
			}
			max, err := jsonparser.GetInt(value, MAX)
			if err != nil {
				return errors.New(fmt.Sprintf("No %s found in %s", MAX, string(key)))
			}

			precision, err := jsonparser.GetInt(value, PRECISION)
			if err != nil {
				return errors.New(fmt.Sprintf("No %s found in %s", MIN, string(key)))
			}
			replacerMap[k] = gen.NewFloatReplacer(min, max, precision)

		case STRING:
			var chars = ""
			min, err := jsonparser.GetInt(value, MIN)
			if err != nil {
				return errors.New(fmt.Sprintf("No %s found in %s", MIN, string(key)))
			}
			max, err := jsonparser.GetInt(value, MAX)
			if err != nil {
				return errors.New(fmt.Sprintf("No %s found in %s", MAX, string(key)))
			}

			if c, err := jsonparser.GetString(value, CHARS); err == nil {
				chars = c
			}
			replacerMap[k] = gen.NewStringReplacer(chars, min, max)

		case LOOKSREAL:
			c, err := jsonparser.GetString(value, METHOD)
			if err != nil {
				return errors.New(fmt.Sprintf("No %s found in %s", METHOD, string(key)))
			}
			replacerMap[k] = gen.NewLooksReal(c)
		}
		return err
	}

	err := jsonparser.ObjectEach(replace, handler)
	return replacerMap, err
}

// PopNewLogs generates new logs with the replacement policies, in a infinite loop.
func PopNewLogs(replacers map[string]gen.Replacer, matches []string, names []string, wg sync.WaitGroup) {
	var newLog string
	defer wg.Done()

	// Gaussian distribution
	grng := rng.NewGaussianGenerator(time.Now().UnixNano())

	for {
		for k, v := range replacers {
			idx := StrIndex(names, k)
			if idx == -1 {
				continue
			}
			if s, err := v.ReplacedValue(grng); err == nil {
				matches[idx] = s
			}
		}

		newLog = strings.Join(matches, "")
		// Print to stdout, you may redirect it to anywhere else you want
		fmt.Fprintln(os.Stdout, newLog)
		var sleepMsec = 1000
		if maxInterval == minInterval {
			sleepMsec = minInterval
		} else {
			gap := maxInterval - minInterval
			sleepMsec = minInterval + gen.SimpleGaussian(grng, gap)
		}
		// We will populate events as fast as possible in high tide mode. (Watch out your CPU!)
		if highTide == false {
			time.Sleep(time.Millisecond * time.Duration(sleepMsec))
		}
	}
	// I never quit...
}
