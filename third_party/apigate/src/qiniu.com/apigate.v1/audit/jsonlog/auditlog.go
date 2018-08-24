package jsonlog

import (
	"bytes"
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/qiniu/bytes/seekable"
	"github.com/qiniu/encoding1.2/json"
	"github.com/qiniu/errors"
	"qiniu.com/auth/authstub.v1"

	. "github.com/qiniu/apigate.v1/proto"
	qbytes "github.com/qiniu/bytes"
)

// ----------------------------------------------------------

type M map[string]interface{}

type responseWriter struct {
	http.ResponseWriter
	body      *qbytes.Writer
	extra     M
	written   int64
	startT    int64
	mod       string
	code      int
	xlog      bool
	skip      bool
	noLogBody bool // Affect response with 2xx only.

	//
	Logger
	url     string
	headerM M
	paramsM M
}

const xlogKey = "X-Log"
const maxXlogLen = 509 // 512 - len("...")

func (r *responseWriter) Write(buf []byte) (n int, err error) {

	if r.xlog {
		r.logDuration(r.code)
		fullXlog, trunced := r.xlogMerge()
		if trunced {
			defer func() {
				r.setXlog(fullXlog)
			}()
		}
		r.xlog = false
		if r.code/100 == 2 && r.noLogBody {
			r.skip = true
		}
	}
	n, err = r.ResponseWriter.Write(buf)
	r.written += int64(n)
	if n == len(buf) && !r.skip {
		n2, _ := r.body.Write(buf)
		if n2 == n {
			return
		}
	}
	r.skip = true
	return
}

func (r *responseWriter) getBody() []byte {

	if r.skip {
		return nil
	}
	return r.body.Bytes()
}

func (r *responseWriter) xlogMerge() (fullXlog string, trunc bool) {

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

func (r *responseWriter) CloseNotify() <-chan bool {

	if cn, ok := r.ResponseWriter.(http.CloseNotifier); ok {
		return cn.CloseNotify()
	}

	return make(chan bool, 1)
}

func (r *responseWriter) setXlog(xlog string) {

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
func (r *responseWriter) WriteHeader(code int) {

	if r.xlog {
		r.logDuration(code)
		fullXlog, trunced := r.xlogMerge()
		if trunced {
			defer func() {
				r.setXlog(fullXlog)
			}()
		}
		r.xlog = false
		if r.code/100 == 2 && r.noLogBody {
			r.skip = true
		}
	}
	r.ResponseWriter.WriteHeader(code)
	r.code = code
}

func (r *responseWriter) GetStatusCode() int {

	return r.code
}

func (r *responseWriter) logDuration(code int) {

	h := r.Header()
	dur := (time.Now().UnixNano() - r.startT) / 1e6
	xlog := r.mod
	if dur > 0 {
		xlog += ":" + strconv.FormatInt(dur, 10)
	}
	if code/100 != 2 {
		xlog += "/" + strconv.Itoa(code)
	}
	h[xlogKey] = append(h[xlogKey], xlog)
}

// ----------------------------------------------------------

func set(h M, header http.Header, key string) {
	if v, ok := header[key]; ok {
		h[key] = v[0]
	}
}

func ip(addr string) string {
	pos := strings.Index(addr, ":")
	if pos < 0 {
		return addr
	}
	return addr[:pos]
}

func queryToJson(m map[string][]string) (h M, err error) {

	h = make(M)
	for k, v := range m {
		if len(v) == 1 {
			h[k] = v[0]
		} else {
			h[k] = v
		}
	}
	return
}

func decodeRequest(req *http.Request) (url string, h, params M) {

	h = M{"IP": ip(req.RemoteAddr), "Host": req.Host}
	ct, ok := req.Header["Content-Type"]
	if ok {
		h["Content-Type"] = ct[0]
	}
	if req.URL.RawQuery != "" {
		h["RawQuery"] = req.URL.RawQuery
	}

	set(h, req.Header, "User-Agent")
	set(h, req.Header, "Range")
	set(h, req.Header, "Refer")
	set(h, req.Header, "Content-Length")
	set(h, req.Header, "If-None-Match")
	set(h, req.Header, "If-Modified-Since")
	set(h, req.Header, "X-Real-Ip")
	set(h, req.Header, "X-Forwarded-For")
	set(h, req.Header, "X-Scheme")
	set(h, req.Header, "X-Remote-Ip")
	set(h, req.Header, "X-Reqid")
	set(h, req.Header, "X-Id")
	set(h, req.Header, "X-From-Cdn")
	set(h, req.Header, "Cdn-Src-Ip")
	set(h, req.Header, "Cdn-Scheme")
	set(h, req.Header, "X-Upload-Encoding")
	set(h, req.Header, "Accept-Encoding")

	url = req.URL.Path
	if ok {
		switch ct[0] {
		case "application/x-www-form-urlencoded":
			seekable, err := seekable.New(req)
			if err == nil {
				req.ParseForm()
				params, _ = queryToJson(req.Form)
				seekable.SeekToBegin()
			}
		case "application/json":
			seekable, err := seekable.New(req)
			if err == nil {
				json.NewDecoder(req.Body).Decode(&params)
				seekable.SeekToBegin()
			}
		}
	}
	if params == nil {
		params = make(M)
	}
	return
}

func decodeResponse(header http.Header, bodyThumb []byte, h, params M) (resph M, body []byte) {

	if h == nil {
		h = make(M)
	}

	ct, ok := header["Content-Type"]
	if ok {
		h["Content-Type"] = ct[0]
	}
	if xlog, ok := header["X-Log"]; ok {
		h["X-Log"] = xlog
	}
	set(h, header, "X-Reqid")
	set(h, header, "X-Id")
	set(h, header, "Content-Length")
	set(h, header, "Content-Encoding")

	if ok && ct[0] == "application/json" && header.Get("Content-Encoding") != "gzip" {
		if -1 == bytes.IndexAny(bodyThumb, "\n\r") {
			body = bodyThumb
		}
	}
	resph = h
	return
}

// ----------------------------------------------------------

type LogWriter interface {
	Log(mod string, msg []byte) error
}

type Logger struct {
	w     LogWriter
	limit int
}

func New(w LogWriter, limit int) *Logger {

	if limit == 0 {
		limit = 1024
	}
	return &Logger{w, limit}
}

func (r *Logger) OpenRequest(ctx context.Context, w *http.ResponseWriter, req *http.Request) RequestEvent {

	url, headerM, paramsM := decodeRequest(req)
	startTime := time.Now().UnixNano()
	mod := ModFromContextSafe(ctx)

	e := &responseWriter{
		ResponseWriter: *w,
		body:           qbytes.NewWriter(make([]byte, r.limit)),
		code:           200,
		xlog:           true,
		mod:            mod,
		startT:         startTime,

		Logger:  *r,
		url:     url,
		headerM: headerM,
		paramsM: paramsM,
	}
	*w = e

	return e
}

type reqLogger struct {
	Logger
	mod       string
	url       string
	headerM   M
	paramsM   M
	startTime int64
}

func (r *responseWriter) AuthParsed(w http.ResponseWriter, req *http.Request) {

	auth, ok := req.Header["Authorization"]
	if !ok {
		return
	}

	user, err := authstub.Parse(auth[0])
	if err != nil {
		return
	}

	token := M{
		"uid":   user.Uid,
		"utype": user.Utype,
	}
	if user.Appid != 0 {
		token["appid"] = user.Appid
	}
	if user.Sudoer != 0 || user.UtypeSu != 0 {
		token["suid"] = user.Sudoer
		token["sut"] = user.UtypeSu
	}
	r.headerM["Token"] = token
}

func (r *responseWriter) End(w http.ResponseWriter, req *http.Request) {

	url, headerM, paramsM := r.url, r.headerM, r.paramsM

	var header, params, resph []byte
	if len(headerM) != 0 {
		header, _ = json.Marshal(headerM)
	}
	if len(paramsM) != 0 {
		params, _ = json.Marshal(paramsM)
		if len(params) > r.limit {
			params, _ = json.Marshal(M{"discarded": len(params)})
		}
	}

	startTime := r.startT / 100
	endTime := time.Now().UnixNano() / 100

	b := bytes.NewBuffer(nil)

	b.WriteString("REQ\t")
	b.WriteString(r.mod)
	b.WriteByte('\t')

	b.WriteString(strconv.FormatInt(startTime, 10))
	b.WriteByte('\t')
	b.WriteString(req.Method)
	b.WriteByte('\t')
	b.WriteString(url)
	b.WriteByte('\t')
	b.Write(header)
	b.WriteByte('\t')
	b.Write(params)
	b.WriteByte('\t')

	w1 := r
	resphM, respb := decodeResponse(w1.Header(), w1.getBody(), w1.extra, paramsM)
	if len(resphM) != 0 {
		resph, _ = json.Marshal(resphM)
	}

	b.WriteString(strconv.Itoa(w1.code))
	b.WriteByte('\t')
	b.Write(resph)
	b.WriteByte('\t')
	b.Write(respb)
	b.WriteByte('\t')
	b.WriteString(strconv.FormatInt(w1.written, 10))
	b.WriteByte('\t')
	b.WriteString(strconv.FormatInt(endTime-startTime, 10))

	err := r.w.Log(r.mod, b.Bytes())
	if err != nil {
		errors.Info(err, "Log failed").Detail(err).Warn()
	}
}

// ----------------------------------------------------------
