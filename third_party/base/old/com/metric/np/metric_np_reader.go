package metric

import (
	"errors"
	"io"
	"log"
	"os"
)

var EIoCanceled = errors.New("io canceled")

// -------------------------------------------------------------------------

type CancelableReader struct {
	Source   io.Reader
	Canceled bool
}

func (r *CancelableReader) Cancel() {
	r.Canceled = true
}

func (r *CancelableReader) Read(buf []byte) (n int, err os.Error) {

	if r.Canceled {
		err = EIoCanceled
		return
	}

	return r.Source.Read(buf)
}

// -------------------------------------------------------------------------

type Reader struct {
	Source   io.Reader
	Offset   int64
	Canceled bool
}

func (r *Reader) Cancel() {
	r.Canceled = true
}

func (r *Reader) Read(buf []byte) (n int, err os.Error) {

	if r.Canceled {
		err = EIoCanceled
		return
	}

	n, err = r.Source.Read(buf)
	r.Offset += int64(n)
	return
}

// -------------------------------------------------------------------------

type ReadSeeker struct {
	Source   io.ReadSeeker
	Offset   int64
	Canceled bool
}

func (r *ReadSeeker) Cancel() {
	r.Canceled = true
}

func (r *ReadSeeker) Read(buf []byte) (n int, err os.Error) {

	if r.Canceled {
		err = EIoCanceled
		return
	}

	n, err = r.Source.Read(buf)
	r.Offset += int64(n)
	return
}

func (r *ReadSeeker) Seek(offset int64, whence int) (ret int64, err os.Error) {

	ret, err = r.Source.Seek(offset, whence)
	if err == nil {
		r.Offset = ret
	}
	return
}

// -------------------------------------------------------------------------

type ReaderAt struct {
	Source   io.ReaderAt
	Offset   int64
	Canceled bool
}

func (r *ReaderAt) Cancel() {
	r.Canceled = true
}

func (r *ReaderAt) Read(buf []byte) (n int, err os.Error) {

	if r.Canceled {
		err = EIoCanceled
		return
	}

	n, err = r.Source.ReadAt(buf, r.Offset)
	r.Offset += int64(n)
	if err != nil {
		if err == os.EOF {
			return
		}
		if e1, ok := err.(*os.PathError); ok {
			if e2, ok := e1.Error.(os.Errno); ok { // windows
				if int(e2) == 38 {
					err = os.EOF
				}
			}
		}
	}
	return
}

func (r *ReaderAt) Seek(offset int64, whence int) (ret int64, err os.Error) {

	switch whence {
	case 0:
		r.Offset = offset
	case 1:
		r.Offset += offset
	default:
		log.Println("metric.ReaderAt.Seek: invalid arguments")
		err = os.EINVAL
		return
	}
	return r.Offset, nil
}

// -------------------------------------------------------------------------
