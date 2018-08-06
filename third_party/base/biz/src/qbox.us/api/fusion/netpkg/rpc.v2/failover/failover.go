package failover

import (
	"errors"
	"io"
	"net/http"

	"github.com/qiniu/rpc.v2"
)

var (
	ErrServiceNotAvailable = errors.New("service not available")
)

const (
	DefaultTryTimes = 2
)

// --------------------------------------------------------------------
// func ShouldRetry

func ShouldRetry(err error) bool {

	if _, ok := err.(*rpc.ErrorInfo); ok {
		return false
	}
	return true
}

// --------------------------------------------------------------------

type Config struct {
	Http        *http.Client
	ShouldRetry func(error) bool
	TryTimes    int
}

var defaultCfg = new(Config)

// --------------------------------------------------------------------
// type Client

type Client struct {
	hosts       []string
	client      rpc.Client
	shouldRetry func(error) bool
	tryTimes    int
}

func New(hosts []string, cfg *Config) *Client {

	if cfg == nil {
		cfg = defaultCfg
	}

	client := rpc.Client{cfg.Http}
	if client.Client == nil {
		client.Client = http.DefaultClient
	}

	p := &Client{hosts: hosts, client: client, shouldRetry: cfg.ShouldRetry, tryTimes: cfg.TryTimes}
	if p.shouldRetry == nil {
		p.shouldRetry = ShouldRetry
	}
	if p.tryTimes == 0 {
		p.tryTimes = DefaultTryTimes
	}
	return p
}

func (p *Client) Call(
	l rpc.Logger, ret interface{}, method, path string) (err error) {

	err = ErrServiceNotAvailable
	for i := 0; i < int(p.tryTimes); i++ {
		if i >= len(p.hosts) {
			break
		}
		err = p.client.Call(l, ret, method, p.hosts[i]+path)
		if err != nil && p.shouldRetry(err) {
			continue
		}
		return
	}
	return
}

func (p *Client) CallWith(
	l rpc.Logger, ret interface{}, method, path, bodyType string, body io.Reader, bodyLength int) (err error) {

	err = ErrServiceNotAvailable
	for i := 0; i < int(p.tryTimes); i++ {
		if i >= len(p.hosts) {
			break
		}
		err = p.client.CallWith(l, ret, method, p.hosts[i]+path, bodyType, body, bodyLength)
		if err != nil && p.shouldRetry(err) {
			continue
		}
		return
	}
	return
}

func (p *Client) CallWith64(
	l rpc.Logger, ret interface{}, method, path, bodyType string, body io.Reader, bodyLength int64) (err error) {

	err = ErrServiceNotAvailable
	for i := 0; i < int(p.tryTimes); i++ {
		if i >= len(p.hosts) {
			break
		}
		err = p.client.CallWith64(l, ret, method, p.hosts[i]+path, bodyType, body, bodyLength)
		if err != nil && p.shouldRetry(err) {
			continue
		}
		return
	}
	return
}

func (p *Client) CallWithForm(
	l rpc.Logger, ret interface{}, method, path string, params map[string][]string) (err error) {

	err = ErrServiceNotAvailable
	for i := 0; i < int(p.tryTimes); i++ {
		if i >= len(p.hosts) {
			break
		}
		err = p.client.CallWithForm(l, ret, method, p.hosts[i]+path, params)
		if err != nil && p.shouldRetry(err) {
			continue
		}
		return
	}
	return
}

func (p *Client) CallWithJson(
	l rpc.Logger, ret interface{}, method, path string, params interface{}) (err error) {

	err = ErrServiceNotAvailable
	for i := 0; i < int(p.tryTimes); i++ {
		if i >= len(p.hosts) {
			break
		}
		err = p.client.CallWithJson(l, ret, method, p.hosts[i]+path, params)
		if err != nil && p.shouldRetry(err) {
			continue
		}
		return
	}
	return
}

func (p *Client) DoRequestWithJson(l rpc.Logger,
	method, path string, params interface{}) (resp *http.Response, err error) {

	err = ErrServiceNotAvailable
	for i := 0; i < int(p.tryTimes); i++ {
		if i >= len(p.hosts) {
			break
		}
		resp, err = p.client.DoRequestWithJson(l, method, p.hosts[i]+path, params)
		if err != nil && p.shouldRetry(err) {
			continue
		}
		return
	}
	return
}

func (p *Client) DoRequestWithForm(l rpc.Logger,
	method, path string, data map[string][]string) (resp *http.Response, err error) {

	err = ErrServiceNotAvailable
	for i := 0; i < int(p.tryTimes); i++ {
		if i >= len(p.hosts) {
			break
		}
		resp, err = p.client.DoRequestWithForm(l, method, p.hosts[i]+path, data)
		if err != nil && p.shouldRetry(err) {
			continue
		}
		return
	}
	return
}

// --------------------------------------------------------------------
