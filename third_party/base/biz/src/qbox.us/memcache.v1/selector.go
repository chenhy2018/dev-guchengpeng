package memcache

import (
	"errors"
	"net"
	"strings"

	"qbox.us/dht"
)

type Conn struct {
	Keys []string `json:"keys"`
	Host string   `json:"host"`
}

type mcSelector struct {
	router dht.Interface
	addrs  map[string]net.Addr
}

func newMcSelector(conns []Conn) (*mcSelector, error) {

	nodes := make([]dht.NodeInfo, 0, len(conns)*3)
	addrs := make(map[string]net.Addr)
	for _, conn := range conns {
		addr, err := resolveAddr(conn.Host)
		if err != nil {
			return nil, err
		}
		addrs[conn.Host] = addr
		for _, key := range conn.Keys {
			nodes = append(nodes, dht.NodeInfo{conn.Host, []byte(key)})
		}
	}
	return &mcSelector{
		router: dht.NewCarp(nodes),
		addrs:  addrs,
	}, nil
}

func resolveAddr(server string) (net.Addr, error) {

	if strings.Contains(server, "/") {
		return net.ResolveUnixAddr("unix", server)
	}
	return net.ResolveTCPAddr("tcp", server)
}

var errNoMcServers = errors.New("mcSelector: no servers configured or available")

func (p *mcSelector) PickServer(key string) (net.Addr, error) {

	node := p.router.RouteOne([]byte(key))
	if node == nil {
		return nil, errNoMcServers
	}
	return p.addrs[node.Host], nil
}
