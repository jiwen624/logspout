// Utils

package main

import (
	"fmt"
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

var globalLevel DebugLevel = INFO

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
		fmt.Fprintf(os.Stderr, "Unknown err type: %T", err)
		return
	}

	var p string
	pc, f, l, ok := runtime.Caller(1)
	if globalLevel > DEBUG {
		p = ""
	} else if ok {
		path := strings.Split(runtime.FuncForPC(pc).Name(), ".")
		name := path[len(path)-1]
		p = fmt.Sprintf("%s %s#%d %s(): ", dbgLevels[level], filepath.Base(f), l, name)
	} else {
		p = fmt.Sprintf("%s %s#%s %s(): ", level, "na", "na", "na")
	}
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "%s%s", p, msg)
	} else {
		fmt.Fprintf(os.Stderr, "%s%s %v", p, msg, args)
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
