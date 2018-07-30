package api

import (
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/qiniu/rpc.v1/lb.v2.1"
	"qbox.us/iam/enums"
)

type Client struct {
	client *lb.Client
}

func NewWithMultiHosts(hosts []string, t http.RoundTripper) *Client {
	return &Client{
		client: lb.New(&lb.Config{
			Hosts:    hosts,
			TryTimes: uint32(len(hosts)),
		}, t),
	}
}

type QueryInput interface {
	GetQueryString() string
}

type SpoofMethod struct {
	Method enums.SpoofMethod `json:"_method"`
}

type Paginator struct {
	Page     int
	PageSize int
}

func (p *Paginator) GetQuery() *url.Values {
	q := &url.Values{}
	q.Set("page", strconv.Itoa(p.Page))
	q.Set("page_size", strconv.Itoa(p.PageSize))
	return q
}

type CommonResponse struct {
	Code    int      `json:"code"`
	Message string   `json:"message"`
	Errors  []string `json:"errors"`
}

func formatTime(t time.Time) string {
	return t.Format(time.RFC3339)
}
