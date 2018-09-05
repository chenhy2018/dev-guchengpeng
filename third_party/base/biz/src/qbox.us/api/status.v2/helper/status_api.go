package helper

import (
	"encoding/json"
	"net/http"

	"github.com/qiniu/rpc.v1"
)

const NoSuchEntry = 612 // 指定的 Entry 不存在或已经 Deleted

type Client struct {
	client *rpc.Client
	host   string
}

func New(host string, t http.RoundTripper) *Client {
	client := &http.Client{Transport: t}
	return &Client{
		client: &rpc.Client{client},
		host:   host,
	}
}

func (s *Client) Get(l rpc.Logger, ret interface{}, sid, id string) error {
	url := s.host + "/get/" + sid + "?id=" + id
	return s.client.Call(l, ret, url)
}

func (s *Client) Put(l rpc.Logger, ctx interface{}, sid, id string) error {
	url := s.host + "/put/" + sid + "?id=" + id
	return s.client.CallWithJson(l, nil, url, ctx)
}

func (s *Client) Delete(l rpc.Logger, sid, id string) error {
	url := s.host + "/delete/" + sid + "?id=" + id
	return s.client.Call(l, nil, url)
}

// 调用 Count 需要建索引，否则会拖垮数据库
func (s *Client) Count(l rpc.Logger, sid string, m map[string]interface{}) (n int, err error) {

	url := s.host + "/count/" + sid
	var ret map[string]int
	err = s.client.CallWithJson(l, &ret, url, m)
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

// 调用 List 需要建索引，否则会拖垮数据库
func (s *Client) List(l rpc.Logger, sid string, cond ListCond, results interface{}) (marker2 string, err error) {

	url := s.host + "/list/" + sid
	var rst ListRet
	rst.Results = []byte("[]")
	err = s.client.CallWithJson(l, &rst, url, cond)
	if err != nil {
		return
	}

	marker2 = rst.Marker

	err = json.Unmarshal(rst.Results, results)

	return
}

// 调用 List 需要建索引，否则会拖垮数据库
func (s *Client) GlobalList(l rpc.Logger, sid string, cond ListCond, results interface{}) (marker2 string, err error) {

	url := s.host + "/glb/status/list/" + sid
	var rst ListRet
	rst.Results = []byte("[]")
	err = s.client.CallWithJson(l, &rst, url, cond)
	if err != nil {
		return
	}

	marker2 = rst.Marker

	err = json.Unmarshal(rst.Results, results)

	return
}
