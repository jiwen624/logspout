package replacer

import (
	"math"
	"time"

	"github.com/leesper/go_rng"
)

// RandomGenerator defines a random number generator
type RandomGenerator interface {
	// Next returns the next random number with the upper limit as `max`.
	// The lower limit is always 0, so the range is [0, max)
	Next(max int) int
}

type TruncatedGaussian struct {
	// the Gaussian random generator
	g *rng.GaussianGenerator

	// the coefficient of mean
	meanC float64

	// the coefficient of standard deviation
	stddevC float64
}

func NewTruncatedGaussian(meanC, stddevC float64) *TruncatedGaussian {
	return &TruncatedGaussian{
		g:       rng.NewGaussianGenerator(time.Now().UnixNano()),
		meanC:   meanC,
		stddevC: stddevC,
	}
}

func (tg *TruncatedGaussian) Next(max int) int {
	if tg.g == nil || max <= 0 {
		return 0
	}
	mean := tg.meanC * float64(max)
	stddev := tg.stddevC * float64(max)
	return int(math.Abs(tg.g.Gaussian(mean, stddev))) % max
}
