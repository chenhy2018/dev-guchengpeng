package foosvr

import (
	"github.com/qiniu/http/httputil.v1"
	"github.com/qiniu/xlog.v1"

	"labix.org/v2/mgo"

	account "qbox.us/http/account.v2"
)

var ErrInvalidArgs = httputil.NewError(400, "invalid arguments")

// ------------------------------------------------------------------------

type Config struct {
	Coll       *mgo.Collection
	AuthParser account.AuthParser
}

type Service struct {
	account.Manager
	Config
}

// ------------------------------------------------------------------------

func New(cfg *Config) (p *Service, err error) {

	p = &Service{Config: *cfg}
	p.InitAccount(cfg.AuthParser)
	return
}

// ------------------------------------------------------------------------

type fooArgs struct {
	Mode int    `flag:"_"`
	Arg1 string `flag:"arg1"`
	Arg2 string `flag:"arg2,base64"`
}

type fooRet struct {
	Ret1 int    `json:"ret1"`
	Ret2 string `json:"ret2,omitempty"`
}

/*
	POST /foo/<Mode>/arg1/<Arg1>/arg2/<Base64EncodedArg2>
	这里假设是公开的服务，需要进行授权验证。
	如果是非公开的内部服务（监听内网端口，那么改为env *wsrpc.Env；或者干脆不要 env 参数，如果同时也不需要写xlog日志的话）
*/
func (p *Service) CmdFoo_(args *fooArgs, env *account.Env) (ret fooRet, err error) {

	log := xlog.New(env.W, env.Req)

	if args.Mode == 0 {
		log.Warn("Foo: invalidargs -", *args)
		err = ErrInvalidArgs
		return
	}

	ret.Ret1 = 1
	ret.Ret2 = "hello, foo"
	return
}

// ------------------------------------------------------------------------

type barArgs struct {
	Mode int    `json:"mode"`
	Arg1 string `json:"arg1"`
	Arg2 string `json:"arg2"`
}

type barRet struct {
	Ret1 int    `json:"ret1"`
	Ret2 string `json:"ret2,omitempty"`
}

/*
	POST /bar?mode=<Mode>&arg1=<Arg1>&arg2=<Arg2>
*/
func (p *Service) WspBar(args *barArgs, env *account.Env) (ret barRet, err error) {

	if args.Mode == 0 {
		err = ErrInvalidArgs
		return
	}

	ret.Ret1 = 1
	ret.Ret2 = "hello, bar"
	return
}

// ------------------------------------------------------------------------
