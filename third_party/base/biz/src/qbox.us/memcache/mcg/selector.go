package mcg

import (
	"errors"
	"net"
	"qbox.us/memcache/dht"
	"strings"
)

type serverSelector struct {
	router dht.Carp
}

func newServerSelector(conns []Node) (serverSelector, error) {

	nodes := make([]dht.NodeInfo, 0, len(conns)*3)
	for _, conn := range conns {
		addr, err := resolveAddr(conn.Host)
		if err != nil {
			return serverSelector{}, err
		}
		for _, key := range conn.Keys {
			nodes = append(nodes, dht.NodeInfo{[]byte(key), addr})
		}
	}
	return serverSelector{
		router: dht.NewCarp(nodes),
	}, nil
}

func resolveAddr(server string) (net.Addr, error) {

	if strings.Contains(server, "/") {
		return net.ResolveUnixAddr("unix", server)
	}
	return net.ResolveTCPAddr("tcp", server)
}

var errNoServers = errors.New("serverSelector: no servers configured or available")

func (p serverSelector) PickServer(key string) (net.Addr, error) {

	node := p.router.Route([]byte(key))
	return node.Node.(net.Addr), nil
}
