package proto

import (
	"net/http"
	"time"

	"github.com/qiniu/rpc.v1"
)

type CallbackConfig struct {
	Uid     uint32
	Timeout time.Duration
}

var (
	CallbackUserAgent        = "qiniu-callback/1.0"
	MirrorUserAgent          = "qiniu-imgstg-spider-1.0"
	CommonMirrorUserAgent    = "qiniu-imgstg-spider"
	ReentrantMirrorUserAgent = "reentrant-qiniu-imgstg-spider"
)

type CallbackProxy interface {
	Callback(l rpc.Logger, URLs []string, host, bodyType string, body string, accessKey string, config CallbackConfig) (resp *http.Response, err error)
}

type MirrorProxy interface {
	Mirror(l rpc.Logger, URLs []string, host, userAgent, srchost string, uid uint32) (resp *http.Response, err error)
}
