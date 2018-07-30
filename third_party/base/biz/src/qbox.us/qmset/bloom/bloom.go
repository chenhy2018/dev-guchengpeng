/*
A Bloom filter is a representation of a set of _n_ items, where the main
requirement is to make membership queries; _i.e._, whether an item is a
member of a set.

A Bloom filter has two parameters: _m_, a maximum size (typically a reasonably large
multiple of the cardinality of the set to represent) and _k_, the number of hashing
functions on elements of the set. (The actual hashing functions are important, too,
but this is not a parameter for this implementation). A Bloom filter is backed by
a BitSet; a key is represented in the filter by setting the bits at each value of the
hashing functions (modulo _m_). Set membership is done by _testing_ whether the
bits at each value of the hashing functions (again, modulo _m_) are set. If so,
the item is in the set. If the item is actually in the set, a Bloom filter will
never fail (the true positive rate is 1.0); but it is susceptible to false
positives. The art is to choose _k_ and _m_ correctly.

This implementation accepts keys for setting as testing as []byte. Thus, to
add a string item, "Love":

	const n = 1000
	filter := bloom.New(20*n, 5) // load of 20, 5 keys
	filter.Add([]byte("Love"))

Similarly, to test if "Love" is in bloom:

	if filter.Test([]byte("Love"))
*/
package bloom

import (
	"unsafe"
)

const (
	log2WordSize = uint(5)
	wordSize     = uint(1 << log2WordSize)
	wordSizeSub1 = wordSize - 1
)

// ------------------------------------------------------------------------

type location struct {
	v        []uint32
	data     []byte
	i, round uint
}

var zero [8]byte

func (p *location) init(data []byte) {

	mod := len(data) & 7
	if mod != 0 {
		data = append(data, zero[:8-mod]...)
	}
	p.data = data
	p.round = 1
	p.v = ((*[0x10000000]uint32)(unsafe.Pointer(&data[0])))[:len(data)>>2]
}

func (p *location) next() uint {

	if p.i >= uint(len(p.v)) {
		p.i = 0
		p.round++
	}
	a, b := uint(p.v[p.i]), uint(p.v[p.i+1])
	p.i += 2
	return a + b*p.round
}

// ------------------------------------------------------------------------

type Filter struct {
	m      uint
	k      uint
	bitset []uint32
}

// Create a new Bloom filter with _m_ bits and _k_ hashing functions
func New(m uint, k uint) *Filter {
	bitset := make([]uint32, (m+wordSizeSub1)>>log2WordSize)
	return &Filter{m, k, bitset}
}

// Return the capacity, _m_, of a Bloom filter
func (f *Filter) Cap() uint {
	return f.m
}

// Return the number of hash functions used
func (f *Filter) K() uint {
	return f.k
}

// Add data to the Bloom Filter. Returns the filter (allows chaining)
func (f *Filter) Add(data []byte) *Filter {

	var l location
	l.init(data)
	for i := uint(0); i < f.k; i++ {
		loc := l.next() % f.m
		f.bitset[loc>>log2WordSize] |= (1 << (loc & wordSizeSub1))
	}
	return f
}

// Tests for the presence of data in the Bloom filter
func (f *Filter) Test(data []byte) bool {
	var l location
	l.init(data)
	for i := uint(0); i < f.k; i++ {
		loc := l.next() % f.m
		if f.bitset[loc>>log2WordSize]&(1<<(loc&wordSizeSub1)) == 0 {
			return false
		}
	}
	return true
}

// Equivalent to calling Test(data) then Add(data). Returns the result of Test.
func (f *Filter) TestAndAdd(data []byte) bool {
	present := true
	var l location
	l.init(data)
	for i := uint(0); i < f.k; i++ {
		loc := l.next() % f.m
		if f.bitset[loc>>log2WordSize]&(1<<(loc&wordSizeSub1)) == 0 {
			f.bitset[loc>>log2WordSize] |= (1 << (loc & wordSizeSub1))
			present = false
		}
	}
	return present
}

// Clear all the data in a Bloom filter, removing all keys
func (f *Filter) ClearAll() *Filter {
	for i := range f.bitset {
		f.bitset[i] = 0
	}
	return f
}

// ------------------------------------------------------------------------

type FilterState struct {
	Cap  uint
	K    uint
	Data []uint32
}

func (f *Filter) GetState() FilterState {
	return FilterState{f.m, f.k, f.bitset}
}

func NewWith(state FilterState) *Filter {
	return &Filter{state.Cap, state.K, state.Data}
}

// ------------------------------------------------------------------------
