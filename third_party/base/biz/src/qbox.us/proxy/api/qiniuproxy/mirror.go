package qiniuproxy

import (
	"net/http"
	"net/url"
	"strconv"

	"github.com/qiniu/errors"
	"github.com/qiniu/rpc.v1"
	"github.com/qiniu/rpc.v1/lb.v2"
	"qbox.us/proxy/api/proto"
)

type mirrorInstance struct {
	conn *lb.Client
}

func NewMirrorInstance(proxyHosts []string) proto.MirrorProxy {

	conn, err := lb.New(proxyHosts, &lb.Config{
		Http:              OneSecondClient,
		TryTimes:          uint32(len(proxyHosts)),
		FailRetryInterval: -1,
		ShouldRetry:       shouldRetry,
	})
	if err != nil {
		panic(err)
	}
	return &mirrorInstance{
		conn: conn,
	}
}

func (self *mirrorInstance) Mirror(l rpc.Logger, URLs []string, host, userAgent, srchost string, uid uint32) (resp *http.Response, err error) {
	m := url.Values{}
	m["url"] = URLs
	if host != "" {
		m.Add("host", host)
	}
	m.Add("uid", strconv.Itoa(int(uid)))
	req, err := lb.NewRequest("GET", "/mirror?"+m.Encode(), nil)
	if err != nil {
		return
	}
	if host != "" {
		req.Host = host
	}
	if srchost != "" {
		req.Header.Add("X-Qiniu-Src-Host", srchost)
	}
	if userAgent != "" {
		req.Header.Set("User-Agent", userAgent)
	}
	resp, err = self.conn.Do(l, req)
	if err != nil {
		err = errors.Info(err, "mirror via qiniuproxy").Detail(err)
	}
	return
}
