package mbloom

import (
	"qbox.us/qmset/bloom"
	"qbox.us/qmset/bloom/bloomutil"
	"sync"
)

// ------------------------------------------------------------------------

type Filter struct {
	bf1    *bloom.Filter
	bf2    *bloom.Filter
	mutex1 sync.RWMutex
	mutex2 sync.Mutex
}

func New(max uint, fp float64) (p *Filter) {

	bf1 := bloomutil.NewWithEstimates(max, fp)
	bf2 := bloom.New(bf1.Cap(), bf1.K())
	return &Filter{
		bf1: bf1, bf2: bf2,
	}
}

func (p *Filter) Exists(vals [][]byte) []int {

	p.mutex1.RLock()
	defer p.mutex1.RUnlock()

	idxs := make([]int, 0, len(vals))
	bf1 := p.bf1
	for i, val := range vals {
		if bf1.Test(val) {
			idxs = append(idxs, i)
		}
	}
	return idxs
}

func (p *Filter) Add(vals [][]byte) {

	p.mutex1.Lock()
	for _, val := range vals {
		p.bf1.Add(val)
	}
	p.mutex1.Unlock()

	p.mutex2.Lock()
	for _, val := range vals {
		p.bf2.Add(val)
	}
	p.mutex2.Unlock()
}

func (p *Filter) Flip() {

	p.mutex2.Lock()

	p.mutex1.Lock()
	p.bf1, p.bf2 = p.bf2, p.bf1
	p.mutex1.Unlock()

	p.bf2.ClearAll()
	p.mutex2.Unlock()

	return
}

// ------------------------------------------------------------------------
