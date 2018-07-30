package cc

import (
	"crypto/rand"
	"io/ioutil"
	"os"
	"github.com/qiniu/xlog.v1"
	"strconv"
	"testing"
)

func TestSimplePool(t *testing.T) {
	xl := xlog.NewDummy()

	root := os.TempDir()
	pool := NewSimplePool(root)

	key := "test"

	_, err := pool.Put(xl, key, rand.Reader, 64)
	if err != nil {
		t.Fatal("simple pool put failed", err)
	}

	r, _, err := pool.Get(xl, key, 0)
	if err != nil {
		t.Fatal("simple pool get failed", err)
	}
	defer r.Close()

	buf, err := ioutil.ReadAll(r)
	if err != nil || len(buf) != 64 {
		t.Fatal("read from simple pool failed", len(buf), err)
	}

	err = pool.Delete(xl, key)
	if err != nil {
		t.Fatal("simple pool delete failed", err)
	}

	r, _, err = pool.Get(xl, key, 0)
	if err == nil || !os.IsNotExist(err) {
		t.Fatal("simple pool should not exist after delete", err)
	}
}

func BenchmarkSimplePool(b *testing.B) {
	xl := xlog.NewDummy()
	b.StopTimer()
	pool := NewSimplePool(os.TempDir())

	var i int
	for i = 1e4; i < 1e4+2e3; i++ {
		pool.Put(xl, strconv.Itoa(i), rand.Reader, int64(1))
	}

	b.StartTimer()
	keys, _ := pool.Keys()
	b.Log(len(keys))
}
