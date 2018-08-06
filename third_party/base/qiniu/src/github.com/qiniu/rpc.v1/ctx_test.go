// +build go1.7

package rpc_test

import (
	"context"
	"crypto/rand"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/qiniu/rpc.v1"
	"github.com/qiniu/xlog.v1"
	"github.com/stretchr/testify/assert"
)

// 测试 l 中带有上游的 ctx
// 测试先关闭连接，然后再读取response.body
func TestClient_DoCtx(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(take50msHandler))
	defer func() {
		svr.Close()
	}()
	url := svr.URL
	rpcClient := &rpc.Client{http.DefaultClient}

	data := []struct {
		hasCtx bool // xlog 是否带有 ctx 信息
		cancel int  // 上游是否取消了请求, 0:不取消；1：开始的时候取消；2：请求过程中取消。
		err    bool
	}{
		{false, 0, false},
		{false, 1, false},
		{true, 0, false},
		{true, 1, true},
		{true, 2, true},
	}

	for i, d := range data {
		index := strconv.Itoa(i)
		ctx, cancel := context.WithCancel(context.Background())
		var xl *xlog.Logger
		if d.hasCtx {
			xl = xlog.NewDummyWithCtx(ctx)
		} else {
			xl = xlog.NewDummy()
		}
		switch d.cancel {
		case 0:
		case 1:
			cancel()
		case 2:
			go func() {
				time.Sleep(time.Millisecond * 30)
				cancel()
			}()
		}

		req, err := http.NewRequest("POST", url, nil)
		assert.NoError(t, err, index)
		resp, err := rpcClient.Do(xl, req)

		if d.err {
			if d.cancel == 1 {
				assert.Error(t, err)
				errInfo, ok := err.(httpCode)
				assert.True(t, ok)
				assert.Equal(t, 499, errInfo.HttpCode())
				assert.True(t, strings.Contains(errInfo.Error(), "context canceled"))
			} else {
				// 读取一半时取消
				assert.Equal(t, 2, d.cancel, index)
				assert.NoError(t, err, index)
				_, err = ioutil.ReadAll(resp.Body)
				assert.Error(t, err, index)
				errInfo, ok := err.(httpCode)
				assert.True(t, ok, index)
				assert.Equal(t, 499, errInfo.HttpCode(), index)
				assert.True(t, strings.Contains(errInfo.Error(), "context canceled"), index)
			}
		} else {
			assert.NoError(t, err, index)
			_, err1 := ioutil.ReadAll(resp.Body)
			resp.Body.Close()
			assert.NoError(t, err1, index)
		}
	}
}

// 测试一些特殊情况
func TestClient_DoCtx_Special(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(simpleHandler))
	defer func() {
		svr.Close()
	}()
	url := svr.URL
	rpcClient := &rpc.Client{http.DefaultClient}

	// 先cancel upCtx，然后再关闭response，关闭不应该出问题。
	req, err := http.NewRequest("GET", url, nil)
	assert.NoError(t, err)
	ctx, cancel := context.WithCancel(context.Background())
	xl := xlog.NewDummyWithCtx(ctx)
	resp, err := rpcClient.Do(xl, req)
	assert.NoError(t, err)
	cancel()
	err = resp.Body.Close()
	assert.NoError(t, err)

	// 读完数据之后，cancel upCtx，再读时应该返回EOF，而不是rpc.EOF
	req, err = http.NewRequest("GET", url, nil)
	assert.NoError(t, err)
	ctx, cancel = context.WithCancel(context.Background())
	xl = xlog.NewDummyWithCtx(ctx)
	resp, err = rpcClient.Do(xl, req)
	assert.NoError(t, err)
	_, err = io.Copy(ioutil.Discard, resp.Body)
	cancel()
	b := make([]byte, 1)
	n, err := resp.Body.Read(b)
	assert.Equal(t, 0, n)
	assert.True(t, err == io.EOF)
}

// 测试 req 重试
func TestClient_DoCtx_Retry(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(simpleHandler))
	defer svr.Close()

	url := svr.URL
	req, err := http.NewRequest("GET", url, nil)
	assert.NoError(t, err)
	rpcClient := &rpc.Client{http.DefaultClient}
	xl := xlog.NewDummy()

	resp, err := rpcClient.Do(xl, req)
	assert.NoError(t, err)
	io.Copy(ioutil.Discard, resp.Body)
	resp.Body.Close()

	resp, err = rpcClient.Do(xl, req)
	assert.NoError(t, err)
	io.Copy(ioutil.Discard, resp.Body)
	resp.Body.Close()
}

type Transport struct {
	tr http.RoundTripper
}

type WrapError struct {
	error
}

func (t *Transport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	resp, err = t.tr.RoundTrip(req)
	if err != nil {
		err = WrapError{err}
	}
	return
}

func assertHttpError(t *testing.T, err error) {
	assert.Error(t, err)
	httpError, ok := err.(httpCode)
	assert.True(t, ok)
	assert.Equal(t, 499, httpError.HttpCode())
}

// 如果client.do返回的context canceled被包装了，doCtx还是能够返回499
// see https://jira.qiniu.io/browse/KODO-3741
func TestClient_DoCtx_WrapError(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(take50msHandler))
	defer svr.Close()
	httpClient := &http.Client{Transport: &Transport{http.DefaultTransport}}
	rpcClient := &rpc.Client{httpClient}
	url := svr.URL

	// cancel at first
	req, err := http.NewRequest("GET", url, nil)
	assert.NoError(t, err)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	xl := xlog.NewDummyWithCtx(ctx)
	_, err = rpcClient.Do(xl, req)
	assertHttpError(t, err)

	// cancel at middle
	req, err = http.NewRequest("GET", url, nil)
	assert.NoError(t, err)
	ctx, cancel = context.WithCancel(context.Background())
	xl = xlog.NewDummyWithCtx(ctx)
	go func() {
		time.Sleep(time.Millisecond * 30)
		cancel()
	}()
	resp, err := rpcClient.Do(xl, req)
	assert.NoError(t, err)
	_, err = io.Copy(ioutil.Discard, resp.Body)
	assertHttpError(t, err)
}

type httpCode interface {
	Error() string
	HttpCode() int
}

// 响应时间50ms，每10ms 发送一次数据
func take50msHandler(w http.ResponseWriter, req *http.Request) {
	data := make([]byte, 64*1024)
	io.ReadFull(rand.Reader, data)
	for i := 0; i < 5; i++ {
		time.Sleep(time.Millisecond * 10)
		w.Write(data)
	}
}

func simpleHandler(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte("hello"))
}
