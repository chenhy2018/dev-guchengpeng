package sha1bd

import (
	"io"
	"syscall"

	"github.com/qiniu/bytes"
	"github.com/qiniu/xlog.v1"
	"qbox.us/fh/proto"
)

type stream struct {
	xl    *xlog.Logger
	g     proto.Sha1bdGetter
	fh    []byte
	raw   []byte
	off   int64
	fsize int64
	bds   [4]uint16
}

func NewStream(xl *xlog.Logger, g proto.Sha1bdGetter, raw []byte, fsize int64) (*stream, error) {

	fh, bds := DecodeFh(raw)
	if len(fh)%20 != 1 {
		return nil, proto.ErrInvalidFh
	}
	if (fh[0] & 0x80) != 0 {
		if len(fh) != 21 {
			return nil, proto.ErrInvalidFh
		}
	}

	return &stream{
		xl:    xl,
		g:     g,
		fh:    fh,
		raw:   raw,
		fsize: fsize,
		bds:   bds,
	}, nil
}

func (s *stream) QueryFhandle() ([]byte, error) {

	return s.raw, nil
}

func (s *stream) UploadedSize() (int64, error) {

	return s.fsize, nil
}

func (p *stream) Read(buf []byte) (n int, err error) {

	to := p.off + int64(len(buf))
	if to > p.fsize {
		to = p.fsize
	}
	if p.off >= to {
		err = io.EOF
		return
	}

	w := bytes.NewWriter(buf)
	err = StreamRead(p.xl, p.g, p.fh, w, p.off, to, p.bds)
	n = w.Len()
	p.off += int64(n)
	return
}

func (p *stream) Seek(offset int64, whence int) (ret int64, err error) {

	switch whence {
	case 0:
		ret = offset
	case 1:
		ret = p.off + offset
	case 2:
		ret = p.fsize + offset
	default:
		ret = p.off
		err = syscall.EINVAL
		return
	}
	p.off = ret
	return
}

type writer struct {
	io.Writer
	written int64
}

func (w *writer) Write(p []byte) (n int, err error) {

	n, err = w.Writer.Write(p)
	w.written += int64(n)
	return
}

func (p *stream) WriteTo(w1 io.Writer) (n int64, err error) {

	w := &writer{w1, 0}
	err = p.RangeRead(w, p.off, p.fsize)
	p.off += w.written
	n = w.written
	return
}

func (s *stream) RangeRead(w io.Writer, from, to int64) (err error) {

	if s.fsize == 0 {
		return nil
	}
	return StreamRead(s.xl, s.g, s.fh, w, from, to, s.bds)
}

// -----------------------------------------------------------------------------

func StreamRead(xl *xlog.Logger, g proto.Sha1bdGetter, fh []byte, w io.Writer, from int64, to int64, bds [4]uint16) (err error) {

	if (fh[0] & 0x80) == 0 {
		return StreamRead2(xl, g, fh[1:], uint(fh[0]), w, from, to, bds)
	}

	if from >= to {
		return syscall.EINVAL
	}

	chunkBits := uint(fh[0]) & 0x7f

	fromIdx := int(from>>chunkBits) * 20
	toIdx := int((to-1)>>chunkBits)*20 + 20

	b := make([]byte, toIdx-fromIdx)
	w2 := bytes.NewWriter(b)

	err = g.Get(xl, fh[1:], w2, fromIdx, toIdx, bds)
	if err != nil {
		return
	}
	if w2.Len() != toIdx-fromIdx {
		return syscall.EINVAL
	}

	fromBase := (from >> chunkBits) << chunkBits

	return StreamRead2(xl, g, b, chunkBits, w, from-fromBase, to-fromBase, bds)
}

func StreamRead2(xl *xlog.Logger, g proto.Sha1bdGetter, keys []byte, chunkBits uint, w io.Writer, from int64, to int64, bds [4]uint16) error {

	if from > to {
		return syscall.EINVAL
	}

	if from == to { // 合法空内容，比如文件大小为 0 的情况。
		return nil
	}

	var err error
	chunkSize := 1 << chunkBits

	fromIdx := int(from>>chunkBits) * 20
	toIdx := int(to>>chunkBits) * 20
	fromOff := int(from) & (chunkSize - 1)
	toOff := int(to) & (chunkSize - 1)

	if fromIdx == toIdx {
		return g.Get(xl, keys[fromIdx:fromIdx+20], w, fromOff, toOff, bds)
	}
	if fromOff != 0 {
		err = g.Get(xl, keys[fromIdx:fromIdx+20], w, fromOff, chunkSize, bds)
		if err != nil {
			return err
		}
		fromIdx += 20
	}
	for fromIdx < toIdx {
		err = g.Get(xl, keys[fromIdx:fromIdx+20], w, 0, chunkSize, bds)
		if err != nil {
			return err
		}
		fromIdx += 20
	}
	if toOff != 0 {
		err = g.Get(xl, keys[fromIdx:fromIdx+20], w, 0, toOff, bds)
	}
	return err
}
