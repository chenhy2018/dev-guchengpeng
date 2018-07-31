package clients

import (
	"github.com/qiniu/log.v1"
	"github.com/qiniu/xlog.v1"
	"io"
	"net/http/httptest"
	"os"
	"qbox.us/fop"
	"qbox.us/fop/mock"
	"testing"
)

func doTestConn(t *testing.T, url string) {
	Client := NewConn(url, nil)
	fopCtx := &fop.FopCtx{RawQuery: "test/a/a"}
	r, length, _, _, err, _ := Client.Op(xlog.NewDummy(), []byte("fopdtest"), 100, fopCtx, false)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	n, _ := io.Copy(os.Stdout, r)
	log.Println("length:", length)
	if n != length {
		t.Error("n:", n)
	}
}

func Test(t *testing.T) {
	mux := mock.Fopd{}.Mux()
	server := httptest.NewServer(mux)
	defer server.Close()
	doTestConn(t, server.URL)
}
