package proto

import (
	"context"
	"net/http"
)

// --------------------------------------------------------------------

type RequestEvent interface {
	AuthParsed(w http.ResponseWriter, req *http.Request)
	End(w http.ResponseWriter, req *http.Request)
}

type ServiceEvent interface {
	OpenRequest(ctx context.Context, w *http.ResponseWriter, req *http.Request) RequestEvent
}

type nilRequestEvent struct{}
type nilServiceEvent struct{}

func (p nilServiceEvent) OpenRequest(
	ctx context.Context, w *http.ResponseWriter, req *http.Request) RequestEvent {
	return NilReqEvent
}

func (p nilRequestEvent) AuthParsed(w http.ResponseWriter, req *http.Request) {}
func (p nilRequestEvent) End(w http.ResponseWriter, req *http.Request)        {}

var NilReqEvent RequestEvent = nilRequestEvent{}
var NilSvcEvent ServiceEvent = nilServiceEvent{}

// --------------------------------------------------------------------

type modKeyT int

const modKey modKeyT = 0

func ModFromContext(ctx context.Context) (mod string, ok bool) {
	mod, ok = ctx.Value(modKey).(string)
	return
}

func ModFromContextSafe(ctx context.Context) (mod string) {
	mod, ok := ctx.Value(modKey).(string)
	if !ok {
		mod = "UNKNOWN"
	}
	return
}

func NewContextByMod(ctx context.Context, mod string) context.Context {
	return context.WithValue(ctx, modKey, mod)
}

// --------------------------------------------------------------------

type patternKeyT int

const patternKey patternKeyT = 0

func PatternFromContext(ctx context.Context) (pattern string, ok bool) {
	pattern, ok = ctx.Value(patternKey).(string)
	return
}

func PatternFromContextSafe(ctx context.Context) string {
	pattern, ok := ctx.Value(patternKey).(string)
	if !ok {
		pattern = "/**"
	}
	return pattern
}

func NewContextByPattern(ctx context.Context, pattern string) context.Context {
	return context.WithValue(ctx, patternKey, pattern)
}

// --------------------------------------------------------------------

type Metric interface {
	Register(moduleName, confStr string) error
}

var NilMetric Metric = nilMetric{}

type nilMetric struct {
}

func (n nilMetric) Register(name, conf string) error {
	return nil
}
