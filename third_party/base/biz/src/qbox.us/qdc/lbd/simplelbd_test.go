package lbd

import (
	"bytes"
	"crypto/rand"
	"crypto/sha1"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"qbox.us/api/bd/bdc"
	"qbox.us/auditlog2"
	"github.com/qiniu/xlog.v1"
	"syscall"
	"testing"
	"time"
)

type mockStg struct {
	data map[string][]byte
}

func newMockStg() *mockStg {
	return &mockStg{make(map[string][]byte)}
}

func (r *mockStg) Get(xl *xlog.Logger, key []byte, w io.Writer, from, to int, bds [4]uint16) (err error) {
	bs, ok := r.data[string(key)]
	if !ok {
		return syscall.ENOENT
	}
	if to >= len(bs) {
		to = len(bs)
	}
	_, err = w.Write(bs[from:to])
	return
}

func (r *mockStg) Put(xl *xlog.Logger, key []byte, reader io.Reader, n int, bds [3]uint16) (err error) {
	bs, err := ioutil.ReadAll(io.LimitReader(reader, int64(n)))
	if err != nil {
		return
	}
	r.data[string(key)] = bs
	return
}

func newLbdForTest(t *testing.T) *Service {
	logCfg := &auditlog2.Config{
		LogFile: "/tmp/audit_lbd",
	}
	service, err := New(&Config{logCfg, os.TempDir() + "/cache", filepath.Join(os.TempDir(), "simple_cache"),
		1024 * 1024, 5 * 60 * 1e9, 1, 10, 22, newMockStg(), nil, 1})
	if err != nil {
		t.Fatal("New lbd service failed:", err)
	}
	return service
}

func TestSimplelbd(t *testing.T) {
	go func() {
		service := newLbdForTest(t)
		err := service.Run("127.0.0.1:23000")
		if err != nil {
			t.Fatal("service.Run failed:", err)
		}
	}()

	time.Sleep(1e9)

	testLocal(t)
	testNoCache(t)
	testDoCache(t)
	testMulti(t)
}

func testLocal(t *testing.T) {

	client := bdc.NewConn("http://127.0.0.1:23000", nil)

	length := 512
	buf := make([]byte, length)
	n, err := io.ReadFull(rand.Reader, buf)
	if err != nil || n != length {
		t.Fatal("rand read failed", n, err)
		return
	}
	sha1_ := sha1.New()
	sha1_.Write(buf)
	key := sha1_.Sum(nil)
	err = client.PutLocal(xlog.NewDummy(), key, bytes.NewBuffer(buf), length)
	if err != nil {
		t.Fatal("put local cache failed", err)
	}

	sha1_.Reset()
	reader, n, err := client.GetLocal(xlog.NewDummy(), key)
	if err != nil {
		t.Fatal("get local cache failed", err)
	}
	defer reader.Close()
	n1, err := io.CopyN(sha1_, reader, int64(n))
	if err != nil || n1 != int64(length) || bytes.Compare(sha1_.Sum(nil), key) != 0 {
		t.Fatal("get local cache failed", n1, err)
	}
}

func testNoCache(t *testing.T) {

	client := bdc.NewConn("http://127.0.0.1:23000", nil)

	length := 512
	buf := make([]byte, length)
	n, err := io.ReadFull(rand.Reader, buf)
	if err != nil || n != length {
		t.Fatal("rand read failed", n, err)
		return
	}
	sha1_ := sha1.New()
	sha1_.Write(buf)
	key := sha1_.Sum(nil)
	key1, err := client.Put2(xlog.NewDummy(), bytes.NewBuffer(buf), length, key, false, [3]uint16{0, 0xffff, 0xffff})
	if err != nil {
		t.Fatal("put local cache failed", err)
	}

	sha1_.Reset()
	n1, err := client.Get(xlog.NewDummy(), key1, sha1_, 0, length, [4]uint16{0, 0xffff, 0xffff, 1})
	if err != nil || n1 != int64(length) || bytes.Compare(sha1_.Sum(nil), key) != 0 {
		t.Fatal("get local cache failed", n1, err)
	}
}

func testDoCache(t *testing.T) {

	client := bdc.NewConn("http://127.0.0.1:23000", nil)

	length := 512
	buf := make([]byte, length)
	n, err := io.ReadFull(rand.Reader, buf)
	if err != nil || n != length {
		t.Fatal("rand read failed", n, err)
		return
	}
	sha1_ := sha1.New()
	sha1_.Write(buf)
	key := sha1_.Sum(nil)
	key1, err := client.Put2(xlog.NewDummy(), bytes.NewBuffer(buf), length, key, true, [3]uint16{0, 0xffff, 0xffff})
	if err != nil {
		t.Fatal("put local cache failed", err)
	}

	sha1_.Reset()
	n1, err := client.Get(xlog.NewDummy(), key1, sha1_, 0, length, [4]uint16{0, 0xffff, 0xffff, 1})
	if err != nil || n1 != int64(length) || bytes.Compare(sha1_.Sum(nil), key) != 0 {
		t.Fatal("get local cache failed", n1, err)
	}
}

func testMulti(t *testing.T) {
	testMultiWDoCache(t)
	testMultiRNoCache(t)
}

func testMultiRNoCache(t *testing.T) {

	client := bdc.NewConn("http://127.0.0.1:23000", nil)

	length := 512
	buf := make([]byte, length)
	n, err := io.ReadFull(rand.Reader, buf)
	if err != nil || n != length {
		t.Fatal("rand read failed", n, err)
		return
	}
	sha1_ := sha1.New()
	sha1_.Write(buf)
	key := sha1_.Sum(nil)
	key1, err := client.Put2(xlog.NewDummy(), bytes.NewBuffer(buf), length, key, false, [3]uint16{0, 0xffff, 0xffff})
	if err != nil {
		t.Fatal("put local cache failed", err)
	}

	reader := func() {
		sha1_ := sha1.New()
		n1, err := client.Get(xlog.NewDummy(), key1, sha1_, 0, length, [4]uint16{0, 0xffff, 0xffff, 1})
		if err != nil || n1 != int64(length) || bytes.Compare(sha1_.Sum(nil), key) != 0 {
			t.Fatal("get local cache failed", n1, err)
		}
	}

	for i := 0; i < 5; i++ {
		go reader()
	}
}

func testMultiWDoCache(t *testing.T) {

	client := bdc.NewConn("http://127.0.0.1:23000", nil)

	length := 512
	buf := make([]byte, length)
	n, err := io.ReadFull(rand.Reader, buf)
	if err != nil || n != length {
		t.Fatal("rand read failed", n, err)
		return
	}
	sha1_ := sha1.New()
	sha1_.Write(buf)
	key := sha1_.Sum(nil)
	ch := make(chan bool, 5)
	writer := func() {
		_, err := client.Put2(xlog.NewDummy(), bytes.NewBuffer(buf), length, key, false, [3]uint16{0, 0xffff, 0xffff})
		ch <- true
		if err != nil {
			t.Fatal("put local cache failed", err)
		}
	}

	for i := 0; i < 5; i++ {
		go writer()
	}
	for i := 0; i < 5; i++ {
		<-ch
	}

	sha1_.Reset()
	n1, err := client.Get(xlog.NewDummy(), key, sha1_, 0, length, [4]uint16{0, 0xffff, 0xffff, 1})
	if err != nil || n1 != int64(length) || bytes.Compare(sha1_.Sum(nil), key) != 0 {
		t.Fatal("get local cache failed", n1, err)
	}
}
