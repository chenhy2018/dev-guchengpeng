// 不要直接使用这个库，使用这个目录下面具体的库
// 推荐使用 status.v2
package status

import (
	"encoding/json"
	"net/http"

	"github.com/qiniu/log.v1"
	"github.com/qiniu/rpc.v1"
	. "qbox.us/api/conf"
)

const NoSuchEntry = 612 // 指定的 Entry 不存在或已经 Deleted

type Client struct {
	Conn rpc.Client
}

func init() {
	log.Error(`this package is DEPRECATED, use "qbox.us/api/status/prefop" and so on`)
}

func New(t http.RoundTripper) Client {
	client := &http.Client{Transport: t}
	return Client{rpc.Client{client}}
}

func (s Client) Get(l rpc.Logger, ret interface{}, sid, id string) error {
	url := STATUS_HOST + "/get/" + sid + "?id=" + id
	return s.Conn.Call(l, ret, url)
}

func (s Client) Put(l rpc.Logger, ctx interface{}, sid, id string) error {
	url := STATUS_HOST + "/put/" + sid + "?id=" + id
	return s.Conn.CallWithJson(l, nil, url, ctx)
}

func (s Client) Delete(l rpc.Logger, sid, id string) error {
	url := STATUS_HOST + "/delete/" + sid + "?id=" + id
	return s.Conn.Call(l, nil, url)
}

func (s Client) Count(l rpc.Logger, sid string, m map[string]interface{}) (n int, err error) {

	url := STATUS_HOST + "/count/" + sid
	var ret map[string]int
	err = s.Conn.CallWithJson(l, &ret, url, m)
	if err != nil {
		return
	}
	n = ret["n"]
	return
}

type ListCond struct {
	Query  map[string]interface{} `json:"query"`
	Limit  int                    `json:"limit"`
	Marker string                 `json:"marker"`
}

type ListRet struct {
	Results json.RawMessage `json:"results"`
	Marker  string          `json:"marker,omitempty"`
}

func (s Client) List(l rpc.Logger, sid string, cond ListCond, results interface{}) (marker2 string, err error) {

	url := STATUS_HOST + "/list/" + sid
	var rst ListRet
	rst.Results = []byte("[]")
	err = s.Conn.CallWithJson(l, &rst, url, cond)
	if err != nil {
		return
	}

	marker2 = rst.Marker

	err = json.Unmarshal(rst.Results, results)

	return
}
