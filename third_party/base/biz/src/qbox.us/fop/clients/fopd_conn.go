package clients

import (
	"encoding/base64"
	"github.com/qiniu/rpc.v1"
	"github.com/qiniu/xlog.v1"
	"io"
	"net/http"
	gourl "net/url"
	"qbox.us/fop"
	"strconv"
	"sync/atomic"
	"time"
)

type FopdConn struct {
	host string
	conn rpc.Client
	// tasks processing simultaneously on the host
	processingNum int64
	// when encountering network error, mark host is down temporarily. recover when network is ok.
	lastFailedTime int64
}

func NewFopdConn(host string, transport http.RoundTripper) *FopdConn {
	c := &FopdConn{
		host:          host,
		conn:          rpc.Client{&http.Client{Transport: transport}},
		processingNum: 0,
	}
	return c
}

func (p *FopdConn) ProcessingNum() int64 {
	return atomic.LoadInt64(&p.processingNum)
}

func (p *FopdConn) LastFailedTime() int64 {
	return atomic.LoadInt64(&p.lastFailedTime)
}

func (p *FopdConn) Op(xl *xlog.Logger, f io.Reader, fsize int64, fopCtx *fop.FopCtx) (
	resp *http.Response, err error, retry bool) {

	atomic.AddInt64(&p.processingNum, 1)
	defer atomic.AddInt64(&p.processingNum, -1)

	url := p.host + "/op?"
	v := gourl.Values{}
	v.Set("fsize", strconv.FormatInt(fsize, 10))
	if fopCtx != nil {
		v.Set("cmd", fopCtx.RawQuery)
		v.Set("sp", fopCtx.StyleParam)
		v.Set("url", fopCtx.URL)
		v.Set("token", fopCtx.Token)
		if fopCtx.Mode != 0 {
			v.Set("mode", strconv.FormatUint(uint64(fopCtx.Mode), 10))
			v.Set("uid", strconv.FormatUint(uint64(fopCtx.Uid), 10))
			v.Set("bucket", fopCtx.Bucket)
			v.Set("key", fopCtx.Key)
			v.Set("fh", base64.URLEncoding.EncodeToString(fopCtx.Fh))
		}
	}
	url += v.Encode()

	xl.Info("fopd url:", url)

	// make http.Request's ContentLength = -1.
	// because only when ContentLength = -1, bdSource.WriteTo method can be called.
	// see: net/http.transferWriter.WriteBody method.
	fsize = -1
	resp, err = p.conn.PostWith64(xl, url, fopCtx.MimeType, f, fsize)
	if err != nil {
		retry = true // only retry when network error
		atomic.StoreInt64(&p.lastFailedTime, time.Now().Unix())
		return
	}
	atomic.StoreInt64(&p.lastFailedTime, 0)

	if resp.StatusCode != 200 {
		defer resp.Body.Close()
		err = rpc.ResponseError(resp)
		return
	}
	return resp, nil, false
}
