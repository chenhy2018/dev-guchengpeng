package teapotlog

import (
	"strconv"
	"strings"
	"time"

	qbytes "github.com/qiniu/bytes"
	. "github.com/qiniu/http/audit/proto"
	"github.com/teapots/teapot"
)

// ----------------------------------------------------------

type ResponseWriter struct {
	teapot.ResponseWriter
	Body         *qbytes.Writer
	Extra        M
	WrittenBytes int64
	StartT       int64
	Mod          string
	Code         int
	Xlog         bool
	Skip         bool
	NoLogBody    bool // Affect response with 2xx only.
}

const xlogKey = "X-Log"
const xwanKey = "X-Warn"
const maxXlogLen = 509 // 512 - len("...")

func (r *ResponseWriter) Write(buf []byte) (n int, err error) {
	if r.Xlog {
		r.logDuration(r.Code)
		fullXlog, trunced := r.xlogMerge()
		if trunced {
			defer func() {
				r.setXlog(fullXlog)
			}()
		}
		r.Xlog = false
		if r.Code/100 == 2 && r.NoLogBody {
			r.Skip = true
		}
	}
	n, err = r.ResponseWriter.Write(buf)
	r.WrittenBytes += int64(n)
	if n == len(buf) && !r.Skip {
		n2, _ := r.Body.Write(buf)
		if n2 == n {
			return
		}
	}
	r.Skip = true
	return
}

func (r *ResponseWriter) getBody() []byte {
	if r.Skip {
		return nil
	}
	return r.Body.Bytes()
}

func (r *ResponseWriter) ExtraDisableBodyLog() {
	r.NoLogBody = true
}

func (r *ResponseWriter) xlogMerge() (fullXlog string, trunc bool) {
	headers := r.Header()
	v, ok := headers[xlogKey]
	if !ok {
		return
	}

	defer func() {
		if len(fullXlog) > maxXlogLen {
			trunc = true
			headers[xlogKey] = []string{"..." + fullXlog[len(fullXlog)-maxXlogLen:]}
		}
	}()

	if len(v) <= 1 {
		fullXlog = v[0]
		return
	}
	fullXlog = strings.Join(v, ";")
	headers[xlogKey] = []string{fullXlog}
	return
}

func (r *ResponseWriter) setXlog(xlog string) {
	headers := r.Header()
	_, ok := headers[xlogKey]
	if !ok {
		return
	}
	headers[xlogKey] = []string{xlog}
}

//
// X-Log: xxx; MOD[:duration][/code]
//
func (r *ResponseWriter) WriteHeader(code int) {
	if r.Xlog {
		r.logDuration(code)
		fullXlog, trunced := r.xlogMerge()
		if trunced {
			defer func() {
				r.setXlog(fullXlog)
			}()
		}
		r.Xlog = false
		if r.Code/100 == 2 && r.NoLogBody {
			r.Skip = true
		}
	}
	r.ResponseWriter.WriteHeader(code)
	r.Code = code
}

func (r *ResponseWriter) ExtraWrite(key string, val interface{}) {
	if r.Extra == nil {
		r.Extra = make(M)
	}
	r.Extra[key] = val
}

func (r *ResponseWriter) ExtraAddInt64(key string, val int64) {
	if r.Extra == nil {
		r.Extra = make(M)
	}
	if v, ok := r.Extra[key]; ok {
		val += v.(int64)
	}
	r.Extra[key] = val
}

func (r *ResponseWriter) ExtraAddString(key string, val string) {
	if r.Extra == nil {
		r.Extra = make(M)
	}
	var v []string
	if v1, ok := r.Extra[key]; ok {
		v = v1.([]string)
	}
	r.Extra[key] = append(v, val)
}

func (r *ResponseWriter) GetStatusCode() int {
	return r.Code
}

func (r *ResponseWriter) logDuration(code int) {
	h := r.Header()
	dur := (time.Now().UnixNano() - r.StartT) / 1e6
	xlog := r.Mod
	if dur > 0 {
		xlog += ":" + strconv.FormatInt(dur, 10)
	}
	if code/100 != 2 {
		xlog += "/" + strconv.Itoa(code)
	}
	h[xlogKey] = append(h[xlogKey], xlog)
}

// ----------------------------------------------------------
