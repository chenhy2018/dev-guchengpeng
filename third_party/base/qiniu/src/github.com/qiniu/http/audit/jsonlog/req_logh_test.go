package jsonlog

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"strconv"

	"github.com/qiniu/io/crc32util"
	qrpc "github.com/qiniu/rpc.v1"
	"github.com/stretchr/testify/assert"
	"qbox.us/net/httputil"
	"qbox.us/servestk"
	"qiniupkg.com/x/errors.v8"
)

type testlog struct {
}

func (l testlog) Log(msg []byte) error {
	return errors.New(string(msg))
}

func TestServeStack(t *testing.T) {

	logf := testlog{}
	var dec Decoder
	al := New("FOO", logf, dec, 512)
	ss := servestk.New(http.NewServeMux(), al.Handler)

	// test xBody works(no panic)
	ss.HandleFunc("/xbody", func(w http.ResponseWriter, req *http.Request) {
		fmt.Println("call /xbody")
		fmt.Println("content-length:", req.ContentLength)
		req.Body = ioutil.NopCloser(req.Body)
		httputil.Reply(w, 200, map[string]string{"foo": "bar"})
	})

	svr := httptest.NewServer(ss)
	svrUrl := svr.URL
	defer svr.Close()

	rpc := qrpc.Client{http.DefaultClient}
	req2Body := bytes.NewReader([]byte{1, 2})
	req2, err := http.NewRequest("POST", svrUrl+"/xbody", req2Body)
	if err != nil {
		fmt.Println(err)
	}
	req2.ContentLength = -1
	rpc.Do(nil, req2)
}

func TestBaseDecoder_DecodeRequest(t *testing.T) {
	dec := BaseDecoder{}

	req, _ := http.NewRequest("POST", "http://localhost:8000", bytes.NewBufferString("test"))
	req.Header.Set("Content-Length", strconv.FormatInt(req.ContentLength, 10))
	_, h, _ := dec.DecodeRequest(req)
	assert.Equal(t, "4", h["Content-Length"])

	req.Header.Set(qrpc.CrcEncodedHeader, "1")
	if req.Body != nil {
		enc := crc32util.SimpleEncoder(req.Body, nil)
		req.Body = ioutil.NopCloser(enc)
	}
	if req.ContentLength >= 0 {
		req.ContentLength = crc32util.EncodeSize(req.ContentLength)
	}
	req.Header.Set("Content-Length", strconv.FormatInt(req.ContentLength, 10))
	_, h, _ = dec.DecodeRequest(req)
	assert.Equal(t, "8", h["Content-Length"])
	assert.Equal(t, "1", h[qrpc.CrcEncodedHeader])
}
