package fopd

import (
	"bytes"
	"code.google.com/p/go.net/context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/qiniu/xlog.v1"
	"github.com/stretchr/testify/assert"

	"qbox.us/fop"
)

func TestCancel(t *testing.T) {

	doneCh := make(chan bool, 1)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		println("remote:", r.RemoteAddr)
		<-doneCh
		time.Sleep(time.Second / 5)
	}))
	defer ts.Close()
	baseCtx := xlog.NewContextWith(context.Background(), "TestCancel")

	conn := NewConn(ts.URL, nil)

	// test op with cancel
	ctx1, cancel1 := context.WithCancel(baseCtx)
	go func() {
		time.Sleep(time.Second / 10)
		cancel1()
		doneCh <- true
	}()
	_, err := conn.Op(ctx1, bytes.NewReader([]byte{}), 0, &fop.FopCtx{})
	assert.Equal(t, context.Canceled, err)

	// test op2 with cancel
	ctx2, cancel2 := context.WithCancel(baseCtx)
	go func() {
		time.Sleep(time.Second / 10)
		cancel2()
		doneCh <- true
	}()
	_, err = conn.Op2(ctx2, []byte{}, 0, &fop.FopCtx{})
	assert.Equal(t, context.Canceled, err)
}
