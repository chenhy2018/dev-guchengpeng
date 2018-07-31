package api

import (
	"fmt"

	"github.com/qiniu/rpc.v1"
	"qbox.us/iam/entity"
)

const (
	ListRootUsersPath   = "/iam/u?%s"
	GetRootUserPath     = "/iam/u/%d"
	EnableRootUserPath  = "/iam/u/%d"
	DisableRootUserPath = "/iam/u/%d"
	UpdateRootUserPath  = "/iam/u/%d"
)

type EnableRootUserInput struct {
	Alias *string `json:"alias"`
}

type UpdateRootUserInput struct {
	Alias *string `json:"alias"`
}

type ListRootUsersInput struct {
	Paginator
	Alias string
}

func (i *ListRootUsersInput) GetQueryString() string {
	q := i.GetQuery()
	q.Set("alias", i.Alias)
	return q.Encode()
}

type ListRootUsersOutput struct {
	List  []*entity.RootUser `json:"list"`
	Count int                `json:"count"`
}

func (c *Client) GetRootUser(l rpc.Logger, uid uint32) (out *entity.RootUser, err error) {
	var output struct {
		CommonResponse `json:",inline"`
		Data           *entity.RootUser `json:"data"`
	}
	reqPath := fmt.Sprintf(GetRootUserPath, uid)
	if err = c.client.GetCall(l, &output, reqPath); err != nil {
		return
	}
	out = output.Data
	return
}

func (c *Client) ListRootUsers(l rpc.Logger, query *ListRootUsersInput) (out *ListRootUsersOutput, err error) {
	var output struct {
		CommonResponse `json:",inline"`
		Data           *ListRootUsersOutput `json:"data"`
	}
	reqPath := fmt.Sprintf(ListRootUsersPath, query.GetQueryString())
	if err = c.client.GetCall(l, &output, reqPath); err != nil {
		return
	}
	out = output.Data
	return
}

func (c *Client) EnableRootUser(l rpc.Logger, uid uint32, query *EnableRootUserInput) (rootUser *entity.RootUser, err error) {
	var output struct {
		CommonResponse `json:",inline"`
		Data           *entity.RootUser `json:"data"`
	}
	reqPath := fmt.Sprintf(EnableRootUserPath, uid)
	if err = c.client.CallWithJson(l, &output, reqPath, query); err != nil {
		return
	}
	rootUser = output.Data
	return
}

func (c *Client) DisableRootUser(l rpc.Logger, uid uint32) (err error) {
	var output CommonResponse
	reqPath := fmt.Sprintf(DisableRootUserPath, uid)
	if err = c.client.GetCall(l, &output, reqPath); err != nil {
		return
	}
	return
}

func (c *Client) UpdateRootUser(l rpc.Logger, uid uint32, param *UpdateRootUserInput) (rootUser *entity.RootUser, err error) {
	var output struct {
		CommonResponse `json:",inline"`
		Data           *entity.RootUser `json:"data"`
	}
	reqPath := fmt.Sprintf(UpdateRootUserPath, uid)
	if err = c.client.CallWithJson(l, &output, reqPath, param); err != nil {
		return
	}
	rootUser = output.Data
	return
}
