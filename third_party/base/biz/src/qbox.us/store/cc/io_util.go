package cc

import (
	"errors"
	"io"
)

// ---------------------------------------------------

type ReadWriterAt interface {
	io.ReaderAt
	io.WriterAt
}

// -----------------------------------------------------------

type Writer struct {
	io.WriterAt
	Offset int64
}

func (p *Writer) Write(val []byte) (n int, err error) {
	n, err = p.WriteAt(val, p.Offset)
	p.Offset += int64(n)
	return
}

// ---------------------------------------------------

type reader struct {
	io.ReaderAt
	offset int64
	base   int64
}

func NewReader(r io.ReaderAt, offset int64) *reader {

	return &reader{r, offset, offset}
}

func (p *reader) Read(val []byte) (n int, err error) {
	n, err = p.ReadAt(val, p.offset)
	p.offset += int64(n)
	return
}

var ErrWhence = errors.New("Seek: invalid whence")
var ErrOffset = errors.New("Seek: invalid offset")

func (p *reader) Seek(offset int64, whence int) (ret int64, err error) {

	switch whence {
	case 0:
		offset += p.base
	case 1:
		offset += p.offset
	default:
		return 0, ErrWhence
	}

	if offset < p.base {
		return 0, ErrOffset
	}

	p.offset = offset

	return offset - p.base, nil
}

// ---------------------------------------------------

type NilReader struct{}

func (r NilReader) Read(val []byte) (n int, err error) {
	return 0, io.EOF
}

// ---------------------------------------------------

type BytesReader []byte // see strings.Reader

func NewBytesReader(val []byte) *BytesReader {
	return (*BytesReader)(&val)
}

func (p *BytesReader) Read(val []byte) (n int, err error) {
	s := *p
	if len(s) == 0 {
		return 0, io.EOF
	}
	n = copy(val, s)
	*p = s[n:]
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

func (p *BytesWriter) Write(val []byte) (n int, err error) {
	n = copy(p.b[p.n:], val)
	if n == 0 && len(val) > 0 {
		err = io.EOF
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

// ---------------------------------------------------

type optimisticMultiWriter struct {
	writers []io.Writer
	errs    []error
}

func (t *optimisticMultiWriter) Write(p []byte) (n int, err error) {
	errNum := 0
	for i, w := range t.writers {
		if t.errs[i] != nil {
			errNum++
			continue
		}
		n, err = w.Write(p)
		if err != nil {
			t.errs[i] = err
		}
		if n != len(p) {
			t.errs[i] = io.ErrShortWrite
		}

		if t.errs[i] != nil {
			errNum++
		}
	}
	if errNum == len(t.writers) {
		return
	}
	return len(p), nil
}

func (t *optimisticMultiWriter) Errors() []error {
	return t.errs
}

func OptimisticMultiWriter(writers ...io.Writer) *optimisticMultiWriter {
	return &optimisticMultiWriter{writers, make([]error, len(writers))}
}

// ---------------------------------------------------
