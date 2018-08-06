package qiniuproxy

import (
	"net/http"
	"time"

	"github.com/qiniu/rpc.v1"
	"github.com/qiniu/rpc.v1/lb.v2"
)

var (
	OneSecondClient = &http.Client{Transport: rpc.NewTransportTimeout(time.Second, 0)}
)

func shouldRetry(code int, err error) bool {
	if code == 570 {
		return true
	}
	return lb.ShouldRetry(code, err)
}
