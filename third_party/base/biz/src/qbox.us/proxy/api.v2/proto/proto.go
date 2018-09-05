package proto

import (
	"net/http"
	"time"

	"github.com/qiniu/rpc.v1"
)

var (
	CallbackUserAgent        = "qiniu-callback/1.0"
	MirrorUserAgent          = "qiniu-imgstg-spider-1.0"
	CommonMirrorUserAgent    = "qiniu-imgstg-spider"
	ReentrantMirrorUserAgent = "reentrant-qiniu-imgstg-spider"
)

type CallbackConfig struct {
	Uid       uint32
	AccessKey string
	Timeout   time.Duration
	Retry     int
}

type CallbackProxy interface {
	Callback(l rpc.Logger, URLs []string, host, bodyType string, body string, config *CallbackConfig) (resp *http.Response, err error)
}

type MirrorConfig struct {
	Uid       uint32
	Bucket    string
	SrcHost   string // Mirror的来源地址
	Nocache   bool
	UserAgent string
	Md5       string
	Etag      string
	Header    http.Header
	Retry     int
}

type MirrorProxy interface {
	Mirror(l rpc.Logger, URLs []string, host string, config *MirrorConfig) (resp *http.Response, err error)
}
