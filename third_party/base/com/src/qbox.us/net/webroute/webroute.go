package webroute

import (
	"net/http"
	"qbox.us/errors"
	"github.com/qiniu/log.v1"
	"reflect"
	"strings"
)

//
// TODO: 这个包已经被废除，请改用 github.com/qiniu/http/webroute
//

var ErrNoSessionManager = errors.Register("no session manager")

// ---------------------------------------------------------------------------

type Mux interface {
	Handle(pattern string, handler http.Handler)
}

type SessionManager interface {
	Get(w http.ResponseWriter, req *http.Request) (http.ResponseWriter, map[string]interface{})
}

// ---------------------------------------------------------------------------

type handler struct {
	rcvr   reflect.Value
	method reflect.Value
}

func (h *handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	w1 := reflect.ValueOf(w)
	req1 := reflect.ValueOf(req)
	h.method.Call([]reflect.Value{h.rcvr, w1, req1})
}

// ---------------------------------------------------------------------------

type handlerWithSession struct {
	rcvr       reflect.Value
	method     reflect.Value
	sessionMgr SessionManager
}

func (h *handlerWithSession) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	w2, sessions := h.sessionMgr.Get(w, req)

	w1 := reflect.ValueOf(w2)
	req1 := reflect.ValueOf(req)
	sessions1 := reflect.ValueOf(sessions)
	h.method.Call([]reflect.Value{h.rcvr, w1, req1, sessions1})
}

// ---------------------------------------------------------------------------

type Router struct {
	NamePrefix    string
	PatternPrefix string
	Mux           Mux
	Sessions      SessionManager
	Style         byte
}

// Precompute the reflect type for http.ResponseWriter. Can't use http.ResponseWriter directly
// because Typeof takes an empty interface value. This is annoying.
var unusedResponseWriter *http.ResponseWriter
var unusedRequest *http.Request
var unusedSessions map[string]interface{}
var typeOfHttpResponseWriter = reflect.TypeOf(unusedResponseWriter).Elem()
var typeOfHttpRequest = reflect.TypeOf(unusedRequest)
var typeOfSessions = reflect.TypeOf(unusedSessions)

func (r *Router) Register(rcvr interface{}) error {

	return doRegister(r.Mux, rcvr, r.Sessions, r.NamePrefix, r.PatternPrefix, r.Style)
}

func doRegister(mux Mux, rcvr interface{}, sessions SessionManager, namePrefix, routePrefix string, sep byte) error {

	if namePrefix == "" {
		namePrefix = "Do"
	}

	if sep == 0 {
		sep = '-'
	}

	if mux == nil {
		mux = http.DefaultServeMux
	}

	typ := reflect.TypeOf(rcvr)
	rcvr1 := reflect.ValueOf(rcvr)

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
		//  1) (rcvr *XXXX) DoYYYY(w http.ResponseWriter, req *http.Request)
		//  2) (rcvr *XXXX) DoYYYY(w http.ResponseWriter, req *http.Request, sessions map[string]interface{})
		if mtype.NumOut() != 0 || (narg != 3 && narg != 4) {
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
		if narg == 4 {
			if sessionsType := mtype.In(3); sessionsType != typeOfSessions {
				log.Debug("method", mname, "third arguement type not map[string]interface{}:", sessionsType)
				continue
			}
		}
		pattern := routePrefix + Pattern(mname[len(namePrefix):], sep)
		log.Debug("Install", pattern, "=>", mname)
		if narg == 3 {
			mux.Handle(pattern, &handler{rcvr1, method.Func})
		} else {
			if sessions == nil {
				return ErrNoSessionManager
			}
			mux.Handle(pattern, &handlerWithSession{rcvr1, method.Func, sessions})
		}
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

//
// TODO: 这个包已经被废除，请改用 github.com/qiniu/webroute
//
func Register(mux Mux, rcvr interface{}, sessions SessionManager) error {

	router := &Router{
		Mux:      mux,
		Sessions: sessions,
	}
	return router.Register(rcvr)
}

//
// TODO: 这个包已经被废除，请改用 github.com/qiniu/webroute
//
func ListenAndServe(addr string, rcvr interface{}, sessions SessionManager) error {

	mux := http.NewServeMux()
	err := Register(mux, rcvr, sessions)
	if err != nil {
		return err
	}
	return http.ListenAndServe(addr, mux)
}

// ---------------------------------------------------------------------------
