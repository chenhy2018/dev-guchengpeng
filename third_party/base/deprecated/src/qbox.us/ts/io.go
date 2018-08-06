package ts

import (
	"bytes"
	crand "crypto/rand"
	"io"
	mrand "math/rand"
)

type BytesWriter struct {
	B []byte
	N int
}

func NewBytesWriter(buff []byte) *BytesWriter {
	return &BytesWriter{buff, 0}
}

func (p *BytesWriter) Write(val []byte) (n int, err error) {
	n = copy(p.B[p.N:], val)
	if n == 0 && len(val) > 0 {
		err = io.EOF
		return
	}
	p.N += n
	return
}

func (p *BytesWriter) Bytes() []byte {
	return p.B[:p.N]
}

// ---------------------------------------------------

type RandReader struct {
	b    []byte
	off  int
	blen int
	size int64
}

func NewBigRandBytes(size int64) ([]byte, error) {

	buf := new(bytes.Buffer)
	_, err := io.CopyN(buf, crand.Reader, size)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func NewRandReader(bigRandBytes []byte, size int64) *RandReader {

	blen := len(bigRandBytes)
	off := mrand.Int() % blen
	return &RandReader{bigRandBytes, off, blen, size}
}

func (p *RandReader) Read(v []byte) (nread int, err error) {

	if p.size <= 0 {
		return 0, io.EOF
	}
	if int64(len(v)) >= p.size {
		v = v[0:p.size]
	}
	for nread < len(v) {
		n := copy(v[nread:], p.b[p.off:])
		p.off += n
		if p.off == p.blen {
			p.off = 0
		}
		nread += n
	}
	p.size -= int64(nread)
	return
}

// ---------------------------------------------------

type WReader struct {
	r io.Reader
	w io.Writer
}

func NewWReader(r io.Reader, w io.Writer) *WReader {
	return &WReader{r, w}
}

func (p *WReader) Read(v []byte) (nread int, err error) {
	nread, err = p.r.Read(v)
	if err == nil {
		_, err = p.w.Write(v[0:nread])
	}
	return
}
