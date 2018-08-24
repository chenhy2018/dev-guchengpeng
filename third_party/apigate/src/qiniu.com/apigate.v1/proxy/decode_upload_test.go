package proxy

import (
	"bytes"
	"math/rand"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	. "code.google.com/p/go.net/context"
	"github.com/qiniu/api/auth/digest"
	"github.com/qiniu/api/rs"
	"github.com/qiniu/rpc.v1"
	"github.com/stretchr/testify/assert"
	authp "qiniu.com/auth/proto.v1"
)

var (
	ak = "accessKey"
	sk = "secretKey"
)

type mockAcc struct {
}

func (acc mockAcc) GetAccessInfo(ctx Context, accessKey string) (info authp.AccessInfo, err error) {
	info = authp.AccessInfo{
		Secret: []byte(sk),
	}
	return
}

func (acc mockAcc) GetUtype(ctx Context, uid uint32) (utype uint32, err error) {
	return authp.USER_TYPE_STDUSER, nil
}

func TestUpload(t *testing.T) {
	ast := assert.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method == "OPTIONS" || req.Method == "HEAD" {
			return
		}
		req.ParseMultipartForm(0)
		ast.Equal("a", req.FormValue("a"))
		ast.NotEqual("", req.FormValue("stubtoken"))
		return
	}))
	defer ts.Close()

	tp := NewUpTransport(http.DefaultTransport, mockAcc{})
	client := &http.Client{Transport: tp}

	// test: HEAD
	resp, err := client.Head(ts.URL)
	ast.Nil(err)
	ast.Equal(200, resp.StatusCode)

	// test: post invalid multipart
	for i := 0; i < 3; i++ {
		resp, err = client.Post(ts.URL, "application/text", strings.NewReader("ss"))
		ast.Nil(err)
		ast.Equal(400, resp.StatusCode)
		err = rpc.ResponseError(resp)
		resp.Body.Close()
		ast.Equal("request Content-Type isn't multipart/form-data but application/text", err.Error())
	}

	// test: multipart
	policy := &rs.PutPolicy{
		Scope: "test",
	}
	token := policy.Token(&digest.Mac{ak, []byte(sk)})
	b := &bytes.Buffer{}
	writer := multipart.NewWriter(b)
	writer.WriteField("a", "a")
	writer.WriteField("token", token)
	writer.Close()
	resp, err = client.Post(ts.URL, writer.FormDataContentType(), b)
	ast.Nil(err)
	ast.Equal(200, resp.StatusCode)
}

func TestUploadBlock(t *testing.T) {
	ast := assert.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method == "OPTIONS" || req.Method == "HEAD" {
			return
		}
		req.ParseMultipartForm(0)
		ast.Equal("a", req.FormValue("a"))
		ast.NotEqual("", req.FormValue("stubtoken"))
		return
	}))
	defer ts.Close()

	tp := NewUpTransport(http.DefaultTransport, mockAcc{})
	client := &http.Client{Transport: tp}

	// test: HEAD
	resp, err := client.Head(ts.URL)
	ast.Nil(err)
	ast.Equal(200, resp.StatusCode)

	// test: post invalid multipart
	for i := 0; i < 3; i++ {
		resp, err = client.Post(ts.URL, "application/text", strings.NewReader("ss"))
		ast.Nil(err)
		ast.Equal(400, resp.StatusCode)
		err = rpc.ResponseError(resp)
		resp.Body.Close()
		ast.Equal("request Content-Type isn't multipart/form-data but application/text", err.Error())
	}

	// test: multipart
	policy := &rs.PutPolicy{
		Scope: "test",
	}
	token := policy.Token(&digest.Mac{ak, []byte(sk)})
	b := &bytes.Buffer{}
	writer := multipart.NewWriter(b)
	writer.WriteField("a", "a")
	writer.WriteField("token", token)
	writer.WriteField("token", token)
	writer.WriteField("rand", string(krand(peekSize)))
	writer.WriteField("token", token)
	writer.Close()
	b.Write(krand(peekSize))
	testDone := make(chan struct{})
	go func() {
		select {
		case <-testDone:
			break
		case <-time.After(3 * time.Second):
			panic("run too long")
		}
	}()
	resp, err = client.Post(ts.URL, writer.FormDataContentType(), b)
	ast.Nil(err)
	ast.Equal(200, resp.StatusCode)
	close(testDone)
}

func krand(size int) []byte {
	kinds, result := [][]int{[]int{10, 48}, []int{26, 97}, []int{26, 65}}, make([]byte, size)
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < size; i++ {
		ikind := rand.Intn(3)
		scope, base := kinds[ikind][0], kinds[ikind][1]
		result[i] = uint8(base + rand.Intn(scope))
	}
	return result
}

func TestGoroutineLeak(t *testing.T) {
	done := make(chan struct{}, 1)
	testHook = func() {
		close(done)
	}
	defer func() {
		testHook = func() {}
	}()
	ast := assert.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method == "OPTIONS" || req.Method == "HEAD" {
			return
		}
		req.ParseMultipartForm(0)
		ast.Equal("a", req.FormValue("a"))
		ast.NotEqual("", req.FormValue("stubtoken"))
		return
	}))
	ts.Close()

	tp := NewUpTransport(http.DefaultTransport, mockAcc{})
	client := &http.Client{Transport: tp}
	// test: multipart
	policy := &rs.PutPolicy{
		Scope: "test",
	}
	token := policy.Token(&digest.Mac{ak, []byte(sk)})
	b := &bytes.Buffer{}
	writer := multipart.NewWriter(b)
	writer.WriteField("a", "a")
	writer.WriteField("token", token)
	writer.Close()
	resp, err := client.Post(ts.URL, writer.FormDataContentType(), b)
	ast.Nil(err)
	ast.Equal(570, resp.StatusCode)

	select {
	case <-done:
		break
	case <-time.After(3 * time.Second):
		panic("goroutine leak")
	}

}
