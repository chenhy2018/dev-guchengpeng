package ios

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"qbox.us/cc/time"
	"syscall"
)

var EIoCanceled = errors.New("io canceled")
var MetricExp = int64(5e8)

// -------------------------------------------------------------------------

type MetricReader struct {
	Source         io.ReadSeeker
	Tag            string
	OnUpdated      func(r *MetricReader)
	Offset         int64 // -1 means canceled
	Fsize          int64
	StartTime      int64
	LastOffset     int64
	LastMetricTime int64
	Speed          int64 // unit: B/s
	Progress       int   // unit: 1/10000
	canceled       bool
}

func onUpdate(r *MetricReader) {

	if r.Progress < 0 {
		fmt.Println("\ncanceled")
	} else {
		fmt.Printf("\r%2d%% - %v KB/s", r.Progress/100, r.Speed/1000)
	}
}

func NewMetricReader(src io.ReadSeeker, fsize int64) *MetricReader {

	tick0 := time.Nanoseconds()
	return &MetricReader{
		StartTime:      tick0,
		LastMetricTime: tick0,
		Source:         src,
		Fsize:          fsize,
		OnUpdated:      onUpdate,
	}
}

func (r *MetricReader) AvgSpeed() int {

	dtick := r.LastMetricTime - r.StartTime
	if dtick < 1e8 {
		return 0
	}
	return int((r.Offset * 10) / (dtick / 1e8))
}

func (r *MetricReader) progress(off int64) {

	if r.Fsize <= 0 {
		r.Progress = 10000
		return
	}

	tick := time.Nanoseconds()
	dtick := tick - r.LastMetricTime
	r.Progress = int(off * 10000 / r.Fsize)
	if dtick < MetricExp {
		return
	}
	doff := off - r.LastOffset
	if doff <= 0 {
		r.Speed = 0
	} else {
		r.Speed = (doff * 10) / (dtick / 1e8)
	}
	r.LastOffset, r.LastMetricTime = off, tick
}

func (r *MetricReader) Cancel() {
	r.canceled = true
}

func (r *MetricReader) Read(buf []byte) (n int, err error) {

	if r.canceled {
		r.Progress = -1
		r.OnUpdated(r)
		err = EIoCanceled
		return
	}

	n, err = r.Source.Read(buf)
	if n == 0 {
		return
	}

	r.Offset += int64(n)
	r.progress(r.Offset)
	r.OnUpdated(r)
	return
}

func (r *MetricReader) Seek(offset int64, whence int) (ret int64, err error) {

	ret, err = r.Source.Seek(offset, whence)
	if err == nil {
		r.Offset = ret
		r.LastOffset, r.LastMetricTime = r.Offset, time.Nanoseconds()
	}
	return
}

// -------------------------------------------------------------------------

type MetricReaderEx struct {
	Source   io.ReadSeeker
	Offset   int64
	Progress func(off int64)
	Canceled bool
}

func (r *MetricReaderEx) Cancel() {
	r.Canceled = true
}

func (r *MetricReaderEx) Read(buf []byte) (n int, err error) {

	if r.Canceled {
		r.Progress(-1)
		err = EIoCanceled
		return
	}

	n, err = r.Source.Read(buf)
	if n != 0 {
		r.Offset += int64(n)
		r.Progress(r.Offset)
	}
	return
}

func (r *MetricReaderEx) Seek(offset int64, whence int) (ret int64, err error) {

	ret, err = r.Source.Seek(offset, whence)
	if err == nil {
		r.Offset = ret
	}
	return
}

// -------------------------------------------------------------------------

type MetricReaderAt struct {
	Source   io.ReaderAt
	Offset   int64
	Progress func(off int64)
	Canceled bool
}

func (r *MetricReaderAt) Cancel() {
	r.Canceled = true
}

func (r *MetricReaderAt) Read(buf []byte) (n int, err error) {

	if r.Canceled {
		r.Progress(-1)
		err = EIoCanceled
		return
	}

	n, err = r.Source.ReadAt(buf, r.Offset)
	if n != 0 {
		r.Offset += int64(n)
		r.Progress(r.Offset)
	}
	if err != nil {
		if err == io.EOF {
			return
		}
		if e1, ok := err.(*os.PathError); ok {
			if e2, ok := e1.Err.(syscall.Errno); ok { // windows
				if int(e2) == 38 {
					err = io.EOF
				}
			}
		}
	}
	return
}

func (r *MetricReaderAt) Seek(offset int64, whence int) (ret int64, err error) {

	switch whence {
	case 0:
		r.Offset = offset
	case 1:
		r.Offset += offset
	default:
		log.Println("MetricReaderEx.Seek: invalid arguments")
		err = syscall.EINVAL
		return
	}
	return r.Offset, nil
}

// -------------------------------------------------------------------------

type CancelableReader struct {
	Source   io.Reader
	Canceled bool
}

func (r *CancelableReader) Cancel() {
	r.Canceled = true
}

func (r *CancelableReader) Read(buf []byte) (n int, err error) {

	if r.Canceled {
		err = EIoCanceled
		return
	}

	return r.Source.Read(buf)
}

// -------------------------------------------------------------------------
