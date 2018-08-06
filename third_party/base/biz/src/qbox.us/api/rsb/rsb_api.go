package rsb

import (
	"github.com/qiniu/rpc.v1"
	"io"
	"net/http"
	"qbox.us/api"
	"github.com/qiniu/xlog.v1"
	"strconv"
)

type Service struct {
	Host string
	Conn *rpc.Client
}

func New(host string, t http.RoundTripper) *Service {
	client := &rpc.Client{&http.Client{Transport: t}}
	return &Service{host, client}
}

func (p *Service) Dump(xl *xlog.Logger, bucket string, from, to int64) (r io.ReadCloser, code int, err error) {
	url := p.Host + "/dump?tbl=" + bucket + "&ft=" + strconv.FormatInt(from, 10) + "&tt=" + strconv.FormatInt(to, 10)
	resp, err := p.Conn.Get(xl, url)
	if err != nil {
		return nil, -1, err
	}
	if resp.StatusCode != 200 {
		err = api.NewError(code)
	}
	return resp.Body, resp.StatusCode, err
}
