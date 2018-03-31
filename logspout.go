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
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"io/ioutil"
	"log"
	"log/syslog"
	"math"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

// Options in the configure file.
const (
	LOGTYPE              = "logtype"
	OUTPUTSTDOUT         = "output-stdout"
	OUTPUTFILE           = "output-file"
	OUTPUTSYSLOG         = "output-syslog"
	SAMPLEFILE           = "sample-file"
	PATTERN              = "pattern"
	REPLACEMENT          = "replacement"
	TYPE                 = "type"
	METHOD               = "method"
	LIST                 = "list"
	MIN                  = "min"
	MAX                  = "max"
	MININTERVAL          = "min-interval"
	MAXINTERVAL          = "max-interval"
	DURATION             = "duration"
	LISTFILE             = "list-file"
	FIXEDLIST            = "fixed-list"
	TIMESTAMP            = "timestamp"
	INTEGER              = "integer"
	FLOAT                = "float"
	PRECISION            = "precision"
	STRING               = "string"
	CHARS                = "chars"
	LOOKSREAL            = "looks-real"
	FORMAT               = "format"
	CONCURRENY           = "concurrency"
	UNIFORM              = "uniform"
	HIGHTIDE             = "hightide"
	RECONVERT            = "re-convert"
	TRANSACTION          = "transaction"
	TRANSACTIONIDS       = "transaction-ids"
	MAXINTRATRANSLATENCY = "max-intra-transaction-latency"
)

// For output-file
const (
	FILENAME   = "file-name"
	MAXSIZE    = "max-size"
	MAXBACKUPS = "max-backups"
	MAXAGE     = "max-age"
	COMPRESS   = "compress"
)

// For output-syslog
const (
	PROTOCOL  = "protocol"
	NETADDR   = "netaddr"
	SYSLOGTAG = "tag"
)

// Control the speed of log bursts, in milliseconds.
var minInterval = 1000.0
var maxInterval = 1000.0
var duration = 0
var concurrency = 1
var highTide = false
var reconvert = true
var uniform = true
var trans = false
var transIds = make([]string, 0)
var rawMsgs = make([]string, 0)
var intraTransLat = 10

// The default log event output stream: stdout
var logger = log.New(os.Stdout, "", 0)

func main() {
	confPath := flag.String("f", "logspout.json", "specify the config file in json format.")
	level := flag.String("v", "info", "Print level: debug, info, warning, error.")
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

	var dests []io.Writer
	// We support regular files, syslog and stdout.
	if stdo, err := jsonparser.GetBoolean(conf, OUTPUTSTDOUT); err == nil {
		if stdo {
			dests = append(dests, os.Stdout)
		}
	}

	if ofile, _, _, err := jsonparser.Get(conf, OUTPUTFILE); err == nil {
		dests = append(dests, BuildOutputFileParms(ofile))
	}

	if osyslog, _, _, err := jsonparser.Get(conf, OUTPUTSYSLOG); err == nil {
		dests = append(dests, BuildOutputSyslogParms(osyslog))
	}

	defer func(dst []io.Writer) {
		for _, d := range dst {
			if d != nil {
				if dc, ok := d.(io.Closer); ok {
					dc.Close()
				}
			}
		}
	}(dests)

	// Set multiple destinations, if any
	logger.SetOutput(io.MultiWriter(dests...))

	if c, err := jsonparser.GetInt(conf, CONCURRENY); err == nil {
		concurrency = int(c)
	}

	if b, err := jsonparser.GetBoolean(conf, UNIFORM); err == nil {
		uniform = b
	}

	var logType, sampleFile string
	if logType, err = jsonparser.GetString(conf, LOGTYPE); err != nil {
		LevelLog(ERROR, err)
		return
	}

	if sampleFile, err = jsonparser.GetString(conf, SAMPLEFILE); err != nil {
		LevelLog(ERROR, err)
		return
	}

	if t, err := jsonparser.GetBoolean(conf, TRANSACTION); err == nil {
		trans = t
	}

	if i, err := jsonparser.GetInt(conf, MAXINTRATRANSLATENCY); err == nil && i != 0 {
		intraTransLat = int(i)
	}

	_, err = jsonparser.ArrayEach(conf, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		transIds = append(transIds, string(value))
	}, TRANSACTIONIDS)

	var ptns = make([]string, 0)
	_, err = jsonparser.ArrayEach(conf, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		ptns = append(ptns, string(value))
	}, PATTERN)

	if err != nil && trans == false {
		var ptn string
		if ptn, err = jsonparser.GetUnsafeString(conf, PATTERN); err != nil {
			LevelLog(ERROR, err)
			return
		}
		ptns = append(ptns, ptn)
	}

	if reconvert == true {
		for idx, ptn := range ptns {
			ptns[idx] = ReConvert(ptn)
		}
	}

	LevelLog(DEBUG, fmt.Sprintf("loaded configurations from %s\n", *confPath))

	LevelLog(DEBUG, fmt.Sprintf("  - logtype = %s\n", logType))
	LevelLog(DEBUG, fmt.Sprintf("  - file = %s\n", sampleFile))
	for idx, ptn := range ptns {
		LevelLog(DEBUG, fmt.Sprintf("  - pattern #%d = %s", idx, ptn))
	}

	file, err := os.Open(sampleFile)
	if err != nil {
		LevelLog(ERROR, err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	var buffer bytes.Buffer
	var vs string
	for scanner.Scan() {
		// Use blank line as the delimiter of a log event.
		if vs = scanner.Text(); vs == "" {
			rawMsgs = append(rawMsgs, strings.TrimRight(buffer.String(), "\n"))
			buffer.Reset()
			continue
		}
		buffer.WriteString(scanner.Text())
		buffer.WriteString("\n") //Multi-line log support
	}

	if buffer.Len() != 0 {
		rawMsgs = append(rawMsgs, strings.TrimRight(buffer.String(), "\n"))
	}

	if len(rawMsgs) != len(ptns) {
		LevelLog(ERROR, fmt.Sprintf("%d sample event(s) but %d pattern(s) found.", len(rawMsgs), len(ptns)))
		return
	}

	for idx, rawMsg := range rawMsgs {
		LevelLog(DEBUG, fmt.Sprintf("**raw message#%d**: %s\n\n", idx, rawMsg))
	}

	var matches = make([][]string, 0)
	var names = make([][]string, 0)

	for idx, ptn := range ptns {
		re := regexp.MustCompile(ptn)
		matches = append(matches, re.FindStringSubmatch(rawMsgs[idx]))
		names = append(names, re.SubexpNames())

		if len(matches[idx]) == 0 {
			LevelLog(ERROR, fmt.Sprintf("the re pattern doesn't match the sample log in #%d.", idx))
			return
		} else {
			// Remove the first one as it is the whole string.
			matches[idx] = matches[idx][1:]
			names[idx] = names[idx][1:]
		}
	}

	for idx, match := range matches {
		LevelLog(DEBUG, fmt.Sprintf("   pattern #%d", idx))
		for i, group := range match {
			LevelLog(DEBUG, fmt.Sprintf("       - %s: %s\n", names[idx][i], group))
		}
	}

	LevelLog(DEBUG, "check above matches and change patterns if something is wrong.\n")

	replace, _, _, err := jsonparser.Get(conf, REPLACEMENT)
	if err != nil {
		LevelLog(ERROR, err)
		return
	}

	if minI, err := jsonparser.GetInt(conf, MININTERVAL); err == nil {
		minInterval = float64(minI)
	}
	if maxI, err := jsonparser.GetInt(conf, MAXINTERVAL); err == nil {
		maxInterval = float64(maxI)
	}
	if d, err := jsonparser.GetInt(conf, DURATION); err == nil {
		duration = int(d) //I suppose you won't set a large number that makes an int overflow.
	}

	if minInterval > maxInterval {
		LevelLog(ERROR, errors.New("minInterval should be less than maxInterval"))
		return
	}

	var replacerMap map[string]gen.Replacer
	if replacerMap, err = BuildReplacerMap(replace); err != nil {
		LevelLog(ERROR, err)
		return
	}

	// goroutine for future use, not necessary for now.
	var wg sync.WaitGroup
	wg.Add(concurrency) // Add it before you start the goroutine.

	for i := 0; i < concurrency; i++ {
		LevelLog(DEBUG, fmt.Sprintf("spawned worker #%d\n", i))
		go PopNewLogs(logger, replacerMap, matches, names, &wg)
	}

	LevelLog(DEBUG, "LogSpout started.\n")
	wg.Wait()
	LevelLog(DEBUG, fmt.Sprintf("LogSpout ended after %d seconds.", duration))
}

// BuildReplacerMap builds and returns an string-Replacer map for future use.
func BuildReplacerMap(replace []byte) (map[string]gen.Replacer, error) {
	var replacerMap = make(map[string]gen.Replacer)

	handler := func(key []byte, value []byte, dataType jsonparser.ValueType, offset int) error {
		var err error = nil
		k := string(key)
		t, err := jsonparser.GetString(value, TYPE)
		if err != nil {
			return errors.New(fmt.Sprintf("no type found in %s", string(key)))
		}

		switch t {
		case FIXEDLIST:
			c, err := jsonparser.GetString(value, METHOD)
			if err != nil {
				return errors.New(fmt.Sprintf("no method found in %s", string(key)))
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
			min, err := jsonparser.GetFloat(value, MIN)
			if err != nil {
				return errors.New(fmt.Sprintf("No %s found in %s", MIN, string(key)))
			}
			max, err := jsonparser.GetFloat(value, MAX)
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

// BuildOutputSyslogParms extracts output parameters from the config file for the syslog output
func BuildOutputSyslogParms(out []byte) io.Writer {
	var protocol = "udp"
	var netaddr = "localhost:514"
	var level = syslog.LOG_INFO
	var tag = "logspout"

	if p, err := jsonparser.GetString(out, PROTOCOL); err == nil {
		protocol = p
	}

	if n, err := jsonparser.GetString(out, NETADDR); err == nil {
		netaddr = n
	}
	// TODO: The syslog default level is hardcoded for now.
	//if l, err := jsonparser.GetString(out, SYSLOGLEVEL); err == nil {
	//	level = l
	//}
	if t, err := jsonparser.GetString(out, SYSLOGTAG); err == nil {
		tag = t
	}
	w, err := syslog.Dial(protocol, netaddr, level, tag)
	if err != nil {
		LevelLog(ERROR, fmt.Sprintf("failed to connect to syslog destination: %s", netaddr))
	}
	return w
}

// BuildOutputFileParms extracts output parameters from the config file, if any.
func BuildOutputFileParms(out []byte) io.Writer {
	var fileName = "logspout_default.log"
	var maxSize = 100  // 100 Megabytes
	var maxBackups = 5 // 5 backups
	var maxAge = 7     // 7 days
	var compress = false
	var localTime = true

	if f, err := jsonparser.GetString(out, FILENAME); err == nil {
		fileName = f
	}
	if ms, err := jsonparser.GetInt(out, MAXSIZE); err == nil {
		maxSize = int(ms)
	}
	if mb, err := jsonparser.GetInt(out, MAXBACKUPS); err == nil {
		maxBackups = int(mb)
	}
	if ma, err := jsonparser.GetInt(out, MAXAGE); err == nil {
		maxAge = int(ma)
	}
	if c, err := jsonparser.GetBoolean(out, COMPRESS); err == nil {
		compress = c
	}
	return &lumberjack.Logger{
		Filename:   fileName,
		MaxSize:    maxSize, // megabytes
		MaxBackups: maxBackups,
		MaxAge:     maxAge,    //days
		Compress:   compress,  // disabled by default.
		LocalTime:  localTime, // always true for now.
	}
}

// PopNewLogs generates new logs with the replacement policies, in a infinite loop.
func PopNewLogs(logger *log.Logger, replacers map[string]gen.Replacer, m [][]string, names [][]string, wg *sync.WaitGroup) {
	var newLog string
	defer wg.Done()

	// Gaussian distribution
	grng := rng.NewGaussianGenerator(time.Now().UnixNano())

	matches := StrSlice2DCopy(m)
	var timeout <-chan time.Time = nil
	if duration != 0 {
		timeout = time.After(time.Second * time.Duration(duration))
	}
	var currMsg = 0

	for {
		// The first message of a transaction
		for k, v := range replacers {
			idx := StrIndex(names[currMsg], k)
			if idx == -1 {
				continue
			} else if currMsg == 0 || StrIndex(transIds, k) == -1 {
				if s, err := v.ReplacedValue(grng); err == nil {
					matches[currMsg][idx] = s
				}
			} else {
				matches[currMsg][idx] = matches[0][idx]
			}
		}

		newLog = strings.Join(matches[currMsg], "")
		// Print to stdout, you may redirect it to anywhere else you want
		logger.Println(newLog)

		// It never sleeps in hightide mode.
		if trans == true && highTide == false {
			time.Sleep(time.Millisecond * time.Duration(gen.SimpleGaussian(grng, intraTransLat)))
		}

		currMsg += 1
		if currMsg >= len(rawMsgs) {
			currMsg = 0

			// We will populate events as fast as possible in high tide mode. (Watch out your CPU!)
			if highTide == false {
				// Sleep for a short while.
				var sleepMsec = minInterval
				if maxInterval == minInterval {
					sleepMsec = minInterval
				} else {
					if uniform == true {
						sleepMsec = minInterval + float64(gen.SimpleGaussian(grng, int(maxInterval-minInterval)))
					} else { // There should be a better algorithm here.
						x := float64((time.Now().Unix() % 86400) / 13751)
						y := (math.Pow(math.Sin(x), 2) + math.Pow(math.Sin(x/2), 2) + 0.2) / 1.7619
						sleepMsec = minInterval / y
						if sleepMsec > maxInterval {
							sleepMsec = maxInterval
						}
					}
				}
				time.Sleep(time.Millisecond * time.Duration(int(sleepMsec)))
			}
		}

		select {
		case <-timeout:
			return
		default:
		}
	}
}
