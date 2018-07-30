// +build go1.5

package fopg

import (
	"sync"
	"time"

	"code.google.com/p/go.net/context"
	"github.com/qiniu/errors"
	"github.com/qiniu/xlog.v1"
)

type ClientProxy struct {
	*Client

	fops map[string]bool
	mu   sync.RWMutex
}

func NewClientProxy(cfg *Config, updateFopsDuration int) (*ClientProxy, error) {
	c, err := NewClient(cfg)
	if err != nil {
		return nil, err
	}

	cp := &ClientProxy{
		Client: c,
		fops:   make(map[string]bool),
	}
	if updateFopsDuration > 0 {
		cp.Refresh(xlog.NewContextWith(context.Background(), "refresh."+xlog.GenReqId()))
		go cp.updateFops(updateFopsDuration)
	}
	return cp, nil
}

func (cp *ClientProxy) Refresh(tpCtx context.Context) (err error) {
	fops, err := cp.Client.List(tpCtx)
	if err != nil {
		err = errors.Info(err, "Client.List").Detail(err).Warn()
		return
	}
	xl := xlog.FromContextSafe(tpCtx)
	xl.Info("ClientProxy.Refresh:", fops)

	m := make(map[string]bool, len(fops))
	for _, fop := range fops {
		m[fop] = true
	}
	cp.mu.Lock()
	cp.fops = m
	cp.mu.Unlock()

	return nil
}

func (cp *ClientProxy) updateFops(updateFopsDuration int) {
	dur := time.Duration(updateFopsDuration) * time.Second
	tpCtx := xlog.NewContextWith(context.Background(), "fopgClientProxy.updateFops")
	for {
		time.Sleep(dur)
		cp.Refresh(tpCtx)
	}
}

func (cp *ClientProxy) Has(fop string) bool {
	cp.mu.RLock()
	fops := cp.fops
	cp.mu.RUnlock()
	return fops[fop]
}
