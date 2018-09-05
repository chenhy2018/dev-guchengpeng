package gobrpc

import (
	"fmt"
	"github.com/qiniu/ts"
	"net/http"
	"net/http/httptest"
	"qbox.us/api"
	"qbox.us/errors"
	"testing"
)

// ---------------------------------------------------------------------------

type Service struct {
}

type Args struct {
	A int
	B int
}

func (r *Service) Add(args *Args, ret *int) error {
	*ret = args.A + args.B
	return nil
}

func (r *Service) Sub(args *Args, ret *int) error {
	return errors.Info(api.ENotImpl, "Service.Sub", args)
}

// ---------------------------------------------------------------------------

func server() *httptest.Server {

	rpc := DefaultServer
	arith := new(Service)
	rpc.Register(arith)
	mux := http.NewServeMux()
	mux.Handle("/rpc/", rpc)
	svr := httptest.NewServer(mux)
	return svr
}

func TestGobRpc(t *testing.T) {

	svr := server()
	defer svr.Close()
	svrUrl := svr.URL

	client := NewClient(svrUrl + "/rpc/")

	args := &Args{7, 8}
	var reply int
	for i := 0; i < 10; i++ {
		err := client.Call("Service.Add", args, &reply)
		if err != nil {
			ts.Fatal(t, "arith error:", err)
		}
		fmt.Printf("Add(%d, %d) => %d\n", args.A, args.B, reply)
	}

	err := client.Call("Service.Sub", args, &reply)
	if err == nil || errors.Err(err) != api.ENotImpl {
		ts.Fatal(t, "Service.Sub: arith error:", err, api.ENotImpl)
	}
	fmt.Println("Service.Sub ret:", err)
}

// ---------------------------------------------------------------------------
