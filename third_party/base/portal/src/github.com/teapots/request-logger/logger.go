package reqlogger

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/teapots/inject"
	"github.com/teapots/teapot"
)

const (
	HeaderReqid         = "X-Reqid"
	HeaderResponseTime  = "X-Response-Time"
	HeaderXRealIp       = "X-Real-Ip"
	HeaderXForwardedFor = "X-Forwarded-For"
)

type LoggerOption struct {
	ColorMode     bool
	LineInfo      bool
	ShortLine     bool
	FlatLine      bool
	LogStackLevel teapot.Level
}

// ReqLogger filter
func ReqLoggerFilter(out teapot.LogPrinter, opts ...LoggerOption) inject.Provider {
	var opt LoggerOption
	if len(opts) > 0 {
		opt = opts[0]
	}
	return func(ctx teapot.Context, rw http.ResponseWriter, req *http.Request) {

		// use origin request id or create new request id
		reqId := req.Header.Get(HeaderReqid)
		reqIdLength := len(reqId)
		if reqIdLength < 10 || reqIdLength > 32 {
			reqId = NewReqId()
		}

		log := teapot.NewReqLogger(out, HeaderReqid, reqId)

		// write request id to request and response
		req.Header.Set(HeaderReqid, reqId)
		rw.Header().Set(HeaderReqid, reqId)

		log.SetLineInfo(opt.LineInfo)
		log.SetShortLine(opt.ShortLine)
		log.SetFlatLine(opt.FlatLine)
		log.SetColorMode(opt.ColorMode)
		log.EnableLogStack(opt.LogStackLevel)

		ctx.ProvideAs(log, (*teapot.Logger)(nil))
		ctx.ProvideAs(log, (*teapot.ReqLogger)(nil))

		remoteAddr := realIp(req)

		start := time.Now()
		log.Infof("[REQ_BEG] %s %s%s %s", req.Method, req.Host, req.URL, remoteAddr)

		res := rw.(teapot.ResponseWriter)
		res.Before(func(rw teapot.ResponseWriter) {
			rw.Header().Del(HeaderReqid)
			rw.Header().Set(HeaderReqid, reqId)
			times := fmt.Sprintf("%0.3fms", float64(time.Since(start).Nanoseconds())/1e6)
			rw.Header().Set(HeaderResponseTime, times)
		})

		ctx.Next()

		status := res.Status()
		times := fmt.Sprintf("%0.3fms", float64(time.Since(start).Nanoseconds())/1e6)
		log.Infof("[REQ_END] %d %0.3fk %s", status, float64(res.Size())/1024.0, times)
	}
}

func realIp(req *http.Request) string {
	if ip := req.Header.Get(HeaderXRealIp); ip != "" {
		return strings.TrimSpace(ip)
	}

	parts := strings.Split(req.Header.Get(HeaderXForwardedFor), ",")
	if len(parts) > 0 && parts[0] != "" {
		parts = strings.Split(parts[0], ":")
		return strings.TrimSpace(parts[0])
	}

	parts = strings.Split(req.RemoteAddr, ":")
	return parts[0]
}
