package noproxy

import (
	"errors"
	"net"
	"net/http"
	"path"
	"time"
)

var DefaultCallbackBlockIPs = []string{"127.*.*.*", "192.168.*.*"}

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

var errIPBlocked = errors.New("ip is blocked")

// -----------------------------------------------------------------------------

// if timeout is zero, it means no timeout.
func blockIPTransport(blockIPs []string, timeout time.Duration, userAgent string) *uaTransport {

	dialTimeout := func(network, address string) (conn net.Conn, err error) {
		d := net.Dialer{Timeout: timeout}
		conn, err = d.Dial(network, address)
		if err != nil {
			return
		}
		remoteAddr := conn.RemoteAddr().String()
		if h1, _, err := net.SplitHostPort(remoteAddr); err == nil {
			remoteAddr = h1
		}
		for _, blockIP := range blockIPs {
			if match, _ := path.Match(blockIP, remoteAddr); match {
				err = errIPBlocked
				conn.Close()
				return
			}
		}
		return
	}

	return &uaTransport{
		userAgent: userAgent,
		RoundTripper: &http.Transport{
			Dial: dialTimeout,
			ResponseHeaderTimeout: timeout, // zero means no timeout
		},
	}
}
