package api

import (
	"fmt"
	"time"

	"github.com/qiniu/rpc.v1"
	"qbox.us/iam/entity"
	"qbox.us/iam/enums"
)

const (
	CreateUserPath        = "/iam/u/%d/users"
	ListUsersPath         = "/iam/u/%d/users?%s"
	GetUserPath           = "/iam/u/%d/users/%s"
	UpdateUserPath        = "/iam/u/%d/users/%s"
	DeleteUserPath        = "/iam/u/%d/users/%s"
	CheckUserPasswordPath = "/iam/u/%d/users/%s/password"
)

type CreateUserInput struct {
	Alias    string `json:"alias"`
	Password string `json:"password"`
}

type ListUsersInput struct {
	Paginator
	Alias string
}

func (i *ListUsersInput) GetQueryString() string {
	q := i.GetQuery()
	q.Set("alias", i.Alias)
	return q.Encode()
}

type ListUsersOutput struct {
	List  []*entity.User `json:"list"`
	Count int            `json:"count"`
}

type UpdateUserInput struct {
	SpoofMethod
	Enabled       *bool      `json:"enabled"`
	Password      *string    `json:"password"`
	LastLoginTime *time.Time `json:"last_login_time"`
}

type CheckUserPasswordInput struct {
	Password string `json:"password"`
}
type CheckUserPasswordOutput struct {
	Valid bool `json:"valid"`
}

func (c *Client) CreateUser(l rpc.Logger, rootUID uint32, params *CreateUserInput) (user *entity.User, err error) {
	var output struct {
		CommonResponse `json:",inline"`
		Data           *entity.User `json:"data"`
	}
	reqPath := fmt.Sprintf(CreateUserPath, rootUID)
	if err = c.client.CallWithJson(l, &output, reqPath, params); err != nil {
		return
	}
	user = output.Data
	return
}

func (c *Client) ListUsers(l rpc.Logger, rootUID uint32, query *ListUsersInput) (out *ListUsersOutput, err error) {
	var output struct {
		CommonResponse `json:",inline"`
		Data           *ListUsersOutput `json:"data"`
	}
	reqPath := fmt.Sprintf(ListUsersPath, rootUID, query.GetQueryString())
	if err = c.client.GetCall(l, &output, reqPath); err != nil {
		return
	}
	out = output.Data
	return
}

func (c *Client) GetUser(l rpc.Logger, rootUID uint32, alias string) (user *entity.User, err error) {
	var output struct {
		CommonResponse `json:",inline"`
		Data           *entity.User `json:"data"`
	}
	reqPath := fmt.Sprintf(GetUserPath, rootUID, alias)
	if err = c.client.GetCall(l, &output, reqPath); err != nil {
		return
	}
	user = output.Data
	return
}

func (c *Client) UpdateUser(l rpc.Logger, rootUID uint32, alias string, params *UpdateUserInput) (user *entity.User, err error) {
	var output struct {
		CommonResponse `json:",inline"`
		Data           *entity.User `json:"data"`
	}
	params.Method = enums.SpoofMethodPatch
	reqPath := fmt.Sprintf(UpdateUserPath, rootUID, alias)
	if err = c.client.CallWithJson(l, &output, reqPath, params); err != nil {
		return
	}
	user = output.Data
	return
}

func (c *Client) DeleteUser(l rpc.Logger, rootUID uint32, alias string) (err error) {
	var output CommonResponse
	reqPath := fmt.Sprintf(DeleteUserPath, rootUID, alias)
	if err = c.client.DeleteCall(l, &output, reqPath); err != nil {
		return
	}
	return
}

func (c *Client) CheckUserPassword(l rpc.Logger, rootUID uint32, alias string, params *CheckUserPasswordInput) (out *CheckUserPasswordOutput, err error) {
	var output struct {
		CommonResponse `json:",inline"`
		Data           *CheckUserPasswordOutput `json:"data"`
	}
	reqPath := fmt.Sprintf(CheckUserPasswordPath, rootUID, alias)
	if err = c.client.CallWithJson(l, &output, reqPath, params); err != nil {
		return
	}
	out = output.Data
	return
}
