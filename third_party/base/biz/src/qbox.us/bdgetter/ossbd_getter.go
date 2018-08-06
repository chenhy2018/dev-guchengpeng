package bdgetter

import (
	"crypto/sha1"

	rpc "github.com/qiniu/rpc.v1"
	"qbox.us/bdgetter/cached"
	qfh "qbox.us/fh"
	"qbox.us/fh/ossbd"
	"qbox.us/fh/proto"
	ossapi "qbox.us/ossbd/api"
)

func newOssGetter(cfg ossapi.Config, ci *cachedInfo) (g proto.CommonGetter, err error) {
	oss, err := ossapi.New(&cfg)
	if err != nil {
		return nil, err
	}
	g = cached.New(oss, &cacheAll{ci}, cfg.RetryTimes)
	return
}

type cacheAll struct {
	ci *cachedInfo
}

func (c *cacheAll) Cached(l rpc.Logger, fh []byte) (key []byte, cached cached.Cached, blocks bool) {
	ssdCached, sataCached, ssdCacheSize := c.ci.Get()
	size := ossbd.Instance(fh).Fsize()
	if size < ssdCacheSize {
		// 4MB以下的缓存，直接拿hash作为key
		return qfh.Etag(fh)[1:], ssdCached, false
	} else {
		key := sha1.Sum(qfh.Etag(fh))
		return key[:], sataCached, true
	}
}
