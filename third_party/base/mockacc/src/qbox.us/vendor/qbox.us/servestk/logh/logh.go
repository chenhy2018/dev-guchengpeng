package logh

import (
	"net/http"
	"github.com/qiniu/bytes"
	"qbox.us/cc"
	"github.com/qiniu/log.v1"
	"qbox.us/servestk"
)

// ----------------------------------------------------------

var disabledURLs []string

func DisableURL(url string) {
	disabledURLs = append(disabledURLs, url)
}

func IsURLDisabled(url string) bool {
	for _, m := range disabledURLs {
		if m == url {
			return true
		}
	}
	return false
}

// ----------------------------------------------------------

type responseWriter struct {
	http.ResponseWriter
	body   *bytes.Writer
	length int64
	code   int
}

func (r *responseWriter) Write(buf []byte) (n int, err error) {
	n, err = r.ResponseWriter.Write(buf)
	r.length += int64(n)
	r.body.Write(buf[:n])
	return
}

func (r *responseWriter) WriteHeader(code int) {
	r.ResponseWriter.WriteHeader(code)
	r.code = code
}

func Instance(w http.ResponseWriter, req *http.Request, f func(w http.ResponseWriter, req *http.Request)) {
	if IsURLDisabled(req.URL.Path) {
		servestk.SafeHandler(w, req, f)
		return
	}
	log.Println("==> Request:", req.Method, req.URL.String())
	var buf [1024]byte
	w1 := &responseWriter{w, cc.NewBytesWriter(buf[:]), 0, 200}
	servestk.SafeHandler(w1, req, f)
	if req.Method != "GET" {
		if ct, ok := w1.Header()["Content-Type"]; ok && ct[0] == "application/octet-stream" {
			log.Println("==> Reply:", w1.code, w1.length, w1.body.Bytes(), "\n")
		} else {
			log.Println("==> Reply:", w1.code, w1.length, string(w1.body.Bytes()), "\n")
		}
	}
}

// ----------------------------------------------------------

func Info(w http.ResponseWriter, key, val string) {
}

// ----------------------------------------------------------
