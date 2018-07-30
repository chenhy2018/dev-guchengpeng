package logh

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"qbox.us/rpc"
	"qbox.us/servestk"
	"testing"
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

	ss := servestk.New(nil, ss_bar, ss_foo, Instance)

	ss.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		fmt.Println("call /")
		w.Write([]byte("call /"))
	})
	ss.HandleFunc("/abc/", func(w http.ResponseWriter, req *http.Request) {
		fmt.Println("call /abc/")
		panic("panic:" + req.URL.Path)
		rpc.ResponseWriter{w}.ReplyWith(400, map[string]string{"error": "invalid arguments"})
	})

	svr := httptest.NewServer(ss)
	defer svr.Close()
	svrUrl := svr.URL

	rpc.Call(nil, svrUrl+"/abcd")
	rpc.Call(nil, svrUrl+"/abc/def")
}
