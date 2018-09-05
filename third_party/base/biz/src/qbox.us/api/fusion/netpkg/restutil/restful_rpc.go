package restutil

import (
	"github.com/qiniu/http/formutil.v1"
	"github.com/qiniu/http/hfac.v1"
	"github.com/qiniu/http/httputil.v1"
	"github.com/qiniu/http/rpcutil.v1"
	"github.com/qiniu/log.v1"

	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
	"strconv"
	"syscall"
)

// copy from base/qiniu/src/github.com/qiniu/http/restrpc.v1/restful_rpc.go
// add detail error code support.

// ------------------------------------------------------------------------
// Reply

var defaultRepl = &rpcutil.Replier{
	Reply:         httputil.Reply,
	ReplyWithCode: httputil.ReplyWithCode,
	Error:         ReplyError,
}

// ------------------------------------------------------------------------
// ErrorInfo

type detailError interface {
	Error() string
	ErrorCode() int
}

type ErrorInfo struct {
	Code int    `json:"code"`
	Err  string `json:"error"`
}

func (e *ErrorInfo) Error() string {
	return e.Err
}

func (e *ErrorInfo) ErrorCode() int {
	return e.Code
}

func NewError(code int, err string) *ErrorInfo {
	return &ErrorInfo{code, err}
}

func ReplyError(w http.ResponseWriter, err error) {
	if de, ok := err.(detailError); ok {
		errInfo := NewError(de.ErrorCode(), de.Error())
		msg, _ := json.Marshal(errInfo)
		log.Std.Output(w.Header().Get("X-Reqid"), log.Lwarn, 3, string(msg))

		h := w.Header()
		h.Set("Content-Length", strconv.Itoa(len(msg)))
		h.Set("Content-Type", "application/json")
		w.WriteHeader(httpCode(errInfo.Code))
		w.Write(msg)
		return
	}

	httputil.Error(w, err)
	return
}

func httpCode(code int) int {
	return code / 1000
}

// ------------------------------------------------------------------------
// ParseRequest

func parseReqDefault(ret reflect.Value, req *http.Request) error {

	if cmdArgs := req.Header["*"]; cmdArgs != nil {
		v := ret.Elem().FieldByName("CmdArgs")
		if v.IsValid() {
			v.Set(reflect.ValueOf(cmdArgs))
			if ret.Elem().NumField() == 1 {
				return nil
			}
		}
	}

	switch req.Header.Get("Content-Type") {
	case "application/json", "application/json; charset=UTF-8", "application/json; charset=utf-8":
		if req.ContentLength == 0 {
			return nil
		}
		return json.NewDecoder(req.Body).Decode(ret.Interface())
	default:
		err := req.ParseForm()
		if err != nil {
			return err
		}
		return formutil.ParseValue(ret, req.Form, "json")
	}
}

/* ---------------------------------------------------------------------------

在少数情况下，需用 ReqBody 来存储参数。样例：

	type Args struct {
		CmdArgs []string
		ReqBody map[string]interface{}
	}

	func (rcvr *XXXX) PostFoo_Bar_(args *Args, env *rpcutil.Env) {
		...
	}

如果请求为：

	POST /foo/COMMAND1/bar/COMMAND2
	Content-Type: application/json

	{
		"domain1": "IP1",
		"domain2": "IP2"
	}

那么解析出来的 args 为：

	args = &Args{
		CmdArgs: []string{"COMMAND1","COMMAND2"},
		ReqBody: map[string]interface{}{
			"domain1": "IP1",
			"domain2": "IP2",
		},
	}

// -------------------------------------------------------------------------*/

func parseReqWithBody(ret reflect.Value, req *http.Request) error {

	retElem := ret.Elem()
	if cmdArgs := req.Header["*"]; cmdArgs != nil {
		v := retElem.FieldByName("CmdArgs")
		if v.IsValid() {
			v.Set(reflect.ValueOf(cmdArgs))
		}
	}

	ret = retElem.FieldByName("ReqBody").Addr()

	switch req.Header.Get("Content-Type") {
	case "application/json":
		if req.ContentLength == 0 {
			return nil
		}
		return json.NewDecoder(req.Body).Decode(ret.Interface())
	default:
		return syscall.EINVAL
	}
}

func parseReqWithReader(ret reflect.Value, req *http.Request) error {

	retElem := ret.Elem()
	if cmdArgs := req.Header["*"]; cmdArgs != nil {
		v := retElem.FieldByName("CmdArgs")
		if v.IsValid() {
			v.Set(reflect.ValueOf(cmdArgs))
		}
	}
	retElem.FieldByName("ReqBody").Set(reflect.ValueOf(req.Body))
	return nil
}

func parseReqWithBytes(ret reflect.Value, req *http.Request) error {

	retElem := ret.Elem()
	if cmdArgs := req.Header["*"]; cmdArgs != nil {
		v := retElem.FieldByName("CmdArgs")
		if v.IsValid() {
			v.Set(reflect.ValueOf(cmdArgs))
		}
	}

	b, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return err
	}
	retElem.FieldByName("ReqBody").Set(reflect.ValueOf(b))
	return nil
}

// ---------------------------------------------------------------------------

var unusedReadCloser *io.ReadCloser
var typeOfIoReadCloser = reflect.TypeOf(unusedReadCloser).Elem()

func selParseReq(reqType reflect.Type) func(ret reflect.Value, req *http.Request) error {

	if sf, ok := reqType.FieldByName("ReqBody"); ok {
		t := sf.Type
		switch t.Kind() {
		case reflect.Map:
		case reflect.Interface:
			if typeOfIoReadCloser.Implements(sf.Type) { // io.ReadCloser
				return parseReqWithReader
			}
		case reflect.Slice:
			if t.Elem().Kind() == reflect.Uint8 { // []byte
				return parseReqWithBytes
			}
		}
		return parseReqWithBody
	}
	return parseReqDefault
}

// ---------------------------------------------------------------------------

var newHandler = rpcutil.HandlerCreator{SelParseReq: selParseReq, Repl: defaultRepl}.New

var Factory = hfac.HandlerFactory{
	{"Post", newHandler},
	{"Put", newHandler},
	{"Delete", newHandler},
	{"Get", newHandler},
}

// ---------------------------------------------------------------------------
