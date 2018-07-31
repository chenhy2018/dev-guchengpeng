package log

import (
	"net/http"

	"github.com/qiniu/rpc.v1"
	"github.com/qiniu/xlog.v1"

	"github.com/teapots/teapot"
)

const (
	logKey = "X-Log"
)

type Logger interface {
	teapot.Logger
}

type ReqLogger interface {
	teapot.ReqLogger
}

type RpcWrapper struct {
	ReqLogger
	header http.Header
}

var _ rpc.Logger = new(RpcWrapper)

func (r *RpcWrapper) Xput(logs []string) {
	if r.header == nil {
		r.header = make(http.Header)
	}
	r.header[logKey] = append(r.header[logKey], logs...)
}

func (r *RpcWrapper) Header() http.Header {
	if r.header == nil {
		r.header = make(http.Header)
	}
	return r.header
}

func NewRpcWrapper(logger ReqLogger) *RpcWrapper {
	wrapper := new(RpcWrapper)
	wrapper.ReqLogger = logger
	return wrapper
}

type xlogReqlogger struct {
	log *xlog.Logger
}

func (l *xlogReqlogger) Emergencyf(format string, v ...interface{}) {
	l.log.Errorf(format, v...)
}

func (l *xlogReqlogger) Emergency(v ...interface{}) {
	l.log.Error(v...)
}

func (l *xlogReqlogger) Alertf(format string, v ...interface{}) {
	l.log.Errorf(format, v...)
}

func (l *xlogReqlogger) Alert(v ...interface{}) {
	l.log.Error(v...)
}

func (l *xlogReqlogger) Critcialf(format string, v ...interface{}) {
	l.log.Errorf(format, v...)
}

func (l *xlogReqlogger) Critcial(v ...interface{}) {
	l.log.Error(v...)
}

func (l *xlogReqlogger) Errorf(format string, v ...interface{}) {
	l.log.Errorf(format, v...)
}

func (l *xlogReqlogger) Error(v ...interface{}) {
	l.log.Error(v...)
}

func (l *xlogReqlogger) Warnf(format string, v ...interface{}) {
	l.log.Warnf(format, v...)
}

func (l *xlogReqlogger) Warn(v ...interface{}) {
	l.log.Warn(v...)
}

func (l *xlogReqlogger) Noticef(format string, v ...interface{}) {
	l.log.Infof(format, v...)
}

func (l *xlogReqlogger) Notice(v ...interface{}) {
	l.log.Info(v...)
}

func (l *xlogReqlogger) Infof(format string, v ...interface{}) {
	l.log.Infof(format, v...)
}

func (l *xlogReqlogger) Info(v ...interface{}) {
	l.log.Info(v...)
}

func (l *xlogReqlogger) Debugf(format string, v ...interface{}) {
	l.log.Debugf(format, v...)
}

func (l *xlogReqlogger) Debug(v ...interface{}) {
	l.log.Debug(v...)
}

func (l *xlogReqlogger) Recover(f func(), level ...teapot.Level) (err interface{}) {
	panic("not implemented")
}

func (l *xlogReqlogger) ReqId() string {
	return l.log.ReqId()
}

func (l *xlogReqlogger) Header() http.Header {
	return l.log.Header()
}

func NewXlogReqLogger(log *xlog.Logger) ReqLogger {
	return &xlogReqlogger{log: log}
}
