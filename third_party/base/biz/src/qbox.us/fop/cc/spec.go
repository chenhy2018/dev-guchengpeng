package cc

import (
	"github.com/qiniu/xlog.v1"
	"io"
)

// --------------------------------------------------------------------

type Getter interface {
	Get(xl *xlog.Logger, key []byte, from, to int, bds [4]uint16) (r io.ReadCloser, length int64, err error) // 要求在 to 参数超出范围时不认为是错误
}

type CacheGetter interface {
	Getter
}

// --------------------------------------------------------------------

type DiskCache interface {
	Get(xl *xlog.Logger, key []byte) (r io.ReadCloser, length int64, err error)
	RangeGet(xl *xlog.Logger, key []byte, from, to int64) (r io.ReadCloser, length int64, err error)
	Set(xl *xlog.Logger, key []byte, r io.Reader, length int64) (err error)
	SetEx(xl *xlog.Logger, key []byte, r io.Reader, length int64, sha1 []byte) (err error)
}
