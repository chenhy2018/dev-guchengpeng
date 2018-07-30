package getter

import (
	"io"

	"github.com/qiniu/errors"
	"github.com/qiniu/http/httputil.v1"
	"github.com/qiniu/rpc.v1"

	"qbox.us/fh/proto"
	pfdapi "qbox.us/pfdstg/api"
	"qbox.us/ptfd/api.v1"
	"qbox.us/ptfd/masterapi.v1"
	"qbox.us/ptfd/stgapi.v1"
)

type Client struct {
	ptfd *api.Client
	pfd  proto.ReaderGetter
}

func New(cfg *stgapi.Config, masterCfg *masterapi.Config, pfd proto.ReaderGetter) (*Client, error) {

	ptfd, err := api.New(cfg, masterCfg)
	if err != nil {
		return nil, err
	}
	return &Client{ptfd: ptfd, pfd: pfd}, nil
}

func NewWith(stg api.Stg, master api.Master, pfd proto.ReaderGetter) *Client {

	ptfd := api.NewWith(stg, master)
	return &Client{ptfd: ptfd, pfd: pfd}
}

func (p *Client) Get(l rpc.Logger, fh []byte, from, to int64) (io.ReadCloser, int64, error) {

	rc, n, err := p.pfd.Get(l, fh, from, to)
	if err == nil {
		return rc, n, err
	}
	if httputil.DetectCode(err) != pfdapi.StatusAllocedEntry {
		return nil, 0, errors.Info(err, "pfd.Get").Detail(err)
	}
	return p.ptfd.Get(l, fh, from, to)
}
