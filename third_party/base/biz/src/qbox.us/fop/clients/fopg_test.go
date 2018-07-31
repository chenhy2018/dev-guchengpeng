package clients

import (
	"bytes"
	"github.com/qiniu/xlog.v1"
	"io"
	"net/http"
	"net/http/httptest"
	"qbox.us/digest_auth"
	"qbox.us/fop/mock"
	"qbox.us/mockacc"
	"testing"
)

const (
	accessKey = "4_odedBxmrAHiu4Y0Qp0HPG0NANCf6VAsAjWL_k9"
	secretKey = "SrRuUVfDX6drVRvpyN8mv8Vcm9XnMZzlbDfvVfMe"
)

var svrUrlOne string
var svrURlTwo string

func doTestOp(t *testing.T) {
	transport := digest_auth.NewTransport(accessKey, secretKey, http.DefaultTransport)
	hosts := []string{
		"http://127.0.0.1:4460",
		svrUrlOne,
		svrURlTwo,
		"http://127.0.0.1:4463",
	}
	client := NewFopg(hosts, 3, transport)
	r, length, _, _, err := client.Op(xlog.NewDummy(), []byte("testfh"), 100, nil, false)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	b := new(bytes.Buffer)
	io.CopyN(b, r, length)
	if b.String() != "mockfopgate.Op" {
		t.Error(b.String())
	}
}

func doTestNop(t *testing.T) {
	transport := digest_auth.NewTransport(accessKey, secretKey, http.DefaultTransport)
	hosts := []string{
		"http://127.0.0.1:4460",
		svrUrlOne,
		svrURlTwo,
		"http://127.0.0.1:4463",
	}
	client := NewFopg(hosts, 4, transport)
	r, length, _, _, err := client.Op(xlog.NewDummy(), []byte("testfh"), 100, nil, true)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	b := new(bytes.Buffer)
	io.CopyN(b, r, length)
	if b.String() != "mockfopgate.Nop" {
		t.Error(b.String())
	}
}

func doTestRepeat(t *testing.T) {
	transport := digest_auth.NewTransport(accessKey, secretKey, http.DefaultTransport)
	hosts := []string{
		"http://127.0.0.1:4460",
		svrUrlOne,
		svrURlTwo,
		"http://127.0.0.1:4463",
	}
	client := NewFopg(hosts, 3, transport)
	for i := 0; i < 5; i++ {
		_, _, _, _, err := client.Op(xlog.NewDummy(), []byte("testfs"), 100, nil, false)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func doTestFail(t *testing.T) {
	transport := digest_auth.NewTransport(accessKey, secretKey, http.DefaultTransport)
	hosts := []string{
		"http://127.0.0.1:14460",
		"http://127.0.0.1:14461",
		"http://127.0.0.1:14462",
		"http://127.0.0.1:14463",
	}
	client := NewFopg(hosts, 2, transport)
	_, _, _, _, err := client.Op(xlog.NewDummy(), []byte("testfs"), 100, nil, false)
	if err == nil {
		t.Error(err)
	}

}

func runServer(sa mockacc.SimpleAccount) (svr *httptest.Server) {
	mux := http.NewServeMux()
	mockacc.RegisterHandlers(mux, sa)
	server := mock.NewFopg(mockacc.Account{})
	server.RegisterHandlers(mux)
	svr = httptest.NewServer(mux)
	return
}

func _TestFopg(t *testing.T) { // TODO: testcase fix

	sa := mockacc.SaInstance
	svrOne := runServer(sa)
	svrTwo := runServer(sa)
	svrUrlOne = svrOne.URL
	svrURlTwo = svrTwo.URL
	defer svrOne.Close()
	defer svrTwo.Close()

	doTestOp(t)
	doTestNop(t)
	doTestRepeat(t)
	doTestFail(t)
}
