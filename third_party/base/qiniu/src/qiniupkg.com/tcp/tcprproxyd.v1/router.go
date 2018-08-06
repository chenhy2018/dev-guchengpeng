package tcprproxyd

import (
	"net"
	"sync/atomic"
)

// -----------------------------------------------------------------------------

type singleBackend struct {
	addr string
}

func SingleBackend(addr string) Router {

	return &singleBackend{addr}
}

func (p *singleBackend) Pick(raddr net.Addr) (backend string, err error) {

	return p.addr, nil
}

func (p *singleBackend) Unpick(backend string, raddr net.Addr) {
}

// -----------------------------------------------------------------------------

type roundRobin struct {
	addrs []string
	i     uint32
}

func RoundRobin(addrs ...string) Router {

	return &roundRobin{addrs, 0}
}

func (p *roundRobin) Pick(raddr net.Addr) (backend string, err error) {

	idx := atomic.AddUint32(&p.i, 1) % uint32(len(p.addrs))
	return p.addrs[idx], nil
}

func (p *roundRobin) Unpick(backend string, raddr net.Addr) {
}

// -----------------------------------------------------------------------------

