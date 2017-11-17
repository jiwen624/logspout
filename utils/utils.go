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

func init() {

}

// l is the global log print object.
var l = log.New(os.Stderr, "", log.LstdFlags)

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
		l.Printf("Unknown err type: %T\n", err)
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
		l.Printf("%s%s", p, msg)
	} else {
		l.Printf("%s%s %v", p, msg, args)
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

// ReConvert does the pre-process of the regular expression literal.
// It does the following things:
// 1. Remove P before captured group names
// 2. Add parenthesises to the other parts of the log event string.
func ReConvert(ptn string) string {
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
