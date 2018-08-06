package dcutil

import (
	"bytes"
	"encoding/binary"
	"github.com/qiniu/ts"
	"github.com/qiniu/xlog.v1"
	"io"
	"io/ioutil"
	"strings"
	"syscall"
	"testing"
)

// --------------------------------------------------------------------

func (s CacheExt) GetOld(key []byte) (r io.ReadCloser, length int64, metas M, err error) {

	r, l, err := s.dc.Get(xlog.NewDummy(), key)
	length = int64(l)
	if err != nil {
		return
	}

	buf := make([]byte, gMinHeader)
	if _, err = io.ReadFull(r, buf); err != nil {
		return
	}

	ver := binary.LittleEndian.Uint16(buf[0:2])
	hl := binary.LittleEndian.Uint32(buf[2:6])

	buf = make([]byte, hl)
	if _, err = io.ReadFull(r, buf); err != nil {
		return
	}

	metas = M{}
	if parser := metaParsers[ver]; parser != nil {
		metas = parser(string(buf))
	}

	length -= int64(hl + gMinHeader)
	return
}

func (s CacheExt) SetOld(key []byte, r io.Reader, length int64, metas M) (err error) {

	strMetas := ""
	if producer := metaProducers[gVersion]; producer != nil {
		strMetas = producer(metas)
	}
	lenMetas := len(strMetas)

	b := make([]byte, lenMetas+gMinHeader)
	binary.LittleEndian.PutUint16(b[0:2], gVersion)
	binary.LittleEndian.PutUint32(b[2:6], uint32(lenMetas))
	copy(b[gMinHeader:], strMetas)

	buf := bytes.NewReader(b)
	err = s.dc.Put(xlog.NewDummy(), key, io.MultiReader(buf, r), int(length)+gMinHeader+lenMetas)
	return
}

var metaParsers = map[uint16]func(string) M{
	0: parseMetas0,
}

var metaProducers = map[uint16]func(M) string{
	0: produceMetas0,
}

func parseMetas0(s string) M {

	m := make(M)
	words := strings.Split(s, "&")
	for _, word := range words {
		if strings.HasPrefix(word, "m=") {
			m["mime"] = word[2:]
		}
	}
	return m
}

func produceMetas0(m M) string {

	metas := ""
	first := true
	for k, v := range m {
		if !first {
			metas += "&"
		}

		if k == "mime" {
			metas += "m=" + v
		}

		first = false
	}
	return metas
}

// --------------------------------------------------------------------

type cache map[string]string

func (r cache) Get(xl *xlog.Logger, key []byte) (rc io.ReadCloser, n int, err error) {

	if v, ok := r[string(key)]; ok {
		rc = ioutil.NopCloser(strings.NewReader(v))
		n = len(v)
		return
	}
	err = syscall.ENOENT
	return
}

func (r cache) GetWithHost(xl *xlog.Logger, host string, key []byte) (rc io.ReadCloser, length int64, err error) {
	rc, n, err := r.Get(xl, key)
	return rc, int64(n), err
}

func (r cache) RangeGet(xl *xlog.Logger, key []byte, from, to int64) (rc io.ReadCloser, n int, err error) {

	if v, ok := r[string(key)]; ok {
		v = v[from:to]
		rc = ioutil.NopCloser(strings.NewReader(v))
		n = len(v)
		return
	}
	err = syscall.ENOENT
	return
}

func (r cache) RangeGetAndHost(xl *xlog.Logger, key []byte, from, to int64) (host string, rc io.ReadCloser, n int, err error) {

	if v, ok := r[string(key)]; ok {
		v = v[from:to]
		rc = ioutil.NopCloser(strings.NewReader(v))
		n = len(v)
		return
	}
	err = syscall.ENOENT
	host = ""
	return
}

func (r cache) RangeGetWithHost(xl *xlog.Logger, host string, key []byte, from int64, to int64) (rc io.ReadCloser, length int64, err error) {
	rc, n, err := r.RangeGet(xl, key, from, to)
	return rc, int64(n), err
}

func (r cache) Put(xl *xlog.Logger, key []byte, rc io.Reader, n int) error {

	b := make([]byte, n)
	_, err := io.ReadFull(rc, b)
	if err != nil {
		return err
	}

	r[string(key)] = string(b)
	return nil
}

func (r cache) PutWithHostRet(xl *xlog.Logger, key []byte, rc io.Reader, n int) (string, error) {
	return "", r.Put(xl, key, rc, n)
}

// --------------------------------------------------------------------

func put(t *testing.T, c Cache, k, v string) {
	err := c.Put(xlog.NewDummy(), []byte(k), strings.NewReader(v), len(v))
	if err != nil {
		ts.Fatal(t, "r.Put failed:", err)
	}
}

func putExt(t *testing.T, c CacheExt, k, v string, m M) {
	err := c.Set(xlog.NewDummy(), []byte(k), strings.NewReader(v), int64(len(v)), m)
	if err != nil {
		ts.Fatal(t, "r.Put failed:", err)
	}
}

func putExtOld(t *testing.T, c CacheExt, k, v string, m M) {
	err := c.SetOld([]byte(k), strings.NewReader(v), int64(len(v)), m)
	if err != nil {
		ts.Fatal(t, "r.Put failed:", err)
	}
}

type getFunc func(xl *xlog.Logger, key []byte) (r io.ReadCloser, length int64, metas M, err error)
type rangeGetFunc func(xl *xlog.Logger, key []byte, from, to int64) (r io.ReadCloser, length int64, metas M, err error)

func check(t *testing.T, f getFunc, k, v string, m M) {
	rc, l, m2, err := f(xlog.NewDummy(), []byte(k))
	if err != nil || l != int64(len(v)) {
		ts.Fatal(t, "r.Get failed:", err, l)
	}
	v2, err := ioutil.ReadAll(rc)
	if err != nil {
		ts.Fatal(t, "r.Get result failed:", v, v2)
	}
	if len(m) != len(m2) {
		ts.Fatal(t, "r.Get result failed:", m, m2)
	}
	for k1, v1 := range m {
		if m2[k1] != v1 {
			ts.Fatal(t, "r.Get result failed:", m, m2)
		}
	}
}

func checkRange(t *testing.T, f rangeGetFunc, k, v string, m M, from, to int64) {
	rc, l, m2, err := f(xlog.NewDummy(), []byte(k), from, to)
	if err != nil || l != to-from {
		ts.Fatal(t, "RangeGet failed:", err, l)
	}
	v2, err := ioutil.ReadAll(rc)
	if err != nil || v[from:to] != string(v2) {
		ts.Fatal(t, "RangeGet result failed:", v, v2)
	}
	if len(m) != len(m2) {
		ts.Fatal(t, "RangeGet result failed:", m, m2)
	}
	for k1, v1 := range m {
		if m2[k1] != v1 {
			ts.Fatal(t, "RangeGet result failed:", m, m2)
		}
	}
}

func Test1(t *testing.T) {
	c := cache{}
	ce := NewExt(c)
	putExtOld(t, ce, "k", "foo", M{"mime": "1"})
	check(t, ce.Get, "k", "foo", M{"m": "1"})

	putExtOld(t, ce, "kr", "foobar", M{"mime": "1r"})
	checkRange(t, ce.RangeGet, "kr", "foobar", M{"m": "1r"}, 0, 6)
	checkRange(t, ce.RangeGet, "kr", "foobar", M{"m": "1r"}, 3, 5)
	checkRange(t, ce.RangeGet, "kr", "foobar", M{"m": "1r"}, 3, 3)
}

func Test2(t *testing.T) {
	c := cache{}
	ce := NewExt(c)
	putExt(t, ce, "k", "foo", M{"m": "2"})
	check(t, ce.Get, "k", "foo", M{"m": "2"})

	putExt(t, ce, "kr", "foobar", M{"m": "2r"})
	checkRange(t, ce.RangeGet, "kr", "foobar", M{"m": "2r"}, 0, 6)
	checkRange(t, ce.RangeGet, "kr", "foobar", M{"m": "2r"}, 3, 5)
	checkRange(t, ce.RangeGet, "kr", "foobar", M{"m": "2r"}, 3, 3)
}

func Test3(t *testing.T) {
	c := cache{}
	defaultConn = c
	ce := NewExt(c)
	putExt(t, ce, "k", "foo", M{"m": "2"})

	f := func(xl *xlog.Logger, key []byte) (r io.ReadCloser, length int64, metas M, err error) {
		return GetWithHost(xl, "", key)
	}
	check(t, f, "k", "foo", M{"m": "2"})

	putExt(t, ce, "kr", "foobar", M{"m": "2r"})

	rf := func(xl *xlog.Logger, key []byte, from, to int64) (r io.ReadCloser, length int64, metas M, err error) {
		return RangeGetWithHost(xl, "", key, from, to)
	}
	checkRange(t, rf, "kr", "foobar", M{"m": "2r"}, 0, 6)
	checkRange(t, rf, "kr", "foobar", M{"m": "2r"}, 3, 5)
	checkRange(t, rf, "kr", "foobar", M{"m": "2r"}, 3, 3)
}

// --------------------------------------------------------------------
