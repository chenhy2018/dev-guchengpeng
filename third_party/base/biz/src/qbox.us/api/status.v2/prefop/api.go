package prefop

import (
	"net/http"

	"github.com/qiniu/rpc.v1"

	"qbox.us/api/status.v2/helper"
)

const (
	SID = "prefop"
)

type Client struct {
	client *helper.Client
}

func New(host string, t http.RoundTripper) *Client {
	return &Client{
		client: helper.New(host, t),
	}
}

func (s Client) Get(l rpc.Logger, id string) (result Status, err error) {

	err = s.client.Get(l, &result, SID, id)
	return
}

func (s Client) Put(l rpc.Logger, st Status, id string) error {

	return s.client.Put(l, st, SID, id)
}

func (s Client) Delete(l rpc.Logger, id string) error {

	return s.client.Delete(l, SID, id)
}

func (s *Client) List(l rpc.Logger, pipeline string, marker string, limit int) (results []Status, marker2 string, err error) {
	cond := helper.ListCond{
		Query:  map[string]interface{}{"pipeline": pipeline}, // 更改 query 需要更改索引
		Marker: marker,
		Limit:  limit,
	}
	marker2, err = s.client.List(l, SID, cond, &results)
	return
}

func (s *Client) GlobalList(l rpc.Logger, pipeline string, marker string, limit int) (results []Status, marker2 string, err error) {
	cond := helper.ListCond{
		Query:  map[string]interface{}{"pipeline": pipeline}, // 更改 query 需要更改索引
		Marker: marker,
		Limit:  limit,
	}
	marker2, err = s.client.GlobalList(l, SID, cond, &results)
	return
}
