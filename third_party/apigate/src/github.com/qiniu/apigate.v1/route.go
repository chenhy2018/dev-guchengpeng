package apigate

import (
	"net/http"
	"strings"

	"github.com/qiniu/log.v1"
)

/* --------------------------------------------------------------------

<pattern> = <method pattern> <path pattern>
<method pattern> = GET | POST | DELETE | PUT | *
<path pattern> = 路径匹配模式，* 表示匹配任意多个非 '/' 字符，** 表示匹配后续所有字符，只能出现在尾部

// ------------------------------------------------------------------*/

// POST /servers/<ServerId>/action => []string{"POST", "servers", "*", "action"}
type Pattern []string

func (p Pattern) Match(method string, cmds []string) (ok bool) {

	n := len(p)
	if p[n-1] == "**" {
		if n <= 1 || len(cmds)+1 < n {
			return
		}
		p = p[:n-1]
		cmds = cmds[:n-2]
	} else if len(cmds)+1 != n {
		return
	}

	method = strings.ToUpper(method)
	if p[0] != "*" && method != "OPTIONS" && method != "HEAD" && method != "TRACE" {
		if !strings.EqualFold(p[0], method) {
			return
		}
	}

	for i := 1; i < len(p); i++ {
		if p[i] == "*" {
			continue
		}
		if !strings.EqualFold(p[i], cmds[i-1]) {
			return
		}
	}
	ok = true
	return
}

// "POST /servers/*/action"
func NewPattern(pattern string) Pattern {

	parts := strings.Split(pattern, "/")

	// trim [METHOD]'s leading or trailing white space
	parts[0] = strings.TrimSpace(parts[0])

	return parts
}

type route struct {
	pattern Pattern
	handler http.Handler
}

type ServeMux struct {
	routes          []*route
	notFoundHandler http.Handler
}

func NewServeMux() *ServeMux {
	return new(ServeMux)
}

func (h *ServeMux) handle(pattern Pattern, handler http.Handler) {

	h.routes = append(h.routes, &route{pattern, handler})
}

func (h *ServeMux) Handle(pattern string, handler http.Handler) *ServeMux {

	h.handle(NewPattern(pattern), handler)
	return h
}

func (h *ServeMux) HandleNotFound(notFoundHandler http.Handler) *ServeMux {
	h.notFoundHandler = notFoundHandler
	return h
}

func (h *ServeMux) ServeNotFound(w http.ResponseWriter, r *http.Request) {
	if h.notFoundHandler != nil {
		h.notFoundHandler.ServeHTTP(w, r)
	} else {
		log.Warn("NotFound: url -", r.Host, r.URL.String())
		http.NotFound(w, r)
	}
}

func (h *ServeMux) HandleFunc(
	pattern string, handler func(http.ResponseWriter, *http.Request)) *ServeMux {

	h.handle(NewPattern(pattern), http.HandlerFunc(handler))
	return h
}

func (h *ServeMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	parts := strings.Split(r.URL.Path[1:], "/")
	h.doServeHTTP(w, r, parts)
}

func (h *ServeMux) doServeHTTP(w http.ResponseWriter, r *http.Request, parts []string) {

	for _, route := range h.routes {
		if ok := route.pattern.Match(r.Method, parts); ok {
			route.handler.ServeHTTP(w, r)
			return
		}
	}

	h.ServeNotFound(w, r)
}

// --------------------------------------------------------------------

type serviceMux struct {
	prefix []string
	mux    *ServeMux
}

func (p *serviceMux) tryServeHTTP(w http.ResponseWriter, r *http.Request, parts []string) (ok bool) {

	n := len(p.prefix)
	if n > 0 {
		if len(parts) < n {
			return
		}
		for i := 0; i < n; i++ {
			if parts[i] != p.prefix[i] {
				return
			}
		}
		parts = parts[n:]
		r.Header["*"] = parts
	}
	p.mux.doServeHTTP(w, r, parts)
	return true
}

// --------------------------------------------------------------------

type MultiServeMux map[string][]*serviceMux

func NewMultiServeMux() MultiServeMux {

	return make(MultiServeMux)
}

func (p MultiServeMux) ServeMux(routes ...string) *ServeMux {

	mux := NewServeMux()
	for _, route := range routes {
		parts := strings.Split(route, "/")
		service := &serviceMux{prefix: parts[1:], mux: mux}
		host := parts[0]
		p[host] = append(p[host], service)
	}
	return mux
}

func (p MultiServeMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	for _, h := range []string{r.Host, getPortStr(r.Host), "*"} {
		if services, ok := p[h]; ok {
			parts := strings.Split(r.URL.Path[1:], "/")
			for _, service := range services {
				if service.tryServeHTTP(w, r, parts) {
					return
				}
			}
		}
	}

	log.Warn("NotFound: url -", r.Host, r.URL.String())
	http.NotFound(w, r)
}

func getPortStr(host string) string {
	if idx := strings.Index(host, ":"); idx != -1 {
		return host[idx:]
	} else {
		return ":80"
	}
}

// --------------------------------------------------------------------
