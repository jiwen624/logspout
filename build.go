package main

import (
	"bufio"
	"fmt"
	"github.com/buger/jsonparser"
	"github.com/jiwen624/logspout/gen"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"log/syslog"
	"os"
	"strconv"
	"path/filepath"
)

// BuildReplacerMap builds and returns an string-Replacer map for future use.
func BuildReplacerMap(replace []byte) (map[string]gen.Replacer, error) {
	var replacerMap = make(map[string]gen.Replacer)

	handler := func(key []byte, value []byte, dataType jsonparser.ValueType, offset int) error {
		var err error
		k := string(key)
		notFound := func(string) error {
			return fmt.Errorf("no %s found in %s", MIN, k)
		}

		t, err := jsonparser.GetString(value, TYPE)
		if err != nil {
			return notFound(TYPE)
		}

		var parms = make(map[string]interface{})

		pHandler := func(key []byte, value []byte, dataType jsonparser.ValueType, offset int) error {
			// We support only int, string or array.
			k := string(key)
			if dataType == jsonparser.Number {
				tn, _ := jsonparser.ParseInt(value)
				parms[k] = int(tn)
			} else if dataType == jsonparser.String {
				ts, _ := jsonparser.ParseString(value)
				parms[k] = ts
			} else if dataType == jsonparser.Array {
				var ts []string
				jsonparser.ArrayEach(value, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
					ts = append(ts, string(value)) // `value` should be a string value but no checks here.
				})
				parms[k] = ts
			}
			return nil
		}

		p, _, _, errParms := jsonparser.Get(value, PARMS)

		// It is a normal case if PARMS is not found
		if errParms == nil {
			jsonparser.ObjectEach(p, pHandler)
		}

		switch t {
		case FIXEDLIST:
			c, err := jsonparser.GetString(value, METHOD)
			if err != nil {
				return notFound(METHOD)
			}
			var vr = make([]string, 0)
			_, err = jsonparser.ArrayEach(value, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
				vr = append(vr, string(value))
			}, LIST)
			// No list found
			if err != nil {
				f, err := jsonparser.GetString(value, LISTFILE)
				if err != nil {
					return err
				}
				//Open sample file and fill into vr
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
				return notFound(METHOD)
			}
			min, err := jsonparser.GetInt(value, MIN)
			if err != nil {
				return notFound(MIN)
			}
			max, err := jsonparser.GetInt(value, MAX)
			if err != nil {
				return notFound(MAX)
			}
			replacerMap[k] = gen.NewIntegerReplacer(c, min, max, min)

		case FLOAT:
			min, err := jsonparser.GetFloat(value, MIN)
			if err != nil {
				return notFound(MIN)
			}
			max, err := jsonparser.GetFloat(value, MAX)
			if err != nil {
				return notFound(MAX)
			}

			precision, err := jsonparser.GetInt(value, PRECISION)
			if err != nil {
				return notFound(MIN)
			}
			replacerMap[k] = gen.NewFloatReplacer(min, max, precision)

		case STRING:
			var chars = ""
			min, err := jsonparser.GetInt(value, MIN)
			if err != nil {
				return notFound(MIN)
			}
			max, err := jsonparser.GetInt(value, MAX)
			if err != nil {
				return notFound(MAX)
			}

			if c, err := jsonparser.GetString(value, CHARS); err == nil {
				chars = c
			}
			replacerMap[k] = gen.NewStringReplacer(chars, min, max)

		case LOOKSREAL:
			c, err := jsonparser.GetString(value, METHOD)
			if err != nil {
				return notFound(METHOD)
			}
			gen.InitLooksRealParms(parms, c)
			replacerMap[k] = gen.NewLooksReal(c, parms)
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
		sugar.Errorf("failed to connect to syslog destination: %s", netaddr)
	}
	return w
}

// BuildOutputFileParms extracts output parameters from the config file, if any.
func BuildOutputFileParms(out []byte) []*lumberjack.Logger {
	var fileName = "logspout_default.log"
	var directory = "."
	var maxSize = 100  // 100 Megabytes
	var maxBackups = 5 // 5 backups
	var maxAge = 7     // 7 days
	var compress = false
	var loggers = make([]*lumberjack.Logger, 0)

	if f, err := jsonparser.GetString(out, FILENAME); err == nil {
		fileName = f
	}
	if d, err := jsonparser.GetString(out, DIRECTORY); err == nil {
		directory = d
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
	for i := 0; i < duplicate; i++ {
		loggers = append(loggers, &lumberjack.Logger{
			Filename:   filepath.Join(directory, strconv.Itoa(i) + "_" + fileName),
			MaxSize:    maxSize, // megabytes
			MaxBackups: maxBackups,
			MaxAge:     maxAge,   // days
			Compress:   compress, // disabled by default.
			LocalTime:  true,
		})
	}
	return loggers
}
