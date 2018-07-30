package stream

import (
	"io"
	"syscall"

	"qbox.us/fh/proto"

	"github.com/qiniu/bytes"
	"github.com/qiniu/xlog.v1"
)

type stream struct {
	xl    *xlog.Logger
	g     proto.CommonGetter
	fh    []byte
	off   int64
	fsize int64
}

func New(xl *xlog.Logger, g proto.CommonGetter, fh []byte, fsize int64) proto.Source {

	return &stream{
		xl:    xl,
		g:     g,
		fh:    fh,
		fsize: fsize,
	}
}

func (s *stream) QueryFhandle() ([]byte, error) {

	return s.fh, nil
}

func (s *stream) UploadedSize() (int64, error) {

	return s.fsize, nil
}

func (s *stream) Read(buf []byte) (n int, err error) {

	s.xl.Warn("stream.Read, should use WriteTo")
	to := s.off + int64(len(buf))
	if to > s.fsize {
		to = s.fsize
	}
	if s.off >= to {
		err = io.EOF
		return
	}

	w := bytes.NewWriter(buf)
	_, err = s.g.Get(s.xl, s.fh, w, s.off, to)
	n = w.Len()
	s.off += int64(n)
	return
}

func (s *stream) Seek(offset int64, whence int) (ret int64, err error) {

	switch whence {
	case 0:
		ret = offset
	case 1:
		ret = s.off + offset
	case 2:
		ret = s.fsize + offset
	default:
		ret = s.off
		err = syscall.EINVAL
		return
	}
	s.off = ret
	return
}

func (s *stream) WriteTo(w io.Writer) (n int64, err error) {

	if s.fsize == 0 && s.off == 0 {
		// Special case.
		return
	}
	n, err = s.g.Get(s.xl, s.fh, w, s.off, s.fsize)
	s.off += n
	return
}

func (s *stream) RangeRead(w io.Writer, from, to int64) (err error) {

	if s.fsize == 0 && from == 0 {
		// Special case.
		return
	}
	if to > s.fsize {
		to = s.fsize
	}
	if from >= to {
		err = syscall.EINVAL
		return
	}
	_, err = s.g.Get(s.xl, s.fh, w, from, to)
	return
}
