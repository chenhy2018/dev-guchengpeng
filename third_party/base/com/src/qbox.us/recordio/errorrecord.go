package recordio

import (
	"io"
	"net/http"
)

type ErrorRecordReader struct {
	r   io.Reader
	err error
}

func NewErrorRecordReader(r io.Reader) *ErrorRecordReader { return &ErrorRecordReader{r: r} }

func (self *ErrorRecordReader) Read(p []byte) (int, error) {
	n, err := self.r.Read(p)
	if err != nil {
		self.err = err
	}
	return n, err
}

func (self *ErrorRecordReader) Error() error { return self.err }

type ErrorRecordWriter struct {
	w   io.Writer
	err error
}

func NewErrorRecordWriter(w io.Writer) *ErrorRecordWriter { return &ErrorRecordWriter{w: w} }

func (self *ErrorRecordWriter) Write(p []byte) (int, error) {
	n, err := self.w.Write(p)
	if err != nil {
		self.err = err
	}
	return n, err
}

func (self *ErrorRecordWriter) Error() error { return self.err }

type ErrorRecordResponseWriter struct {
	http.ResponseWriter
	err error
}

func NewErrorRecordResponseWriter(w http.ResponseWriter) *ErrorRecordResponseWriter {
	return &ErrorRecordResponseWriter{ResponseWriter: w}
}

func (self *ErrorRecordResponseWriter) Write(p []byte) (int, error) {
	n, err := self.ResponseWriter.Write(p)
	if err != nil {
		self.err = err
	}
	return n, err
}

func (self *ErrorRecordResponseWriter) Error() error { return self.err }
