package apigate

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/qiniu/apigate.v1/proto"
	"github.com/qiniu/encoding.v2/jsonutil"
)

// --------------------------------------------------------------------

type ApiConfig struct {
	Patterns      []string `json:"patterns"`
	Pattern       string   `json:"pattern"`
	Allow         string   `json:"allow"`
	NotAllow      string   `json:"notallow"`
	SuOnly        bool     `json:"suonly"`
	Forward       string   `json:"forward"`
	Auths         []string `json:"auths"`
	Proxy         string   `json:"proxy"`
	MaxConcurency int64    `json:"max_concurency"`
}

type ServiceConfig struct {
	Metric        json.RawMessage `json:"metric"`
	Module        string          `json:"module"`
	Routes        []string        `json:"routes"`
	Auths         []string        `json:"auths"`
	Forward       string          `json:"forward"`
	ForwardHost   string          `json:"forward_host"`
	MaxConcurency int64           `json:"max_concurency"`
	APIs          []*ApiConfig    `json:"apis"`

	AllowFrozenWithAdminAuth bool `json:"allow_frozen_with_admin_auth"`
}

type Config struct {
	Services []*ServiceConfig `json:"services"`
}

// --------------------------------------------------------------------

func NewWith(cfg *Config, mg proto.Metric) (p *Service, err error) {

	p = New()

	for _, sconf := range cfg.Services {

		// register metric
		// 理论上如果支持多个统计服务的话，不同的统计配置应该支持不一样，这里先不做
		err = mg.Register(sconf.Module, string(sconf.Metric))
		if err != nil {
			return nil, err
		}

		if sconf.Forward == "" {
			return nil, ErrNoServiceForwardHost
		}
		service := p.Service(sconf.Module, sconf.MaxConcurency, sconf.Routes...)
		auths, err1 := GetAuthStubers(sconf.Auths, sconf.AllowFrozenWithAdminAuth)
		if err1 != nil {
			return nil, err1
		}
		service.HandleNotFound()
		service.Auths(auths...).Forward(sconf.Forward).ForwardHost(sconf.ForwardHost)
		for _, apiconf := range sconf.APIs {
			ai, err2 := ParseAccessInfo(apiconf.Allow, apiconf.NotAllow, apiconf.SuOnly)
			if err2 != nil {
				return nil, err2
			}
			for _, pattern := range append(apiconf.Patterns, apiconf.Pattern) {
				api := service.Api(pattern, apiconf.MaxConcurency)
				if len(apiconf.Auths) > 0 {
					auths, err2 = GetAuthStubers(apiconf.Auths, sconf.AllowFrozenWithAdminAuth)
					if err2 != nil {
						return nil, err2
					}
					api.Auths(auths...)
				}
				if apiconf.Proxy != "" {
					proxy, ok := GetProxy(apiconf.Proxy)
					if !ok {
						return nil, fmt.Errorf("proxy [%s] not found", apiconf.Proxy)
					}
					api.Proxy(proxy)
				}
				api.Access(ai).Forward(apiconf.Forward)
			}
		}
	}
	return
}

func NewFromFile(configFile string, metricR proto.Metric) (p *Service, err error) {

	f, err := os.Open(configFile)
	if err != nil {
		return
	}
	defer f.Close()

	var cfg Config
	err = json.NewDecoder(f).Decode(&cfg)
	if err != nil {
		return
	}
	return NewWith(&cfg, metricR)
}

func NewFromString(config string, metricR proto.Metric) (p *Service, err error) {

	var cfg Config
	err = jsonutil.Unmarshal(config, &cfg)
	if err != nil {
		return
	}
	return NewWith(&cfg, metricR)
}

// --------------------------------------------------------------------
