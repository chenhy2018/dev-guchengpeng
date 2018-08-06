package client

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"

	"github.com/qiniu/rpc.v1"

	"qbox.us/biz/utils.v2/log"
)

var (
	DefaultTransport = &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 60 * time.Second,
		Dial: (&net.Dialer{
			Timeout:   60 * time.Second,
			KeepAlive: 60 * time.Second,
		}).Dial,
	}
)

type Transport struct {
	*http.Transport
}

func (t *Transport) DisableTLS() {
	t.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
}

func newTransport() *Transport {
	return &Transport{Transport: DefaultTransport}
}

type TransportWithReqLogger struct {
	*Transport

	reqLogger log.ReqLogger
}

func NewTransportWithReqLogger(reqLogger log.ReqLogger) *TransportWithReqLogger {
	return &TransportWithReqLogger{
		Transport: newTransport(),
		reqLogger: reqLogger,
	}
}

func (t *TransportWithReqLogger) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	start := time.Now()
	resp, err = t.Transport.RoundTrip(req)
	logTransportRequest(t.reqLogger.ReqId(), t.reqLogger, req, resp, start, err)
	return
}

type TransportWithLogger struct {
	*Transport

	logger log.Logger
}

func NewTransportWithLogger(logger log.Logger) *TransportWithLogger {
	return &TransportWithLogger{
		Transport: newTransport(),
		logger:    logger,
	}
}

func (t *TransportWithLogger) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	start := time.Now()
	resp, err = t.Transport.RoundTrip(req)
	logTransportRequest("", t.logger, req, resp, start, err)
	return
}

func logTransportRequest(reqId string, reqLogger log.Logger, req *http.Request, resp *http.Response, start time.Time, err error) {
	var (
		respReqId, xlog string
		code            int
		extra           string

		addr     = req.Host + req.URL.Path
		elaplsed = time.Since(start)
	)

	if resp != nil {
		respReqId = resp.Header.Get("X-Reqid")
		xlog = resp.Header.Get("X-Log")
		code = resp.StatusCode
	}

	if len(respReqId) > 0 && respReqId != reqId {
		extra = ", RespReqId: " + respReqId
	}

	if len(xlog) > 0 {
		extra += ", Xlog: " + xlog
	}

	if err != nil {
		extra += ", Err: " + err.Error()
		if er, ok := err.(rpc.RespError); ok {
			extra += ", " + er.ErrorDetail()
		}
	}

	reqLogger.Infof("Service: %s %s, Code: %d%s, Time: %dms", req.Method, addr, code, extra, elaplsed.Nanoseconds()/1e6)
}
