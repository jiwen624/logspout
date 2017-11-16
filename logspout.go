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
	"github.com/Pallinder/go-randomdata"
	"github.com/buger/jsonparser"
	"github.com/jiwen624/logspout/gen"
	. "github.com/jiwen624/logspout/utils"
	"github.com/leesper/go_rng"
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
	LOOKSREAL   = "looks-real"
	NEXT        = "next"
	PREV        = "prev"
	RANDOM      = "random"
	FORMAT      = "format"
	CONCURRENY  = "concurrency"
	HIGHTIDE    = "hightide"
	RECONVERT   = "re-convert"
)

// LooksReal data methods
const (
	IPV4           = "ipv4"
	IPV4CHINA      = "ipv4china"
	CELLPHONECHINA = "cellphone-china"
	IPV6           = "ipv6"
	MAC            = "mac"
	UA             = "user-agent"
	COUNTRY        = "country"
	EMAIL          = "email"
	NAME           = "name"
	CHINESENAME    = "chinese-name"
	UUID           = "uuid"
)

// Control the speed of log bursts, in milliseconds.
var minInterval = 1000
var maxInterval = 1000
var concurrency = 1
var highTide = false
var reconvert = true

// The silly big all-in-one main function. Yes I will refactor it when I have some time. :-P
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
		ptn = reConvert(ptn)
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
		buffer.WriteString("\n") //Multi-line log support
	}

	rawMsg := strings.TrimRight(buffer.String(), "\n")

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
			replacerMap[k] = newFloatReplacer(min, max, precision)

		case STRING:
			min, err := jsonparser.GetInt(value, MIN)
			if err != nil {
				return errors.New(fmt.Sprintf("No %s found in %s", MIN, string(key)))
			}
			max, err := jsonparser.GetInt(value, MAX)
			if err != nil {
				return errors.New(fmt.Sprintf("No %s found in %s", MAX, string(key)))
			}
			replacerMap[k] = newStringReplacer(min, max)

		case LOOKSREAL:
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

// reConvert does the pre-process of the regular expression literal.
// It does the following things:
// 1. Remove P before captured group names
// 2. Add parenthesises to the other parts of the log event string.
func reConvert(ptn string) string {
	s := strings.Split(ptn, "")
	for i := 1; i < len(s)-1; i++ {
		if s[i] == "?" && s[i-1] == "(" && s[i+1] == "<" {
			// Replace "?" with "?P", it has a bug but works for 99% of the cases.
			// TODO: I'll keep it before I have time to write a better one.
			s[i] = "?P"
		}
	}
	return strings.Join(s, "")
}

// Replacer is the interface which must be implemented by a particular replacement policy.
type Replacer interface {
	// ReplacedValue returns the new replaced value.
	ReplacedValue(*rng.GaussianGenerator) (string, error)
}

// FixedListReplacer is a struct to record config options of a fixed-list replacement type.
type FixedListReplacer struct {
	method   string
	valRange []string
	currIdx  int
}

// newFixedListReplacer returns a new FixedListReplacer struct instance
func newFixedListReplacer(c string, v []string, ci int) Replacer {
	return &FixedListReplacer{
		method:   c,
		valRange: v,
		currIdx:  ci,
	}
}

// ReplacedValue returns a new replacement value of fixed-list type.
func (fl *FixedListReplacer) ReplacedValue(g *rng.GaussianGenerator) (string, error) {
	var newVal string

	switch fl.method {
	case NEXT:
		fl.currIdx = (fl.currIdx + 1) % len(fl.valRange)

	case RANDOM:
		fallthrough
	default:
		fl.currIdx = gen.SimpleGaussian(g, len(fl.valRange))
	}
	newVal = fl.valRange[fl.currIdx]
	return newVal, nil
}

type TimeStampReplacer struct {
	format string
}

// newTimeStampReplacer returns a new TimeStampReplacer struct instance.
func newTimeStampReplacer(f string) Replacer {
	return &TimeStampReplacer{
		format: f,
	}
}

// ReplacedValue populates a new timestamp with current time.
func (ts *TimeStampReplacer) ReplacedValue(*rng.GaussianGenerator) (string, error) {
	return jodaTime.Format(ts.format, time.Now()), nil
}

type StringReplacer struct {
	min int64
	max int64
}

func newStringReplacer(min int64, max int64) Replacer {
	return &StringReplacer{
		min: min,
		max: max,
	}
}

func (s *StringReplacer) ReplacedValue(g *rng.GaussianGenerator) (string, error) {
	var str string
	var err error
	if s.min == s.max {
		str = gen.GetRandomString(int(s.min))
	} else {
		l := rand.Intn(int(s.max-s.min)) + int(s.min)
		str = gen.GetRandomString(l)
	}
	return str, err
}

type FloatReplacer struct {
	min       int64
	max       int64
	precision int64
}

func newFloatReplacer(min int64, max int64, precision int64) Replacer {
	return &FloatReplacer{
		min:       min,
		max:       max,
		precision: precision,
	}
}

func (f *FloatReplacer) ReplacedValue(g *rng.GaussianGenerator) (string, error) {
	v := float64(f.min) + rand.Float64()*float64(f.max-f.min)
	s := fmt.Sprintf("%%.%df", f.precision)
	return fmt.Sprintf(s, v), nil
}

type IntegerReplacer struct {
	method  string
	min     int64
	max     int64
	currVal int64
}

// newIntegerReplacer returns a new IntegerReplacer struct instance
func newIntegerReplacer(c string, minV int64, maxV int64, cv int64) Replacer {
	return &IntegerReplacer{
		method:  c,
		min:     minV,
		max:     maxV,
		currVal: cv,
	}
}

// ReplacedValue is the main function to populate replacement value of an integer type.
func (i *IntegerReplacer) ReplacedValue(g *rng.GaussianGenerator) (string, error) {
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
		i.currVal = int64(gen.SimpleGaussian(g, int(i.max-i.min))) + i.min
	}
	return strconv.FormatInt(i.currVal, 10), nil
}

// LooksReal is a struct to record the configured method to generate data.
type LooksReal struct {
	method string
}

// newLooksReal returns a new LooksReal struct instance
func newLooksReal(m string) Replacer {
	return &LooksReal{
		method: m,
	}
}

// ReplacedValue returns random data based on the data type selection.
func (ia *LooksReal) ReplacedValue(g *rng.GaussianGenerator) (data string, err error) {
	switch ia.method {
	case IPV4:
		data = randomdata.IpV4Address()
	case IPV4CHINA:
		data = gen.GetRandomChinaIP(g)
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
	case CELLPHONECHINA:
		data = gen.GetRandomChinaCellPhoneNo(g)
	case CHINESENAME:
		data = gen.GetRandomChineseName(g)
	case MAC:
		data = randomdata.MacAddress()
	case UUID:
		data = gen.GetRandomUUID()
	}
	return data, nil
}
