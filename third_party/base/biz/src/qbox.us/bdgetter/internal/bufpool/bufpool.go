package bufpool

import (
	"io"
	"sync"
)

var bufPool = sync.Pool{
	New: func() interface{} {
		b := make([]byte, 32*1024)
		return &b
	},
}

func Copy(w io.Writer, r io.Reader) (n int64, err error) {
	buf := bufPool.Get().(*[]byte)
	n, err = io.CopyBuffer(w, r, *buf)
	bufPool.Put(buf)
	return
}
