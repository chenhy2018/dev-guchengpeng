package bloomutil

import (
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"qbox.us/qmset/bloom"
	"testing"
)

func keyOf(i uint32) []byte {

	n1 := make([]byte, 4)
	binary.BigEndian.PutUint32(n1, i)

	h := md5.New()
	h.Write(n1)
	return h.Sum(nil)
}

// Estimate, for a Filter with a limit of m bytes
// and k hash functions, what the false positive rate will be
// whilst storing n entries; runs 10k tests
func EstimateFalsePositiveRate(f *bloom.Filter, n, ncase uint) (fp_rate float64) {
	f.ClearAll()
	for i := uint32(0); i < uint32(n); i++ {
		f.Add(keyOf(i))
	}
	fp := 0
	for i := uint32(0); i < uint32(ncase); i++ {
		if f.Test(keyOf(i + uint32(n) + 1)) {
			fp++
		}
	}
	fp_rate = float64(fp) / float64(ncase)
	f.ClearAll()
	return
}

// ------------------------------------------------------------------------

func Test(t *testing.T) {

	const n = 150000

	fp := 1e-6
	f := NewWithEstimates(n, fp)
	fmt.Println("m, k, bits:", f.Cap(), f.K(), f.Cap()/n)

	fp_rate := EstimateFalsePositiveRate(f, n, n)
	fmt.Println(fp_rate, fp_rate/fp)

	if fp_rate/fp > 1.1 {
		t.Fatal("EstimateFalsePositiveRate failed")
	}
}

// ------------------------------------------------------------------------
