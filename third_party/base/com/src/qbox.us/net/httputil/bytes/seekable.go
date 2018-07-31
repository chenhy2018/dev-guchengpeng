// [DEPRECATED] please use "github.com/qiniu/bytes/seekable" .
package bytes

import (
	"io"
	"net/http"

	"github.com/qiniu/bytes"
	"qbox.us/errors"
)

var MaxSeekableLength int64 = 16 * 1024 * 1024

// ---------------------------------------------------

func readAll(src io.Reader, length int) (b []byte, err error) {
	b = make([]byte, length)
	_, err = io.ReadFull(src, b)
	return
}

// ---------------------------------------------------

type SeekToBeginer interface {
	SeekToBegin() error
}

type Seekabler interface {
	io.Reader
	SeekToBeginer
	Bytes() []byte
}

type SeekableCloser interface {
	Seekabler
	io.Closer
}

// ---------------------------------------------------

type readCloser struct {
	Seekabler
	io.Closer
}

func Seekable(req *http.Request) (r SeekableCloser, err error) {
	if req.ContentLength <= 0 || req.ContentLength > MaxSeekableLength {
		err = errors.Info(errors.EINVAL, "Seekable")
		return
	}
	if req.Body != nil {
		var ok bool
		if r, ok = req.Body.(SeekableCloser); ok {
			return
		}
		b, err2 := readAll(req.Body, int(req.ContentLength))
		if err2 != nil {
			err = errors.Info(err2, "Seekable").Detail(err)
			return
		}
		r = bytes.NewReader(b)
		req.Body = readCloser{r, req.Body}
	} else {
		err = errors.Info(errors.EINVAL, "Seekable")
	}
	return
}

// ---------------------------------------------------
