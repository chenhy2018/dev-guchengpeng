package dc

import (
	"io"

	"github.com/qiniu/xlog.v1"
)

type DiskCache interface {
	Get(xl *xlog.Logger, key []byte) (r io.ReadCloser, length int64, err error)
	GetHint(xl *xlog.Logger, key []byte) (r io.ReadCloser, length int64, hint bool, err error)
	RangeGet(xl *xlog.Logger, key []byte, from, to int64) (r io.ReadCloser, length int64, err error)
	RangeGetHint(xl *xlog.Logger, key []byte, from, to int64) (r io.ReadCloser, length int64, hint bool, err error)
	RangeGetAndHost(xl *xlog.Logger, key []byte, from, to int64) (host string, r io.ReadCloser, length int64, err error)
	KeyHost(xl *xlog.Logger, key []byte) (host string, err error)
	Set(xl *xlog.Logger, key []byte, r io.Reader, length int64) (err error)
	SetEx(xl *xlog.Logger, key []byte, r io.Reader, length int64, sha1 []byte) (err error)
	SetWithHostRet(xl *xlog.Logger, key []byte, r io.Reader, length int64) (host string, err error)
	SetExWithHostRet(xl *xlog.Logger, key []byte, r io.Reader, length int64, checksum []byte) (host string, err error)
}

type DiskCacheExt struct {
	dc DiskCache
}

func NewDiskCacheExt(dc DiskCache) *DiskCacheExt {

	return &DiskCacheExt{dc}
}

func (p *DiskCacheExt) Get(xl *xlog.Logger, key []byte) (rc io.ReadCloser, n int, err error) {

	rc, n64, err := p.dc.Get(xl, key)
	return rc, int(n64), err
}

func (p *DiskCacheExt) RangeGet(xl *xlog.Logger, key []byte, from, to int64) (rc io.ReadCloser, n int, err error) {

	rc, n64, err := p.dc.RangeGet(xl, key, from, to)
	return rc, int(n64), err
}

func (p *DiskCacheExt) RangeGetAndHost(xl *xlog.Logger, key []byte, from, to int64) (host string, rc io.ReadCloser, n int, err error) {
	host, rc, n64, err := p.dc.RangeGetAndHost(xl, key, from, to)
	return host, rc, int(n64), err
}

func (p *DiskCacheExt) Put(xl *xlog.Logger, key []byte, r io.Reader, n int) error {

	return p.dc.Set(xl, key, r, int64(n))
}

func (p *DiskCacheExt) PutWithHostRet(xl *xlog.Logger, key []byte, r io.Reader, n int) (string, error) {

	return p.dc.SetWithHostRet(xl, key, r, int64(n))
}
