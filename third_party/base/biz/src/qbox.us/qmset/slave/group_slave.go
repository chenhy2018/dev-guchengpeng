package slave

import (
	"errors"
	"sync"
	"syscall"
)

var ErrSetFull = errors.New("set full")

// ------------------------------------------------------------------------

type smallSet struct {
	values []string
	mutex  sync.Mutex
}

func exists(v string, pvalues []string) bool {

	for _, v2 := range pvalues {
		if v2 == v {
			return true
		}
	}
	return false
}

func (p *smallSet) Add(value string, max int) (err error) {

	p.mutex.Lock()
	defer p.mutex.Unlock()

	if exists(value, p.values) {
		return syscall.EEXIST
	}

	if len(p.values) >= max {
		return ErrSetFull
	}

	p.values = append(p.values, value)
	return
}

// ------------------------------------------------------------------------

type msetGroup struct {
	map1  map[string]*smallSet
	map2  map[string]*smallSet
	mutex sync.RWMutex
	max   int
}

func newMsetGroup(max int) (p *msetGroup) {

	map1 := make(map[string]*smallSet)
	map2 := make(map[string]*smallSet)
	return &msetGroup{
		map1: map1, map2: map2, max: max,
	}
}

func (p *msetGroup) Add(key string, value string) (err error) {

	set1, set2 := p.requireSet(key)
	err = set1.Add(value, p.max)
	set2.Add(value, p.max)
	return
}

func (p *msetGroup) requireSet(key string) (set1, set2 *smallSet) {

	p.mutex.RLock()
	set1 = p.map1[key]
	set2 = p.map2[key]
	p.mutex.RUnlock()

	if set1 != nil && set2 != nil {
		return
	}

	p.mutex.Lock()
	defer p.mutex.Unlock()

	if set1 == nil {
		set1 = new(smallSet)
		p.map1[key] = set1
	}
	if set2 == nil {
		set2 = new(smallSet)
		p.map2[key] = set2
	}
	return
}

func (p *msetGroup) Flip() {

	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.map1, p.map2 = p.map2, make(map[string]*smallSet)
	return
}

func (p *msetGroup) Clear() {

	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.map1, p.map2 = make(map[string]*smallSet), make(map[string]*smallSet)
	return
}

// ------------------------------------------------------------------------
