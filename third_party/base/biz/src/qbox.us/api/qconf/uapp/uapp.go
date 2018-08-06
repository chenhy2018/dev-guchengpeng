package uapp

import (
	"qbox.us/qconf/qconfapi"
	qufop "qbox.us/ufop"
	"github.com/qiniu/xlog.v1"
)

type Client struct {
	Conn *qconfapi.Client
}

func (p *Client) Get(xl *xlog.Logger, uapp string) (r *qufop.UappInfo, err error) {
	r = &qufop.UappInfo{}
	err = p.Conn.Get(xl, r, qufop.QconfUappID(uapp), qconfapi.Cache_NoSuchEntry)
	return
}

func (p *Client) SetProp(xl *xlog.Logger, uapp string, prop string, val interface{}) error {
	return p.Conn.SetProp(xl, qufop.QconfUappID(uapp), prop, val)
}

func (p *Client) Modify(xl *xlog.Logger, uapp string, ui *qufop.UappInfo) error {
	return p.Conn.Modify(xl, qufop.QconfUappID(uapp), ui)
}
