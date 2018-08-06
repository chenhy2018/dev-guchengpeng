package api

import (
	"fmt"

	"github.com/qiniu/rpc.v1"
	"qbox.us/iam/entity"
)

const (
	ListActionsPath  = "/iam/actions?%s"
	ListServicesPath = "/iam/services"
)

type ListActionsInput struct {
	Paginator
	Service string
}

func (i *ListActionsInput) GetQueryString() string {
	q := i.GetQuery()
	q.Set("service", i.Service)
	return q.Encode()
}

type ListActionsOutput struct {
	List  []*entity.Action `json:"list"`
	Count int              `json:"count"`
}

func (c *Client) ListActions(l rpc.Logger, query *ListActionsInput) (out *ListActionsOutput, err error) {
	var output struct {
		CommonResponse `json:",inline"`
		Data           *ListActionsOutput `json:"data"`
	}
	reqPath := fmt.Sprintf(ListActionsPath, query.GetQueryString())
	if err = c.client.GetCall(l, &output, reqPath); err != nil {
		return
	}
	out = output.Data
	return
}

func (c *Client) ListServices(l rpc.Logger) (out []string, err error) {
	var output struct {
		CommonResponse `json:",inline"`
		Data           []string `json:"data"`
	}
	if err = c.client.GetCall(l, &output, ListServicesPath); err != nil {
		return
	}
	out = output.Data
	return
}
