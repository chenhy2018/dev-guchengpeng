package nginxproxy

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/qiniu/xlog.v1"
	"github.com/stretchr/testify/assert"
	"qbox.us/mockacc"
	"qbox.us/proxy/api.v2/proto"
)

func copyHeaders(dst, src http.Header) {
	for k, _ := range dst {
		dst.Del(k)
	}
	for k, vs := range src {
		for _, v := range vs {
			dst.Add(k, v)
		}
	}
}
func TestHttpPosts(t *testing.T) {
	xl := xlog.NewWith("TestHttpPosts")

	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

		assert.Equal(t, req.Host, "www.qiniu.com")
		querys := strings.Split(req.URL.Path[1:], "/")
		code, _ := strconv.Atoi(querys[0])
		w.WriteHeader(code)
		b, _ := ioutil.ReadAll(req.Body)
		w.Write(b)
	}))
	defer svr.Close()
	url := svr.URL

	closedSvr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(200)
	}))
	closedUrl := closedSvr.URL
	closedSvr.Close()

	odd := 0
	oddSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if odd == 1 {
			w.WriteHeader(200)
			b, _ := ioutil.ReadAll(req.Body)
			w.Write(b)
		} else {
			w.WriteHeader(500)
		}
		odd = 1 - odd
	}))
	oddUrl := oddSrv.URL

	proxySvr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

		req.RequestURI = ""
		xl := xlog.New(w, req)
		if urlHost := req.Header.Get("X-Proxy-For"); urlHost != "" {
			req.URL.Host = urlHost
		}
		xl.Info(req.URL.String())
		client := &http.Client{
			Transport: &http.Transport{
				ResponseHeaderTimeout: time.Second,
			},
			Timeout: time.Second,
		}
		resp, err := client.Do(req)
		if err != nil {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()
		copyHeaders(w.Header(), resp.Header)
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
	}))
	defer proxySvr.Close()
	proxyURL := proxySvr.URL

	access := "4_odedBxmrAHiu4Y0Qp0HPG0NANCf6VAsAjWL_k9"

	ins := NewCallbackInstance([]string{proxyURL}, time.Second, 1, mockacc.Instance)

	goodURLs := [][]string{
		[]string{url + "/200/"},
		[]string{url + "/200/", closedUrl},
		[]string{url + "/500/", url + "/400/"},
		[]string{url + "/504/", closedUrl, url + "/200/"},
		[]string{url + "/504/", closedUrl, url + "/401/"},
		[]string{url + "/504/", "http://1.1.1.1:7664/", url + "/401/"},
		[]string{oddUrl},
	}
	codes := []int{200, 200, 400, 200, 401, 401, 200}
	for i, URLs := range goodURLs {
		buf := bytes.NewBuffer(nil)
		resp, err := ins.Callback(xl, URLs, "www.qiniu.com", "", "abc", &proto.CallbackConfig{Uid: 0, AccessKey: access})
		assert.NoError(t, err, "%v", URLs)
		assert.Equal(t, codes[i], resp.StatusCode, "%v", URLs)
		io.Copy(buf, resp.Body)
		assert.Equal(t, "abc", string(buf.Bytes()), "%v", URLs)
	}

	resp, err := ins.Callback(xl, []string{closedUrl}, "www.qiniu.com", "", "abc", &proto.CallbackConfig{Uid: 0, AccessKey: access})
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadGateway, resp.StatusCode)
}
