package multiebd

import (
	"io"

	"github.com/qiniu/errors"
	"github.com/qiniu/rpc.v1"

	ebdapi "qbox.us/ebd/api"
	ebddnapi "qbox.us/ebddn/api"
	pfdcfg "qbox.us/pfdcfg/api"
)

type Getter interface {
	Get(l rpc.Logger, fh []byte, from, to int64) (rc io.ReadCloser, fsize int64, err error)
	GetType(l rpc.Logger, fh []byte) (typ pfdcfg.DiskType, err error)
}

type GetterChooser interface {
	Choose(group string) Getter
}

type Config struct {
	DefaultGroup string `json:"default"`

	Ebd   map[string]ebdapi.Config   `json:"ebdmap"`   //group ->ebd
	Ebddn map[string]ebddnapi.Config `json:"ebddnmap"` //group ->ebddn
}

type Client struct {
	getters map[string]Getter // group ->getters
}

func NewClient(cfg *Config) (c *Client, err error) {
	c = new(Client)
	c.getters = make(map[string]Getter)
	var ebd Getter
	for group, cfg := range cfg.Ebd {
		if _, ok := c.getters[group]; ok {
			err = errors.New("duplicate group")
			return
		}
		ebd, err = ebdapi.New(&cfg)
		if err != nil {
			err = errors.Info(err, "ebdapi.New").Detail(err)
			return
		}
		c.getters[group] = &groupGetter{group: group, getter: ebd}
	}
	for group, cfg := range cfg.Ebddn {
		if _, ok := c.getters[group]; ok {
			err = errors.New("duplicate group")
			return
		}
		ebd, err = ebddnapi.New(&cfg)
		if err != nil {
			err = errors.Info(err, "ebddnapi.New").Detail(err)
			return
		}
		c.getters[group] = &groupGetter{group: group, getter: ebd}
	}
	if cfg.DefaultGroup != "" {
		if _, ok := c.getters[cfg.DefaultGroup]; !ok {
			err = errors.New("no default getters")
			return
		}
		c.getters[""] = c.getters[cfg.DefaultGroup]
	}
	return
}

func (c *Client) Choose(group string) Getter {
	if getter, ok := c.getters[group]; ok {
		return getter
	}
	return nilGetter{}
}

type nilGetter struct {
}

func (g nilGetter) Get(l rpc.Logger, fh []byte, from, to int64) (rc io.ReadCloser, fsize int64, err error) {
	return nil, 0, errors.New("not support")
}

func (g nilGetter) GetType(l rpc.Logger, fh []byte) (typ pfdcfg.DiskType, err error) {
	return 0, errors.New("not support")
}

type SingleEbd struct {
	getter Getter
}

func NewSingleEbd(getter Getter) *SingleEbd {
	return &SingleEbd{getter: getter}
}

func (c *SingleEbd) Choose(group string) Getter {
	return c.getter
}

type groupGetter struct {
	group  string
	getter Getter
}

func (gg *groupGetter) Get(l rpc.Logger, fh []byte, from, to int64) (rc io.ReadCloser, fsize int64, err error) {
	l.Xput([]string{gg.group + "EBD"})
	return gg.getter.Get(l, fh, from, to)
}

func (gg *groupGetter) GetType(l rpc.Logger, fh []byte) (typ pfdcfg.DiskType, err error) {
	return gg.getter.GetType(l, fh)
}
