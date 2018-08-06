package keystone

import (
	"github.com/qiniu/rpc.v2"
	"ustack.com/admin_api.v1/ustack"
)

// --------------------------------------------------

type Client struct {
	Conn ustack.Conn
}

func New(services ustack.Services) Client {

	conn, ok := services.Find("keystone.v3.us-internal")
	if !ok {
		panic("keystone.v3.us-internal api not found")
	}
	return Client{conn}
}

func fakeError(err error) bool {

	if rpc.HttpCodeOf(err)/100 == 2 {
		return true
	}
	return false
}

// --------------------------------------------------
// 注册新用户

type RegisterInfo struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Enabled  bool   `json:"enabled"`
}

type RegisteredUser struct {
	Id        string `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Enabled   bool   `json:"enabled"`
	DomainId  string `json:"domain_id"`
	ProjectId string `json:"default_project_id"`
}

type registerArgs struct {
	Info *RegisterInfo `json:"register_info"`
}

type registerRet struct {
	User RegisteredUser `json:"user"`
}

func (p Client) Register(l rpc.Logger, info *RegisterInfo) (user RegisteredUser, err error) {

	var ret registerRet
	err = p.Conn.CallWithJson(l, &ret, "POST", "/register", registerArgs{info})
	if err != nil && fakeError(err) {
		err = nil
	}
	if err == nil {
		user = ret.User
	}
	return
}

// --------------------------------------------------
// 重置用户密码

type user struct {
	Id          string `json:"id"`
	NewPassword string `json:"new_password"`
}

type resetPasswdArgs struct {
	User user `json:"user"`
}

func (p Client) RestPassword(l rpc.Logger, id, newPasswd string) (err error) {

	type resetPasswdArgs struct {
		User struct {
			Id          string `json:"id"`
			NewPassword string `json:"new_password"`
		} `json:"user"`
	}

	var args resetPasswdArgs
	args.User.Id = id
	args.User.NewPassword = newPasswd

	err = p.Conn.CallWithJson(l, nil, "POST", "/reset_password", args)
	if err != nil && fakeError(err) {
		err = nil
	}
	return
}

// --------------------------------------------------
// 查询用户信息

func (p Client) GetUser(l rpc.Logger, email string) (user RegisteredUser, err error) {

	type getUserArgs struct {
		Query struct {
			Email string `json:"email"`
		} `json:"query"`
	}

	var args getUserArgs
	args.Query.Email = email

	err = p.Conn.CallWithJson(l, &user, "POST", "/get_users", args)
	if err != nil && fakeError(err) {
		err = nil
	}
	return
}

// --------------------------------------------------
