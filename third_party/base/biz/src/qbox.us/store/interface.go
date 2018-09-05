package store

import (
	"io"
	ioadmin "qbox.us/admin_api/io"
	"github.com/qiniu/xlog.v1"
)

// -----------------------------------------------------------

type Getter interface {
	Get(xl *xlog.Logger, key []byte, w io.Writer, from, to int, bds [4]uint16) error // 要求在 to 参数超出范围时不认为是错误
}

type Putter interface {
	Put(xl *xlog.Logger, key []byte, r io.Reader, n int, bds [3]uint16) (err error)
}

type Interface interface {
	Getter
	Putter
}

type CachedPutter interface {
	Put(xl *xlog.Logger, key []byte, r io.Reader, n int, doCache bool, bds [3]uint16) (err error)
}

type ServiceStater interface {
	ServiceStat() (info []*ioadmin.CacheInfoEx, err error)
}

type MultiBdInterface interface {
	Getter
	CachedPutter
	ServiceStater
}

// -----------------------------------------------------------
