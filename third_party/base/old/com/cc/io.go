package cc

import (
	"io"
	"os"
	"qbox.us/log"
)

// ---------------------------------------------------

type ReadWriterAt interface {
	io.ReaderAt
	io.WriterAt
}

// ---------------------------------------------------

type Reseter interface {
	Reset() os.Error
}

type ReadReseter interface {
	io.Reader
	Reseter
}

type WriteReseter interface {
	io.Writer
	Reseter
}

// ---------------------------------------------------

type Writer struct {
	io.WriterAt
	Offset int64
}

func (p *Writer) Write(val []byte) (n int, err os.Error) {
	log.Debug("cc.Writer,offset:", p.Offset)
	n, err = p.WriteAt(val, p.Offset)
	p.Offset += int64(n)
	return
}

// ---------------------------------------------------

type Reader struct {
	io.ReaderAt
	Offset int64
}

func (p *Reader) Read(val []byte) (n int, err os.Error) {
	n, err = p.ReadAt(val, p.Offset)
	p.Offset += int64(n)
	return
}

// ---------------------------------------------------

type NilReader struct{}
type NilWriter struct{}

func (r NilReader) Read(val []byte) (n int, err os.Error) {
	return 0, os.EOF
}

func (r NilWriter) Write(val []byte) (n int, err os.Error) {
	return len(val), nil
}

// ---------------------------------------------------

type BytesReader struct {
	b   []byte // see strings.Reader
	off int
}

func NewBytesReader(val []byte) *BytesReader {
	return &BytesReader{val, 0}
}

func (r *BytesReader) Seek(offset int64, whence int) (ret int64, err os.Error) {
	switch whence {
	case 0:
	case 1:
		offset += int64(r.off)
	case 2:
		offset += int64(len(r.b))
	default:
		err = os.EINVAL
		return
	}
	if offset < 0 {
		err = os.EINVAL
		return
	}
	if offset >= int64(len(r.b)) {
		r.off = len(r.b)
	} else {
		r.off = int(offset)
	}
	ret = int64(r.off)
	return
}

func (r *BytesReader) Read(val []byte) (n int, err os.Error) {
	n = copy(val, r.b[r.off:])
	if n == 0 && len(val) != 0 {
		err = os.EOF
		return
	}
	r.off += n
	return
}

func (r *BytesReader) Close() (err os.Error) {
	return
}

// ---------------------------------------------------

type BytesWriter struct {
	b []byte
	n int
}

func NewBytesWriter(buff []byte) *BytesWriter {
	return &BytesWriter{buff, 0}
}

func (p *BytesWriter) Write(val []byte) (n int, err os.Error) {
	n = copy(p.b[p.n:], val)
	if n == 0 && len(val) > 0 {
		err = os.EOF
		return
	}
	p.n += n
	return
}

func (p *BytesWriter) Len() int {
	return p.n
}

func (p *BytesWriter) Bytes() []byte {
	return p.b[:p.n]
}

func (p *BytesWriter) Reset() {
	p.n = 0
}

// ---------------------------------------------------
