package timeio

import (
	"io"
	"net/http"
	"time"
)

type Reader struct {
	t   int64
	err error
	r   io.Reader
}

// [DEPRECATED] please use github.com/qiniu/xlog
func NewReader(reader io.Reader) *Reader {
	return &Reader{r: reader}
}

func (r *Reader) Read(p []byte) (n int, err error) {
	begin := time.Now().UnixNano()
	defer func() {
		end := time.Now().UnixNano()
		r.t += end - begin
	}()
	n, err = r.r.Read(p)
	if err != nil {
		r.err = err
	}
	return
}

func (r *Reader) Duration() time.Duration {
	return time.Duration(r.t)
}

func (r *Reader) Time() int64 {
	return r.t
}

func (r *Reader) Error() error {
	return r.err
}

//---------------------------------------------------------------------------//

type Writer struct {
	t   int64
	err error
	w   io.Writer
}

// [DEPRECATED] please use github.com/qiniu/xlog
func NewWriter(writer io.Writer) *Writer {
	return &Writer{w: writer}
}

func (r *Writer) Write(p []byte) (n int, err error) {
	begin := time.Now().UnixNano()
	defer func() {
		end := time.Now().UnixNano()
		r.t += end - begin
	}()
	n, err = r.w.Write(p)
	if err != nil {
		r.err = err
	}
	return
}

func (r *Writer) Duration() time.Duration {
	return time.Duration(r.t)
}

func (r *Writer) Time() int64 {
	return r.t
}

func (r *Writer) Error() error {
	return r.err
}

//---------------------------------------------------------------------------//

type ResponseWriter struct {
	w *Writer
	http.ResponseWriter
}

// [DEPRECATED] please use github.com/qiniu/xlog
func NewResponseWriter(resp http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{w: NewWriter(resp), ResponseWriter: resp}
}

func (r *ResponseWriter) Write(p []byte) (n int, err error) {
	return r.w.Write(p)
}

func (r *ResponseWriter) Duration() time.Duration {
	return r.w.Duration()
}

func (r *ResponseWriter) Time() int64 {
	return r.w.Time()
}

func (r *ResponseWriter) Error() error {
	return r.w.Error()
}

func GetResponseTime(resp http.ResponseWriter) int64 {
	if w, ok := resp.(*ResponseWriter); ok {
		return w.Time()
	}
	return -1
}
