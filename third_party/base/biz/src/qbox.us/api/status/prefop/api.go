package prefop

import (
	"net/http"

	"github.com/qiniu/rpc.v1"

	"qbox.us/api/status/helper"
)

const (
	SID = "prefop"
)

type Client struct {
	helper.Client
}

func New(t http.RoundTripper) Client {
	return Client{helper.New(t)}
}

func (s Client) Get(l rpc.Logger, id string) (result Status, err error) {

	err = s.Client.Get(l, &result, SID, id)
	return
}

func (s Client) Put(l rpc.Logger, st Status, id string) error {

	return s.Client.Put(l, st, SID, id)
}

func (s Client) Delete(l rpc.Logger, id string) error {

	return s.Client.Delete(l, SID, id)
}

func (s Client) CountPipelineMsg(l rpc.Logger, pipeline string, code StatusCode) (n int, err error) {

	q := map[string]interface{}{
		"pipeline": pipeline,
		"code":     code,
	}

	return s.Count(l, SID, q)
}

func (s Client) ListPipelineMsg(l rpc.Logger, pipeline string, code StatusCode, marker string, limit int) (results []Status, marker2 string, err error) {

	cond := helper.ListCond{
		map[string]interface{}{
			"pipeline": pipeline,
			"code":     code,
		},
		limit,
		marker,
	}
	marker2, err = s.List(l, SID, cond, &results)
	return
}
