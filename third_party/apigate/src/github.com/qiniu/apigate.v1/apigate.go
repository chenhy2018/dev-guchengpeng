package apigate

import (
	"context"
	"net"
	"net/http"
	gohttputil "net/http/httputil"
	"strings"
	"sync/atomic"

	"github.com/qiniu/xlog.v1"

	"github.com/qiniu/http/httputil.v1"
	"github.com/qiniu/rtutil"

	. "github.com/qiniu/apigate.v1/proto"
)

var (
	ErrUnauthorized            = httputil.NewError(401, "unauthorized")
	ErrInvalidAccess           = httputil.NewError(400, "invalid access mode")
	ErrNoServiceForwardHost    = httputil.NewError(400, "no service forward host")
	ErrNotFound                = httputil.NewError(404, "no routes found")
	ErrMaxConcurencyExcceed    = httputil.NewError(500, "max concurency excceed of service")
	ErrMaxApiConcurencyExcceed = httputil.NewError(500, "max concurency excceed of api")
)

// ------------------------------------------------------------------------

func nilDirector(req *http.Request) {}

var theProxy = &gohttputil.ReverseProxy{
	Director:  nilDirector,
	Transport: rtutil.New570RT(http.DefaultTransport),
}

func SetDefaultProxyTransport(t http.RoundTripper) {

	theProxy.Transport = t
}

func (p *ApiHandler) doProxy(w http.ResponseWriter, req *http.Request) {

	if p.forward != "" {
		req.URL.Path = p.forward
	} else if parts, ok := req.Header["*"]; ok {
		path := make([]byte, 0, len(req.URL.Path))
		for _, part := range parts {
			path = append(path, '/')
			path = append(path, part...)
		}
		req.URL.Path = string(path)
	}
	req.URL.Path = p.parent.pathPrefix + req.URL.Path

	delete(req.Header, "*")

	req.URL.Scheme = "http"
	req.URL.Host = p.parent.fwdAddr
	if p.parent.fwdHost != "" {
		req.Host = p.parent.fwdHost
	}
	if clientIP, _, err := net.SplitHostPort(req.RemoteAddr); err == nil {
		if prior, ok := req.Header["X-Forwarded-For"]; ok {
			clientIP = strings.Join(prior, ", ") + ", " + clientIP
		}
		req.Header.Set("X-Forwarded-For", clientIP)
	}
	p.proxy.ServeHTTP(w, req)
}

// --------------------------------------------------------------------

type ApiHandler struct {
	forward        string
	pattern        string
	auths          []AuthStuber
	access         AccessInfo
	parent         *ServiceHandler
	proxy          http.Handler
	maxConcurrency int64
	concurency     int64
}

func newApiHandler(parent *ServiceHandler, pattern string, maxConcurrency int64) *ApiHandler {

	p := &ApiHandler{
		pattern:        pattern,
		parent:         parent,
		auths:          parent.auths,
		proxy:          theProxy,
		maxConcurrency: maxConcurrency,
	}
	parent.mux.Handle(pattern, p)
	return p
}

func (p *ApiHandler) TryServe() bool {

	if p.maxConcurrency > 0 {
		n := atomic.AddInt64(&p.concurency, 1)
		if n > p.maxConcurrency {
			atomic.AddInt64(&p.concurency, -1)
			return false
		}
		return true
	}
	return true
}

func (p *ApiHandler) ServeDone() {
	atomic.AddInt64(&p.concurency, -1)
	return
}

func (p *ApiHandler) Proxy(h http.Handler) *ApiHandler {

	p.proxy = h
	return p
}

func (p *ApiHandler) Access(ai AccessInfo) *ApiHandler {

	p.access = ai
	return p
}

func (p *ApiHandler) Forward(forward string) *ApiHandler {

	p.forward = forward
	return p
}

func (p *ApiHandler) Auths(auths ...AuthStuber) *ApiHandler {

	p.auths = auths
	return p
}

func (p *ApiHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	_ = xlog.New(w, req) // init reqid for all requests

	service := p.parent

	ctx := NewContextByMod(context.Background(), service.module)
	ctx = NewContextByPattern(ctx, p.pattern)
	events := service.parent.OpenRequests(ctx, &w, req)

	if !p.TryServe() {
		httputil.Error(w, ErrMaxApiConcurencyExcceed)
		events.End(w, req)
		return
	}
	defer p.ServeDone()

	if !service.TryServe() {
		httputil.Error(w, ErrMaxConcurencyExcceed)
		events.End(w, req)
		return
	}
	defer service.ServeDone()

	if p.access.Allow > 0 {
		authi, ok, err := AuthStub(req, p.auths)
		events.AuthParsed(w, req)
		if ok && err != nil {
			httputil.Error(w, err)
			events.End(w, req)
			return
		}
		if !ok || !p.access.Can(authi) {
			httputil.Error(w, ErrUnauthorized)
			events.End(w, req)
			return
		}
	} else if p.access.Allow != Access_Public {
		httputil.Error(w, ErrInvalidAccess)
		events.End(w, req)
		return
	}

	p.doProxy(w, req)
	events.End(w, req)
}

// --------------------------------------------------------------------
// handle 404
type NotFoundHandler struct {
	parent *ServiceHandler
}

func newNotFoundandler(parent *ServiceHandler) *NotFoundHandler {

	p := &NotFoundHandler{
		parent: parent,
	}
	return p
}

const (
	pattern404 = "404"
)

func (p *NotFoundHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	_ = xlog.New(w, req) // init reqid for all requests
	service := p.parent

	ctx := NewContextByMod(context.Background(), service.module)
	ctx = NewContextByPattern(ctx, pattern404)

	events := service.parent.OpenRequests(ctx, &w, req)
	httputil.Error(w, ErrNotFound)
	events.End(w, req)
}

// --------------------------------------------------------------------

type ServiceHandler struct {
	module     string
	fwdAddr    string
	fwdHost    string
	pathPrefix string
	auths      []AuthStuber
	parent     *Service
	mux        *ServeMux

	maxConcurrency int64
	concurency     int64
}

func newServiceHandler(parent *Service, module string, maxConcurrency int64, routes ...string) *ServiceHandler {

	return &ServiceHandler{
		module:         module,
		parent:         parent,
		mux:            parent.ServeMux(routes...),
		maxConcurrency: maxConcurrency,
	}
}

func (p *ServiceHandler) Auths(auths ...AuthStuber) *ServiceHandler {

	p.auths = auths
	return p
}

func (p *ServiceHandler) Forward(fwd string) *ServiceHandler {

	l := strings.SplitN(fwd, "/", 2)
	p.fwdAddr = l[0]
	if len(l) == 2 {
		l[1] = strings.TrimRight(l[1], "/")
		p.pathPrefix = "/" + l[1]
	}
	return p
}

func (p *ServiceHandler) TryServe() bool {

	if p.maxConcurrency > 0 {
		n := atomic.AddInt64(&p.concurency, 1)
		if n > p.maxConcurrency {
			atomic.AddInt64(&p.concurency, -1)
			return false
		}
		return true
	}
	return true
}

func (p *ServiceHandler) ServeDone() {
	atomic.AddInt64(&p.concurency, -1)
}

func (p *ServiceHandler) ForwardHost(host string) *ServiceHandler {
	p.fwdHost = host
	return p
}

func (p *ServiceHandler) Api(pattern string, maxConcurrency int64) *ApiHandler {

	return newApiHandler(p, pattern, maxConcurrency)
}

func (p *ServiceHandler) HandleNotFound() *ServiceHandler {
	p.mux.HandleNotFound(newNotFoundandler(p))
	return p
}

// --------------------------------------------------------------------

type Service struct {
	MultiServeMux
	events []ServiceEvent
}

func New() (p *Service) {

	return &Service{
		MultiServeMux: NewMultiServeMux(),
	}
}

type requestEvents []RequestEvent

func (r requestEvents) AuthParsed(w http.ResponseWriter, req *http.Request) {

	for _, e := range r {
		e.AuthParsed(w, req)
	}
}

func (r requestEvents) End(w http.ResponseWriter, req *http.Request) {

	for _, e := range r {
		e.End(w, req)
	}
}

func (p *Service) OpenRequests(ctx context.Context, w *http.ResponseWriter, req *http.Request) requestEvents {

	re := make([]RequestEvent, len(p.events))
	for i, event := range p.events {
		re[i] = event.OpenRequest(ctx, w, req)
	}
	return re
}

func (p *Service) Sink(event ServiceEvent) {

	p.events = append(p.events, event)
}

func (p *Service) Service(module string, maxConcurrency int64, routes ...string) *ServiceHandler {

	return newServiceHandler(p, module, maxConcurrency, routes...)
}

// --------------------------------------------------------------------
