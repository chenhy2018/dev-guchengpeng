package services

const (
	OplogActionSignin         OplogAction = "signin"
	OplogActionPermCreate     OplogAction = "perm:create"
	OplogActionPermUpdate     OplogAction = "perm:update"
	OplogActionRoleCreate     OplogAction = "role:create"
	OplogActionRoleUpdate     OplogAction = "role:update"
	OplogActionUserPermCreate OplogAction = "user:perm:create"
	OplogActionUserPermUpdate OplogAction = "user:perm:update"
	OplogActionUserPermRemove OplogAction = "user:perm:remove"
	OplogActionUserAssignRole OplogAction = "user:perm:assignRole"
	OplogActionUserSuSignIn   OplogAction = "user:su:signin" //模拟登陆
)

type OplogAction string
