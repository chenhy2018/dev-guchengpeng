package api

import (
	"fmt"

	"github.com/qiniu/rpc.v1"
	"qbox.us/iam/entity"
)

const (
	ListUserKeypairsPath   = "/iam/u/%d/users/%s/keypairs?%s"
	CreateUserKeypairPath  = "/iam/u/%d/users/%s/keypairs"
	DeleteUserKeypairPath  = "/iam/u/%d/users/%s/keypairs/%s"
	EnableUserKeypairPath  = "/iam/u/%d/users/%s/keypairs/%s/enable"
	DisableUserKeypairPath = "/iam/u/%d/users/%s/keypairs/%s/disable"
)

type ListUserKeypairsInput struct {
	Paginator
}

func (i *ListUserKeypairsInput) GetQueryString() string {
	return i.GetQuery().Encode()
}

type ListUserKeypairsOutput struct {
	List  []*entity.Keypair `json:"list"`
	Count int               `json:"count"`
}

func (c *Client) ListUserKeypairs(l rpc.Logger, rootUID uint32, alias string, query *ListUserKeypairsInput) (out *ListUserKeypairsOutput, err error) {
	var output struct {
		CommonResponse `json:",inline"`
		Data           *ListUserKeypairsOutput `json:"data"`
	}
	reqPath := fmt.Sprintf(ListUserKeypairsPath, rootUID, alias, query.GetQueryString())
	if err = c.client.GetCall(l, &output, reqPath); err != nil {
		return
	}
	out = output.Data
	return
}

func (c *Client) CreateUserKeypair(l rpc.Logger, rootUID uint32, alias string) (keypair *entity.Keypair, err error) {
	var output struct {
		CommonResponse `json:",inline"`
		Data           *entity.Keypair `json:"data"`
	}
	reqPath := fmt.Sprintf(CreateUserKeypairPath, rootUID, alias)
	if err = c.client.Call(l, &output, reqPath); err != nil {
		return
	}
	keypair = output.Data
	return
}

func (c *Client) DeleteUserKeypair(l rpc.Logger, rootUID uint32, alias string, accessKey string) (err error) {
	var output CommonResponse
	reqPath := fmt.Sprintf(DeleteUserKeypairPath, rootUID, alias, accessKey)
	if err = c.client.DeleteCall(l, &output, reqPath); err != nil {
		return
	}
	return
}

func (c *Client) EnableUserKeypair(l rpc.Logger, rootUID uint32, alias string, accessKey string) (err error) {
	var output CommonResponse
	reqPath := fmt.Sprintf(EnableUserKeypairPath, rootUID, alias, accessKey)
	if err = c.client.Call(l, &output, reqPath); err != nil {
		return
	}
	return
}

func (c *Client) DisableUserKeypair(l rpc.Logger, rootUID uint32, alias string, accessKey string) (err error) {
	var output CommonResponse
	reqPath := fmt.Sprintf(DisableUserKeypairPath, rootUID, alias, accessKey)
	if err = c.client.Call(l, &output, reqPath); err != nil {
		return
	}
	return
}
