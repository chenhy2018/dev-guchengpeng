package api

import (
	"fmt"

	"github.com/qiniu/rpc.v1"
	"qbox.us/iam/entity"
	"qbox.us/iam/enums"
)

const (
	CreatePolicyPath               = "/iam/u/%d/policies"
	ListPoliciesPath               = "/iam/u/%d/policies?%s"
	ListAllPoliciesPath            = "/iam/u/%d/allpolicies?%s"
	ListUserForPolicyPath          = "/iam/u/%d/policies/%s/users?%s"
	ListUserGroupForPolicyPath     = "/iam/u/%d/policies/%s/groups?%s"
	GetPolicyPath                  = "/iam/u/%d/policies/%s"
	UpdatePolicyPath               = "/iam/u/%d/policies/%s"
	DeletePolicyPath               = "/iam/u/%d/policies/%s"
	ListPolicyForUserPath          = "/iam/u/%d/users/%s/policies?%s"
	AttachPolicyToUserPath         = "/iam/u/%d/users/%s/policies"
	DetachPolicyFromUserPath       = "/iam/u/%d/users/%s/policies"
	ReassignPolicyForUserPath      = "/iam/u/%d/users/%s/policies"
	ListPolicyForUserGroupPath     = "/iam/u/%d/groups/%s/policies?%s"
	AttachPolicyToUserGroupPath    = "/iam/u/%d/groups/%s/policies"
	DetachPolicyFromUserGroupPath  = "/iam/u/%d/groups/%s/policies"
	ReassignPolicyForUserGroupPath = "/iam/u/%d/groups/%s/policies"
	ReassignUserForPolicyPath      = "/iam/u/%d/policies/%s/users"
	ReassignUserGroupForPolicyPath = "/iam/u/%d/policies/%s/groups"
	ListSystemPoliciesPath         = "/iam/policies?%s"
	GetSystemPolicyPath            = "/iam/policies/%s"
)

type CreatePolicyInput struct {
	Alias       string               `json:"alias"`
	Description string               `json:"description"`
	EditType    enums.PolicyEditType `json:"edit_type"`
	Statement   []*entity.Statement  `json:"statement"`
}

type ListPoliciesInput struct {
	Paginator
	Alias string
}

func (i *ListPoliciesInput) GetQueryString() string {
	q := i.GetQuery()
	q.Set("alias", i.Alias)
	return q.Encode()
}

type ListPoliciesOutput struct {
	List  []*entity.Policy `json:"list"`
	Count int              `json:"count"`
}

type ListUserForPolicyInput struct {
	Paginator
	UserAlias string
}

func (i *ListUserForPolicyInput) GetQueryString() string {
	q := i.GetQuery()
	q.Set("user_alias", i.UserAlias)
	return q.Encode()
}

type ListUserGroupForPolicyInput struct {
	Paginator
	GroupAlias string
}

func (i *ListUserGroupForPolicyInput) GetQueryString() string {
	q := i.GetQuery()
	q.Set("group_alias", i.GroupAlias)
	return q.Encode()
}

type UpdatePolicyInput struct {
	SpoofMethod
	Alias       *string              `json:"alias"`
	Description *string              `json:"description"`
	Statement   *[]*entity.Statement `json:"statement"`
}

type ListPolicyForUserInput struct {
	Paginator
	PolicyAlias string
}

func (i *ListPolicyForUserInput) GetQueryString() string {
	q := i.GetQuery()
	q.Set("policy_alias", i.PolicyAlias)
	return q.Encode()
}

type AttachPolicyToUserInput struct {
	SpoofMethod
	PolicyAliases []string `json:"policy_aliases"`
}

type DetachPolicyFromUserInput struct {
	SpoofMethod
	PolicyAliases []string `json:"policy_aliases"`
}

type ReassignPolicyForUserInput struct {
	PolicyAliases []string `json:"policy_aliases"`
}

type ListPolicyForUserGroupInput struct {
	Paginator
	PolicyAlias string
}

func (i *ListPolicyForUserGroupInput) GetQueryString() string {
	q := i.GetQuery()
	q.Set("policy_alias", i.PolicyAlias)
	return q.Encode()
}

type AttachPolicyToUserGroupInput struct {
	SpoofMethod
	PolicyAliases []string `json:"policy_aliases"`
}

type DetachPolicyFromUserGroupInput struct {
	SpoofMethod
	PolicyAliases []string `json:"policy_aliases"`
}

type ReassignPolicyForUserGroupInput struct {
	PolicyAliases []string `json:"policy_aliases"`
}

type ReassignUserForPolicyInput struct {
	UserAliases []string `json:"user_aliases"`
}

type ReassignUserGroupForPolicyInput struct {
	GroupAliases []string `json:"group_aliases"`
}

func (c *Client) CreatePolicy(l rpc.Logger, rootUID uint32, params *CreatePolicyInput) (policy *entity.Policy, err error) {
	var output struct {
		CommonResponse `json:",inline"`
		Data           *entity.Policy `json:"data"`
	}
	reqPath := fmt.Sprintf(CreatePolicyPath, rootUID)
	if err = c.client.CallWithJson(l, &output, reqPath, params); err != nil {
		return
	}
	policy = output.Data
	return
}

func (c *Client) ListPolicies(l rpc.Logger, rootUID uint32, query *ListPoliciesInput) (out *ListPoliciesOutput, err error) {
	var output struct {
		CommonResponse `json:",inline"`
		Data           *ListPoliciesOutput `json:"data"`
	}
	reqPath := fmt.Sprintf(ListPoliciesPath, rootUID, query.GetQueryString())
	if err = c.client.GetCall(l, &output, reqPath); err != nil {
		return
	}
	out = output.Data
	return
}

func (c *Client) ListAllPolicies(l rpc.Logger, rootUID uint32, query *ListPoliciesInput) (out *ListPoliciesOutput, err error) {
	var output struct {
		CommonResponse `json:",inline"`
		Data           *ListPoliciesOutput `json:"data"`
	}
	reqPath := fmt.Sprintf(ListAllPoliciesPath, rootUID, query.GetQueryString())
	if err = c.client.GetCall(l, &output, reqPath); err != nil {
		return
	}
	out = output.Data
	return
}

func (c *Client) ListUserForPolicy(l rpc.Logger, rootUID uint32, alias string, query *ListUserForPolicyInput) (out *ListUsersOutput, err error) {
	var output struct {
		CommonResponse `json:",inline"`
		Data           *ListUsersOutput `json:"data"`
	}
	reqPath := fmt.Sprintf(ListUserForPolicyPath, rootUID, alias, query.GetQueryString())
	if err = c.client.GetCall(l, &output, reqPath); err != nil {
		return
	}
	out = output.Data
	return
}

func (c *Client) ListUserGroupForPolicy(l rpc.Logger, rootUID uint32, alias string, query *ListUserGroupForPolicyInput) (out *ListUserGroupsOutput, err error) {
	var output struct {
		CommonResponse `json:",inline"`
		Data           *ListUserGroupsOutput `json:"data"`
	}
	reqPath := fmt.Sprintf(ListUserGroupForPolicyPath, rootUID, alias, query.GetQueryString())
	if err = c.client.GetCall(l, &output, reqPath); err != nil {
		return
	}
	out = output.Data
	return
}

func (c *Client) GetPolicy(l rpc.Logger, rootUID uint32, alias string) (policy *entity.Policy, err error) {
	var output struct {
		CommonResponse `json:",inline"`
		Data           *entity.Policy `json:"data"`
	}
	reqPath := fmt.Sprintf(GetPolicyPath, rootUID, alias)
	if err = c.client.GetCall(l, &output, reqPath); err != nil {
		return
	}
	policy = output.Data
	return
}

func (c *Client) UpdatePolicy(l rpc.Logger, rootUID uint32, alias string, params *UpdatePolicyInput) (policy *entity.Policy, err error) {
	var output struct {
		CommonResponse `json:",inline"`
		Data           *entity.Policy `json:"data"`
	}
	params.Method = enums.SpoofMethodPatch
	reqPath := fmt.Sprintf(UpdatePolicyPath, rootUID, alias)
	if err = c.client.CallWithJson(l, &output, reqPath, params); err != nil {
		return
	}
	policy = output.Data
	return
}

func (c *Client) DeletePolicy(l rpc.Logger, rootUID uint32, alias string) (err error) {
	var output CommonResponse
	reqPath := fmt.Sprintf(DeletePolicyPath, rootUID, alias)
	if err = c.client.DeleteCall(l, &output, reqPath); err != nil {
		return
	}
	return
}

func (c *Client) ListPolicyForUser(l rpc.Logger, rootUID uint32, alias string, query *ListPolicyForUserInput) (out *ListPoliciesOutput, err error) {
	var output struct {
		CommonResponse `json:",inline"`
		Data           *ListPoliciesOutput `json:"data"`
	}
	reqPath := fmt.Sprintf(ListPolicyForUserPath, rootUID, alias, query.GetQueryString())
	if err = c.client.GetCall(l, &output, reqPath); err != nil {
		return
	}
	out = output.Data
	return
}

func (c *Client) AttachPolicyToUser(l rpc.Logger, rootUID uint32, alias string, params *AttachPolicyToUserInput) (err error) {
	var output CommonResponse
	params.Method = enums.SpoofMethodPatch
	reqPath := fmt.Sprintf(AttachPolicyToUserPath, rootUID, alias)
	if err = c.client.CallWithJson(l, &output, reqPath, params); err != nil {
		return
	}
	return
}

func (c *Client) DetachPolicyFromUser(l rpc.Logger, rootUID uint32, alias string, params *DetachPolicyFromUserInput) (err error) {
	var output CommonResponse
	params.Method = enums.SpoofMethodDELETE
	reqPath := fmt.Sprintf(DetachPolicyFromUserPath, rootUID, alias)
	if err = c.client.CallWithJson(l, &output, reqPath, params); err != nil {
		return
	}
	return
}

func (c *Client) ReassignPolicyForUser(l rpc.Logger, rootUID uint32, alias string, params *ReassignPolicyForUserInput) (err error) {
	var output CommonResponse
	reqPath := fmt.Sprintf(ReassignPolicyForUserPath, rootUID, alias)
	if err = c.client.CallWithJson(l, &output, reqPath, params); err != nil {
		return
	}
	return
}

func (c *Client) ListPolicyForUserGroup(l rpc.Logger, rootUID uint32, alias string, query *ListPolicyForUserGroupInput) (out *ListPoliciesOutput, err error) {
	var output struct {
		CommonResponse `json:",inline"`
		Data           *ListPoliciesOutput `json:"data"`
	}
	reqPath := fmt.Sprintf(ListPolicyForUserGroupPath, rootUID, alias, query.GetQueryString())
	if err = c.client.GetCall(l, &output, reqPath); err != nil {
		return
	}
	out = output.Data
	return
}

func (c *Client) AttachPolicyToUserGroup(l rpc.Logger, rootUID uint32, alias string, params *AttachPolicyToUserGroupInput) (err error) {
	var output CommonResponse
	params.Method = enums.SpoofMethodPatch
	reqPath := fmt.Sprintf(AttachPolicyToUserGroupPath, rootUID, alias)
	if err = c.client.CallWithJson(l, &output, reqPath, params); err != nil {
		return
	}
	return
}

func (c *Client) DetachPolicyFromUserGroup(l rpc.Logger, rootUID uint32, alias string, params *DetachPolicyFromUserGroupInput) (err error) {
	var output CommonResponse
	params.Method = enums.SpoofMethodDELETE
	reqPath := fmt.Sprintf(DetachPolicyFromUserGroupPath, rootUID, alias)
	if err = c.client.CallWithJson(l, &output, reqPath, params); err != nil {
		return
	}
	return
}

func (c *Client) ReassignPolicyForUserGroup(l rpc.Logger, rootUID uint32, alias string, params *ReassignPolicyForUserGroupInput) (err error) {
	var output CommonResponse
	reqPath := fmt.Sprintf(ReassignPolicyForUserGroupPath, rootUID, alias)
	if err = c.client.CallWithJson(l, &output, reqPath, params); err != nil {
		return
	}
	return
}

func (c *Client) ReassignUserForPolicy(l rpc.Logger, rootUID uint32, alias string, params *ReassignUserForPolicyInput) (err error) {
	var output CommonResponse
	reqPath := fmt.Sprintf(ReassignUserForPolicyPath, rootUID, alias)
	if err = c.client.CallWithJson(l, &output, reqPath, params); err != nil {
		return
	}
	return
}

func (c *Client) ReassignUserGroupForPolicy(l rpc.Logger, rootUID uint32, alias string, params *ReassignUserGroupForPolicyInput) (err error) {
	var output CommonResponse
	reqPath := fmt.Sprintf(ReassignUserGroupForPolicyPath, rootUID, alias)
	if err = c.client.CallWithJson(l, &output, reqPath, params); err != nil {
		return
	}
	return
}

func (c *Client) ListSystemPolicies(l rpc.Logger, query *ListPoliciesInput) (out *ListPoliciesOutput, err error) {
	var output struct {
		CommonResponse `json:",inline"`
		Data           *ListPoliciesOutput `json:"data"`
	}
	reqPath := fmt.Sprintf(ListSystemPoliciesPath, query.GetQueryString())
	if err = c.client.GetCall(l, &output, reqPath); err != nil {
		return
	}
	out = output.Data
	return
}

func (c *Client) GetSystemPolicy(l rpc.Logger, alias string) (policy *entity.Policy, err error) {
	var output struct {
		CommonResponse `json:",inline"`
		Data           *entity.Policy `json:"data"`
	}
	reqPath := fmt.Sprintf(GetSystemPolicyPath, alias)
	if err = c.client.GetCall(l, &output, reqPath); err != nil {
		return
	}
	policy = output.Data
	return
}
