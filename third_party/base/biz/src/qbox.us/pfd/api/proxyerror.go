// +build go1.8

package api

import (
	"net"
	"net/url"
)

func isProxyError(err error) bool {
	if e, ok := err.(*url.Error); ok {
		if e, ok := e.Err.(*net.OpError); ok {
			return e.Op == "proxyconnect"
		}
	}
	return false
}
