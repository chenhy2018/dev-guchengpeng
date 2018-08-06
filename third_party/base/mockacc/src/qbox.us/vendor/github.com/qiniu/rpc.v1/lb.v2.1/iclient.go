package lb

import (
	"io"
	"net/http"
	"github.com/qiniu/rpc.v1"
)

type IClient interface {
	Get(l rpc.Logger, url string) (resp *http.Response, err error)
	GetCall(l rpc.Logger, ret interface{}, url1 string) (err error)
	DeleteCall(l rpc.Logger, ret interface{}, url1 string) (err error)
	Do(l rpc.Logger, req *Request) (resp *http.Response, err error)
	CallWith64(l rpc.Logger, ret interface{}, path string, bodyType string, body io.ReaderAt, bodyLength int64) (err error)
	CallWithForm(l rpc.Logger, ret interface{}, path string, params map[string][]string) (err error)
	CallWithJson(l rpc.Logger, ret interface{}, path string, params interface{}) (err error)
	CallWith(l rpc.Logger, ret interface{}, path string, bodyType string, body io.ReaderAt, bodyLength int) (err error)
	Call(l rpc.Logger, ret interface{}, path string) (err error)
	PostEx(l rpc.Logger, path string) (resp *http.Response, err error)
	PostWith64(l rpc.Logger, path, bodyType string, body io.ReaderAt, bodyLength int64) (resp *http.Response, err error)
	PostWith(l rpc.Logger, path, bodyType string, body io.ReaderAt, bodyLength int) (resp *http.Response, err error)
	PostWithForm(l rpc.Logger, path string, params map[string][]string) (resp *http.Response, err error)
	PostWithJson(l rpc.Logger, path string, params interface{}) (resp *http.Response, err error)
	PostWithHostRet(l rpc.Logger, path, bodyType string, body io.ReaderAt, bodyLength int) (host string, resp *http.Response, err error)
	PutWithHostRet(l rpc.Logger, path, bodyType string, body io.ReaderAt, bodyLength int) (host string, resp *http.Response, err error)
	PutWithJson(l rpc.Logger, path string, params interface{}) (err error)
	GetLBCfg() Config
}
