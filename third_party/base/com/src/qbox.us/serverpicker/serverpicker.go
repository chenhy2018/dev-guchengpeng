package serverpicker

import (
	"errors"
	"hash/crc32"
	"net"
	"sort"
	"strconv"
	"strings"
)

var (
	ErrNoServers = errors.New("serverpicker: no servers configured or available")
)

type hashNode struct {
	node net.Addr
	hash uint32
}

type nodeArray []hashNode

type instance struct {
	nodes nodeArray
}

// multiple: a physical server will appear under 'multiple' virtual nodes
func New(servers []string, multiple int) (r *instance, err error) {

	r = &instance{}
	err = r.assign(servers, multiple)
	return
}

// ----------------------------------------------------------------------------

// Assign add servers to the ring, the old nodes will be discarded.
func (r *instance) assign(servers []string, multiple int) error {

	r.nodes = make(nodeArray, 0, len(servers)*multiple)
	for _, server := range servers {
		if err := r.add(server, multiple); err != nil {
			return err
		}
	}
	sort.Sort(r.nodes)
	return nil
}

func (r *instance) add(server string, multiple int) error {

	node, err := resolveAddr(server)
	if err != nil {
		return err
	}

	for i := 0; i < multiple; i++ {
		hash := hashKey(server + ":" + strconv.Itoa(i))
		r.nodes = append(r.nodes, hashNode{node, hash})
	}
	return nil
}

func resolveAddr(server string) (net.Addr, error) {

	if strings.Contains(server, "/") {
		return net.ResolveUnixAddr("unix", server)
	}
	return net.ResolveTCPAddr("tcp", server)
}

// PickServer searchs preferable node for the specified key.
// If there is an error, it will be ErrNoServers.
func (r *instance) PickServer(key string) (net.Addr, error) {

	l := len(r.nodes)
	if l == 0 {
		return nil, ErrNoServers
	}

	h := hashKey(key)
	pos := sort.Search(l, func(i int) bool {
		return r.nodes[i].hash >= h
	})
	if pos == l {
		pos = 0
	}
	return r.nodes[pos].node, nil
}

func hashKey(key string) uint32 {

	return crc32.ChecksumIEEE([]byte(key))
}

// ----------------------------------------------------------------------------

func (a nodeArray) Len() int {
	return len(a)
}

func (a nodeArray) Less(i, j int) bool {

	return a[i].hash < a[j].hash
}

func (a nodeArray) Swap(i, j int) {

	a[i], a[j] = a[j], a[i]
}
