package directio

import (
	"io"
	"qbox.us/errors"
	//	"github.com/qiniu/log.v1"
)

type Reader struct {
	buf          []byte
	lastErr      error
	f            io.ReaderAt
	off, limit   int64
	pos, bufSize int
}

func NewReaderSize(f io.ReaderAt, off, n int64, bufBits uint) *Reader {

	bufSize := 1 << bufBits
	buf := make([]byte, bufSize)
	r := &Reader{
		buf:     buf,
		f:       f,
		off:     off,
		limit:   off + n,
		bufSize: bufSize,
	}

	mod := int(off) & (bufSize - 1)
	r.fetch(bufSize - mod)
	return r
}

func (r *Reader) fetch(nn int) {

	if r.off+int64(nn) > r.limit {
		nn = int(r.limit - r.off)
		r.lastErr = io.EOF
	}

	//	log.Debug("directio.Reader.fetch", nn)

	nn, err := readAtFull(r.f, r.buf[:nn], r.off)
	r.off += int64(nn)
	r.pos = 0
	r.buf = r.buf[:nn]
	if err != nil {
		r.lastErr = err
	}
}

func readAtFull(f io.ReaderAt, buf []byte, off int64) (n int, err error) {

	var nn int

	for n < len(buf) {
		nn, err = f.ReadAt(buf[n:], off)
		n += nn
		off += int64(nn)
		if err != nil {
			break
		}
	}
	return
}

func (r *Reader) Read(b []byte) (n int, err error) {

	n = copy(b, r.buf[r.pos:])
	r.pos += n
	if r.pos == len(r.buf) {
		if r.lastErr != nil {
			err = r.lastErr
			return
		}
		r.fetch(r.bufSize)
	}
	return
}

func (r *Reader) Close() error {
	return nil
}

//---------------------------------------------------------------------------//

type Writer struct {
	buf          []byte
	f            io.WriterAt
	off, limit   int64
	pos, bufSize int
}

func NewWriterSize(f io.WriterAt, off, n int64, bufBits uint) *Writer {

	bufSize := 1 << bufBits
	buf := make([]byte, bufSize)
	mod := int(off) & (bufSize - 1)
	r := &Writer{
		buf:     buf[:bufSize-mod],
		f:       f,
		off:     off,
		limit:   off + n,
		bufSize: bufSize,
	}

	return r
}

func (w *Writer) Close() (err error) {

	if w.off+int64(w.pos) != w.limit {
		return errors.ErrShortWrite
	}
	_, err = w.f.WriteAt(w.buf[:w.pos], w.off)
	w.f = nil
	return
}

func (w *Writer) Write(p []byte) (n int, err error) {

	for {
		nn := copy(w.buf[w.pos:], p)
		n += nn
		w.pos += nn
		if len(p) == nn {
			break
		}
		p = p[nn:]
		if w.off+int64(w.pos) > w.limit {
			err = errors.ErrNoSpace
			return
		}
		_, err = w.f.WriteAt(w.buf, w.off)
		if err != nil {
			return
		}
		w.off += int64(w.pos)
		w.buf = w.buf[:w.bufSize]
		w.pos = 0
	}
	return
}
