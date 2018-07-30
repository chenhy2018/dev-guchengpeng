package nginxproxy

import (
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync/atomic"
	"time"
)

// =====================================================

type uaTransport struct {
	userAgent string
	http.RoundTripper
}

func (p *uaTransport) RoundTrip(req *http.Request) (*http.Response, error) {

	req.Header.Set("User-Agent", p.userAgent)
	return p.RoundTripper.RoundTrip(req)
}

func (p *uaTransport) NestedObject() interface{} {
	return p.RoundTripper
}

type transports struct {
	trs []http.RoundTripper
	idx uint32
}

func newTransports(trs []http.RoundTripper) *transports {
	return &transports{
		trs: trs,
	}
}

func (p *transports) Pick() (http.RoundTripper, uint32) {

	idx := atomic.AddUint32(&p.idx, 1) % uint32(len(p.trs))
	return p.trs[idx], idx
}

func (p *transports) Next(idx1 uint32) (http.RoundTripper, uint32) {

	idx := (idx1 + 1) % uint32(len(p.trs))
	return p.trs[idx], idx
}

func (p *transports) Len() int {

	return len(p.trs)
}

// -----------------------------------------------------------------------------

type proxyTransport struct {
	*http.Transport
}

func (p *proxyTransport) RoundTrip(req *http.Request) (resp *http.Response, err error) {

	req.Header.Set("X-Proxy-For", req.URL.Host)
	return p.Transport.RoundTrip(req)
}

func (p *proxyTransport) NestedObject() interface{} {
	return p.Transport
}

func newProxyTransports(proxies []string, timeout time.Duration, userAgent string) *transports {

	if len(proxies) == 0 {
		panic("len(proxies) == 0")
	}

	dialer := net.Dialer{Timeout: timeout}
	trs := make([]http.RoundTripper, len(proxies))
	for i, proxy := range proxies {
		if strings.Index(proxy, "://") == -1 {
			proxy = "http://" + proxy
		}
		url, err := url.Parse(proxy)
		if err != nil {
			panic("newProxyTransports: url parse failed " + err.Error())
		}
		trs[i] = &uaTransport{
			userAgent: userAgent,
			RoundTripper: &proxyTransport{Transport: &http.Transport{
				Dial:  dialer.Dial,
				Proxy: http.ProxyURL(url),
				ResponseHeaderTimeout: timeout,
			}},
		}
	}
	return newTransports(trs)
}
