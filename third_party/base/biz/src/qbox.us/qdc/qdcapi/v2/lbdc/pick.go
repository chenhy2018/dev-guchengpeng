package lbdc

import (
	"strings"

	"qbox.us/dht"
)

type picker struct {
	hosts []string
	index int
}

func (p *picker) One() string {

	idx := p.index
	num := len(p.hosts)

	p.index++

lzOuter:
	for i := idx; i < num; i++ {
		host := p.hosts[i]
		prefix := host[:strings.LastIndex(host, ":")+1] // trim port
		for j := 0; j < idx; j++ {
			if strings.HasPrefix(p.hosts[j], prefix) {
				continue lzOuter
			}
		}
		if i != idx {
			// Swap for next retry.
			p.hosts[i] = p.hosts[idx]
			p.hosts[idx] = host
		}
		return host
	}
	return p.hosts[idx%num]
}

// -----------------------------------------------------------------------------

type pickerProxy struct {
	routers dht.RouterInfos
	picker  *picker
	picked  bool
}

func newPickerProxy(routers dht.RouterInfos) (*pickerProxy, error) {

	if len(routers) == 0 {
		return nil, EServerNotAvailable
	}
	return &pickerProxy{routers: routers}, nil
}

func (p *pickerProxy) One() string {

	if !p.picked {
		p.picked = true
		return p.routers[0].Host
	}
	if p.picker == nil {
		hosts := make([]string, len(p.routers))
		for i, rt := range p.routers {
			hosts[i] = rt.Host
		}
		p.picker = &picker{hosts: hosts}
		p.picker.One()
	}
	return p.picker.One()
}
