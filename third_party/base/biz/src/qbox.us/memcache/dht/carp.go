package dht

import (
	"bytes"
	"crypto/md5"
	"hash"
	"sort"
)

// ----------------------------------------------------------------------------

type NodeInfo struct {
	Key  []byte
	Node interface{}
}

type RouterInfo struct {
	Node interface{}
	// TODO: Metrics的口径在Route和RouteEx里不一致
	Metrics int
}

// ----------------------------------------------------------------------------

type Carp struct {
	Nodes []NodeInfo
}

// 注意：这个类从 "qbox.us/dht" 演化，但是算法不同（用md5换掉了原来的sha1）。
// 建议：如果之前用 "qbox.us/dht"，应该沿用；如果是新集群，应该优先用这里的。
func NewCarp(nodes []NodeInfo) Carp {
	return Carp{nodes}
}

func (p Carp) Route(key []byte) (router RouterInfo) {
	h := md5.New()
	n := len(p.Nodes)
	isel, dis := 0, hashOf(h, key, p.Nodes[0].Key)
	for i := 1; i < n; i++ {
		dis2 := hashOf(h, key, p.Nodes[i].Key)
		if bytes.Compare(dis2, dis) < 0 {
			isel, dis = i, dis2
		}
	}
	return RouterInfo{p.Nodes[isel].Node, isel + 1}
}

func hashOf(h hash.Hash, key1, key2 []byte) []byte {
	h.Reset()
	h.Write(key1)
	h.Write(key2)
	return h.Sum(nil)
}

// len(routes) <= ttl
func (p Carp) RouteEx(key []byte, ttl int) (routers []RouterInfo) {
	n := ttl
	if n > len(p.Nodes) {
		n = len(p.Nodes)
	}
	h := md5.New()
	nodes := make([]carpNode, len(p.Nodes))
	for i, node := range p.Nodes {
		nodes[i] = carpNode{node.Node, hashOf(h, key, node.Key)}
	}
	sort.Sort(carpSlice(nodes))
	rs := make([]RouterInfo, n)
	for i := 0; i < n; i++ {
		rs[i].Node = nodes[i].node
		rs[i].Metrics = i + 1
	}
	return rs
}

// ----------------------------------------------------------------------------

type carpNode struct {
	node interface{}
	hash []byte
}

type carpSlice []carpNode

func (p carpSlice) Len() int           { return len(p) }
func (p carpSlice) Less(i, j int) bool { return bytes.Compare(p[i].hash, p[j].hash) < 0 }
func (p carpSlice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

// ----------------------------------------------------------------------------
