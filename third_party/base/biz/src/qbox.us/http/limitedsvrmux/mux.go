package limitedsvrmux

import (
	"net/http"
	"qiniupkg.com/http/httputil.v2"
	"strings"
)

var ErrOverload = httputil.NewError(503, "service is overload")

type Limit struct {
	sem chan struct{}
}

func newLimit(n int) Limit {
	return Limit{make(chan struct{}, n)}
}

func (l *Limit) acquire() error {
	select {
	case l.sem <- struct{}{}:
		return nil
	default:
		return ErrOverload
	}
}

func (l *Limit) release() {
	<-l.sem
}

type Muxer interface {
	Handler(r *http.Request) (h http.Handler, pattern string)
}

type ServeMux struct {
	Muxer
	apiHandleLimitMap map[string]Limit
}

// @apiHandleLimitMap 是一个tag -> 句柄限制的map，如果不同的pattern共用一个句柄限制
// 那么他们就以“｜”分割
func NewServeMux(mux Muxer, apiHandleLimitMap map[string]int) *ServeMux {
	if mux == nil {
		mux = http.DefaultServeMux
	}
	limitMap := make(map[string]Limit)
	for tag, limit := range apiHandleLimitMap {
		patterns := strings.Split(tag, "|")
		l := newLimit(limit)
		for _, pattern := range patterns {
			limitMap[pattern] = l
		}
	}
	return &ServeMux{mux, limitMap}
}

func (mux *ServeMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.RequestURI == "*" {
		if r.ProtoAtLeast(1, 1) {
			w.Header().Set("Connection", "close")
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	h, pattern := mux.Handler(r)
	l, ok := mux.apiHandleLimitMap[pattern]
	if ok {
		err := l.acquire()
		if err != nil {
			httputil.Error(w, err)
			return
		}
	}
	h.ServeHTTP(w, r)
	if ok {
		l.release()
	}
}
