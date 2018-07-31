package filelog

import (
	"qbox.us/errors"
	"qbox.us/largefile"
	"sync/atomic"
)

type Writer struct {
	Offset int64
	*largefile.Instance
}

func Open(name string, chunkBits uint) (r *Writer, err error) {

	f, err := largefile.Open(name, chunkBits)
	if err != nil {
		err = errors.Info(err, "largefile.filelog.Open", name, chunkBits).Detail(err)
		return
	}
	fsize, err := f.Size()
	if err != nil {
		err = errors.Info(err, "largefile.filelog.Open", name, chunkBits).Detail(err)
		return
	}
	return &Writer{Instance: f, Offset: fsize}, nil
}

func (p *Writer) Write(val []byte) (n int, err error) {

	nw := int64(len(val))
	off := atomic.AddInt64(&p.Offset, nw)
	return p.WriteAt(val, off-nw)
}
