package multi

import (
	"github.com/qiniu/lbnsq.v1"
	"github.com/qiniu/rpc.v1"
	"github.com/qiniu/xlog.v1"
	ebdpfd "qbox.us/ebdpfd/api"
	"qbox.us/pfd/api/types"
)

type GroupConfig struct {
	DeleteProducerCfg lbnsq.ProducerConfig `json:"delete_producer_cfg"`
	EbdDeleteTopic    string               `json:"ebd_delete_topic"`
}
type Config struct {
	DefaultGroup string                 `json:"default_group"`
	GroupsConfig map[string]GroupConfig `json:"groups_config"`
}

type Client struct {
	conns map[string]ebdpfd.Deleter
}

type delete struct {
	ebdDeleteProducer *lbnsq.Client
	deleteTopic       string
}

func (d *delete) Delete(l rpc.Logger, fh []byte) (err error) {
	_, err = types.DecodeFh(fh)
	if err != nil {
		return
	}
	return d.ebdDeleteProducer.PublishEx(l.(*xlog.Logger), d.deleteTopic, fh)
}

func NewClient(cfg *Config) *Client {

	conns := make(map[string]ebdpfd.Deleter)
	for group, c := range cfg.GroupsConfig {
		producerClient, err := lbnsq.New(&c.DeleteProducerCfg)
		if err != nil {
			panic("lbnsq.New failed, error:" + err.Error())
		}
		conns[group] = &delete{ebdDeleteProducer: producerClient, deleteTopic: c.EbdDeleteTopic}
	}
	if cfg.DefaultGroup != "" {
		if conns[cfg.DefaultGroup] == nil {
			panic("default group is nil")
		}
		conns[""] = conns[cfg.DefaultGroup]
	}

	return &Client{conns: conns}
}

func (c *Client) Choose(group string) ebdpfd.Deleter {
	return c.conns[group]
}
