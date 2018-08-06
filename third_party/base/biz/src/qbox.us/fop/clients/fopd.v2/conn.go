package fopd

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"code.google.com/p/go.net/context"
	"github.com/qiniu/http/httputil.v1"
	"github.com/qiniu/rpc.v3"
	"github.com/qiniu/xlog.v1"

	"qbox.us/fop"
)

type Conn struct {
	Host string
	conn rpc.Client
	// tasks processing simultaneously on the host
	processingNum int64
	// when encountering network error, mark host is down temporarily. recover when network is ok.
	lastFailedTime int64
}

var DcCodes = []int{400}

func NewConn(host string, transport http.RoundTripper) *Conn {
	c := &Conn{
		Host: host,
		conn: rpc.Client{&http.Client{Transport: transport}},
	}
	return c
}

func (p *Conn) ProcessingNum() int64 {
	return atomic.LoadInt64(&p.processingNum)
}

func (p *Conn) LastFailedTime() int64 {
	return atomic.LoadInt64(&p.lastFailedTime)
}

func (p *Conn) Op2(tpCtx context.Context, fh []byte, fsize int64, ctx *fop.FopCtx) (
	resp *http.Response, err error) {

	atomic.AddInt64(&p.processingNum, 1)
	defer atomic.AddInt64(&p.processingNum, -1)

	u := p.Host + "/op2?" + EncodeQuery(fh, fsize, ctx)

	l := xlog.FromContextSafe(tpCtx)
	xl := xlog.NewWith(l)
	xl.Info("fopd.Op2 url:", u)

	resp, err = p.conn.DoRequest(tpCtx, "POST", u)
	if err != nil {
		if isServerFail(err) {
			atomic.StoreInt64(&p.lastFailedTime, time.Now().Unix())
		}
		return
	}
	atomic.StoreInt64(&p.lastFailedTime, 0)

	if resp.StatusCode != 200 {
		defer resp.Body.Close()
		err = rpc.ResponseError(resp)
	}
	return
}

func IsDcCode(code int) bool {
	for i := 0; i < len(DcCodes); i++ {
		if DcCodes[i] == code {
			return true
		}
	}
	return false
}

var (
	FopCacheErrHeader = "X-Fop-Cache-Err"
)

func SetCacheErrHeader(w http.ResponseWriter) {
	w.Header().Set(FopCacheErrHeader, "1")
}

func ShouldCacheError(err error, resp *http.Response, tpCtx *fop.FopCtx) (code int, errMsg string, ok bool) {

	if tpCtx.OutType != "dc" {
		return
	}

	respErr, ok := err.(rpc.RespError)
	if ok {
		code = respErr.HttpCode()
		if !IsDcCode(code) {
			ok = false
			return
		}

		if resp != nil && resp.Header.Get(FopCacheErrHeader) != "1" {
			ok = false
			return
		}

		e := map[string]string{
			"error": err.Error(),
		}
		errB, _ := json.Marshal(e)
		errMsg = string(errB)
	}
	return
}

func (p *Conn) Op(tpCtx context.Context, f io.Reader, fsize int64, ctx *fop.FopCtx) (
	resp *http.Response, err error) {

	atomic.AddInt64(&p.processingNum, 1)
	defer atomic.AddInt64(&p.processingNum, -1)

	u := p.Host + "/op?" + EncodeQuery(ctx.Fh, fsize, ctx)

	l := xlog.FromContextSafe(tpCtx)
	xl := xlog.NewWith(l)
	xl.Info("fopd.Op url:", u)

	var m map[string]string
	if ctx.PreviousXstats != "" {
		m = map[string]string{xQiniuFop: ctx.PreviousXstats}
	}
	resp, err = p.conn.DoRequestWith64Header(tpCtx, "POST", u, ctx.MimeType, f, -1, m) // use chunk
	if err != nil {
		if isServerFail(err) {
			select {
			case <-tpCtx.Done():
				err = httputil.NewError(499, "request was canceled, got: "+err.Error())
				return
			default:
			}
			atomic.StoreInt64(&p.lastFailedTime, time.Now().Unix())
		}
		return
	}
	atomic.StoreInt64(&p.lastFailedTime, 0)

	if resp.StatusCode != 200 {
		defer resp.Body.Close()
		err = rpc.ResponseError(resp)
		return
	}
	return
}

func isServerFail(err error) bool {
	return !isResponseTimeout(err) && err != context.DeadlineExceeded && err != context.Canceled
}

func isResponseTimeout(err error) bool {
	return strings.Contains(err.Error(), "timeout awaiting response headers")
}
