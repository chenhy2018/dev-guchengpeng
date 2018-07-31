package api

import (
	"fmt"

	"github.com/qiniu/rpc.v1"
)

const (
	CheckUserPermissionsPath                = "/iam/u/%d/users/%s/permission/check"
	CheckUserGroupPermissionsPath           = "/iam/u/%d/groups/%s/permission/check"
	ListResourcesOfUserUnderActionPath      = "/iam/u/%d/users/%s/services/%s/actions/%s/resources"
	ListResourcesOfUserGroupUnderActionPath = "/iam/u/%d/groups/%s/services/%s/actions/%s/resources"
	ListAvailableServiceOfUserPath          = "/iam/u/%d/users/%s/services"
)

type CheckPermissionInput struct {
	Action   []string `json:"action"`
	Resource []string `json:"resource"`
}

type CheckPermissionOutput []struct {
	Action   string `json:"action"`
	Resource string `json:"resource,omitempty"`
	Effect   string `json:"effect"`
}

type ListResourceOutput struct {
	Allow []string `json:"allow"`
	Deny  []string `json:"deny"`
}

func (c *Client) ListResourcesOfUserUnderAction(l rpc.Logger, uid uint32, uAlias, service, aAlias string) (out *ListResourceOutput, err error) {
	var output struct {
		CommonResponse `json:",inline"`
		Data           *ListResourceOutput `json:"data"`
	}
	reqPath := fmt.Sprintf(ListResourcesOfUserUnderActionPath, uid, uAlias, service, aAlias)
	if err = c.client.GetCall(l, &output, reqPath); err != nil {
		return
	}
	out = output.Data
	return
}

func (c *Client) ListResourcesOfUserGroupUnderAction(l rpc.Logger, uid uint32, gAlias, service, aAlias string) (out *ListResourceOutput, err error) {
	var output struct {
		CommonResponse `json:",inline"`
		Data           *ListResourceOutput `json:"data"`
	}
	reqPath := fmt.Sprintf(ListResourcesOfUserGroupUnderActionPath, uid, gAlias, service, aAlias)
	if err = c.client.GetCall(l, &output, reqPath); err != nil {
		return
	}
	out = output.Data
	return
}

func (c *Client) CheckUserPermissions(l rpc.Logger, uid uint32, alias string, params *CheckPermissionInput) (out *CheckPermissionOutput, err error) {
	var output struct {
		CommonResponse `json:",inline"`
		Data           *CheckPermissionOutput `json:"data"`
	}
	reqPath := fmt.Sprintf(CheckUserPermissionsPath, uid, alias)
	if err = c.client.CallWithJson(l, &output, reqPath, params); err != nil {
		return
	}
	out = output.Data
	return
}

func (c *Client) CheckUserGroupPermissions(l rpc.Logger, uid uint32, alias string, params *CheckPermissionInput) (out *CheckPermissionOutput, err error) {
	var output struct {
		CommonResponse `json:",inline"`
		Data           *CheckPermissionOutput `json:"data"`
	}
	reqPath := fmt.Sprintf(CheckUserGroupPermissionsPath, uid, alias)
	if err = c.client.CallWithJson(l, &output, reqPath, params); err != nil {
		return
	}
	out = output.Data
	return
}

func (c *Client) ListAvailableServiceOfUser(l rpc.Logger, uid uint32, alias string) (out []string, err error) {
	var output struct {
		CommonResponse `json:",inline"`
		Data           []string `json:"data"`
	}
	reqPath := fmt.Sprintf(ListAvailableServiceOfUserPath, uid, alias)
	if err = c.client.GetCall(l, &output, reqPath); err != nil {
		return
	}
	out = output.Data
	return
}
