package cfgapi

import (
	"sync"
	"sync/atomic"
	"time"

	"qbox.us/ptfd/cfgapi.v1/api"

	"github.com/qiniu/errors"
	"github.com/qiniu/xlog.v1"
)

var errServiceUnavailable = errors.New("service unavailable")

type dgHost struct {
	Hosts []string
	Index uint32
	Idc   string
}

type activeHost struct {
	Hosts []string
	Index uint32
}

type Config struct {
	Guid            string   `json:"guid"`
	Hosts           []string `json:"hosts"`
	UpdateIntervalS int      `json:"update_interval_s"`
}

type Client struct {
	Config
	client      *api.Client
	dgLock      sync.RWMutex
	dgHosts     map[uint32]*dgHost
	activeLock  sync.RWMutex
	activeHosts map[string]*activeHost // idc -> active hosts
}

func New(cfg *Config) (*Client, error) {

	client, err := api.New(cfg.Hosts, nil)
	if err != nil {
		return nil, errors.Info(err, "api.New").Detail(err)
	}
	p := &Client{
		Config: *cfg,
		client: client,
	}
	if cfg.UpdateIntervalS == 0 {
		cfg.UpdateIntervalS = 300 // 5min
	}
	p.updateDgs(xlog.NewDummy())
	go p.loopUpdate()
	return p, nil
}

// deprecated: please use HostsIdc
func (p *Client) Hosts(xl *xlog.Logger, dgid uint32) (hosts []string, ihost int, err error) {
	hosts, ihost, _, err = p.HostsIdc(xl, dgid)
	return
}

func (p *Client) HostsIdc(xl *xlog.Logger, dgid uint32) (hosts []string, ihost int, idc string, err error) {

	p.dgLock.RLock()
	dg, ok := p.dgHosts[dgid]
	p.dgLock.RUnlock()
	if ok {
		idx := atomic.AddUint32(&dg.Index, 1)
		return dg.Hosts, int(idx % uint32(len(dg.Hosts))), dg.Idc, nil
	}

	hosts, idc, err = p.client.HostsIdc(xl, p.Guid, dgid)
	if err != nil {
		return nil, 0, "", errors.Info(err, "cfg.Hosts").Detail(err)
	}

	p.dgLock.Lock()
	defer p.dgLock.Unlock()

	dg, ok = p.dgHosts[dgid]
	if ok {
		idx := atomic.AddUint32(&dg.Index, 1)
		return dg.Hosts, int(idx % uint32(len(dg.Hosts))), dg.Idc, nil
	}

	if len(hosts) == 0 {
		return nil, 0, "", errServiceUnavailable
	}

	p.dgHosts[dgid] = &dgHost{Hosts: hosts, Idc: idc}
	return hosts, 0, idc, nil
}

func (p *Client) Actives(xl *xlog.Logger, idc string) ([]string, int, error) {

	p.activeLock.RLock()
	active, ok := p.activeHosts[idc]
	p.activeLock.RUnlock()
	if ok {
		idx := atomic.AddUint32(&active.Index, 1)
		return active.Hosts, int(idx % uint32(len(active.Hosts))), nil
	}

	dgs, err := p.client.IdcDgs(xl, p.Guid, idc)
	if err != nil {
		return nil, 0, errors.Info(err, "cfg.Dgs").Detail(err)
	}

	hosts := make([]string, 0, len(dgs))
	for _, dg := range dgs {
		if !dg.Writable || dg.Repair {
			continue
		}
		hosts = append(hosts, dg.Hosts[0])
	}

	p.activeLock.Lock()
	defer p.activeLock.Unlock()

	active, ok = p.activeHosts[idc]
	if ok {
		idx := atomic.AddUint32(&active.Index, 1)
		return active.Hosts, int(idx % uint32(len(active.Hosts))), nil
	}

	if len(hosts) == 0 {
		return nil, 0, errServiceUnavailable
	}

	active = &activeHost{Hosts: hosts}
	p.activeHosts[idc] = active
	return active.Hosts, 0, nil
}

// -----------------------------------------------------------------------------

func (p *Client) updateDgs(xl *xlog.Logger) {

	xl.Infof("Client.updateDgs: start...")
	dgs, err := p.client.Dgs(xl, p.Guid)
	if err != nil {
		xl.Errorf("Client.updateDgs: cfg.Dgs %v failed => %v", p.Guid, err)
		return
	}

	actives := make(map[string]*activeHost)
	dgHosts := make(map[uint32]*dgHost, len(dgs))
	for _, dg := range dgs {
		xl.Infof("stg.update: dg => %+v", dg)
		if len(dg.Hosts) == 0 || len(dg.Dgids) == 0 {
			continue
		}
		dg1 := &dgHost{Hosts: dg.Hosts, Idc: dg.Idc}
		for _, dgid := range dg.Dgids {
			dgHosts[dgid] = dg1
		}
		if !dg.Writable || dg.Repair {
			continue
		}
		active, ok := actives[dg.Idc]
		if !ok {
			active = new(activeHost)
			actives[dg.Idc] = active
		}
		active.Hosts = append(active.Hosts, dg.Hosts[0])
	}

	p.dgLock.Lock()
	p.dgHosts = dgHosts
	p.dgLock.Unlock()

	p.activeLock.Lock()
	p.activeHosts = actives
	p.activeLock.Unlock()

	xl.Infof("Client.updateDgs: len(dgids) %v, len(hosts) %v", len(dgHosts), len(actives))
	return
}

func (p *Client) loopUpdate() {

	c := time.Tick(time.Duration(p.UpdateIntervalS) * time.Second)
	for _ = range c {
		p.updateDgs(xlog.NewDummy())
	}
}
