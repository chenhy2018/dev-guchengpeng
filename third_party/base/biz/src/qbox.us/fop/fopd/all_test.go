package fopd_test

import (
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"qbox.us/auditlog2"
	"qbox.us/fop"
	"qbox.us/fop/clients/fopd.v2"
	. "qbox.us/fop/fopd"

	"code.google.com/p/go.net/context"
	"github.com/qiniu/log.v1"
	"github.com/qiniu/xlog.v1"
	"github.com/stretchr/testify/assert"
)

var (
	bindHost = ":43001"
)

func runFopd() string {
	cmdA := func(w fop.Writer, r fop.Reader) {
		io.Copy(w, r.Source)
	}
	cmdB := func(w fop.Writer, r fop.Reader) {
		srcURL := r.SourceURL
		resp, _ := http.Get(srcURL)
		defer resp.Body.Close()
		io.Copy(w, resp.Body)
	}
	fops := map[string]func(w fop.Writer, r fop.Reader){
		"a": cmdA,
		"b": cmdB,
	}
	cfg := &Config{
		Fops:   fops,
		LogCfg: auditlog2.Config{LogFile: os.TempDir()},
	}
	fopdservice := New(cfg)
	mux := http.NewServeMux()
	_, err := fopdservice.RegisterHandlers(mux, "")
	if err != nil {
		log.Fatalln(err)
	}
	svr := httptest.NewServer(mux)
	return svr.URL
}

func runFileServer() string {
	handler := func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte("helloworld"))
	}
	svr := httptest.NewServer(http.HandlerFunc(handler))
	return svr.URL
}

func assertResponse(t *testing.T, resp *http.Response, expectedData string) {
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, string(b), expectedData)
}

func TestFopd(t *testing.T) {
	fopdHost := runFopd()
	srcURL := runFileServer()
	tpCtx := xlog.NewContextWith(context.Background(), "TestFopd")

	conn := fopd.NewConn(fopdHost, nil)
	ctx := &fop.FopCtx{}

	ctx.RawQuery = "a"
	resp, err := conn.Op(tpCtx, strings.NewReader("helloworld"), 10, ctx)
	assert.NoError(t, err)
	assertResponse(t, resp, "helloworld")

	ctx.RawQuery = "b"
	ctx.SourceURL = srcURL
	resp, err = conn.Op2(tpCtx, []byte("fh"), 10, ctx)
	assert.NoError(t, err)
	assertResponse(t, resp, "helloworld")

	ctx.CmdName = "b"
	ctx.RawQuery = "z"
	ctx.SourceURL = srcURL
	resp, err = conn.Op2(tpCtx, []byte("fh"), 10, ctx)
	assert.NoError(t, err)
	assertResponse(t, resp, "helloworld")

	ctx.URL = "http://www.qiniu.com"
	resp, err = conn.Op2(tpCtx, []byte("fh"), 10, ctx)
	assert.Equal(t, 400, resp.StatusCode)
}
