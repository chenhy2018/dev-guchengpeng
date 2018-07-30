package fopg

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	HeaderShouldReturn = "X-Header-Return"
	HeaderWaitMs       = "X-Header-Wait-Ms"
	Contents           = "from target server"
)

func TestClientNotUseProxy(t *testing.T) {
	assert := assert.New(t)

	targetServer := newTargetServer()
	defer targetServer.Close()

	proxyServer := newInvalidProxyServer()
	defer proxyServer.Close()

	failoverCfg := &FailoverConfig{}
	var clientTr http.RoundTripper = &http.Transport{}
	var failoverTr http.RoundTripper = &http.Transport{
		Proxy: func(req *http.Request) (*url.URL, error) {
			return url.Parse(proxyServer.URL)
		},
	}
	failoverClient := NewFailoverClient(failoverCfg, clientTr, failoverTr, nil)

	// 返回正确结果
	reqObj, _ := http.NewRequest("GET", targetServer.URL, nil)
	req := &FailoverRequest{
		req: reqObj,
	}
	resp, err := failoverClient.DoCtx(req)
	assert.NoError(err)
	defer resp.Body.Close()

	contentBytes, err := ioutil.ReadAll(resp.Body)
	assert.NoError(err)
	assert.Equal(Contents, strings.TrimSpace(string(contentBytes)))

	// 目标服务器返回 500，不 failover
	req.req.Header.Add(HeaderShouldReturn, "500")
	resp, err = failoverClient.DoCtx(req)
	assert.NoError(err)
	assert.Equal(500, resp.StatusCode)
}

func TestClientUseProxy(t *testing.T) {
	assert := assert.New(t)

	targetServer := newTargetServer()
	defer targetServer.Close()

	validProxyServer := newValidProxyServer()
	defer validProxyServer.Close()

	failoverCfg := &FailoverConfig{}
	var clientTr http.RoundTripper = &http.Transport{}
	var failoverTr http.RoundTripper = &http.Transport{
		Proxy: func(req *http.Request) (*url.URL, error) {
			return url.Parse(validProxyServer.URL)
		},
	}
	failoverClient := NewFailoverClient(failoverCfg, clientTr, failoverTr, nil)

	reqObj, _ := http.NewRequest("GET", targetServer.URL, nil)
	reqObj.Header.Add(HeaderShouldReturn, "503")
	req := &FailoverRequest{
		req: reqObj,
	}
	resp, err := failoverClient.DoCtx(req)
	assert.NoError(err)
	defer resp.Body.Close()

	contentBytes, err := ioutil.ReadAll(resp.Body)
	assert.NoError(err)
	assert.Equal(Contents, strings.TrimSpace(string(contentBytes)))
}

func TestClientTimeout(t *testing.T) {
	assert := assert.New(t)

	targetServer := newTargetServer()
	defer targetServer.Close()

	validProxyServer := newValidProxyServer()
	defer validProxyServer.Close()

	failoverCfg := &FailoverConfig{
		ClientTimeoutMS: 1000,
	}
	var clientTr http.RoundTripper = &http.Transport{}
	var failoverTr http.RoundTripper = &http.Transport{
		Proxy: func(req *http.Request) (*url.URL, error) {
			return url.Parse(validProxyServer.URL)
		},
	}
	failoverClient := NewFailoverClient(failoverCfg, clientTr, failoverTr, nil)

	reqObj, _ := http.NewRequest("GET", targetServer.URL, nil)
	reqObj.Header.Add(HeaderWaitMs, "2000") // 注意这个时间不要设太长
	req := &FailoverRequest{
		req: reqObj,
	}
	resp, err := failoverClient.DoCtx(req)
	assert.NoError(err)
	defer resp.Body.Close()

	contentBytes, err := ioutil.ReadAll(resp.Body)
	assert.NoError(err)
	assert.Equal(Contents, strings.TrimSpace(string(contentBytes)))
}

func newTargetServer() (svr *httptest.Server) {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Header.Get(HeaderShouldReturn) {
		case "500":
			w.WriteHeader(500)
			return
		case "503":
			w.WriteHeader(503)
			return
		}

		waitMs := r.Header.Get(HeaderWaitMs)
		if waitMs != "" {
			waitMsInt, _ := strconv.Atoi(waitMs)
			time.Sleep(time.Duration(waitMsInt) * time.Millisecond)
		}

		fmt.Fprintln(w, Contents)
	}))
}

func newValidProxyServer() (svr *httptest.Server) {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		client := &http.Client{}
		req, _ := http.NewRequest(r.Method, r.URL.String(), nil)
		req.Header.Del(HeaderShouldReturn)
		resp, err := client.Do(req)
		if err != nil {
			w.WriteHeader(500)
			return
		}
		defer resp.Body.Close()

		io.Copy(w, resp.Body)
	}))
}

func newInvalidProxyServer() (svr *httptest.Server) {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	}))
}
