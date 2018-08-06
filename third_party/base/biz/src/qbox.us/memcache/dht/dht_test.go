package dht

import (
	"fmt"
	"testing"
)

func newDht(nodes []NodeInfo) Carp {
	return NewCarp(nodes)
}

func TestDht(t *testing.T) {
	n1, n2 := testAddNode()
	fmt.Println("test add node:", n2, "/", n1)
	n1, n2 = testDelNode()
	fmt.Println("test del node:", n2, "/", n1)
}

func testAddNode() (n1, n2 int) {
	nodes := make([]NodeInfo, 4)
	nodes[0].Node = "host1"
	nodes[0].Key = []byte("host1")
	nodes[1].Node = "host2"
	nodes[1].Key = []byte("host2")
	nodes[2].Node = "host3"
	nodes[2].Key = []byte("host3")
	nodes[3].Node = "host4"
	nodes[3].Key = []byte("host4")
	dht := newDht(nodes)
	r1 := testRoute(dht)

	nodes = make([]NodeInfo, 5)
	nodes[0].Node = "host1"
	nodes[0].Key = []byte("host1")
	nodes[1].Node = "host2"
	nodes[1].Key = []byte("host2")
	nodes[2].Node = "host3"
	nodes[2].Key = []byte("host3")
	nodes[3].Node = "host4"
	nodes[3].Key = []byte("host4")
	nodes[4].Node = "host5"
	nodes[4].Key = []byte("host5")
	dht = newDht(nodes)
	r2 := testRoute(dht)

	n1, n2 = checkResult(r1, r2)
	return
}

func testDelNode() (n1, n2 int) {
	nodes := make([]NodeInfo, 4)
	nodes[0].Node = "host1"
	nodes[0].Key = []byte("host1")
	nodes[1].Node = "host2"
	nodes[1].Key = []byte("host2")
	nodes[2].Node = "host3"
	nodes[2].Key = []byte("host3")
	nodes[3].Node = "host4"
	nodes[3].Key = []byte("host4")
	dht := newDht(nodes)
	r1 := testRoute(dht)

	nodes = make([]NodeInfo, 3)
	nodes[0].Node = "host1"
	nodes[0].Key = []byte("host1")
	nodes[1].Node = "host2"
	nodes[1].Key = []byte("host2")
	nodes[2].Node = "host3"
	nodes[2].Key = []byte("host3")
	dht = newDht(nodes)
	r2 := testRoute(dht)

	n1, n2 = checkResult(r1, r2)
	return
}

func testRoute(dht Carp) [][]RouterInfo {
	result := make([][]RouterInfo, 100000)
	for i := 0; i < 100000; i++ {
		key := []byte(fmt.Sprintf("key1234567890_%v", i))
		result[i] = []RouterInfo{dht.Route(key)}
		//result[i] = dht.RouteEx(key, 1)
	}
	return result
}

func checkResult(l1, l2 [][]RouterInfo) (n1, n2 int) {
	for i := 0; i < len(l1); i++ {
		n1++
		if l1[i][0].Node == l2[i][0].Node {
			n2++
		}
	}
	return n1, n2
}
