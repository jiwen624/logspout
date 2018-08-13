package utils

import (
	"math/rand"
	"strconv"
	"strings"

	"github.com/Pallinder/go-randomdata"
	"github.com/beevik/etree"
	"github.com/pkg/errors"
)

var cset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// Constants for generating random strings
const (
	lIdxBits = 6               // 6 bits to represent a letter index
	idxMask  = 1<<lIdxBits - 1 // All 1-bits, as many as lIdxBits
	idxMax   = 63 / lIdxBits   // # of letter indices fitting in 63 bits
)

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
func XMLStr(maxDepth int, maxElements int, seed []string) (string, error) {
	if maxDepth == 0 || maxElements == 0 {
		return "", errors.New("invalid maxDepth or maxElements")
	}

	doc := etree.NewDocument()
	doc.CreateProcInst("xml", `version="1.0" encoding="UTF-8"`)
	doc.CreateProcInst("xml-stylesheet", `type="text/xsl" href="style.xsl"`)

	elmentsCnt := make(map[int]int)
	for i := 0; i <= maxDepth; i++ {
		elmentsCnt[i] = 0 // len(elementsCnt) == maxDepth + 1
	}

	xmlStr(&doc.Element, maxDepth, maxElements, 0, elmentsCnt, seed)
	doc.Indent(2)

	return doc.WriteToString()
}

// xmlStr is the internal helper function for XMLStr
// In future we may use a different seed for each recursion.
func xmlStr(doc *etree.Element, maxDepth int, maxElements int, currDepth int, elementsCnt map[int]int, seed []string) int {
	if doc == nil {
		return 0
	}

	doc.CreateComment(strconv.Itoa(currDepth) + "," + strconv.Itoa(elementsCnt[currDepth]))
	if currDepth != 0 {
		if needAttr() {
			doc.CreateAttr(randomAttrK(), randomAttrV())
		}
		if needData() {
			doc.CreateCharData(randomData())
		}
	}
	if currDepth >= maxDepth {
		return 1
	}

	childDepth := currDepth + 1
	// The number of children would be in the range of 1 - 10
	numChildren := rand.Intn(maxElements)%10 + 1
	// The maximum elements for each level would be less then maxElements
	for i := 0; i < numChildren && elementsCnt[childDepth] < maxElements; i++ {
		elementsCnt[childDepth] += xmlStr(doc.CreateElement(randomTag(seed)), maxDepth, maxElements, childDepth, elementsCnt, seed)
		if currDepth == 0 {
			break
		}
	}
	return 1
}

// randomTag is a helper function to generate random tag string
func randomTag(seed []string) string {
	if l := len(seed); l == 0 {
		return randomdata.Noun()
	} else {
		return seed[rand.Intn(l)]
	}
}

func randomAttrK() string {
	return strings.Replace(randomdata.State(randomdata.Large), " ", "", -1)
}

func randomAttrV() string {
	return randomdata.Country(randomdata.ThreeCharCountry)
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
