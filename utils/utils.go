package utils

import (
	"fmt"
	"log"
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

// DbgLevels is a map of level macros and strings.
var DbgLevels = map[DebugLevel]string{
	DEBUG:   "DEBUG",
	INFO:    "INFO ",
	WARNING: "WARN ",
	ERROR:   "ERROR",
}

// LevelsDbg is a reversed map of level macros and strings.
var LevelsDbg = map[string]DebugLevel{
	"debug":   DEBUG,
	"info":    INFO,
	"warning": WARNING,
	"error":   ERROR,
}

// globalLevel is the global variable for global debug level.
var globalLevel = INFO

func SetGlobalDebugLevel(level DebugLevel) {
	globalLevel = level
}

func GlobalDebugLevel() DebugLevel {
	return globalLevel
}

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

	var p string = fmt.Sprintf("%s ", DbgLevels[level])
	if globalLevel == DEBUG {
		if pc, f, l, ok := runtime.Caller(1); ok {
			path := strings.Split(runtime.FuncForPC(pc).Name(), ".")
			name := path[len(path)-1]
			p = fmt.Sprintf("%s %s#%d %s(): ", DbgLevels[level], filepath.Base(f), l, name)
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
