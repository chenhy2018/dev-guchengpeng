package dcutil

import (
	"bytes"
	"errors"
	"io"
)

type multiReaderAt struct {
	head           []byte
	headLen        int64
	r              io.ReaderAt
	offsetOfReader int64
}

func (mr *multiReaderAt) ReadAt(buf []byte, off int64) (n int, err error) {
	if off < 0 {
		return 0, errors.New("multiReaderAt.ReadAt: negative off")
	}
	headReader := bytes.NewReader(mr.head)
	headLeft := mr.headLen - off
	if headLeft <= 0 {
		offNew := off - mr.headLen
		n, err = mr.r.ReadAt(buf, offNew)
		return
	}
	if int64(len(buf)) <= headLeft {
		return headReader.ReadAt(buf, off)
	}
	n, err = headReader.ReadAt(buf[:headLeft], off)
	if err != nil {
		return
	}
	n2, err := mr.r.ReadAt(buf[headLeft:], 0)
	n += n2
	return
}

func (mr *multiReaderAt) Read(buf []byte) (n int, err error) {
	if len(buf) == 0 {
		return 0, nil
	}
	n, err = mr.ReadAt(buf, mr.offsetOfReader)
	mr.offsetOfReader += int64(n)
	return
}

func MultiReaderAt(head []byte, r io.ReaderAt) *multiReaderAt {
	return &multiReaderAt{head, int64(len(head)), r, 0}
}
