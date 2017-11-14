// Utils

package main

import (
	"fmt"
	"github.com/leesper/go_rng"
	"log"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type DebugLevel uint8

const (
	DEBUG DebugLevel = iota
	INFO
	WARNING
	ERROR
)

var dbgLevels = map[DebugLevel]string{
	DEBUG:   "DEBUG",
	INFO:    "INFO ",
	WARNING: "WARN ",
	ERROR:   "ERROR",
}

// Reversed map
var levelsDbg = map[string]DebugLevel{
	"debug":   DEBUG,
	"info":    INFO,
	"warning": WARNING,
	"error":   ERROR,
}

var globalLevel DebugLevel = INFO
var logger = log.New(os.Stderr, "", log.LstdFlags)

// LevelLog print logs based on the debug level.
func LevelLog(level DebugLevel, err interface{}, args ...interface{}) {
	if level < globalLevel {
		return
	}

	var msg string
	switch err.(type) {
	case error:
		msg = err.(error).Error()
	case string:
		msg = err.(string)
	default:
		logger.Printf("Unknown err type: %T\n", err)
		return
	}

	var p string = fmt.Sprintf("%s ", dbgLevels[level])
	if globalLevel == DEBUG {
		if pc, f, l, ok := runtime.Caller(1); ok {
			path := strings.Split(runtime.FuncForPC(pc).Name(), ".")
			name := path[len(path)-1]
			p = fmt.Sprintf("%s %s#%d %s(): ", dbgLevels[level], filepath.Base(f), l, name)
		} else {
			p = fmt.Sprintf("%s %s#%s %s(): ", level, "na", "na", "na")
		}
	}
	if len(args) == 0 {
		logger.Printf("%s%s", p, msg)
	} else {
		logger.Printf("%s%s %v", p, msg, args)
	}
}

// Helper function to find the position of a string in a []string
func StrIndex(vs []string, t string) int {
	for i, v := range vs {
		if v == t {
			return i
		}
	}
	return -1
}

// mean=0.5*the_range, stddev=0.2*the_range
func SimpleGaussian(g *rng.GaussianGenerator, gap int) int {
	return int(math.Abs(g.Gaussian(0.5*float64(gap), 0.2*float64(gap)))) % gap
}
