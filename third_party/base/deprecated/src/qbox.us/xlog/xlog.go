package xlog

import (
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"qbox.us/log"
)

const logKey = "X-Log"
const reqidKey = "X-Reqid"

// ============================================================================

type reqIder interface {
	ReqId() string
}

type header interface {
	Header() http.Header
}

// ============================================================================
// type Instance

type logger struct {
	h     http.Header
	reqId string
}

type Instance struct {
	*logger
}

var pid = uint32(os.Getpid())

func genReqId() string {
	var b [12]byte
	binary.LittleEndian.PutUint32(b[:], pid)
	binary.LittleEndian.PutUint64(b[4:], uint64(time.Now().UnixNano()))
	return base64.URLEncoding.EncodeToString(b[:])
}

// NOTE: 这个包已经迁移到 github.com/qiniu/xlog
//
func New(w http.ResponseWriter, req *http.Request) Instance {
	reqId := req.Header.Get(reqidKey)
	if reqId == "" {
		reqId = genReqId()
		req.Header.Set(reqidKey, reqId)
	}
	h := w.Header()
	h.Set(reqidKey, reqId)
	return Instance{&logger{h, reqId}}
}

func NewWith(a interface{}) Instance {
	var h http.Header
	var reqId string
	if a == nil {
		reqId = genReqId()
	} else {
		l, ok := a.(Instance)
		if ok {
			return l
		}
		reqId, ok = a.(string)
		if !ok {
			if g, ok := a.(reqIder); ok {
				reqId = g.ReqId()
			} else {
				panic("xlog.NewWith: unknown param")
			}
			if g, ok := a.(header); ok {
				h = g.Header()
			}
		}
	}
	if h == nil {
		h = http.Header{reqidKey: []string{reqId}}
	}
	return Instance{&logger{h, reqId}}
}

func NewDummy() Instance {
	return NewWith(genReqId())
}

func (xlog Instance) Spawn() Instance {
	return NewWith(xlog.reqId)
}

// ============================================================================

func (xlog Instance) Xget() []string {
	return xlog.h[logKey]
}

func (xlog Instance) Xput(logs []string) {
	xlog.h[logKey] = append(xlog.h[logKey], logs...)
}

func (xlog Instance) Xlog(v ...interface{}) {
	s := fmt.Sprint(v...)
	xlog.h[logKey] = append(xlog.h[logKey], s)
}

func (xlog Instance) Xlogf(format string, v ...interface{}) {
	s := fmt.Sprintf(format, v...)
	xlog.h[logKey] = append(xlog.h[logKey], s)
}

/*
* 用法示意：

	func Foo(log xlog.Instance) {
		...
		now := time.Now()
		err := longtimeOperation()
		log.Xprof("longtimeOperation", now, err)
		...
	}
*/
func (xlog Instance) Xprof(modFn string, start time.Time, err error) {
	const maxErrorLen = 32
	durMs := time.Since(start).Nanoseconds() / 1e6
	if durMs > 0 {
		modFn += ":" + strconv.FormatInt(durMs, 10)
	}
	if err != nil {
		msg := err.Error()
		if len(msg) > maxErrorLen {
			msg = msg[:maxErrorLen]
		}
		modFn += "/" + msg
	}
	xlog.h[logKey] = append(xlog.h[logKey], modFn)
}

/*
* 用法示意：

	func Foo(log xlog.Instance) (err error) {
		defer log.Xtrack("Foo", time.Now(), &err)
		...
	}

	func Bar(log xlog.Instance) {
		defer log.Xtrack("Bar", time.Now(), nil)
		...
	}
*/
func (xlog Instance) Xtrack(modFn string, start time.Time, errTrack *error) {
	var err error
	if errTrack != nil {
		err = *errTrack
	}
	xlog.Xprof(modFn, start, err)
}

// ============================================================================

func (xlog Instance) ReqId() string {
	return xlog.reqId
}

func (xlog Instance) Header() http.Header {
	return xlog.h
}

// Print calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Print.
func (xlog Instance) Print(v ...interface{}) {
	log.Std.Output(xlog.reqId, log.Linfo, 2, fmt.Sprint(v...))
}

// Printf calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Printf.
func (xlog Instance) Printf(format string, v ...interface{}) {
	log.Std.Output(xlog.reqId, log.Linfo, 2, fmt.Sprintf(format, v...))
}

// Println calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Println.
func (xlog Instance) Println(v ...interface{}) {
	log.Std.Output(xlog.reqId, log.Linfo, 2, fmt.Sprintln(v...))
}

// -----------------------------------------

func (xlog Instance) Debugf(format string, v ...interface{}) {
	if log.Ldebug < log.Std.Level {
		return
	}
	log.Std.Output(xlog.reqId, log.Ldebug, 2, fmt.Sprintf(format, v...))
}

func (xlog Instance) Debug(v ...interface{}) {
	if log.Ldebug < log.Std.Level {
		return
	}
	log.Std.Output(xlog.reqId, log.Ldebug, 2, fmt.Sprintln(v...))
}

// -----------------------------------------

func (xlog Instance) Infof(format string, v ...interface{}) {
	if log.Linfo < log.Std.Level {
		return
	}
	log.Std.Output(xlog.reqId, log.Linfo, 2, fmt.Sprintf(format, v...))
}

func (xlog Instance) Info(v ...interface{}) {
	if log.Linfo < log.Std.Level {
		return
	}
	log.Std.Output(xlog.reqId, log.Linfo, 2, fmt.Sprintln(v...))
}

// -----------------------------------------

func (xlog Instance) Warnf(format string, v ...interface{}) {
	log.Std.Output(xlog.reqId, log.Lwarn, 2, fmt.Sprintf(format, v...))
}

func (xlog Instance) Warn(v ...interface{}) {
	log.Std.Output(xlog.reqId, log.Lwarn, 2, fmt.Sprintln(v...))
}

// -----------------------------------------

func (xlog Instance) Errorf(format string, v ...interface{}) {
	log.Std.Output(xlog.reqId, log.Lerror, 2, fmt.Sprintf(format, v...))
}

func (xlog Instance) Error(v ...interface{}) {
	log.Std.Output(xlog.reqId, log.Lerror, 2, fmt.Sprintln(v...))
}

// -----------------------------------------

// Fatal is equivalent to Print() followed by a call to os.Exit(1).
func (xlog Instance) Fatal(v ...interface{}) {
	log.Std.Output(xlog.reqId, log.Lfatal, 2, fmt.Sprint(v...))
	os.Exit(1)
}

// Fatalf is equivalent to Printf() followed by a call to os.Exit(1).
func (xlog Instance) Fatalf(format string, v ...interface{}) {
	log.Std.Output(xlog.reqId, log.Lfatal, 2, fmt.Sprintf(format, v...))
	os.Exit(1)
}

// Fatalln is equivalent to Println() followed by a call to os.Exit(1).
func (xlog Instance) Fatalln(v ...interface{}) {
	log.Std.Output(xlog.reqId, log.Lfatal, 2, fmt.Sprintln(v...))
	os.Exit(1)
}

// -----------------------------------------

// Panic is equivalent to Print() followed by a call to panic().
func (xlog Instance) Panic(v ...interface{}) {
	s := fmt.Sprint(v...)
	log.Std.Output(xlog.reqId, log.Lpanic, 2, s)
	panic(s)
}

// Panicf is equivalent to Printf() followed by a call to panic().
func (xlog Instance) Panicf(format string, v ...interface{}) {
	s := fmt.Sprintf(format, v...)
	log.Std.Output(xlog.reqId, log.Lpanic, 2, s)
	panic(s)
}

// Panicln is equivalent to Println() followed by a call to panic().
func (xlog Instance) Panicln(v ...interface{}) {
	s := fmt.Sprintln(v...)
	log.Std.Output(xlog.reqId, log.Lpanic, 2, s)
	panic(s)
}

// ============================================================================
