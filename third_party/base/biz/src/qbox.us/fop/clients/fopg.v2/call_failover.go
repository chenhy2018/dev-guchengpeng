package fopg

import (
	"net/http"
	"time"

	"github.com/qiniu/rpc.v1"
	"github.com/qiniu/xlog.v1"
	"golang.org/x/net/context"
)

const (
	DefaultClientTimeoutMS = 20000
)

type FailoverConfig struct {
	TryTimes        uint32 `json:"try_times"` // 对于某个请求 client 或者 failover 各自最大尝试次数，TryTimes = 重试次数 + 1
	ClientTimeoutMS uint32 `json:"client_timeout_ms"`
}

type FailoverClient struct {
	client   *failoverClient
	failover *failoverClient

	shouldFailover func(int, error) bool
}

type failoverClient struct {
	tryTimes uint32
	client   *rpc.Client
}

type FailoverRequest struct {
	req *http.Request

	ctx context.Context
}

func (req *FailoverRequest) Context() context.Context {
	if req.ctx != nil {
		return req.ctx
	}

	return context.Background()
}

func NewFailoverClient(cfg *FailoverConfig, clientTr, failoverTr http.RoundTripper, shouldFailover func(int, error) bool) *FailoverClient {
	if cfg.ClientTimeoutMS == 0 {
		cfg.ClientTimeoutMS = DefaultClientTimeoutMS // 不允许无限等待
	}

	client := newClient(cfg, clientTr)
	failover := newClient(cfg, failoverTr)

	if shouldFailover == nil {
		shouldFailover = ShouldFailover
	}

	return &FailoverClient{client, failover, shouldFailover}
}

func newClient(cfg *FailoverConfig, tr http.RoundTripper) *failoverClient {
	client := &rpc.Client{
		&http.Client{
			Transport: tr,
			Timeout:   time.Duration(cfg.ClientTimeoutMS) * time.Millisecond,
		},
	}

	return &failoverClient{
		tryTimes: cfg.TryTimes,
		client:   client,
	}
}

func (c *FailoverClient) DoCtx(req *FailoverRequest) (resp *http.Response, err error) {
	ctx := req.Context()
	xl := xlog.FromContextSafe(ctx)

	var code int
	resp, code, err = c.client.DoCtx(xl, req.req)
	select {
	case <-ctx.Done():
		return
	default:
	}

	if !c.shouldFailover(code, err) {
		return
	}

	resp, _, err = c.failover.DoCtx(xl, req.req)
	return
}

func (c *failoverClient) DoCtx(l rpc.Logger, req *http.Request) (resp *http.Response, code int, err error) {
	for i := uint32(0); i < c.tryTimes+1; i++ {
		resp, err = c.client.Do(l, req)
		if err == nil {
			return resp, resp.StatusCode, nil
		}
	}

	if err == nil {
		code = resp.StatusCode
	}

	return
}

var ShouldFailover = func(code int, err error) bool {
	return ShouldRetry(code, err)
}

func ShouldRetry(code int, err error) bool {
	if code == 502 || code == 503 || code == 504 {
		return true
	}
	if err == nil {
		return false
	}
	if _, ok := err.(rpc.RespError); ok {
		return false
	}
	return true
}
