package bloomutil

import (
	"math"
	"qbox.us/qmset/bloom"
	"github.com/qiniu/log.v1"
)

// ------------------------------------------------------------------------

// estimate parameters. Based on https://bitbucket.org/ww/bloom/src/829aa19d01d9/bloom.go
// used with permission.
func estimateParameters(n uint, p float64) (m uint, k uint) {
	m = uint(-1 * float64(n) * math.Log(p) / math.Pow(math.Log(2), 2))
	k = uint(math.Ceil(math.Log(2) * float64(m) / float64(n)))
	return
}

// Create a new Bloom filter for about n items with fp
// false positive rate
func NewWithEstimates(n uint, fp float64) *bloom.Filter {
	m, k := estimateParameters(n, fp)
	log.Info("Bloom m, k =", m, k)
	return bloom.New(m, k)
}

// ------------------------------------------------------------------------
