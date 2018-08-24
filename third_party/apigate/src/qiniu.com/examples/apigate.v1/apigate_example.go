package apigate_example

import (
	"github.com/qiniu/http/httputil.v1"
	"qiniu.com/auth/authstub.v1"
)

type fooInfo struct {
	Foo string `json:"foo"`
	A   string `json:"a"'`
	B   string `json:"b"`
	Id  string `json:"id"`
}

// ---------------------------------------------------------------------------

type Config struct {
}

type Service struct {
	foos map[string]fooInfo
}

func New(cfg *Config) (p *Service, err error) {

	p = &Service{
		foos: make(map[string]fooInfo),
	}
	return
}

// ---------------------------------------------------------------------------

type fooBarArgs struct {
	CmdArgs []string
	A       string `json:"a"'`
	B       string `json:"b"`
}

type fooBarRet struct {
	Id string `json:"id"`
}

/*
请求包：

```
POST /foo/<FooArg>/bar
Content-Type: application/json

{a: <A>, b: <B>}
```

返回包：

```
200 OK

{id: <FooId>}
```
*/
func (p *Service) PostFoo_Bar(args *fooBarArgs, env *authstub.Env) (ret fooBarRet, err error) {

	id := args.A + "." + args.B
	p.foos[id] = fooInfo{
		Foo: args.CmdArgs[0],
		A: args.A,
		B: args.B,
		Id: id,
	}
	return fooBarRet{Id: id}, nil
}

// ---------------------------------------------------------------------------

type reqArgs struct {
	CmdArgs []string
}

/*
请求包：

```
GET /foo/<FooId>
```

返回包：

```
200 OK

{a: <A>, b: <B>, foo: <Foo>, id: <FooId>}
```
*/
func (p *Service) GetFoo_(args *reqArgs, env *authstub.Env) (ret fooInfo, err error) {

	id := args.CmdArgs[0]
	if foo, ok := p.foos[id]; ok {
		return foo, nil
	}
	err = httputil.NewError(404, "id not found")
	return
}

// ---------------------------------------------------------------------------

