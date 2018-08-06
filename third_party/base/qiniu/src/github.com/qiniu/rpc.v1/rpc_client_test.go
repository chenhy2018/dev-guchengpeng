package rpc_test

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"crypto/rand"

	"github.com/qiniu/http/httputil.v1"
	"github.com/qiniu/io/crc32util"
	"github.com/qiniu/rpc.v1"
	"github.com/qiniu/xlog.v1"
	"github.com/stretchr/testify/assert"
)

var userAgentTst string

func foo(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	httputil.Reply(w, 200, map[string]interface{}{
		"info":  "Call method foo",
		"url":   req.RequestURI,
		"query": req.Form,
	})
}

func agent(w http.ResponseWriter, req *http.Request) {

	userAgentTst = req.Header.Get("User-Agent")
}

type Object struct {
}

func (p *Object) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	req2, _ := ioutil.ReadAll(req.Body)
	httputil.Reply(w, 200, map[string]interface{}{"info": "Call method object", "req": string(req2)})
}

var done = make(chan bool)

func server(t *testing.T) *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/foo", foo)
	mux.Handle("object", new(Object))
	return httptest.NewServer(mux)
}

func TestCall(t *testing.T) {
	s := server(t)
	defer s.Close()

	//param := "http:**localhost:8888*abc:def,g;+&$=foo*~!*~!"
	r := map[string]interface{}{}
	c := rpc.DefaultClient
	c.GetCall(nil, &r, s.URL+"/foo")
	assert.Equal(t, r, map[string]interface{}{"info": "Call method foo", "query": map[string]interface{}{}, "url": "/foo"})

	c.GetCallWithForm(nil, &r, s.URL+"/foo", map[string][]string{"a": {"1"}})
	assert.Equal(t, r["url"], "/foo?a=1")

	c.GetCallWithForm(nil, &r, s.URL+"/foo?b=2", map[string][]string{"a": {"1"}})
	assert.Equal(t, r["url"], "/foo?b=2&a=1")

	c.GetCallWithForm(nil, &r, s.URL+"/foo?", map[string][]string{"a": {"1"}})
	assert.Equal(t, r["url"], "/foo?&a=1")
}

func TestDo(t *testing.T) {

	svr := httptest.NewServer(http.HandlerFunc(agent))
	defer svr.Close()

	svrUrl := svr.URL
	c := rpc.DefaultClient
	{
		req, _ := http.NewRequest("GET", svrUrl+"/agent", nil)
		c.Do(nil, req)
		assert.Equal(t, userAgentTst, "Golang qiniu/rpc package")
	}
	{
		req, _ := http.NewRequest("GET", svrUrl+"/agent", nil)
		req.Header.Set("User-Agent", "tst")
		c.Do(nil, req)
		assert.Equal(t, userAgentTst, "tst")
	}
}

func TestResponseError(t *testing.T) {

	fmtStr := "{\"error\":\"test error info\"}"
	http.HandleFunc("/ct1", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(599)
		w.Write([]byte(fmt.Sprintf(fmtStr)))
	}))
	http.HandleFunc("/ct2", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(599)
		w.Write([]byte(fmt.Sprintf(fmtStr)))
	}))
	http.HandleFunc("/ct3", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", " application/json ; charset=utf-8")
		w.WriteHeader(599)
		w.Write([]byte(fmt.Sprintf(fmtStr)))
	}))
	ts := httptest.NewServer(nil)
	defer ts.Close()

	resp, _ := http.Get(ts.URL + "/ct1")
	assert.Equal(t, "test error info", rpc.ResponseError(resp).Error())
	resp, _ = http.Get(ts.URL + "/ct2")
	assert.Equal(t, "test error info", rpc.ResponseError(resp).Error())
	resp, _ = http.Get(ts.URL + "/ct3")
	assert.Equal(t, "test error info", rpc.ResponseError(resp).Error())
}

// 1. 正常情况
// 2. contentlength =0
// 3. contentlength =-1
// 4. 错误的host url
// 5. 数据读了一半，然后出错。
func TestClient_CallWithCrcEncoded(t *testing.T) {
	fmt.Println("TestClient_CallWithCrcEncoded")
	ceSever := httptest.NewServer(http.HandlerFunc(crcUpload))
	defer ceSever.Close()
	ceClient := rpc.DefaultClient

	r := map[string]interface{}{}
	xl := xlog.NewWith("CallWithCrcEncoded")
	url := ceSever.URL
	bodyType := "application/octet-stream"
	var body io.Reader
	var contentLength int64

	body, contentLength = randBody(512)
	ceClient.CallAfterCrcEncoded(xl, &r, url, bodyType, body, contentLength)
	assertRet(t, crc32util.EncodeSize(512), crc32util.EncodeSize(512), true, nil, r)

	// 运行结果可能和go版本相关
	body, contentLength = randBody(512)
	ceClient.CallAfterCrcEncoded(xl, &r, url, bodyType, body, 0)
	assertRet(t, -1, crc32util.EncodeSize(512), true, nil, r)

	body, contentLength = randBody(512)
	ceClient.CallAfterCrcEncoded(xl, &r, url, bodyType, body, -1)
	assertRet(t, -1, crc32util.EncodeSize(512), true, nil, r)

	ceClient.CallAfterCrcEncoded(xl, &r, url, bodyType, nil, 0)
	assertRet(t, 0, 0, true, nil, r)

	body, contentLength = randBody(512)
	err := ceClient.CallAfterCrcEncoded(xl, &r, "http://localhost", bodyType, body, contentLength)
	assert.Error(t, err) // host not found
	// reader can reuse
	err = ceClient.CallAfterCrcEncoded(xl, &r, url, bodyType, body, contentLength)
	assert.NoError(t, err)
	assertRet(t, crc32util.EncodeSize(512), crc32util.EncodeSize(512), true, nil, r)

	body, contentLength = randBody(512)
	err = ceClient.CallAfterCrcEncoded(xl, &r, url, bodyType, body, 256)
	assert.Error(t, err) // err body size
}

// 测试返回 err 的情况
func TestClient_DoWithCrcCheck(t *testing.T) {
	errServer := httptest.NewServer(http.HandlerFunc(errHandler))
	defer errServer.Close()
	ceClient := rpc.DefaultClient
	url := errServer.URL

	xl := xlog.NewDummy()
	req, _ := http.NewRequest("POST", url+"s", nil)
	req.URL = nil
	resp, err := ceClient.DoWithCrcCheck(xl, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	fmt.Println(err)
}

// 测试服务端没有升级
func TestClient_CrcEncoded_OldServer(t *testing.T) {
	oldServer := httptest.NewServer(http.HandlerFunc(oldCrcUpload))
	defer oldServer.Close()
	ceClient := rpc.DefaultClient
	url := oldServer.URL

	xl := xlog.NewDummy()
	r := map[string]interface{}{}
	bodyType := "application/octet-stream"
	var body io.Reader
	var contentLength int64

	body, contentLength = randBody(512)
	err := ceClient.CallWith64(xl, &r, url, bodyType, body, contentLength)
	assert.NoError(t, err)
	assertOldRet(t, 512, 512, nil, r)

	// 服务端没有升级报501错误
	body, contentLength = randBody(512)
	err = ceClient.CallAfterCrcEncoded(xl, &r, url, bodyType, body, contentLength)
	assert.Equal(t, http.StatusNotImplemented, err.(*rpc.ErrorInfo).Code)

}

func assertOldRet(t *testing.T, contentLength, bodyLength int64, err error, r map[string]interface{}) {
	assert.Equal(t, contentLength, r["content length"])
	assert.Equal(t, bodyLength, r["body length"])
	assert.Equal(t, nil, r["crc encoded"])
	assert.Equal(t, err, r["err"])
}

func assertRet(t *testing.T, contentLength, bodyLength int64, crcEncoded bool,
	err error, r map[string]interface{}) {
	assert.Equal(t, contentLength, r["content length"])
	assert.Equal(t, bodyLength, r["body length"])
	assert.Equal(t, crcEncoded, r["crc encoded"])
	assert.Equal(t, err, r["err"])
}

func errHandler(w http.ResponseWriter, req *http.Request) {
	httputil.ReplyWithCode(w, 404)
}

func crcUpload(w http.ResponseWriter, req *http.Request) {
	crced := (req.Header.Get(rpc.CrcEncodedHeader) != "")
	if crced {
		w.Header().Set(rpc.AckCrcEncodedHeader, "1")
	}
	clength := req.ContentLength
	var blength int64
	var err error
	if req.Body == nil {
		blength = 0
	} else {
		blength, err = io.Copy(ioutil.Discard, req.Body)
	}

	httputil.Reply(w, 200, map[string]interface{}{
		"content length": clength,
		"body length":    blength,
		"crc encoded":    crced,
		"err":            err,
	})
}

func oldCrcUpload(w http.ResponseWriter, req *http.Request) {
	clength := req.ContentLength
	var blength int64
	var err error
	if req.Body == nil {
		blength = 0
	} else {
		blength, err = io.Copy(ioutil.Discard, req.Body)
	}

	httputil.Reply(w, 200, map[string]interface{}{
		"content length": clength,
		"body length":    blength,
		"err":            err,
	})
}

func randBody(size int) (body io.Reader, bodyLength int64) {
	b := make([]byte, size)
	size2, _ := io.ReadFull(rand.Reader, b[:])
	bodyLength = int64(size2)
	body = bytes.NewReader(b)
	return
}
