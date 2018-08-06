package bdgetter

import (
	"crypto/sha1"

	"github.com/qiniu/errors"
	"github.com/qiniu/log.v1"
	"github.com/qiniu/rpc.v1"

	"qbox.us/bdgetter/cached"
	"qbox.us/bdgetter/retrys"
	ebdapi "qbox.us/ebd/api"
	ebdtypes "qbox.us/ebd/api/types"
	ebddnapi "qbox.us/ebddn/api"
	ebdpfdapi "qbox.us/ebdpfd/api"
	qfh "qbox.us/fh"
	"qbox.us/fh/proto"
	"qbox.us/multiebd"
	pfdapi "qbox.us/pfd/api"
	pfdtypes "qbox.us/pfd/api/types"
	cfgapi "qbox.us/pfdcfg/api"
	"qbox.us/pfdtracker/stater"
	ptfd "qbox.us/ptfd/getter.v1"
)

// -----------------------------------------------------------------------------

type PfdConfig struct {
	Pfd       pfdapi.Config    `json:"pfd"`
	Ebd       ebdapi.Config    `json:"ebd"`
	Ebddn     ebddnapi.Config  `json:"ebddn"`
	Multi     *multiebd.Config `json:"multi"`
	GetRetrys int              `json:"get_retrys"`
}

func newPfdGetter(cfg *PfdConfig, tcfg *PtfdConfig,
	ci *cachedInfo, noNeedLocalCache bool) (g proto.CommonGetter, err error) {

	pfd, err := pfdapi.New(&cfg.Pfd)
	if err != nil {
		return nil, err
	}
	ebdpfdGetter := ebdpfdapi.Getter(pfd)
	if cfg.Multi != nil {
		ebd, err := multiebd.NewClient(cfg.Multi)
		if err != nil {
			return nil, err
		}
		log.Info("support ebd via multi")
		ebdpfdGetter = ebdpfdapi.NewWithChooser(pfd.GidStater, pfd, ebd)
	} else if len(cfg.Ebd.MasterHosts) > 0 {
		ebd, err := ebdapi.New(&cfg.Ebd)
		if err != nil {
			return nil, err
		}
		log.Info("support ebd via ebdapi")
		ebdpfdGetter = ebdpfdapi.New(pfd.GidStater, pfd, ebd)
	} else if len(cfg.Ebddn.Hosts) > 0 {
		ebd, err := ebddnapi.New(&cfg.Ebddn)
		if err != nil {
			return nil, err
		}
		log.Info("support ebd via ebddnapi")
		ebdpfdGetter = ebdpfdapi.New(pfd.GidStater, pfd, ebd)
	}

	var rg proto.ReaderGetter = ebdpfdGetter
	if len(tcfg.Cfg.Hosts) > 0 {
		rg, err = ptfd.New(&tcfg.Config, &tcfg.Master, ebdpfdGetter)
		if err != nil {
			return nil, errors.Info(err, "ptfd.New").Detail(err)
		}
	}

	if ci != nil {
		cc := &ebdpfdCachedController{
			ci:               ci,
			noNeedLocalCache: noNeedLocalCache,
			gidStater:        pfd.GidStater,
			ebdpfdGetter:     ebdpfdGetter,
			ebdGroupBlocks:   newEbdCached(cfg),
		}
		g = cached.New(rg, cc, cfg.GetRetrys)
	} else {
		g = retrys.NewRetrys(rg, cfg.GetRetrys)
	}
	return g, nil
}

type ebdpfdCachedController struct {
	ci               *cachedInfo
	noNeedLocalCache bool
	gidStater        stater.Stater
	ebdpfdGetter     ebdpfdapi.Getter
	ebdGroupBlocks   ebdGroupCached
}

func (p *ebdpfdCachedController) Cached(l rpc.Logger, fh []byte) (key []byte, cached cached.Cached, blocks bool) {

	if typ, _ := p.ebdpfdGetter.GetType(l, fh); typ == cfgapi.SSD {
		return nil, nil, false
	}
	ssdCached, sataCached, ssdCacheSize := p.ci.Get()
	if !p.noNeedLocalCache {
		if fsize := qfh.Fsize(fh); fsize <= ssdCacheSize {
			// 4MB以下的缓存，直接拿hash作为key
			return qfh.Etag(fh)[1:], ssdCached, false
		}
	}
	fhi, _ := ebdtypes.DecodeFh(fh)
	if group, _, isECed, _ := p.gidStater.StateWithGroup(l, pfdtypes.EncodeGid(fhi.Gid)); isECed {
		// ebd的缓存是跨机房的，忽略noNeedLocalCache选项
		// dc的key必须为20字节, 拿etag的sha1为key
		key := sha1.Sum(qfh.Etag(fh))
		isCached, isBlocks := p.ebdGroupBlocks(group)
		if isCached {
			return key[:], sataCached, isBlocks
		} else {
			return key[:], nil, false
		}
	}
	return nil, nil, false
}

type ebdGroupCached func(group string) (cached bool, blocks bool)

func newEbdCached(cfg *PfdConfig) ebdGroupCached {
	if cfg.Multi != nil {
		multi := cfg.Multi
		blocks := make(map[string]bool)
		cached := make(map[string]bool)
		for group, ebddn := range multi.Ebddn {
			blocks[group] = !ebddn.NoBlocks
			cached[group] = !ebddn.NoCached
		}
		for group, ebd := range multi.Ebd {
			blocks[group] = !ebd.NoBlocks
			cached[group] = !ebd.NoCached
		}
		if multi.DefaultGroup != "" {
			blocks[""] = blocks[multi.DefaultGroup]
			cached[""] = cached[multi.DefaultGroup]
		}
		return func(group string) (bool, bool) {
			isBlocks, ok1 := blocks[group]
			isCached, ok2 := cached[group]
			if !(ok1 && ok2) {
				panic("no such ebd group")
			}
			return isCached, isBlocks
		}
	} else if len(cfg.Ebd.MasterHosts) > 0 {
		return func(group string) (bool, bool) {
			return !cfg.Ebd.NoCached, !cfg.Ebd.NoBlocks
		}
	} else if len(cfg.Ebddn.Hosts) > 0 {
		return func(group string) (bool, bool) {
			return !cfg.Ebddn.NoCached, !cfg.Ebddn.NoBlocks
		}
	}
	return func(group string) (bool, bool) {
		return true, true
	}
}
