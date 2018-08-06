package logh

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"io/ioutil"

	qrpc "github.com/qiniu/rpc.v1"
	"github.com/qiniu/xlog.v1"
	"qbox.us/audit/file"
	"qbox.us/net/httputil"
	"qbox.us/servestk"
)

func ss_foo(w http.ResponseWriter, req *http.Request, next func(http.ResponseWriter, *http.Request)) {
	fmt.Println("foo")
	next(w, req)
}

func ss_bar(w http.ResponseWriter, req *http.Request, next func(http.ResponseWriter, *http.Request)) {
	fmt.Println("bar")
	next(w, req)
}

func TestServeStack(t *testing.T) {

	ss := servestk.New(nil)
	l := New("FOO", file.Stderr, nil, 512)
	ss.Push(ss_bar)
	ss.Push(ss_foo)
	ss.Push(l.Handler())
	ss.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		fmt.Println("call /")
		httputil.Reply(w, 200, map[string]string{"foo": "bar"})
	})
	// test xBody works(no panic)
	ss.HandleFunc("/xbody", func(w http.ResponseWriter, req *http.Request) {
		fmt.Println("call /xbody")
		fmt.Println("content-length:", req.ContentLength)
		req.Body = ioutil.NopCloser(req.Body)
		httputil.Reply(w, 200, map[string]string{"foo": "bar"})
	})
	ss.HandleFunc("/abc/", func(w http.ResponseWriter, req *http.Request) {
		log := xlog.New(w, req)
		req.ParseForm()
		log.Xlog(req.Form)
		log.Println("call /abc/", req.Form)
		if req.Form == nil || req.Form["a"][0] != "1" {
			t.Fatal("ParseForm failed")
		}
		panic("panic:" + req.URL.Path)
		httputil.Reply(w, 400, map[string]string{"error": "invalid arguments"})
	})

	svr := httptest.NewServer(ss)
	svrUrl := svr.URL
	defer svr.Close()

	rpc := qrpc.Client{http.DefaultClient}
	rpc.Call(nil, nil, svrUrl+"/abcd?a=1")
	rpc.CallWithForm(nil, nil, svrUrl+"/abc/def", map[string][]string{
		"a": {"1"},
		"b": {"2"},
	})

	req2Body := bytes.NewReader([]byte{1, 2})
	req2, err := http.NewRequest("POST", svrUrl+"/xbody", req2Body)
	if err != nil {
		fmt.Println(err)
	}
	req2.ContentLength = -1
	rpc.Do(nil, req2)

	longParam := strings.Repeat("x", 520)
	resp, err := rpc.PostForm(svrUrl+"/abc/def", map[string][]string{
		"p": {longParam},
	})
	if err != nil {
		t.Fatal("PostForm failed:", err)
	}
	x, ok := resp.Header[xlogKey]
	if !ok {
		t.Fatal("No X-Log found.")
	}
	fmt.Println("xlog->", x[0])
	if len(x[0]) > maxXlogLen+len("...") {
		t.Fatal("X-Log too long.")
	}
}
