package configuration

import (
	"sync"
)

// ------------------------------------------------------------------------------------------

type memGroup struct {
	data map[string]string
	sync.Mutex
}

func (p *memGroup) Group() (items []GroupItem, err error) {

	i, n := 0, len(p.data)
	items = make([]GroupItem, n)
	for k, v := range p.data {
		items[i] = GroupItem{k, v}
		i++
	}
	return
}

// ------------------------------------------------------------------------------------------

type memInstance struct {
	grps  map[string]*memGroup
	mutex sync.RWMutex
}

func newMemInstance() *memInstance {

	return &memInstance{
		grps: make(map[string]*memGroup),
	}
}

func (p *memInstance) dirtyRequire(grp string) *memGroup {

	g, ok := p.grps[grp]
	if !ok {
		g = &memGroup{
			data: make(map[string]string),
		}
		p.grps[grp] = g
	}
	return g
}

func (p *memInstance) Lock(grp string) *memGroup {

	p.mutex.RLock()
	g, ok := p.grps[grp]
	p.mutex.RUnlock()

	if !ok {
		p.mutex.Lock()
		g, ok = p.grps[grp]
		if !ok {
			g = &memGroup{
				data: make(map[string]string),
			}
			p.grps[grp] = g
		}
		p.mutex.Unlock()
	}

	g.Lock()
	return g
}

// ------------------------------------------------------------------------------------------
