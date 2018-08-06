package api

import (
	"fmt"

	"github.com/qiniu/rpc.v1"
	"qbox.us/iam/entity"
	"qbox.us/iam/enums"
)

const (
	CreateUserGroupPath          = "/iam/u/%d/groups"
	ListUserGroupsPath           = "/iam/u/%d/groups?%s"
	ListUsersForGroupPath        = "/iam/u/%d/groups/%s/users?%s"
	GetUserGroupPath             = "/iam/u/%d/groups/%s"
	UpdateUserGroupPath          = "/iam/u/%d/groups/%s"
	DeleteUserGroupPath          = "/iam/u/%d/groups/%s"
	AddUserToUserGroupPath       = "/iam/u/%d/groups/%s/users"
	RemoveUserFromUserGroupPath  = "/iam/u/%d/groups/%s/users"
	ReassignUserGroupForUserPath = "/iam/u/%d/users/%s/groups"
	ListGroupsForUserPath        = "/iam/u/%d/users/%s/groups?%s"
	ReassignUserForUserGroupPath = "/iam/u/%d/groups/%s/users"
)

type CreateUserGroupInput struct {
	Alias       string `json:"alias"`
	Description string `json:"description"`
}

type ListUserGroupsInput struct {
	Paginator
	Alias string
}

func (i *ListUserGroupsInput) GetQueryString() string {
	q := i.GetQuery()
	q.Set("alias", i.Alias)
	return q.Encode()
}

type ListUserGroupsOutput struct {
	List  []*entity.UserGroup `json:"list"`
	Count int                 `json:"count"`
}

type ListUsersForGroupInput struct {
	Paginator
	UserAlias string
}

func (i *ListUsersForGroupInput) GetQueryString() string {
	q := i.GetQuery()
	q.Set("user_alias", i.UserAlias)
	return q.Encode()
}

type UpdateUserGroupInput struct {
	SpoofMethod
	Alias       *string `json:"alias"`
	Description *string `json:"description"`
}

type AddUserToUserGroupInput struct {
	SpoofMethod `json:",inline"`
	UserAliases []string `json:"user_aliases"`
}

type RemoveUserFromUserGroupInput struct {
	SpoofMethod `json:",inline"`
	UserAliases []string `json:"user_aliases"`
}

type ReassignUserForUserGroupInput struct {
	UserAliases []string `json:"user_aliases"`
}

type ReassignUserGroupForUserInput struct {
	GroupAliases []string `json:"group_aliases"`
}

type ListGroupsForUserInput struct {
	Paginator
	GroupAlias string
}

func (i *ListGroupsForUserInput) GetQueryString() string {
	q := i.GetQuery()
	q.Set("group_alias", i.GroupAlias)
	return q.Encode()
}

func (c *Client) CreateUserGroup(l rpc.Logger, rootUID uint32, params *CreateUserGroupInput) (userGroup *entity.UserGroup, err error) {
	var output struct {
		CommonResponse `json:",inline"`
		Data           *entity.UserGroup `json:"data"`
	}
	reqPath := fmt.Sprintf(CreateUserGroupPath, rootUID)
	if err = c.client.CallWithJson(l, &output, reqPath, params); err != nil {
		return
	}
	userGroup = output.Data
	return
}

func (c *Client) ListUserGroups(l rpc.Logger, rootUID uint32, query *ListUserGroupsInput) (out *ListUserGroupsOutput, err error) {
	var output struct {
		CommonResponse `json:",inline"`
		Data           *ListUserGroupsOutput `json:"data"`
	}
	reqPath := fmt.Sprintf(ListUserGroupsPath, rootUID, query.GetQueryString())
	if err = c.client.GetCall(l, &output, reqPath); err != nil {
		return
	}
	out = output.Data
	return
}

func (c *Client) ListUsersForGroup(l rpc.Logger, rootUID uint32, alias string, query *ListUsersForGroupInput) (out *ListUsersOutput, err error) {
	var output struct {
		CommonResponse `json:",inline"`
		Data           *ListUsersOutput `json:"data"`
	}
	reqPath := fmt.Sprintf(ListUsersForGroupPath, rootUID, alias, query.GetQueryString())
	if err = c.client.GetCall(l, &output, reqPath); err != nil {
		return
	}
	out = output.Data
	return
}

func (c *Client) GetUserGroup(l rpc.Logger, rootUID uint32, alias string) (userGroup *entity.UserGroup, err error) {
	var output struct {
		CommonResponse `json:",inline"`
		Data           *entity.UserGroup `json:"data"`
	}
	reqPath := fmt.Sprintf(GetUserGroupPath, rootUID, alias)
	if err = c.client.GetCall(l, &output, reqPath); err != nil {
		return
	}
	userGroup = output.Data
	return
}

func (c *Client) UpdateUserGroup(l rpc.Logger, rootUID uint32, alias string, params *UpdateUserGroupInput) (userGroup *entity.UserGroup, err error) {
	var output struct {
		CommonResponse `json:",inline"`
		Data           *entity.UserGroup `json:"data"`
	}
	params.Method = enums.SpoofMethodPatch
	reqPath := fmt.Sprintf(UpdateUserGroupPath, rootUID, alias)
	if err = c.client.CallWithJson(l, &output, reqPath, params); err != nil {
		return
	}
	userGroup = output.Data
	return
}

func (c *Client) DeleteUserGroup(l rpc.Logger, rootUID uint32, alias string) (err error) {
	var output CommonResponse
	reqPath := fmt.Sprintf(DeleteUserGroupPath, rootUID, alias)
	if err = c.client.DeleteCall(l, &output, reqPath); err != nil {
		return
	}
	return
}

func (c *Client) AddUserToUserGroup(l rpc.Logger, rootUID uint32, alias string, params *AddUserToUserGroupInput) (err error) {
	var output CommonResponse
	params.Method = enums.SpoofMethodPatch
	reqPath := fmt.Sprintf(AddUserToUserGroupPath, rootUID, alias)
	if err = c.client.CallWithJson(l, &output, reqPath, params); err != nil {
		return
	}
	return
}

func (c *Client) RemoveUserFromUserGroup(l rpc.Logger, rootUID uint32, alias string, params *RemoveUserFromUserGroupInput) (err error) {
	var output CommonResponse
	params.Method = enums.SpoofMethodDELETE
	reqPath := fmt.Sprintf(RemoveUserFromUserGroupPath, rootUID, alias)
	if err = c.client.CallWithJson(l, &output, reqPath, params); err != nil {
		return
	}
	return
}

func (c *Client) ReassignUserGroupForUser(l rpc.Logger, rootUID uint32, alias string, params *ReassignUserGroupForUserInput) (err error) {
	var output CommonResponse
	reqPath := fmt.Sprintf(ReassignUserGroupForUserPath, rootUID, alias)
	if err = c.client.CallWithJson(l, &output, reqPath, params); err != nil {
		return
	}
	return
}

func (c *Client) ListGroupsForUser(l rpc.Logger, rootUID uint32, alias string, query *ListGroupsForUserInput) (out *ListUserGroupsOutput, err error) {
	var output struct {
		CommonResponse `json:",inline"`
		Data           *ListUserGroupsOutput `json:"data"`
	}
	reqPath := fmt.Sprintf(ListGroupsForUserPath, rootUID, alias, query.GetQueryString())
	if err = c.client.GetCall(l, &output, reqPath); err != nil {
		return
	}
	out = output.Data
	return
}

func (c *Client) ReassignUserForUserGroup(l rpc.Logger, rootUID uint32, alias string, params *ReassignUserForUserGroupInput) (err error) {
	var output CommonResponse
	reqPath := fmt.Sprintf(ReassignUserForUserGroupPath, rootUID, alias)
	if err = c.client.CallWithJson(l, &output, reqPath, params); err != nil {
		return
	}
	return
}
