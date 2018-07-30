package lbdc

import (
	"strings"
	"sync"

	"github.com/qiniu/http/httputil.v1"
	"github.com/qiniu/xlog.v1"
)

var (
	ErrAccessBd  = httputil.NewError(572, "We encountered an internal error 0. Please try again")
	ErrAccessLbd = httputil.NewError(572, "We encountered an internal error 1. Please try again")
	ErrAccessIp  = httputil.NewError(572, "We encountered an internal error 2. Please try again")
	ErrAccessIdc = httputil.NewError(572, "We encountered an internal error 3. Please try again")
)

func shouldRetry(err error) bool {

	return err == ErrAccessLbd || err == ErrAccessIp
}

func xlogAclError(xl *xlog.Logger, err error) {

	var tag string
	switch err {
	case ErrAccessBd:
		tag = "bd"
	case ErrAccessLbd:
		tag = "lbd"
	case ErrAccessIp:
		tag = "ip"
	case ErrAccessIdc:
		tag = "idc"
	}
	xl.Xput([]string{"lbd.acl." + tag})
}

type AclConfig struct {
	MaxBdCount  int                 `json:"max_bd_count"`
	MaxLbdCount int                 `json:"max_lbd_count"`
	MaxIpCount  int                 `json:"max_ip_count"`
	MaxIdcCount int                 `json:"max_idc_count"`
	IdcBds      map[string][]uint16 `json:"idc_bds"`
}

// Access Controller
type Acl struct {
	AclConfig
	bds    map[uint16]int // 限制单 bd 故障的影响
	lbds   map[string]int // 限制单 lbd 故障的影响
	ips    map[string]int // 限制单 ip 故障的影响
	idcs   map[string]int // 限制单 idc 故障的影响
	bd2idc map[uint16]string
	lock   sync.Mutex
}

func NewAcl(cfg *AclConfig) *Acl {

	bd2idc := make(map[uint16]string)
	for idc, bds := range cfg.IdcBds {
		for _, bd := range bds {
			bd2idc[bd] = idc
		}
	}
	p := &Acl{
		AclConfig: *cfg,
		bds:       make(map[uint16]int),
		lbds:      make(map[string]int),
		ips:       make(map[string]int),
		idcs:      make(map[string]int),
		bd2idc:    bd2idc,
	}
	if p.MaxBdCount == 0 {
		p.MaxBdCount = 500
	}
	if p.MaxLbdCount == 0 {
		p.MaxLbdCount = 500
	}
	if p.MaxIpCount == 0 {
		p.MaxIpCount = 2000
	}
	if p.MaxIdcCount == 0 {
		p.MaxIdcCount = 3000
	}
	return p
}

func (p *Acl) Acquire(host string) (func(), error) {

	p.lock.Lock()
	defer p.lock.Unlock()

	lbdCount := p.lbds[host]
	if lbdCount >= p.MaxLbdCount {
		return nil, ErrAccessLbd
	}

	ip := host[:strings.LastIndex(host, ":")]
	ipCount := p.ips[ip]
	if ipCount >= p.MaxIpCount {
		return nil, ErrAccessIp
	}

	p.lbds[host] = lbdCount + 1
	p.ips[ip] = ipCount + 1

	fn := func() {
		p.lock.Lock()
		p.ips[ip]--
		p.lbds[host]--
		p.lock.Unlock()
	}
	return fn, nil

}

func (p *Acl) AcquireWithBd(host string, bd uint16) (func(), error) {

	p.lock.Lock()
	defer p.lock.Unlock()

	bdCount := p.bds[bd]
	if bdCount >= p.MaxBdCount {
		return nil, ErrAccessBd
	}

	lbdCount := p.lbds[host]
	if lbdCount >= p.MaxLbdCount {
		return nil, ErrAccessLbd
	}

	ip := host[:strings.LastIndex(host, ":")]
	ipCount := p.ips[ip]
	if ipCount >= p.MaxIpCount {
		return nil, ErrAccessIp
	}

	idc := p.bd2idc[bd]
	idcCount := p.idcs[idc]
	if idcCount >= p.MaxIdcCount {
		return nil, ErrAccessIdc
	}

	p.bds[bd] = bdCount + 1
	p.lbds[host] = lbdCount + 1
	p.ips[ip] = ipCount + 1
	p.idcs[idc] = idcCount + 1

	fn := func() {
		p.lock.Lock()
		p.idcs[idc]--
		p.ips[ip]--
		p.lbds[host]--
		p.bds[bd]--
		p.lock.Unlock()
	}
	return fn, nil
}
