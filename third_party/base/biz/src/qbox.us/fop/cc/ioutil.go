package cc

import (
	"io"
)

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

func (p *BytesWriter) Reset() {
	p.n = 0
}

// ---------------------------------------------------
