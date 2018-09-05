package main

import (
	"log"
	"net/http"

	"github.com/qiniu/http/httputil.v1"
	"github.com/qiniu/http/jsonrpc.v1"
	"github.com/qiniu/http/webroute.v1"
	"github.com/qiniu/http/wsrpc.v1"
	"qbox.us/example/demosvr/account"
)

// ------------------------------------------------------------------------

type Service struct {
	account.Manager
}

// ------------------------------------------------------------------------

/*
POST /double
Content-Type: application/json
123

200 OK
Content-Type: application/json
246
*/
func (p *Service) RpcDouble(v int) (int, error) {
	return v * 2, nil
}

// ------------------------------------------------------------------------

type fooArgs struct {
	A string `json:"a"`
	B int    `json:"b"`
}

type fooRet struct {
	C   string `json:"c"`
	Uid uint   `json:"uid"`
}

/*
POST /foo?a=$A&b=$B
Authorization: <Token>

200 OK
Content-Type: application/json
{c: $C, d: $D}
*/
func (p *Service) WsFoo(args *fooArgs, env *account.Env) (ret fooRet, err error) {

	if args.A == "" {
		err = httputil.NewError(400, "invalid argument 'a'")
		return
	}
	return fooRet{args.A, env.Uid}, nil
}

// ------------------------------------------------------------------------

func main() {

	service := new(Service)

	router := &webroute.Router{Factory: wsrpc.Factory.Union(jsonrpc.Factory)}
	router.Register(service)

	log.Fatal(http.ListenAndServe(":8888", nil))
}

// ------------------------------------------------------------------------
