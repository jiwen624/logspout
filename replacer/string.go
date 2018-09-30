package replacer

import "math/rand"

var cset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// Constants for generating random strings
const (
	lIdxBits = 6               // 6 bits to represent a letter index
	idxMask  = 1<<lIdxBits - 1 // All 1-bits, as many as lIdxBits
	idxMax   = 63 / lIdxBits   // # of letter indices fitting in 63 bits
)

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
