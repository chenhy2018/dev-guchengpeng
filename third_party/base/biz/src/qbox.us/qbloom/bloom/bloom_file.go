package bloom

import (
	"github.com/qiniu/osl/vm"
	"sync"
	"syscall"
	"unsafe"
)

const (
	log2PageBytes = uint(18) // 256K
	pageBytes     = uint(1 << log2PageBytes)
	pageBytesSub1 = pageBytes - 1

	log2PageBits = log2PageBytes + 3
	pageBits     = uint(1 << log2PageBits)
	pageBitsSub1 = pageBits - 1
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
	pages []vm.Range
	m     uint
	k     uint
	fd    int
	mutex sync.Mutex
}

// Create a new Bloom filter with _m_ bits and _k_ hashing functions
func Open(file string, m uint, k uint) (f *Filter, err error) {
	fd, err := syscall.Open(file, syscall.O_RDWR|syscall.O_CREAT, 0666)
	if err != nil {
		return
	}
	npage := (m + pageBitsSub1) >> log2PageBits
	fsize := int64(npage) << log2PageBytes
	err = syscall.Ftruncate(fd, fsize)
	if err != nil {
		syscall.Close(fd)
		return
	}
	pages := make([]vm.Range, npage)
	return &Filter{pages: pages, m: m, k: k, fd: fd}, nil
}

func (f *Filter) Close() error {
	for i, page := range f.pages {
		if page != nil {
			page.Close()
			f.pages[i] = nil
		}
	}
	return syscall.Close(f.fd)
}

// Return the capacity, _m_, of a Bloom filter
func (f *Filter) Cap() uint {
	return f.m
}

// Return the number of hash functions used
func (f *Filter) K() uint {
	return f.k
}

func (f *Filter) requirePage(ipage uint) (page vm.Range, err error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	page = f.pages[ipage]
	if page == nil {
		off := int64(ipage) << log2PageBytes
		page, err = vm.Map(f.fd, off, int(pageBytes), syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
		if err != nil {
			return
		}
		f.pages[ipage] = page
	}
	return
}

// Add data to the Bloom Filter. Returns the filter (allows chaining)
func (f *Filter) Add(data []byte) (err error) {
	var l location
	l.init(data)
	for i := uint(0); i < f.k; i++ {
		loc := l.next() % f.m
		ipage := loc >> log2PageBits
		ioff := loc & pageBitsSub1
		page := f.pages[ipage]
		if page == nil {
			page, err = f.requirePage(ipage)
			if err != nil {
				return
			}
		}
		page[ioff>>3] |= (1 << (ioff & 7))
	}
	return nil
}

// Tests for the presence of data in the Bloom filter
func (f *Filter) Test(data []byte) (present bool, err error) {
	var l location
	l.init(data)
	for i := uint(0); i < f.k; i++ {
		loc := l.next() % f.m
		ipage := loc >> log2PageBits
		ioff := loc & pageBitsSub1
		page := f.pages[ipage]
		if page == nil {
			page, err = f.requirePage(ipage)
			if err != nil {
				return
			}
		}
		if page[ioff>>3]&(1<<(ioff&7)) == 0 {
			return false, nil
		}
	}
	return true, nil
}

// Equivalent to calling Test(data) then Add(data). Returns the result of Test.
func (f *Filter) TestAndAdd(data []byte) (present bool, err error) {
	present = true
	var l location
	l.init(data)
	for i := uint(0); i < f.k; i++ {
		loc := l.next() % f.m
		ipage := loc >> log2PageBits
		ioff := loc & pageBitsSub1
		page := f.pages[ipage]
		if page == nil {
			page, err = f.requirePage(ipage)
			if err != nil {
				return false, err
			}
		}
		if page[ioff>>3]&(1<<(ioff&7)) == 0 {
			page[ioff>>3] |= (1 << (ioff & 7))
			present = false
		}
	}
	return present, nil
}

// ------------------------------------------------------------------------
