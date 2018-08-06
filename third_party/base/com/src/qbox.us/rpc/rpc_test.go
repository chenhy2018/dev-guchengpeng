package rpc

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/qiniu/http/httputil.v1"
	"github.com/qiniu/log.v1"
	"github.com/stretchr/testify/assert"
)

func init() {
	log.SetOutputLevel(0)
}

func hello2(w ResponseWriter, req Request) {
	fmt.Fprintln(w, "Call method hello")
	fmt.Fprintln(w, req.Query)
}

func foo2(w ResponseWriter, req Request) {
	w.Reply(map[string]interface{}{"info": "Call method foo", "query": req.Query})
}

type Object2 struct {
}

func (p *Object2) ServeHTTP(w ResponseWriter, req Request) {
	w.Reply(map[string]interface{}{"info": "Call method object", "req": req})
}

func _TestRpcServer(t *testing.T) {
	HandleFunc("hello", hello2)
	HandleFunc("foo", foo2)
	Handle("object", new(Object2))
	go func() {
		err := Run(":8678", nil)
		if err != nil {
			log.Fatal("ListenAndServe: ", err)
		}
	}()
}

func mockReplyRange(w ResponseWriter, r *http.Request) {
	meta := &Metas{
		ETag: "abcdefghijklmn",
	}
	w.ReplyRange(nil, 0, meta, r)
}

func mockReplyFile(w ResponseWriter, r *http.Request) {
	meta := &Metas{
		ETag: "abcdefghijklmn",
	}
	w.ReplyFile(nil, 0, meta, r)
}

func TestGetEtag(t *testing.T) {

	testcase := map[string]string{
		"\"ljaa\"": "ljaa",
		"ljaa\"":   "ljaa\"",
		"\"ljaa":   "\"ljaa",
		"\"":       "\"",
		"\"\"":     "",
		"\"a":      "\"a",
		"a\"":      "a\"",
		"aa":       "aa",
		"":         "",
	}

	for k, v := range testcase {
		assert.Equal(t, getETag(k), v)
	}
}

func TestAddEtag(t *testing.T) {

	mux := http.NewServeMux()
	mux.HandleFunc("/etag/replyrange", func(w http.ResponseWriter, r *http.Request) {
		rw := ResponseWriter{w}
		mockReplyRange(rw, r)
	})
	mux.HandleFunc("/etag/replyfile", func(w http.ResponseWriter, r *http.Request) {
		rw := ResponseWriter{w}
		mockReplyFile(rw, r)
	})
	svr := httptest.NewServer(mux)
	defer svr.Close()
	svrUrl := svr.URL

	req1, err := http.NewRequest("GET", svrUrl+"/etag/replyrange", nil)
	if err != nil {
		t.Fatal(err)
	}

	req2, err := http.NewRequest("GET", svrUrl+"/etag/replyfile", nil)
	if err != nil {
		t.Fatal(err)
	}

	req1.Header.Set("If-None-Match", `abcdefghijklmn`)   // without quote
	req2.Header.Set("If-None-Match", `"abcdefghijklmn"`) // with quote

	client := &http.Client{}
	resp1, err := client.Do(req1)
	if err != nil {
		t.Fatal(err)
	}
	resp2, err := client.Do(req2)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 304, resp1.StatusCode)
	assert.Equal(t, 304, resp2.StatusCode)
	assert.Equal(t, `"abcdefghijklmn"`, resp1.Header.Get("Etag"))
	assert.Equal(t, `"abcdefghijklmn"`, resp2.Header.Get("ETag"))
}

type mockReader struct {
}

func (mockReader) RangeRead(w io.Writer, from, to int64) (err error) {
	return httputil.NewError(573, "req is out of quato")
}

func TestHeaderResponseWriter(t *testing.T) {

	mux := http.NewServeMux()

	mux.HandleFunc("/replyrange", func(w http.ResponseWriter, r *http.Request) {
		rw := ResponseWriter{w}
		meta := &Metas{
			ETag: "abcdefghijklmn",
		}
		rw.ReplyRange(mockReader{}, 0, meta, r)
	})
	svr := httptest.NewServer(mux)
	defer svr.Close()
	svrUrl := svr.URL

	req1, err := http.NewRequest("GET", svrUrl+"/replyrange", nil)
	if err != nil {
		t.Fatal(err)
	}

	req2, err := http.NewRequest("HEAD", svrUrl+"/replyrange", nil)
	if err != nil {
		t.Fatal(err)
	}

	client := &http.Client{}
	resp1, err := client.Do(req1)
	if err != nil {
		t.Fatal(err)
	}
	resp2, err := client.Do(req2)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 200, resp1.StatusCode)
	assert.Equal(t, "573", resp1.Header.Get("X-KODO-RESPCODE"))
	assert.Equal(t, 200, resp2.StatusCode)
}

func TestHandle304(t *testing.T) {
	meta := &Metas{
		ETag:         "v1",
		LastModified: time.Now(),
		Expires:      time.Now().String(),
		CacheControl: "max-age=0",
	}
	{
		req := &http.Request{
			Header: make(http.Header),
		}
		w := httptest.NewRecorder()
		assert.False(t, Handle304(w, meta, req))
	}
	{
		req := &http.Request{
			Header: make(http.Header),
		}
		req.Header.Add("If-None-Match", `"v1"`)
		w := httptest.NewRecorder()
		assert.True(t, Handle304(w, meta, req))
		assert.Equal(t, meta.ETag, getETag(w.Header().Get("ETag")))
		assert.Equal(t, meta.Expires, w.Header().Get("Expires"))
		assert.Equal(t, meta.CacheControl, w.Header().Get("Cache-Control"))
	}
	{
		req := &http.Request{
			Header: make(http.Header),
		}
		req.Header.Add("If-None-Match", `v1`)
		w := httptest.NewRecorder()
		assert.True(t, Handle304(w, meta, req))
		assert.Equal(t, meta.ETag, getETag(w.Header().Get("ETag")))
		assert.Equal(t, meta.Expires, w.Header().Get("Expires"))
		assert.Equal(t, meta.CacheControl, w.Header().Get("Cache-Control"))
	}
	{
		req := &http.Request{
			Header: make(http.Header),
		}
		req.Header.Add("If-None-Match", `"v2"`)
		w := httptest.NewRecorder()
		assert.False(t, Handle304(w, meta, req))
	}
	{
		req := &http.Request{
			Header: make(http.Header),
		}
		req.Header.Add("If-Modified-Since", meta.LastModified.UTC().Format(http.TimeFormat))
		w := httptest.NewRecorder()
		assert.True(t, Handle304(w, meta, req))
		assert.Equal(t, meta.Expires, w.Header().Get("Expires"))
		assert.Equal(t, meta.CacheControl, w.Header().Get("Cache-Control"))
	}
	{
		req := &http.Request{
			Header: make(http.Header),
		}
		req.Header.Add("If-Modified-Since", meta.LastModified.Add(time.Second).UTC().Format(http.TimeFormat))
		w := httptest.NewRecorder()
		assert.True(t, Handle304(w, meta, req))
		assert.Equal(t, meta.Expires, w.Header().Get("Expires"))
		assert.Equal(t, meta.CacheControl, w.Header().Get("Cache-Control"))
	}
	{
		req := &http.Request{
			Header: make(http.Header),
		}
		req.Header.Add("If-Modified-Since", meta.LastModified.Add(-time.Second).UTC().Format(http.TimeFormat))
		w := httptest.NewRecorder()
		assert.False(t, Handle304(w, meta, req))
	}
	{
		req := &http.Request{
			Header: make(http.Header),
		}
		req.Header.Add("If-None-Match", `"v1"`)
		req.Header.Add("If-Modified-Since", meta.LastModified.UTC().Format(http.TimeFormat))
		w := httptest.NewRecorder()
		assert.True(t, Handle304(w, meta, req))
		assert.Equal(t, meta.Expires, w.Header().Get("Expires"))
		assert.Equal(t, meta.CacheControl, w.Header().Get("Cache-Control"))
	}
	{
		req := &http.Request{
			Header: make(http.Header),
		}
		req.Header.Add("If-None-Match", `"v1"`)
		req.Header.Add("If-Modified-Since", meta.LastModified.Add(-time.Second).UTC().Format(http.TimeFormat))
		w := httptest.NewRecorder()
		assert.False(t, Handle304(w, meta, req))
	}
	{
		req := &http.Request{
			Header: make(http.Header),
		}
		req.Header.Add("If-None-Match", `"v2"`)
		req.Header.Add("If-Modified-Since", meta.LastModified.UTC().Format(http.TimeFormat))
		w := httptest.NewRecorder()
		assert.False(t, Handle304(w, meta, req))
	}
}
