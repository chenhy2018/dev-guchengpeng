package apigate

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// --------------------------------------------------------------------

type testcase struct {
	method  string
	path    string
	pattern string
	notOk   bool
}

var matchCases = []testcase{
	{
		method:  "POST",
		path:    "/bar",
		pattern: "POST /bar",
	},
	{
		method:  "GET",
		path:    "/bar/bar-param",
		pattern: "GET /bar/*",
	},
	{
		method:  "DELETE",
		path:    "/bar/bar-param/foo",
		pattern: "DELETE /bar/*/foo",
	},
	{
		method:  "POST",
		path:    "/bar/bar-param/foo/foo-param",
		pattern: "POST /bar/*/foo/*",
	},
	{
		method:  "POST",
		path:    "/bar/bar-param/foo/foo-param/a/b/c/d",
		pattern: "POST   /bar/*/foo/*/a/**",
	},
	{
		method:  "POST",
		path:    "/bar/bar-param/foo/foo-param/a/b/c/d",
		pattern: "* /bar/*/foo/*/a/**",
	},
	{
		method:  "POST",
		path:    "/bar/bar-param/foo/foo-param/a/b/c/d",
		pattern: "* /bar/**/foo/*/a/**",
		notOk:   true,
	},
	{
		method:  "POST",
		path:    "/bar/bar-param/foo/foo-param/a/b/c/d",
		pattern: "**",
		notOk:   true,
	},
	{
		method:  "POST",
		path:    "/bar",
		pattern: "POST /bar/**",
		notOk:   true,
	},
	{
		method:  "POST",
		path:    "/bar",
		pattern: "POST /**",
	},
	{
		method:  "POST",
		path:    "/",
		pattern: "POST /**",
	},
	{
		method:  "POST",
		path:    "/a/b/c/d/e",
		pattern: "POST /**",
	},
}

func TestMatch(t *testing.T) {

	for _, c := range matchCases {
		paths := strings.Split(c.path[1:], "/")
		if ok := NewPattern(c.pattern).Match(c.method, paths); ok == c.notOk {
			t.Fatal("not match =>", c)
		}
	}
}

// --------------------------------------------------------------------

type handler struct {
	resp string
}

func (p handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.Method == "HEAD" {
		w.WriteHeader(200)
		return
	}
	io.WriteString(w, p.resp)
}

type multiServeMuxCase struct {
	method string
	req    string
	resp   string
	code   int
}

func TestMultiServeMux(t *testing.T) {

	cases := []multiServeMuxCase{
		{
			method: "HEAD",
			req:    "http://api.qiniu.com:8888/rs/stat/123",
			resp:   "",
			code:   200,
		},
		{
			method: "POST",
			req:    "http://api.qiniu.com:8888/rs/stat/123",
			resp:   "rs stat",
			code:   200,
		},
		{
			method: "GET",
			req:    "http://api.qiniu.com:8888/rs/stat/123",
			resp:   "",
			code:   404,
		},
		{
			method: "POST",
			req:    "http://rs.qiniuapi.com/mkbucket/123",
			resp:   "rs mkbucket",
			code:   200,
		},
		{
			method: "POST",
			req:    "http://rs.qiniuapi.com/bucketinfo/123",
			resp:   "rs bucketinfo",
			code:   200,
		},
		{
			method: "GET",
			req:    "http://rs.qiniuapi.com/bucketinfo/123",
			resp:   "rs bucketinfo",
			code:   200,
		},
		{
			method: "POST",
			req:    "http://api.qiniu.com:8888/up/mkblk/123",
			resp:   "up mkblk",
			code:   200,
		},
		{
			method: "POST",
			req:    "http://api.qiniu.com:8888/up/mkblk",
			resp:   "upload",
			code:   200,
		},
		{
			method: "GET",
			req:    "http://api.qiniu.com:8888/os/v1/auths/123",
			resp:   "os get auth",
			code:   200,
		},
		{
			method: "POST",
			req:    "http://os.qiniuapi.com/v1/auths/123/rules",
			resp:   "os put auth rules",
			code:   200,
		},
		{
			method: "GET",
			req:    "http://default.qiniuapi.com/default/",
			resp:   "default",
			code:   200,
		},
		{
			method: "GET",
			req:    "http://default.qiniuapi.com/not_found/",
			resp:   "",
			code:   404,
		},
	}

	mux := NewMultiServeMux()

	mux.ServeMux("rs.qiniuapi.com", "api.qiniu.com:8888/rs").
		Handle("POST /stat/**", handler{"rs stat"}).
		Handle("POST /mkbucket/**", handler{"rs mkbucket"}).
		Handle("* /bucketinfo/**", handler{"rs bucketinfo"})

	mux.ServeMux("up.qiniuapi.com", "api.qiniu.com:8888/up").
		Handle("POST /mkblk/**", handler{"up mkblk"}).
		Handle("POST /**", handler{"upload"})

	mux.ServeMux("os.qiniuapi.com", "api.qiniu.com:8888/os").
		Handle("POST /v1/auths/*/rules", handler{"os put auth rules"}).
		Handle("GET /v1/auths/*", handler{"os get auth"})

	mux.ServeMux("*/default").Handle("GET /", handler{"default"})

	for _, c := range cases {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(c.method, c.req, nil)
		mux.ServeHTTP(w, req)
		if w.Code != c.code || (c.resp != "" && w.Body.String() != c.resp) {
			t.Fatal("TestMultiServeMux failed:", c, w)
		}
	}
}

// --------------------------------------------------------------------
