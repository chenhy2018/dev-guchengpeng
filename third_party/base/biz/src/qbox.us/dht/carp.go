package dht

import (
	"bytes"
	"crypto/sha1"
	"hash"
	"sort"
)

type Carp struct {
	nodes NodeInfos
}

// deprecated: 如果是新集群请使用 qbox.us/memcache/dht.
func NewCarp(nodes NodeInfos) Interface {
	crap := &Carp{}
	crap.Setup(nodes)
	return crap
}

func (p *Carp) Setup(nodes NodeInfos) {
	p.nodes = nodes
}

func (p *Carp) Nodes() NodeInfos {
	return p.nodes
}

func hashOf(h hash.Hash, key1, key2 []byte) []byte {
	h.Reset()
	h.Write(key1)
	h.Write(key2)
	return h.Sum(nil)
}

// len(routes) <= ttl
func (p *Carp) Route(key []byte, ttl int) (routers RouterInfos) {
	n := ttl
	if n > len(p.nodes) {
		n = len(p.nodes)
	}
	if n == 1 { // fast path
		r := p.RouteOne(key)
		return []RouterInfo{*r}
	}
	h := sha1.New()
	nodes := make([]carpNode, len(p.nodes))
	for i, node := range p.nodes {
		nodes[i] = carpNode{node.Host, hashOf(h, key, node.Key)}
	}
	sort.Sort(carpSlice(nodes))
	rs := make([]RouterInfo, n)
	for i := 0; i < n; i++ {
		rs[i].Host = nodes[i].host
		rs[i].Metrics = i + 1
	}
	return rs
}

func (p Carp) RouteOne(key []byte) (router *RouterInfo) {
	n := len(p.nodes)
	if n == 0 {
		return nil
	}
	h := sha1.New()
	isel, dis := 0, hashOf(h, key, p.nodes[0].Key)
	for i := 1; i < n; i++ {
		dis2 := hashOf(h, key, p.nodes[i].Key)
		if bytes.Compare(dis2, dis) < 0 {
			isel, dis = i, dis2
		}
	}
	return &RouterInfo{Host: p.nodes[isel].Host, Metrics: 1}
}

type carpNode struct {
	host string
	hash []byte
}
type carpSlice []carpNode

func (p carpSlice) Len() int           { return len(p) }
func (p carpSlice) Less(i, j int) bool { return bytes.Compare(p[i].hash, p[j].hash) < 0 }
func (p carpSlice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
