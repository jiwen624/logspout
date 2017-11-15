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

// The debug levels
const (
	DEBUG DebugLevel = iota
	INFO
	WARNING
	ERROR
)

// dbgLevels is a map of level macros and strings.
var dbgLevels = map[DebugLevel]string{
	DEBUG:   "DEBUG",
	INFO:    "INFO ",
	WARNING: "WARN ",
	ERROR:   "ERROR",
}

// levelsDbg is a reversed map of level macros and strings.
var levelsDbg = map[string]DebugLevel{
	"debug":   DEBUG,
	"info":    INFO,
	"warning": WARNING,
	"error":   ERROR,
}

// globalLevel is the global variable for global debug level.
var globalLevel DebugLevel = INFO

// logger is the global log print object.
var logger = log.New(os.Stderr, "", log.LstdFlags)

// LevelLog prints logs based on the debug level.
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

// StrIndex is the helper function to find the position of a string in a []string
func StrIndex(vs []string, t string) int {
	for i, v := range vs {
		if v == t {
			return i
		}
	}
	return -1
}

// SimpleGaussian returns a random value of Gaussian distribution.
// mean=0.5*the_range, stddev=0.2*the_range
func SimpleGaussian(g *rng.GaussianGenerator, gap int) int {
	return int(math.Abs(g.Gaussian(0.5*float64(gap), 0.2*float64(gap)))) % gap
}
