package dht

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	mrand "math/rand"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func newDht(nodes NodeInfos) Interface {
	return NewCarp(nodes)
}

func TestDht(t *testing.T) {
	n1, n2 := testAddNode(t)
	fmt.Println("test add node:", n2, "/", n1)
	assert.Equal(t, 79984, n2)
	n1, n2 = testDelNode(t)
	fmt.Println("test del node:", n2, "/", n1)
	assert.Equal(t, 74863, n2)
}

func testAddNode(t *testing.T) (n1, n2 int) {
	nodes := make([]NodeInfo, 4)
	nodes[0].Host = "host1"
	nodes[0].Key = []byte("host1")
	nodes[1].Host = "host2"
	nodes[1].Key = []byte("host2")
	nodes[2].Host = "host3"
	nodes[2].Key = []byte("host3")
	nodes[3].Host = "host4"
	nodes[3].Key = []byte("host4")
	dht := newDht(nodes)
	r1 := testRoute(t, dht)

	nodes = make([]NodeInfo, 5)
	nodes[0].Host = "host1"
	nodes[0].Key = []byte("host1")
	nodes[1].Host = "host2"
	nodes[1].Key = []byte("host2")
	nodes[2].Host = "host3"
	nodes[2].Key = []byte("host3")
	nodes[3].Host = "host4"
	nodes[3].Key = []byte("host4")
	nodes[4].Host = "host5"
	nodes[4].Key = []byte("host5")
	dht = newDht(nodes)
	r2 := testRoute(t, dht)

	n1, n2 = checkResult(r1, r2)
	return
}

func testDelNode(t *testing.T) (n1, n2 int) {
	nodes := make([]NodeInfo, 4)
	nodes[0].Host = "host1"
	nodes[0].Key = []byte("host1")
	nodes[1].Host = "host2"
	nodes[1].Key = []byte("host2")
	nodes[2].Host = "host3"
	nodes[2].Key = []byte("host3")
	nodes[3].Host = "host4"
	nodes[3].Key = []byte("host4")
	dht := newDht(nodes)
	r1 := testRoute(t, dht)

	nodes = make([]NodeInfo, 3)
	nodes[0].Host = "host1"
	nodes[0].Key = []byte("host1")
	nodes[1].Host = "host2"
	nodes[1].Key = []byte("host2")
	nodes[2].Host = "host3"
	nodes[2].Key = []byte("host3")
	dht = newDht(nodes)
	r2 := testRoute(t, dht)

	n1, n2 = checkResult(r1, r2)
	return
}

func testRoute(t *testing.T, dht Interface) []RouterInfos {
	result := make([]RouterInfos, 100000)
	for i := 0; i < 100000; i++ {
		key := []byte(fmt.Sprintf("key1234567890_%v", i))
		result[i] = dht.Route(key, 10)
		result2 := dht.RouteOne(key)
		assert.Equal(t, result[i][0], *result2)
	}
	return result
}

func checkResult(l1, l2 []RouterInfos) (n1, n2 int) {
	for i := 0; i < len(l1); i++ {
		n1++
		if l1[i][0].Host == l2[i][0].Host {
			n2++
		}
	}
	return n1, n2
}

// 之前的实现算完sha1后还会做一次hex.EncodeToString,
// 现在把它去掉了，这里的测试验证加不加这个不影响顺序。
func TestCompare(t *testing.T) {
	N := 1000
	mrand.Seed(time.Now().Unix())
	for i := 0; i < N; i++ {
		L := mrand.Int() % 1000
		b1 := make([]byte, L)
		b2 := make([]byte, L)
		rand.Read(b1)
		rand.Read(b2)

		ret1 := bytes.Compare(b1, b2)
		ret2 := strings.Compare(hex.EncodeToString(b1), hex.EncodeToString(b2))
		assert.Equal(t, ret1, ret2)
	}
}

/***********************
*
* test result
*
* BenchmarkTestRoute100	   10000	    118031 ns/op
* BenchmarkTestRoute500	    5000	    606598 ns/op
* BenchmarkTestRoute1000	2000	   1212967 ns/op
*
*************************/
func BenchmarkTestRoute100(b *testing.B) {
	dhtSize := 100

	nodes := make([]NodeInfo, dhtSize)
	for num := 0; num < dhtSize; num++ {
		nodes[num].Host = "host" + strconv.Itoa(num)
		nodes[num].Key = []byte("host" + strconv.Itoa(num))
	}
	dht := newDht(nodes)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := []byte(fmt.Sprintf("abcedf%v", i))
		dht.Route(key, 1)
	}
}

func BenchmarkTestRoute500(b *testing.B) {
	dhtSize := 500

	nodes := make([]NodeInfo, dhtSize)
	for num := 0; num < dhtSize; num++ {
		nodes[num].Host = "host" + strconv.Itoa(num)
		nodes[num].Key = []byte("host" + strconv.Itoa(num))
	}
	dht := newDht(nodes)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := []byte(fmt.Sprintf("abcdef%v", i))
		dht.Route(key, 1)
	}
}

func BenchmarkTestRoute1000(b *testing.B) {
	dhtSize := 1000

	nodes := make([]NodeInfo, dhtSize)
	for num := 0; num < dhtSize; num++ {
		nodes[num].Host = "host" + strconv.Itoa(num)
		nodes[num].Key = []byte("host" + strconv.Itoa(num))
	}
	dht := newDht(nodes)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := []byte(fmt.Sprintf("abcdef%v", i))
		dht.Route(key, 1)
	}
}
