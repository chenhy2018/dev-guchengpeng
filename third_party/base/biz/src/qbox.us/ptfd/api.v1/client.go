package api

import (
	"bytes"
	"crypto/md5"
	"io"
	"io/ioutil"

	"github.com/qiniu/errors"
	"github.com/qiniu/http/httputil.v1"
	"github.com/qiniu/rpc.v1"
	"github.com/qiniu/xlog.v1"

	"qbox.us/fh/pfd"
	"qbox.us/ptfd/masterapi.v1"
	"qbox.us/ptfd/stgapi.v1"
)

var (
	ErrInvalidArgs  = httputil.NewError(400, "invalid args")
	ErrExceedFsize  = httputil.NewError(400, "exceed file size")
	ErrInvalidFsize = httputil.NewError(400, "invalid fsize")
)

type Stg interface {
	Get(xl *xlog.Logger, eblocks []string, from, to int64) (rc io.ReadCloser, err error)
}

type Master interface {
	Query(l rpc.Logger, fh []byte) (*masterapi.Entry, error)
}

type Config struct {
	stgapi.Config
	Master masterapi.Config `json:"master"`
}

type Client struct {
	stg    Stg
	master Master
}

func New(stgCfg *stgapi.Config, masterCfg *masterapi.Config) (*Client, error) {
	stg, err := stgapi.New(stgCfg)
	if err != nil {
		return nil, errors.Info(err, "stgapi.New").Detail(err)
	}
	master := masterapi.New(masterCfg)
	return &Client{stg: stg, master: master}, nil
}

func NewWith(stg Stg, master Master) *Client {
	return &Client{stg: stg, master: master}
}

func (p *Client) Get(l rpc.Logger, fh []byte, from, to int64) (io.ReadCloser, int64, error) {
	inst := pfd.Instance(fh)
	if inst.Fsize() == 0 && from == 0 && to == 0 {
		return ioutil.NopCloser(bytes.NewReader(nil)), 0, nil
	}
	entry, err := p.master.Query(l, fh)
	if err != nil {
		return nil, 0, errors.Info(err, "Get: master.Query").Detail(err)
	}
	if inst.Fsize() != entry.Fsize {
		return nil, 0, errors.Info(ErrInvalidArgs, "Get: inst.Fsize != entry.Fsize", inst.Fsize(), entry.Fsize)
	}
	if to > entry.Fsize {
		return nil, 0, errors.Info(ErrExceedFsize, "Get: to > fsize", to, entry.Fsize)
	}
	rc, err := p.stg.Get(xlog.NewWith(l), entry.Eblocks, from, to)
	if err != nil {
		return nil, 0, errors.Info(err, "Get: stg.Get").Detail(err)
	}
	return rc, to - from, nil
}

func (p *Client) Md5(l rpc.Logger, fh []byte) ([]byte, error) {
	inst := pfd.Instance(fh)
	rc, fsize, err := p.Get(l, fh, 0, inst.Fsize())
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	h := md5.New()
	n, err := io.Copy(h, rc)
	if err != nil {
		return nil, err
	}
	if n != fsize {
		return nil, errors.Info(ErrInvalidFsize, "read size != fsize")
	}
	return h.Sum(nil), nil
}
