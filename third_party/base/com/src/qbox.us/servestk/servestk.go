package servestk

import (
	"io"
	"io/ioutil"
	"net/http"
	"runtime/debug"

	"github.com/qiniu/log.v1"
)

// ----------------------------------------------------------

func SafeHandler(w http.ResponseWriter, req *http.Request, f func(w http.ResponseWriter, req *http.Request)) {
	defer func() {
		p := recover()
		if p != nil {
			w.WriteHeader(597)
			log.Printf("WARN: panic fired in %v.panic - %v\n", f, p)
			log.Println(string(debug.Stack()))
		}
	}()
	f(w, req)
}

// ----------------------------------------------------------

func DiscardHandler(w http.ResponseWriter, req *http.Request, f func(w http.ResponseWriter, req *http.Request)) {
	f(w, req)
	io.Copy(ioutil.Discard, req.Body)
}

// ----------------------------------------------------------

type Handler func(http.ResponseWriter, *http.Request, func(http.ResponseWriter, *http.Request))

type ServeStack struct {
	stk []Handler
}

func (r *ServeStack) Push(f ...Handler) {
	r.stk = append(r.stk, f...)
}

func (r *ServeStack) Build(h func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return build(r.stk, h)
}

func build(stk []Handler,
	h func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {

	if len(stk) == 0 {
		return h
	}
	return build(stk[1:], func(w http.ResponseWriter, req *http.Request) {
		stk[0](w, req, h)
	})
}

// ----------------------------------------------------------

type Mux interface {
	HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request))
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}

type ServeStackMux struct {
	ServeStack
	Mux
}

func New(mux Mux, f ...Handler) (r *ServeStackMux) {
	if mux == nil {
		mux = http.NewServeMux()
	}
	return &ServeStackMux{ServeStack{f}, mux}
}

func (r *ServeStackMux) HandleFuncEx(methodName string, pattern string, handler func(http.ResponseWriter, *http.Request)) {
	r.Mux.HandleFunc(pattern, r.Build(handler))
}

func (r *ServeStackMux) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	r.Mux.HandleFunc(pattern, r.Build(handler))
}

func (r *ServeStackMux) Handle(pattern string, handler http.Handler) {
	r.Mux.HandleFunc(pattern, r.Build(func(w http.ResponseWriter, req *http.Request) {
		handler.ServeHTTP(w, req)
	}))
}

// ----------------------------------------------------------
