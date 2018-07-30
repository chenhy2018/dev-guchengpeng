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
)

type Conn struct {
	host string
	conn *http.Client
}

func NewConn(host string, transport http.RoundTripper) *Conn {
	return &Conn{host, &http.Client{Transport: transport}}
}

func (p *Conn) OpWithRT(xl *xlog.Logger, fh []byte, fsize int64, fopCtx *fop.FopCtx, nop bool, conn *http.Client) (
	r io.ReadCloser, length int64, mime string, needCache bool, err error, retry bool) {

	if conn == nil {
		return p.Op(xl, fh, fsize, fopCtx, nop)
	}

	if nop {
		return p.op(xl, "/nop", fh, fsize, fopCtx, conn)
	}
	return p.op(xl, "/op", fh, fsize, fopCtx, conn)
}

func (p *Conn) PersistentOp(xl *xlog.Logger, ret interface{}, fh []byte, fsize int64, fopCtx *fop.FopCtx, conn *http.Client) (err error) {

	if conn == nil {
		conn = p.conn
	}
	url := p.host + "/pop?" + fopgQuery(fh, fsize, fopCtx)
	return rpc.Client{conn}.Call(xl, &ret, url)
}

func (p *Conn) RangeOp(xl *xlog.Logger, fh []byte, fsize int64, fopCtx *fop.FopCtx, xrange string, conn *http.Client) (
	r io.ReadCloser, length int64, mime string, contentRange string, err error, retry bool) {

	if conn == nil {
		conn = p.conn
	}

	url := p.host + "/rop?" + fopgQuery(fh, fsize, fopCtx)
	xl.Info("Fopg.RangeOp url:", url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		retry = true
		return
	}
	req.Header.Set("Range", xrange)

	resp, err := rpc.Client{conn}.Do(xl, req)
	if err != nil {
		retry = true
		return
	}

	if resp.StatusCode/100 != 2 {
		defer resp.Body.Close()
		err = rpc.ResponseError(resp)
		return
	}

	mime = resp.Header.Get("Content-Type")
	contentRange = resp.Header.Get("Content-Range")
	return resp.Body, resp.ContentLength, mime, contentRange, nil, false
}

func (p *Conn) List(xl rpc.Logger) (fops []string, err error) {

	err = rpc.DefaultClient.Call(xl, &fops, p.host+"/list")
	return
}

func (p *Conn) Op(xl *xlog.Logger, fh []byte, fsize int64, fopCtx *fop.FopCtx, nop bool) (
	r io.ReadCloser, length int64, mime string, needCache bool, err error, retry bool) {

	if nop {
		return p.op(xl, "/nop", fh, fsize, fopCtx, p.conn)
	}
	return p.op(xl, "/op", fh, fsize, fopCtx, p.conn)
}

func (p *Conn) op(xl *xlog.Logger, path string, fh []byte, fsize int64, fopCtx *fop.FopCtx, conn *http.Client) (
	r io.ReadCloser, length int64, mime string, needCache bool, err error, retry bool) {

	url := p.host + path + "?" + fopgQuery(fh, fsize, fopCtx)

	xl.Info("Conn.op: fop url:", url)

	client := rpc.Client{conn}
	resp, err := client.Get(xl, url)
	if err != nil {
		retry = true
		return
	}

	if resp.StatusCode/100 != 2 {
		defer resp.Body.Close()
		err = rpc.ResponseError(resp)
		return
	}
	mime = resp.Header.Get("Content-Type")
	needCache = (resp.Header.Get("X-Cache-Control") != "no-cache")
	return resp.Body, resp.ContentLength, mime, needCache, nil, false
}

func fopgQuery(fh []byte, fsize int64, fopCtx *fop.FopCtx) string {

	v := gourl.Values{}
	v.Set("fh", base64.URLEncoding.EncodeToString(fh))
	v.Set("fsize", strconv.FormatInt(fsize, 10))
	if fopCtx != nil {
		v.Set("cmd", fopCtx.RawQuery)
		v.Set("sp", fopCtx.StyleParam)
		v.Set("url", fopCtx.URL)
		v.Set("token", fopCtx.Token)
		v.Set("uid", strconv.FormatUint(uint64(fopCtx.Uid), 10))
		v.Set("bucket", fopCtx.Bucket)
		v.Set("mime", fopCtx.MimeType)
		v.Set("pipelineId", fopCtx.PipelineId)
		if fopCtx.Mode != 0 {
			v.Set("mode", strconv.FormatUint(uint64(fopCtx.Mode), 10))
			v.Set("force", strconv.FormatUint(uint64(fopCtx.Force), 10))
			v.Set("key", fopCtx.Key)
		}
	}
	return v.Encode()
}
