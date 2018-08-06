package stgapi

import (
	"io"

	"qbox.us/ptfd/cfgapi.v1"
	"qbox.us/ptfd/stgapi.v1/api"

	"github.com/qiniu/errors"
	"github.com/qiniu/http/httputil.v1"
	qio "github.com/qiniu/io"
	"github.com/qiniu/rpc.v1"
	"github.com/qiniu/xlog.v1"
)

var (
	ErrInvalidCtx         = httputil.NewError(400, "invalid ctx")
	ErrInvalidArgs        = httputil.NewError(400, "invalid args")
	ErrServiceUnavailable = httputil.NewError(503, "service unavailable")
)

type Stg interface {
	Create(l rpc.Logger, host string, max uint32, r io.Reader, length uint32) (ret api.StgRet, err error)
	Put(l rpc.Logger, host string, ctx string, off uint32, r io.Reader, length uint32) (ret api.StgRet, err error)
	Get(l rpc.Logger, host string, eblock string, from, to uint32) (rc io.ReadCloser, err error)
	ProxyGet(l rpc.Logger, host string, eblock string, from, to uint32) (rc io.ReadCloser, err error)
}

type Cfg interface {
	HostsIdc(xl *xlog.Logger, dgid uint32) (hosts []string, ihost int, idc string, err error)
	Actives(xl *xlog.Logger, idc string) (hosts []string, ihost int, err error)
}

type Config struct {
	Cfg       cfgapi.Config      `json:"cfg"`
	Timeouts  api.TimeoutOptions `json:"timeouts"`
	Proxys    []string           `json:"proxies"`
	PutRetrys int                `json:"put_retrys"`
	Idc       string             `json:"idc"`
}

type Client struct {
	stg        Stg
	cfg        Cfg
	maxPutTrys int
	idc        string
}

func New(cfg *Config) (*Client, error) {

	pcfg, err := cfgapi.New(&cfg.Cfg)
	if err != nil {
		return nil, errors.Info(err, "cfgapi.New").Detail(err)
	}
	stg := api.NewClient(&api.Options{TimeoutOptions: cfg.Timeouts, Proxys: cfg.Proxys})
	return &Client{stg: stg, cfg: pcfg, maxPutTrys: cfg.PutRetrys + 1, idc: cfg.Idc}, nil
}

func NewTransfer(cfg *Config) (*Client, error) {
	pcfg, err := cfgapi.New(&cfg.Cfg)
	if err != nil {
		return nil, errors.Info(err, "cfgapi.New").Detail(err)
	}
	stg := api.NewClient(&api.Options{TimeoutOptions: cfg.Timeouts, Proxys: cfg.Proxys, Transfer: true})
	return &Client{stg: stg, cfg: pcfg, maxPutTrys: cfg.PutRetrys + 1, idc: cfg.Idc}, nil
}

func NewWith(stg Stg, cfg Cfg, maxPutTrys int, idc string) *Client {

	return &Client{stg: stg, cfg: cfg, maxPutTrys: maxPutTrys, idc: idc}
}

// -----------------------------------------------------------------------------

func (p *Client) Create(xl *xlog.Logger, max uint32, ra io.ReaderAt, size uint32) (ctx string, err error) {

	hosts, idx, err := p.cfg.Actives(xl, p.idc) // 只上传到本 idc 的节点
	if err != nil {
		err = errors.Info(ErrServiceUnavailable, "cfg.Actives").Detail(err)
		return
	}
	maxPutTrys := len(hosts)
	if maxPutTrys > p.maxPutTrys {
		maxPutTrys = p.maxPutTrys
	}
	var ret api.StgRet
	trys := 0
	host := hosts[idx]
	for {
		r := &qio.Reader{ReaderAt: ra}
		ret, err = p.stg.Create(xl, host, max, r, size)
		if err == nil {
			ctx = ret.Ctx
			break
		}
		xl.Errorf("Client.Create: stg.Create %v failed => %v", host, err)
		trys++
		if trys >= maxPutTrys {
			err = errors.Info(err, "stg.Create").Detail(err)
			break
		}
		if trys == 1 {
			hosts = copyExcept(hosts, idx)
		}
		hosts, host = randomShrink(hosts)
	}
	return
}

func (p *Client) Put(xl *xlog.Logger, ctx string, off uint32, ra io.ReaderAt, size uint32) (rctx string, err error) {

	fp, err := api.DecodePositionCtx(ctx)
	if err != nil {
		err = errors.Info(ErrInvalidCtx, "DecodeFilePositionCtx", ctx).Detail(err)
		return
	}
	var ret api.StgRet
	if fp.Eblock != api.ZeroEblock {
		addr, err1 := api.DecodeEblock(fp.Eblock)
		if err1 != nil {
			err = errors.Info(ErrInvalidCtx, "DecodeEblock", fp.Eblock).Detail(err1)
			return
		}
		hosts, _, _, err1 := p.cfg.HostsIdc(xl, addr.Dgid)
		if err1 != nil {
			err = errors.Info(ErrServiceUnavailable, "cfg.Hosts", addr.Dgid).Detail(err1)
			return
		}
		r := &qio.Reader{ReaderAt: ra}
		ret, err = p.stg.Put(xl, hosts[0], ctx, off, r, size)
		if err != nil {
			err = errors.Info(err, "stg.Put", hosts[0]).Detail(err)
			return
		}
		rctx = ret.Ctx
		return
	}

	hosts, idx, err := p.cfg.Actives(xl, p.idc)
	if err != nil {
		err = errors.Info(ErrServiceUnavailable, "cfg.Actives").Detail(err)
		return
	}
	maxPutTrys := len(hosts)
	if maxPutTrys > p.maxPutTrys {
		maxPutTrys = p.maxPutTrys
	}
	trys := 0
	host := hosts[idx]
	for {
		r := &qio.Reader{ReaderAt: ra}
		ret, err = p.stg.Put(xl, host, ctx, off, r, size)
		if err == nil {
			rctx = ret.Ctx
			break
		}
		xl.Errorf("Client.Put: stg.Put %v failed => %v", host, err)
		trys++
		if trys >= maxPutTrys {
			err = errors.Info(err, "stg.Put", host).Detail(err)
			break
		}
		if trys == 1 {
			hosts = copyExcept(hosts, idx)
		}
		hosts, host = randomShrink(hosts)
	}
	return
}

func (p *Client) Get(xl *xlog.Logger, eblocks []string, from, to int64) (rc io.ReadCloser, err error) {

	if int64(len(eblocks))*api.MaxDataSize < to {
		return nil, errors.Info(ErrInvalidArgs, "len(eblocks) to", len(eblocks), to)
	}
	return newReader(xl, p.stg, p.cfg, p.idc, eblocks, from, to), nil
}
