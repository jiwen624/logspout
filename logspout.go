// Package logspout implements a tiny tool to generate machine logs in accordance with
// the format of a sample log and the replacement policies defined in the config file.
// It can be used to generate logs in very high speed as well to do stress tests.
package main

import (
	"bufio"
	"bytes"
	"flag"
	"github.com/buger/jsonparser"
	"github.com/jiwen624/logspout/gen"
	. "github.com/jiwen624/logspout/utils"
	"github.com/leesper/go_rng"
	"io"
	"io/ioutil"
	"log"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// Options in the configure file.
const (
	LOGTYPE              = "logtype"
	OUTPUTSTDOUT         = "output-stdout"
	OUTPUTFILE           = "output-file"
	OUTPUTSYSLOG         = "output-syslog"
	CONSOLEPORT          = "console-port"
	SAMPLEFILE           = "sample-file"
	PATTERN              = "pattern"
	REPLACEMENT          = "replacement"
	TYPE                 = "type"
	PARMS                = "parms"
	METHOD               = "method"
	LIST                 = "list"
	MIN                  = "min"
	MAX                  = "max"
	MININTERVAL          = "min-interval"
	MAXINTERVAL          = "max-interval"
	DURATION             = "duration"
	MAXEVENTS            = "max-events"
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
	DUPLICATE            = "duplicate"
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
	DIRECTORY = "directory"
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

// Counter stores the counter values returned to the client
type Counter struct {
	Workers []uint64 `json:"Workers"`
	Total   uint64   `json:"TotalEPS"`
	Conf    string   `json:"ConfigFile"`
}

// Control the speed of log bursts, in milliseconds.
var minInterval = 1000.0
var maxInterval = 1000.0
var duration = 0
var maxEvents uint64 = math.MaxUint64
var concurrency = 1
var duplicate = 1
var consolePort = "10306"
var highTide = false
var reconvert = true
var uniform = true
var trans = false
var transIds = make([]string, 0)
var rawMsgs = make([]string, 0)
var intraTransLat = 10

// For fetching the counter values
var wgCounter sync.WaitGroup
var mCounter = sync.Mutex{}
var cCounter = sync.NewCond(&mCounter)
var reqCounter = false
var resChan = make(chan uint64)

// termChans stores the channels for close requests
var termChans = make([]chan struct{}, 0, concurrency)

// The default log event output stream: stdout
var logger = log.New(os.Stdout, "", 0)
var confPath *string

var sugar *Logger
var conf []byte
var err error

func main() {
	confPath = flag.String("f", "logspout.json", "specify the config file in json format")
	level := flag.String("v", "info", "Print level: debug, info, warning, error")
	flag.Parse()

	SetGlobalDebugLevel(*level)

	sugar = GetSugaredLogger()
	defer sugar.Sync()

	conf, err = ioutil.ReadFile(*confPath)
	if err != nil {
		sugar.Error(err)
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

	if c, err := jsonparser.GetInt(conf, DUPLICATE); err == nil {
		duplicate = int(c)
	}

	if p, err := jsonparser.GetInt(conf, CONSOLEPORT); err == nil {
		consolePort = strconv.Itoa(int(p))
	}

	if ofile, _, _, err := jsonparser.Get(conf, OUTPUTFILE); err == nil {
		for _, f := range BuildOutputFileParms(ofile) {
			dests = append(dests, f)
		}
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
		sugar.Error(err)
		return
	}

	if sampleFile, err = jsonparser.GetString(conf, SAMPLEFILE); err != nil {
		sugar.Error(err)
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
			sugar.Error(err)
			return
		}
		ptns = append(ptns, ptn)
	}

	if reconvert == true {
		for idx, ptn := range ptns {
			ptns[idx] = ReConvert(ptn)
		}
	}

	sugar.Debugf("loaded configurations from %s", *confPath)

	sugar.Debugf("  - logtype = %s", logType)
	sugar.Debugf("  - file = %s", sampleFile)
	for idx, ptn := range ptns {
		sugar.Debugf("  - pattern #%d = %s", idx, ptn)
	}

	file, err := os.Open(sampleFile)
	if err != nil {
		sugar.Error(err)
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
		sugar.Errorf("%d sample event(s) but %d pattern(s) found", len(rawMsgs), len(ptns))
		return
	}

	for idx, rawMsg := range rawMsgs {
		sugar.Debugf("**raw message#%d**: %s", idx, rawMsg)
	}

	var matches = make([][]string, 0)
	var names = make([][]string, 0)

	for idx, ptn := range ptns {
		re := regexp.MustCompile(ptn)
		matches = append(matches, re.FindStringSubmatch(rawMsgs[idx]))
		names = append(names, re.SubexpNames())

		if len(matches[idx]) == 0 {
			sugar.Errorf("the re pattern doesn't match the sample log in #%d", idx)
			return
		}

		// Remove the first one as it is the whole string.
		matches[idx] = matches[idx][1:]
		names[idx] = names[idx][1:]
	}

	for idx, match := range matches {
		sugar.Debugf("   pattern #%d", idx)
		for i, group := range match {
			sugar.Debugf("       - %s: %s", names[idx][i], group)
		}
	}

	sugar.Debug("check above matches and change patterns if something is wrong")

	replace, _, _, err := jsonparser.Get(conf, REPLACEMENT)
	if err != nil {
		sugar.Error(err)
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

	if me, err := jsonparser.GetInt(conf, MAXEVENTS); err == nil {
		maxEvents = uint64(me)
	}

	if minInterval > maxInterval {
		sugar.Error("minInterval should be less than maxInterval")
		return
	}

	// goroutine for future use, not necessary for now.
	var wg sync.WaitGroup
	wg.Add(concurrency) // Add it before you start the goroutine.

	for i := 0; i < concurrency; i++ {
		sugar.Debugf("spawned worker #%d", i)

		termChans = append(termChans, make(chan struct{}))

		var replacerMap map[string]gen.Replacer
		if replacerMap, err = BuildReplacerMap(replace); err != nil {
			sugar.Error(err)
			return
		}
		go PopNewLogs(logger, replacerMap, matches, names, &wg, termChans[i])
	}

	go console()

	sugar.Debug("LogSpout started")

	if duration != 0 {
		select {
		case <-time.After(time.Second * time.Duration(duration)):
			for _, c := range termChans {
				close(c)
			}
			termChans = make([]chan struct{}, 0)
		}
	}

	wg.Wait()
	sugar.Debugf("LogSpout ended")
}

// PopNewLogs generates new logs with the replacement policies, in a infinite loop.
func PopNewLogs(logger *log.Logger, replacers map[string]gen.Replacer, m [][]string,
	names [][]string, wg *sync.WaitGroup, terminate chan struct{}) {
	var newLog string
	defer wg.Done()

	// Gaussian distribution
	grng := rng.NewGaussianGenerator(time.Now().UnixNano())

	matches := StrSlice2DCopy(m)

	var currMsg int
	var counter uint64
	var totalCnt uint64

	var c uint64

	// This goroutine waits for the request from client to fetch the current counter value.
	go func(res chan uint64) {
		for {
			cCounter.L.Lock()
			for reqCounter == false {
				cCounter.Wait()
			}
			cCounter.L.Unlock()
			wgCounter.Done()

			res <- atomic.LoadUint64(&c)
		}
	}(resChan)

	cTicker := time.NewTicker(time.Second * 1).C
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
		// Print to logger streams, you may redirect it to anywhere else you want
		logger.Println(newLog)
		counter++
		// Exits after it exceeds the predefined maximum events.
		totalCnt++
		if totalCnt >= maxEvents/uint64(concurrency) {
			return
		}

		// It never sleeps in hightide mode.
		if trans == true && highTide == false {
			time.Sleep(time.Millisecond * time.Duration(gen.SimpleGaussian(grng, intraTransLat)))
		}

		currMsg++
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
		case <-terminate:
			return
		case <-cTicker:
			atomic.StoreUint64(&c, counter)
			counter = 0
		default:
		}
	}
}
