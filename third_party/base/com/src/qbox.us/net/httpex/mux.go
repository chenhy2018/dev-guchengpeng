package httpex

import (
	"net/http"
	"path"
	"github.com/qiniu/log.v1"
	"reflect"
	"strings"
	"sync"
)

// ---------------------------------------------------------------------------
// type ServeMux

type ServeMux struct {
	mu sync.RWMutex
	m  map[string]muxEntry
}

type muxEntry struct {
	fn       reflect.Value
	explicit bool
}

// NewServeMux allocates and returns a new ServeMux.
func NewServeMux() *ServeMux {
	return &ServeMux{m: make(map[string]muxEntry)}
}

// Does path match pattern?
func pathMatch(pattern, path string) bool {
	if len(pattern) == 0 {
		// should not happen
		return false
	}
	n := len(pattern)
	if pattern[n-1] != '/' {
		return pattern == path
	}
	return len(path) >= n && path[0:n] == pattern
}

// Return the canonical path for p, eliminating . and .. elements.
func cleanPath(p string) string {
	if p == "" {
		return "/"
	}
	if p[0] != '/' {
		p = "/" + p
	}
	np := path.Clean(p)
	// path.Clean removes trailing slash except for root;
	// put the trailing slash back if necessary.
	if p[len(p)-1] == '/' && np != "/" {
		np += "/"
	}
	return np
}

// Find a handler on a handler map given a path string
// Most-specific (longest) pattern wins
func (mux *ServeMux) match(path string) reflect.Value {
	var fn reflect.Value
	var n = 0
	for k, v := range mux.m {
		if !pathMatch(k, path) {
			continue
		}
		if len(k) > n {
			n = len(k)
			fn = v.fn
		}
	}
	return fn
}

func (mux *ServeMux) handler(r *http.Request) reflect.Value {
	mux.mu.RLock()
	defer mux.mu.RUnlock()

	// Host-specific pattern takes precedence over generic ones
	fn := mux.match(r.Host + r.URL.Path)
	if !fn.IsValid() {
		fn = mux.match(r.URL.Path)
	}
	return fn
}

func (mux *ServeMux) Dispatch(rcvr reflect.Value, w http.ResponseWriter, r *http.Request, env interface{}) bool {
	fn := mux.handler(r)
	if fn.IsValid() {
		w1 := reflect.ValueOf(w)
		req1 := reflect.ValueOf(r)
		env1 := reflect.ValueOf(env)
		fn.Call([]reflect.Value{rcvr, w1, req1, env1})
		return true
	}
	return false
}

func (mux *ServeMux) Handle(pattern string, handler reflect.Value) {
	mux.mu.Lock()
	defer mux.mu.Unlock()

	if pattern == "" {
		panic("http: invalid pattern " + pattern)
	}
	if handler.Kind() != reflect.Func {
		panic("http: nil handler")
	}
	if mux.m[pattern].explicit {
		panic("http: multiple registrations for " + pattern)
	}

	mux.m[pattern] = muxEntry{handler, true}

	// Helpful behavior:
	// If pattern is /tree/, insert an implicit permanent redirect for /tree.
	// It can be overridden by an explicit registration.
	n := len(pattern)
	if pattern[n-1] == '/' {
		pattern1 := pattern[:n-1]
		if !mux.m[pattern1].explicit {
			mux.m[pattern1] = muxEntry{handler, false}
		}
	}
}

// ---------------------------------------------------------------------------
// type Router

type Router struct {
	NamePrefix string
	Style      byte
}

// Precompute the reflect type for http.ResponseWriter. Can't use http.ResponseWriter directly
// because Typeof takes an empty interface value. This is annoying.
var unusedResponseWriter *http.ResponseWriter
var unusedRequest *http.Request
var typeOfHttpResponseWriter = reflect.TypeOf(unusedResponseWriter).Elem()
var typeOfHttpRequest = reflect.TypeOf(unusedRequest)

func Register(mux *ServeMux, rcvr interface{}, envType reflect.Type) error {

	return doRegister(mux, rcvr, envType, "Do", '-')
}

func (r *Router) Register(mux *ServeMux, rcvr interface{}, envType reflect.Type) error {

	return doRegister(mux, rcvr, envType, r.NamePrefix, r.Style)
}

func doRegister(mux *ServeMux, rcvr interface{}, envType reflect.Type, namePrefix string, sep byte) error {

	if namePrefix == "" {
		namePrefix = "Do"
	}

	if sep == 0 {
		sep = '-'
	}

	typ := reflect.TypeOf(rcvr)

	// Install the methods
	for m := 0; m < typ.NumMethod(); m++ {
		method := typ.Method(m)
		mname := method.Name
		if !strings.HasPrefix(mname, namePrefix) {
			continue
		}
		mtype := method.Type
		narg := mtype.NumIn()
		// Method spec:
		//  (rcvr *XXXX) DoYYYY(w http.ResponseWriter, req *http.Request, env EnvType)
		if mtype.NumOut() != 0 || narg != 4 {
			log.Debug("method", mname, "has wrong number arguments or return values:", narg, mtype.NumOut())
			continue
		}
		// First arg muste be http.ResponseWriter
		if wType := mtype.In(1); wType != typeOfHttpResponseWriter {
			log.Debug("method", mname, "first argument type not http.ResponseWriter:", wType)
			continue
		}
		// Second arg must be *http.Request
		if reqType := mtype.In(2); reqType != typeOfHttpRequest {
			log.Debug("method", mname, "second arguement type not *http.Request:", reqType)
			continue
		}
		if thirdType := mtype.In(3); thirdType != envType {
			log.Debug("method", mname, "third arguement type error:", thirdType)
			continue
		}
		pattern := Pattern(mname[len(namePrefix):], sep)
		log.Debug("Install", pattern, "=>", mname)
		mux.Handle(pattern, method.Func)
	}

	return nil
}

func Pattern(method string, sep byte) string {

	var c byte
	route := make([]byte, 0, len(method)+8)
	for i := 0; i < len(method); i++ {
		c = method[i]
		if sep != '/' && c >= 'A' && c <= 'Z' {
			route = append(route, '/')
			c += ('a' - 'A')
		} else if c == '_' {
			c = sep
		}
		route = append(route, c)
	}
	if c == sep {
		route[len(route)-1] = '/'
	}
	return string(route)
}

// ---------------------------------------------------------------------------
