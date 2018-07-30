// +build ignore

package mockdc

import (
	"bytes"
	"crypto/sha1"
	"io"
	"io/ioutil"
	"qbox.us/dc/clients"
	"testing"
	"time"
)

var cfg = &Config{
	Key: []byte("dc1"),
}

var Addr = ":4009"
var Host = "http://127.0.0.1" + Addr

func doTest(t *testing.T) {
	key := []byte("key123")
	client := clients.NewConn(Host, nil)
	r, length, err := client.Get(key)
	if err == nil {
		t.Fatal("Get missing cache:", "err should not be nil")
	}

	buf := bytes.NewBufferString("data123")
	err = client.Set(key, buf, int64(buf.Len()))
	if err != nil {
		t.Fatal("write cache error", err)
	}

	buf = new(bytes.Buffer)
	r, length, err = client.Get(key)
	if err != nil {
		t.Fatal("get cache err", err)
	}
	defer r.Close()
	io.CopyN(buf, r, length)
	if buf.String() != "data123" {
		t.Error("error cache data:", buf.String())
	}
}

func doTestUnknownLength(t *testing.T) {
	key := []byte("key321")
	client := clients.NewConn(Host, nil)
	buf := bytes.NewBufferString("data321")
	err := client.Set(key, buf, -1)
	if err != nil {
		return
	} else {
		t.Error("error should not be nil.")
	}
}

func doTestRewrite(t *testing.T) {
	key := []byte("key456")
	client := clients.NewConn(Host, nil)
	buf := bytes.NewBufferString("data456")
	err := client.Set(key, buf, int64(buf.Len()))
	if err != nil {
		t.Fatal(err)
	}

	buf = bytes.NewBufferString("data789")
	err = client.Set(key, buf, int64(buf.Len()))
	if err != nil {
		t.Fatal(err)
	}

	buf = new(bytes.Buffer)
	r, length, err := client.Get(key)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	io.CopyN(buf, r, length)
	if buf.String() != "data789" {
		t.Error("error rewrite cache:", buf.String())
	}
}

func doTestRange(t *testing.T) {
	key := []byte("key789")
	client := clients.NewConn(Host, nil)
	buf := bytes.NewBufferString("data789")
	err := client.Set(key, buf, int64(buf.Len()))
	if err != nil {
		t.Fatal(err)
	}

	buf = new(bytes.Buffer)
	r, length, err := client.RangeGet(key, 1, 3)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	io.CopyN(buf, r, length)
	if buf.String() != "at" {
		t.Error("error RangeGet:", buf.String())
	}
}

func doTestWriteCheck(t *testing.T) {
	key := []byte("abcd")
	client := clients.NewConn(Host, nil)
	buf := bytes.NewBufferString("dataabcd")
	buf2 := bytes.NewBufferString("dataabcd")

	h := sha1.New()
	io.Copy(h, buf)
	checksum := h.Sum(nil)

	err := client.SetEx(key, buf2, int64(buf2.Len()), checksum)
	if err != nil {
		t.Fatal(err)
	}

	buf = new(bytes.Buffer)
	r, length, err := client.Get(key)
	if err != nil {
		t.Fatal("get cache err", err)
	}
	defer r.Close()
	io.CopyN(buf, r, length)
	if buf.String() != "dataabcd" {
		t.Error("error cache data:", buf.String())
	}

}

func Test(t *testing.T) {

	server, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	dir, _ := ioutil.TempDir("", "mockdc_test")
	go server.Run(Addr, dir)
	time.Sleep(1e9)

	doTest(t)
	doTestUnknownLength(t)
	doTestRewrite(t)
	doTestRange(t)
	doTestWriteCheck(t)
}
