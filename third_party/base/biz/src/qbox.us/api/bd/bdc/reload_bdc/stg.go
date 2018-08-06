package reload_bdc

import (
	"io"
	"strconv"
	"sync"

	"github.com/qiniu/errors"
	"github.com/qiniu/rpc.v1"

	"qbox.us/api/bd/bdc"
	"qbox.us/cc/config"
	"github.com/qiniu/xlog.v1"
)

var errUnknownBd = errors.New("unknown bd")

type MultiStgConfig struct {
	config.ReloadingConfig
	RetryIntervalMs int
}

type MultiStg struct {
	stgs map[uint16]*bdc.BdClient
	sync.RWMutex
	MultiStgConfig
}

func NewMultiStg(cfg *MultiStgConfig) (p *MultiStg, err error) {

	p = &MultiStg{
		MultiStgConfig: *cfg,
	}
	err = config.StartReloading(&cfg.ReloadingConfig, p.onReload)
	if err != nil {
		err = errors.Info(err, "config.StartReloading").Detail(err)
	}
	return
}

func (p *MultiStg) Get(xl *xlog.Logger, key []byte, w io.Writer, from, to int, bds [4]uint16) error {

	p.RWMutex.RLock()
	stgs := p.stgs
	p.RWMutex.RUnlock()
	stg, ok := stgs[bds[0]]
	if !ok {
		return errUnknownBd
	}
	return stg.Get(xl, key, w, from, to, bds)
}

func (p *MultiStg) Put(xl *xlog.Logger, key []byte, r io.Reader, n int, bds [3]uint16) error {

	p.RWMutex.RLock()
	stgs := p.stgs
	p.RWMutex.RUnlock()
	stg, ok := stgs[bds[0]]
	if !ok {
		return errUnknownBd
	}
	return stg.Put(xl, key, r, n, bds)
}

func (p *MultiStg) onReload(l rpc.Logger, data []byte) error {

	xl := xlog.NewWith(l.ReqId())

	var hosts map[string][]string
	err := config.LoadData(&hosts, data)
	if err != nil {
		err = errors.Info(err, "config.LoadData").Detail(err)
		return err
	}

	interval := int64(p.RetryIntervalMs) * 1000
	stgs := make(map[uint16]*bdc.BdClient)
	for k, v := range hosts {
		bd, err := strconv.Atoi(k)
		if err != nil {
			xl.Warn("MultiStg.Reload: strconv.Atoi(" + k + "): " + err.Error())
			return err
		}
		conns := make([]*bdc.Conn, len(v))
		for i, host := range v {
			conns[i] = bdc.NewConn(host, nil)
		}
		stgs[uint16(bd)] = bdc.NewBdClient(conns, interval, len(conns))
		xl.Infof("MultiStg.Reload: bd %v => hosts %v", bd, v)
	}

	p.RWMutex.Lock()
	p.stgs = stgs
	p.RWMutex.Unlock()
	return nil
}
