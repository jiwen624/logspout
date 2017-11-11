// Utils

package main

import (
	"fmt"
	"log"
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

// LevelLog print logs based on the debug level.
func LevelLog(level DebugLevel, err interface{}, args ...interface{}) {
	var msg string
	switch err.(type) {
	case error:
		msg = err.(error).Error()
	case string:
		msg = err.(string)
	default:
		log.Printf("Unknown err type: %T", err)
		return
	}

	var p string
	pc, f, l, ok := runtime.Caller(1)
	if ok {
		path := strings.Split(runtime.FuncForPC(pc).Name(), ".")
		name := path[len(path)-1]
		p = fmt.Sprintf("%s %s#%d %s(): ", dbgLevels[level], filepath.Base(f), l, name)
	} else {
		p = fmt.Sprintf("%s %s#%s %s(): ", level, "na", "na", "na")
	}
	if len(args) == 0 {
		log.Printf("%s%s", p, msg)
	} else {
		log.Printf("%s%s %v", p, msg, args)
	}
}
