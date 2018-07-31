package retrys

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"testing"

	"github.com/qiniu/rpc.v1"
	"github.com/qiniu/xlog.v1"
	"github.com/stretchr/testify/assert"
	errors2 "qbox.us/errors"
)

var errNotFound = errors.New("test: not found")

type mockStg struct {
	datas   map[string][]byte
	maxRead int64
}

type fullReader struct {
	io.Reader
	n int64
}

func (p *fullReader) Read(b []byte) (int, error) {

	n := p.n
	if int64(len(b)) < n {
		n = int64(len(b))
	}
	n1, err := io.ReadFull(p.Reader, b[:n])
	p.n -= int64(n1)
	return n1, err
}

func (p *mockStg) Get(xl rpc.Logger, fh []byte, from, to int64) (rc io.ReadCloser, n int64, err error) {

	data, ok := p.datas[string(fh)]
	if !ok {
		err = errNotFound
		return
	}
	r := io.Reader(bytes.NewReader(data[from:to]))
	if to-from > p.maxRead {
		r = &fullReader{
			Reader: bytes.NewReader(data[from : from+p.maxRead]),
			n:      to - from,
		}
	}
	rc = ioutil.NopCloser(r)
	return
}

func TestIsRateLimitError(t *testing.T) {
	e := errors2.NewRateLimitError("test")
	assert.True(t, isRateLimitError(e))

	assert.False(t, isRateLimitError(errors.New("test")))
}

func TestRetrysGetter(t *testing.T) {

	g := &mockStg{
		datas: make(map[string][]byte),
	}
	fh := []byte("test")
	data := []byte("helloworld")
	from, to := int64(0), int64(len(data))

	xl := xlog.NewDummy()
	p := NewRetrys(g, 0)
	w := bytes.NewBuffer(nil)
	n, err := p.Get(xl, fh, w, from, to)
	assert.Equal(t, errNotFound, err)
	assert.Equal(t, 0, n)

	g.datas[string(fh)] = data
	g.maxRead = 7
	w = bytes.NewBuffer(nil)
	n, err = p.Get(xl, fh, w, from, to)
	assert.Equal(t, io.ErrUnexpectedEOF, err)
	assert.Equal(t, 7, n)
	assert.Equal(t, data[0:7], w.Bytes())

	p = NewRetrys(g, 1)
	w = bytes.NewBuffer(nil)
	n, err = p.Get(xl, fh, w, from, to)
	assert.NoError(t, err)
	assert.Equal(t, 10, n)
	assert.Equal(t, data, w.Bytes())

	from, to = from+1, to-1
	w = bytes.NewBuffer(nil)
	n, err = p.Get(xl, fh, w, from, to)
	assert.NoError(t, err)
	assert.Equal(t, 8, n)
	assert.Equal(t, data[from:to], w.Bytes())

	from, to = from+1, to-1
	w = bytes.NewBuffer(nil)
	n, err = p.Get(xl, fh, w, from, to)
	assert.NoError(t, err)
	assert.Equal(t, 6, n)
	assert.Equal(t, data[from:to], w.Bytes())
}
