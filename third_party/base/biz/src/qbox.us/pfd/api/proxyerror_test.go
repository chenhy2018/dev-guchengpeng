package api

import (
	"net"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestIsProxyError(t *testing.T) {

	c := http.Client{
		Transport: &http.Transport{
			Proxy: func(*http.Request) (*url.URL, error) {
				return url.Parse("http://2.2.2.2:80")
			},
			Dial: (&net.Dialer{Timeout: 10 * time.Millisecond}).Dial,
		},
	}
	_, err := c.Get("http://www.qiniu.com")
	if assert.Error(t, err) {
		assert.True(t, isProxyError(err))
	}
}
