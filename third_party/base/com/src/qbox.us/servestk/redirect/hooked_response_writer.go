package redirect

import (
	"net/http"
)

type hookedResponseWriter struct {
	http.ResponseWriter
	hookedWriteHeader func(code int)
	hookedWrite       func(p []byte) (n int, err error)
}

func (self *hookedResponseWriter) WriteHeader(code int) {
	self.hookedWriteHeader(code)
}

func (self *hookedResponseWriter) Write(p []byte) (int, error) {
	return self.hookedWrite(p)
}
