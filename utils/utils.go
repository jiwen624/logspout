package utils

import (
	"fmt"
	"github.com/Pallinder/go-randomdata"
	"github.com/beevik/etree"
	"github.com/pkg/errors"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
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

var cset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// Constants for generating random strings
const (
	lIdxBits = 6               // 6 bits to represent a letter index
	idxMask  = 1<<lIdxBits - 1 // All 1-bits, as many as lIdxBits
	idxMax   = 63 / lIdxBits   // # of letter indices fitting in 63 bits
)

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
// TODO: replace it with logrus or something alike
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

// StrSlice2DCopy is a simple helper function to make a deep copy of a 2-dimensional string slice.
func StrSlice2DCopy(src [][]string) (cpy [][]string) {
	cpy = make([][]string, len(src))
	for i := range src {
		cpy[i] = make([]string, len(src[i]))
		copy(cpy[i], src[i])
	}
	return
}

// XMLStr returns an XML string with specified maximum depth and elements of each level
func XMLStr(maxDepth int, maxElements int) (string, error) {
	if maxDepth == 0 || maxElements == 0 {
		return "", errors.New("invalid maxDepth or maxElements")
	}

	doc := etree.NewDocument()
	doc.CreateProcInst("xml", `version="1.0" encoding="UTF-8"`)
	doc.CreateProcInst("xml-stylesheet", `type="text/xsl" href="style.xsl"`)

	elmentsCnt := make(map[int]int)
	for i := 1; i <= maxDepth; i++ {
		elmentsCnt[i] = 0 //len(elementsCnt) == maxDepth + 1
	}

	xmlStr(&doc.Element, maxDepth-1, maxElements, 0, elmentsCnt)
	doc.Indent(2)

	return doc.WriteToString()
}

// xmlStr is the internal helper function for XMLStr
func xmlStr(doc *etree.Element, maxDepth int, maxElements int, currDepth int, elementsCnt map[int]int) int {
	if doc == nil {
		return 0
	}

	// TODO: sometimes we don't need a comment
	//if needComment() {
	//	doc.CreateComment(randomComment())
	//}
	doc.CreateComment(strconv.Itoa(currDepth) + "," + strconv.Itoa(elementsCnt[currDepth]))

	if needAttr() {
		doc.CreateAttr(randomAttrK(), randomAttrV())
	}
	if needData() {
		doc.CreateCharData(randomData())
	}

	if currDepth >= maxDepth {
		return 1
	}

	childDepth := currDepth + 1
	// The number of children would be in the range of 1 - 10
	numChildren := rand.Intn(maxElements)%10 + 1
	// The maximum elements for each level would be less then maxElements
	for i := 0; i < numChildren && elementsCnt[childDepth] < maxElements; i++ {
		elementsCnt[childDepth] += xmlStr(doc.CreateElement(randomTag()), maxDepth, maxElements, childDepth, elementsCnt)
	}
	return 1
}

// randomTag is a helper function to generate random tag string
func randomTag() string {
	return randomdata.Noun()
}

func randomAttrK() string {
	s := randomdata.Country(randomdata.FullCountry)
	if rand.Intn(11)%10 == 0 {
		s = strings.Replace(s, " ", "", -1)
	}
	return s
}

func randomAttrV() string {
	return randomdata.LastName()
}

func randomComment() string {
	return randomdata.Adjective()
}

func needComment() bool {
	return randomdata.Boolean()
}

func needAttr() bool {
	return randomdata.Boolean()
}

func needData() bool {
	return rand.Intn(11)%10 != 0
}

func randomData() string {
	if randomdata.Boolean() {
		return randomdata.Adjective()
	} else {
		return strconv.Itoa(rand.Int())
	}
}

// RandomStr generates a random string within charset `chars` and shorter than length
func RandomStr(chars string, length int) string {
	if chars != "" {
		cset = chars
	}

	b := make([]byte, length)
	for i, cache, remain := length-1, rand.Int63(), idxMax; i >= 0; {
		if remain == 0 {
			cache, remain = rand.Int63(), idxMax
		}
		if idx := int(cache & idxMask); idx < len(cset) {
			b[i] = cset[idx]
			i--
		}
		cache >>= lIdxBits
		remain--
	}
	return string(b)
}
