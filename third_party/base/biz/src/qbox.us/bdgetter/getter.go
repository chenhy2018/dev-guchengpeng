// 七牛存储服务的读客户端
package bdgetter

import (
	"io"

	"qbox.us/api/bd/bdc/reload_bdc"
	"qbox.us/api/bd/s3"
	"qbox.us/bdgetter/cached"
	"qbox.us/cc/config"
	"qbox.us/fh/proto"
	"qbox.us/fh/sha1bdebd"
	ossapi "qbox.us/ossbd/api"

	"github.com/qiniu/errors"
	"github.com/qiniu/log.v1"
	"github.com/qiniu/xlog.v1"
)

type S3Config struct {
	Basic   s3.Config      `json:"basic"`
	Buckets []S3BucketInfo `json:"buckets"`
	Retrys  int            `json:"retrys"`
}

type S3BucketInfo struct {
	Id   uint16 `json:"id"`
	Name string `json:"name"`
}

type Config struct {
	S3Conf           S3Config               `json:"s3"`
	PfdConf          PfdConfig              `json:"pfd"`
	PtfdConf         PtfdConfig             `json:"ptfd"`
	OssbdConf        ossapi.Config          `json:"oss"`
	RbdRemoteConf    config.ReloadingConfig `json:"rbd"`
	RBdRetryInterval int64                  `json:"rbd_retry_interval"`
	IbdsReadFromEbd  []uint16               `json:"ibds_read_from_ebd"`
	DcRemoteConf     config.ReloadingConfig `json:"dc"`
	NoNeedLocalCache bool                   `json:"no_need_local_cache"`
	EbdChunkBits     uint                   `json:"ebd_chunk_bits"`
}

func New(cfg *Config) (getter *proto.Getter, err error) {
	var stg proto.Sha1bdGetter
	if count := len(cfg.S3Conf.Buckets); count > 0 {
		conf := cfg.S3Conf
		mbds := make(map[uint16]proto.Sha1bdGetter, count)
		for _, info := range conf.Buckets {
			mbds[info.Id] = s3.New(&conf.Basic, info.Name, conf.Retrys)
		}
		stg = &multiGetter{mbds}
	} else {
		conf := &reload_bdc.MultiStgConfig{
			ReloadingConfig: cfg.RbdRemoteConf,
			RetryIntervalMs: int(cfg.RBdRetryInterval / 1000),
		}
		log.Infof("direct to rbd: %+v\n", conf)
		stg, err = reload_bdc.NewMultiStg(conf)
		if err != nil {
			log.Error("NewMultiStg error:", errors.Detail(err))
			return
		}
		for _, ibd := range cfg.IbdsReadFromEbd {
			sha1bdebd.ReadFromEbd[ibd] = true
		}
		// sha1bd已经几乎不使用了，不用实现对它的缓存逻辑
	}

	var ci *cachedInfo
	if cfg.DcRemoteConf.RemoteURL != "" {
		ci, err = newCachedInfo(&cfg.DcRemoteConf)
		if err != nil {
			log.Error("newCached error", errors.Detail(err))
			return
		}
	}

	var pfd proto.CommonGetter
	if len(cfg.PfdConf.Pfd.CfgHosts) > 0 {
		pfd, err = newPfdGetter(&cfg.PfdConf, &cfg.PtfdConf, ci, cfg.NoNeedLocalCache)
		if err != nil {
			log.Error("newPfdGetter error:", errors.Detail(err))
			return
		}
	}

	oss, err := newOssGetter(cfg.OssbdConf, ci)
	if err != nil {
		log.Error("newOssGetter failed", err)
		return
	}
	getter = &proto.Getter{
		Sha1bd: stg,
		Pfd:    pfd,
		Oss:    oss,
	}

	if cfg.EbdChunkBits != 0 {
		cached.Init(cfg.EbdChunkBits)
	}
	return getter, nil
}

type multiGetter struct {
	mbds map[uint16]proto.Sha1bdGetter
}

func (p *multiGetter) Get(xl *xlog.Logger, key []byte, w io.Writer, from, to int, bds [4]uint16) error {
	return p.mbds[bds[0]].Get(xl, key, w, from, to, bds)
}
