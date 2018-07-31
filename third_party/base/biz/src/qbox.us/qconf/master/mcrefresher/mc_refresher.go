package mcrefresher

import (
	"github.com/bradfitz/gomemcache/memcache"
	"github.com/qiniu/rpc.v1"
	"github.com/qiniu/xlog.v1"
)

type Config struct {
	McHosts []string `json:"mc_hosts"` // 互为镜像的Memcache服务
}

type Service struct {
	mcaches []*memcache.Client
	Config
}

func New(cfg *Config) (p *Service, err error) {

	mcaches := make([]*memcache.Client, len(cfg.McHosts))
	for i, host := range cfg.McHosts {
		mcaches[i] = memcache.New(host)
	}

	p = &Service{Config: *cfg, mcaches: mcaches}
	return
}

func (p *Service) Refresh(l rpc.Logger, id string) (err error) {
	xl := xlog.NewWith(l.ReqId())

	for i, mc := range p.mcaches {
		err := mc.Delete(id)
		if err != nil {
			xl.Warn("qconf slave.refresh failed:", p.McHosts[i], err)
		}
	}
	return nil
}
