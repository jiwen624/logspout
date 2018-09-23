package main

import (
	"bufio"
	"fmt"
	"io"
	"log/syslog"
	"os"
	"path/filepath"
	"strconv"

	"github.com/jiwen624/logspout/output"

	"github.com/buger/jsonparser"
	"github.com/jiwen624/logspout/config"
	"github.com/jiwen624/logspout/gen"
	"github.com/jiwen624/logspout/log"
	"gopkg.in/natefinch/lumberjack.v2"
)

// BuildReplacerMap builds and returns an string-Replacer map for future use.
func BuildReplacerMap(replace []byte) (map[string]gen.Replacer, error) {
	var replacerMap = make(map[string]gen.Replacer)

	handler := func(key []byte, value []byte, dataType jsonparser.ValueType, offset int) error {
		var err error
		k := string(key)
		notFound := func(string) error {
			return fmt.Errorf("no %s found in %s", config.MIN, k)
		}

		t, err := jsonparser.GetString(value, config.TYPE)
		if err != nil {
			return notFound(config.TYPE)
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

		p, _, _, errParms := jsonparser.Get(value, config.PARMS)

		// It is a normal case if PARMS is not found
		if errParms == nil {
			jsonparser.ObjectEach(p, pHandler)
		}

		switch t {
		case config.FIXEDLIST:
			c, err := jsonparser.GetString(value, config.METHOD)
			if err != nil {
				return notFound(config.METHOD)
			}
			var vr = make([]string, 0)
			_, err = jsonparser.ArrayEach(value, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
				vr = append(vr, string(value))
			}, config.LIST)
			// No list found
			if err != nil {
				f, err := jsonparser.GetString(value, config.LISTFILE)
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

		case config.TIMESTAMP:
			if tsFmt, err := jsonparser.GetString(value, config.FORMAT); err == nil {
				replacerMap[k] = gen.NewTimeStampReplacer(tsFmt)
			} else {
				return err
			}

		case config.INTEGER:
			c, err := jsonparser.GetString(value, config.METHOD)
			if err != nil {
				return notFound(config.METHOD)
			}
			min, err := jsonparser.GetInt(value, config.MIN)
			if err != nil {
				return notFound(config.MIN)
			}
			max, err := jsonparser.GetInt(value, config.MAX)
			if err != nil {
				return notFound(config.MAX)
			}
			replacerMap[k] = gen.NewIntegerReplacer(c, min, max, min)

		case config.FLOAT:
			min, err := jsonparser.GetFloat(value, config.MIN)
			if err != nil {
				return notFound(config.MIN)
			}
			max, err := jsonparser.GetFloat(value, config.MAX)
			if err != nil {
				return notFound(config.MAX)
			}

			precision, err := jsonparser.GetInt(value, config.PRECISION)
			if err != nil {
				return notFound(config.MIN)
			}
			replacerMap[k] = gen.NewFloatReplacer(min, max, precision)

		case config.STRING:
			var chars = ""
			min, err := jsonparser.GetInt(value, config.MIN)
			if err != nil {
				return notFound(config.MIN)
			}
			max, err := jsonparser.GetInt(value, config.MAX)
			if err != nil {
				return notFound(config.MAX)
			}

			if c, err := jsonparser.GetString(value, config.CHARS); err == nil {
				chars = c
			}
			replacerMap[k] = gen.NewStringReplacer(chars, min, max)

		case config.LOOKSREAL:
			c, err := jsonparser.GetString(value, config.METHOD)
			if err != nil {
				return notFound(config.METHOD)
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

	if p, err := jsonparser.GetString(out, output.PROTOCOL); err == nil {
		protocol = p
	}

	if n, err := jsonparser.GetString(out, output.NETADDR); err == nil {
		netaddr = n
	}
	// TODO: The syslog default level is hardcoded for now.
	//if l, err := jsonparser.GetString(out, SYSLOGLEVEL); err == nil {
	//	level = l
	//}
	if t, err := jsonparser.GetString(out, output.SYSLOGTAG); err == nil {
		tag = t
	}
	w, err := syslog.Dial(protocol, netaddr, level, tag)
	if err != nil {
		log.Errorf("failed to connect to syslog destination: %s", netaddr)
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

	if f, err := jsonparser.GetString(out, output.FILENAME); err == nil {
		fileName = f
	}
	if d, err := jsonparser.GetString(out, output.DIRECTORY); err == nil {
		directory = d
	}
	if ms, err := jsonparser.GetInt(out, output.MAXSIZE); err == nil {
		maxSize = int(ms)
	}
	if mb, err := jsonparser.GetInt(out, output.MAXBACKUPS); err == nil {
		maxBackups = int(mb)
	}
	if ma, err := jsonparser.GetInt(out, output.MAXAGE); err == nil {
		maxAge = int(ma)
	}
	if c, err := jsonparser.GetBoolean(out, output.COMPRESS); err == nil {
		compress = c
	}
	for i := 0; i < duplicate; i++ {
		loggers = append(loggers, &lumberjack.Logger{
			Filename:   filepath.Join(directory, strconv.Itoa(i)+"_"+fileName),
			MaxSize:    maxSize, // megabytes
			MaxBackups: maxBackups,
			MaxAge:     maxAge,   // days
			Compress:   compress, // disabled by default.
			LocalTime:  true,
		})
	}
	return loggers
}
