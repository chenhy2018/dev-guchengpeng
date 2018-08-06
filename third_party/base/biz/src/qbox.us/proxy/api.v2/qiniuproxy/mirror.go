package qiniuproxy

import (
	"net/http"
	"net/url"
	"strconv"

	"io"

	"github.com/qiniu/errors"
	"github.com/qiniu/io/crc32util"
	"github.com/qiniu/rpc.v1"
	"github.com/qiniu/rpc.v1/lb.v2"
	"qbox.us/proxy/api.v2/proto"
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

type readerCloser struct {
	io.Reader
	io.Closer
}

func (self *mirrorInstance) Mirror(l rpc.Logger, URLs []string, host string, config *proto.MirrorConfig) (resp *http.Response, err error) {
	m := url.Values{}
	m["url"] = URLs
	if host != "" {
		m.Add("host", host)
	}
	if config.Md5 != "" {
		m.Add("md5", config.Md5)
	}
	if config.Nocache {
		m.Add("nocache", "1")
	}
	if config.Bucket != "" {
		m.Add("bucket", config.Bucket)
	}
	if config.Etag != "" {
		m.Add("etag", config.Etag)
	}
	m.Add("uid", strconv.Itoa(int(config.Uid)))
	m.Add("retry", strconv.Itoa(int(config.Retry)))
	req, err := lb.NewRequest("GET", "/mirror?"+m.Encode(), nil)
	if err != nil {
		return
	}
	if host != "" {
		req.Host = host
	}
	if config.SrcHost != "" {
		req.Header.Add("X-Qiniu-Src-Host", config.SrcHost)
	}
	if config.UserAgent != "" {
		req.Header.Set("User-Agent", config.UserAgent)
	}
	for k, vv := range config.Header {
		if req.Header.Get(k) == "" {
			for _, v := range vv {
				req.Header.Add(k, v)
			}
		}
	}

	req.Header.Set(rpc.NeedCrcEncodeHeader, "1")
	resp, err = self.conn.Do(l, req)
	if err != nil {
		err = errors.Info(err, "mirror via qiniuproxy").Detail(err)
		return
	}
	// 这里和 rpc_client 中的代码相同，以后可以考虑合并。
	// 对 body 进行 crc decode.
	if resp.Header.Get(rpc.CrcEncodedHeader) != "" && resp.StatusCode/100 == 2 {
		// resp body is always non-nil
		dec := crc32util.SimpleDecoder(resp.Body, nil)
		resp.Body = readerCloser{dec, resp.Body}
		if resp.ContentLength >= 0 {
			resp.ContentLength = crc32util.DecodeSize(resp.ContentLength)
		}
	}
	return
}
