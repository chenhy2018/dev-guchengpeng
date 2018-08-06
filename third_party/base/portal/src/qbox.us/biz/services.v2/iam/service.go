package iam

import (
	rpc "github.com/qiniu/rpc.v1"
	"qbox.us/iam/api"
	"qbox.us/iam/entity"
	"qbox.us/oauth"
)

// Service iam 服务接口定义
type Service interface {
	// user
	CreateUser(l rpc.Logger, rootUID uint32, params *api.CreateUserInput) (user *entity.User, err error)
	ListUsers(l rpc.Logger, rootUID uint32, query *api.ListUsersInput) (out *api.ListUsersOutput, err error)
	GetUser(l rpc.Logger, rootUID uint32, alias string) (user *entity.User, err error)
	UpdateUser(l rpc.Logger, rootUID uint32, alias string, params *api.UpdateUserInput) (user *entity.User, err error)
	DeleteUser(l rpc.Logger, rootUID uint32, alias string) (err error)
	CheckUserPassword(l rpc.Logger, rootUID uint32, alias string, params *api.CheckUserPasswordInput) (out *api.CheckUserPasswordOutput, err error)

	// key pair
	ListUserKeypairs(l rpc.Logger, rootUID uint32, alias string, query *api.ListUserKeypairsInput) (out *api.ListUserKeypairsOutput, err error)
	CreateUserKeypair(l rpc.Logger, rootUID uint32, alias string) (keypair *entity.Keypair, err error)
	DeleteUserKeypair(l rpc.Logger, rootUID uint32, alias string, accessKey string) (err error)
	EnableUserKeypair(l rpc.Logger, rootUID uint32, alias string, accessKey string) (err error)
	DisableUserKeypair(l rpc.Logger, rootUID uint32, alias string, accessKey string) (err error)

	// user group
	CreateUserGroup(l rpc.Logger, rootUID uint32, params *api.CreateUserGroupInput) (userGroup *entity.UserGroup, err error)
	ListUserGroups(l rpc.Logger, rootUID uint32, query *api.ListUserGroupsInput) (out *api.ListUserGroupsOutput, err error)
	ListUsersForGroup(l rpc.Logger, rootUID uint32, alias string, query *api.ListUsersForGroupInput) (out *api.ListUsersOutput, err error)
	GetUserGroup(l rpc.Logger, rootUID uint32, alias string) (userGroup *entity.UserGroup, err error)
	UpdateUserGroup(l rpc.Logger, rootUID uint32, alias string, params *api.UpdateUserGroupInput) (userGroup *entity.UserGroup, err error)
	DeleteUserGroup(l rpc.Logger, rootUID uint32, alias string) (err error)
	AddUserToUserGroup(l rpc.Logger, rootUID uint32, alias string, params *api.AddUserToUserGroupInput) (err error)
	RemoveUserFromUserGroup(l rpc.Logger, rootUID uint32, alias string, params *api.RemoveUserFromUserGroupInput) (err error)
	ReassignUserGroupForUser(l rpc.Logger, rootUID uint32, alias string, params *api.ReassignUserGroupForUserInput) (err error)
	ListGroupsForUser(l rpc.Logger, rootUID uint32, alias string, query *api.ListGroupsForUserInput) (out *api.ListUserGroupsOutput, err error)
	ReassignUserForUserGroup(l rpc.Logger, rootUID uint32, alias string, params *api.ReassignUserForUserGroupInput) (err error)

	// policy
	CreatePolicy(l rpc.Logger, rootUID uint32, params *api.CreatePolicyInput) (policy *entity.Policy, err error)
	ListPolicies(l rpc.Logger, rootUID uint32, query *api.ListPoliciesInput) (out *api.ListPoliciesOutput, err error)
	ListAllPolicies(l rpc.Logger, rootUID uint32, query *api.ListPoliciesInput) (out *api.ListPoliciesOutput, err error)
	ListUserForPolicy(l rpc.Logger, rootUID uint32, alias string, query *api.ListUserForPolicyInput) (out *api.ListUsersOutput, err error)
	ListUserGroupForPolicy(l rpc.Logger, rootUID uint32, alias string, query *api.ListUserGroupForPolicyInput) (out *api.ListUserGroupsOutput, err error)
	GetPolicy(l rpc.Logger, rootUID uint32, alias string) (policy *entity.Policy, err error)
	UpdatePolicy(l rpc.Logger, rootUID uint32, alias string, params *api.UpdatePolicyInput) (policy *entity.Policy, err error)
	DeletePolicy(l rpc.Logger, rootUID uint32, alias string) (err error)
	ListPolicyForUser(l rpc.Logger, rootUID uint32, alias string, query *api.ListPolicyForUserInput) (out *api.ListPoliciesOutput, err error)
	AttachPolicyToUser(l rpc.Logger, rootUID uint32, alias string, params *api.AttachPolicyToUserInput) (err error)
	DetachPolicyFromUser(l rpc.Logger, rootUID uint32, alias string, params *api.DetachPolicyFromUserInput) (err error)
	ReassignPolicyForUser(l rpc.Logger, rootUID uint32, alias string, params *api.ReassignPolicyForUserInput) (err error)
	ListPolicyForUserGroup(l rpc.Logger, rootUID uint32, alias string, query *api.ListPolicyForUserGroupInput) (out *api.ListPoliciesOutput, err error)
	AttachPolicyToUserGroup(l rpc.Logger, rootUID uint32, alias string, params *api.AttachPolicyToUserGroupInput) (err error)
	DetachPolicyFromUserGroup(l rpc.Logger, rootUID uint32, alias string, params *api.DetachPolicyFromUserGroupInput) (err error)
	ReassignPolicyForUserGroup(l rpc.Logger, rootUID uint32, alias string, params *api.ReassignPolicyForUserGroupInput) (err error)
	ReassignUserForPolicy(l rpc.Logger, rootUID uint32, alias string, params *api.ReassignUserForPolicyInput) (err error)
	ReassignUserGroupForPolicy(l rpc.Logger, rootUID uint32, alias string, params *api.ReassignUserGroupForPolicyInput) (err error)
	ListSystemPolicies(l rpc.Logger, query *api.ListPoliciesInput) (out *api.ListPoliciesOutput, err error)
	GetSystemPolicy(l rpc.Logger, alias string) (policy *entity.Policy, err error)

	// root user
	ListRootUsers(l rpc.Logger, query *api.ListRootUsersInput) (out *api.ListRootUsersOutput, err error)
	GetRootUser(l rpc.Logger, uid uint32) (out *entity.RootUser, err error)
	EnableRootUser(l rpc.Logger, uid uint32, query *api.EnableRootUserInput) (rootUser *entity.RootUser, err error)
	DisableRootUser(l rpc.Logger, uid uint32) (err error)
	UpdateRootUser(l rpc.Logger, uid uint32, param *api.UpdateRootUserInput) (rootUser *entity.RootUser, err error)

	// audit log
	ListAuditLogs(l rpc.Logger, rootUID uint32, query *api.ListAuditLogsInput) (out *api.ListAuditLogsOutput, err error)
	PostAuditLogs(l rpc.Logger, logs []api.PostAuditLogInput) (int, error)

	// action
	ListActions(l rpc.Logger, query *api.ListActionsInput) (out *api.ListActionsOutput, err error)
	ListServices(l rpc.Logger) (out []string, err error)

	// permission
	ListResourcesOfUserUnderAction(l rpc.Logger, uid uint32, uAlias, service, aAlias string) (*api.ListResourceOutput, error)
	ListResourcesOfUserGroupUnderAction(l rpc.Logger, uid uint32, gAlias, service, aAlias string) (*api.ListResourceOutput, error)
	CheckUserPermissions(l rpc.Logger, uid uint32, alias string, params *api.CheckPermissionInput) (*api.CheckPermissionOutput, error)
	CheckUserGroupPermissions(l rpc.Logger, uid uint32, alias string, params *api.CheckPermissionInput) (*api.CheckPermissionOutput, error)
	ListAvailableServiceOfUser(l rpc.Logger, uid uint32, alias string) ([]string, error)
}

func NewService(hosts []string, adminOAuth *oauth.Transport) Service {
	return api.NewWithMultiHosts(hosts, adminOAuth)
}
