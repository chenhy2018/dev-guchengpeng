package slave

import (
	"github.com/qiniu/errors"
	"github.com/qiniu/rpc.v1"
	"github.com/qiniu/xlog.v1"
	account "qbox.us/http/account.v2"
	"qbox.us/qmset/master/mbloom"
	. "qbox.us/qmset/proto"
	"strings"
	"sync"
	"syscall"
)

// ------------------------------------------------------------------------

type AddNotifier interface {
	AddNotify(l rpc.Logger, id string, kvs []string)
}

type Config struct {
	AddNotifier
	FlipCfgs   []*FlipConfig
	AuthParser account.AuthParser
	UidMgr     uint32 // 只接受这个管理员发过来的请求
}

type Service struct {
	account.Manager

	grps  map[string]Flipper
	mutex sync.RWMutex

	Config
}

// ------------------------------------------------------------------------

func New(cfg *Config) (p *Service, err error) {

	grps := make(map[string]Flipper)

	p = &Service{Config: *cfg, grps: grps}
	p.InitAccount(cfg.AuthParser)

	for _, flipCfg := range cfg.FlipCfgs {
		err = p.initFlip(flipCfg)
		if err != nil {
			err = errors.Info(err, "mset/slave.New: initFlip failed", *flipCfg).Detail(err)
			return
		}
	}
	return
}

func (p *Service) initFlip(flipCfg *FlipConfig) (err error) {

	for _, cfg := range flipCfg.Msets {
		id := cfg.Id
		if _, ok := p.grps[id]; ok {
			return syscall.EEXIST
		}
		grp := newMsetGroup(cfg.Max)
		p.grps[id] = grp
	}

	for _, cfg := range flipCfg.Mblooms {
		id := cfg.Id
		if _, ok := p.grps[id]; ok {
			return syscall.EEXIST
		}
		grp := mbloom.New(cfg.Max, cfg.Fp)
		p.grps[id] = grp
	}
	return nil
}

// ------------------------------------------------------------------------

type addsArgs struct {
	Id        string   `json:"c"`
	KeyValues []string `json:"kv"`
}

//
// POST /adds?c=<GrpName>&kv=<KeyValue1>&kv=<KeyValue2>&kv=...
//
func (p *Service) WspAdds(args *addsArgs, env *account.AdminEnv) (err error) {

	if env.Uid != p.UidMgr {
		return ErrUnacceptable
	}

	grp, ok := p.grps[args.Id]
	if !ok {
		err = errors.Info(syscall.ENOENT, "qmset.Slave: group not found", args.Id)
		return
	}

	msetg, ok := grp.(*msetGroup)
	if !ok {
		return ErrNotMsetGroup
	}

	log := xlog.New(env.W, env.Req)
	kvs := args.KeyValues
	n := 0

	for _, kv := range kvs {
		pos := strings.Index(kv, ":")
		if pos < 0 {
			err = errors.Info(ErrInvalidKeyValues, "qset.Slave.Adds: invalid key-values", kv)
			return
		}
		err2 := msetg.Add(kv[:pos], kv[pos+1:])
		if err2 != nil { // exists or full?
			if err2 != syscall.EEXIST {
				log.Warn("qmset.Slave.Adds failed:", err2)
			}
			continue
		}
		kvs[n] = kv
		n++
	}
	if n != 0 {
		p.AddNotify(log, args.Id, kvs[:n])
	}
	return
}

// ------------------------------------------------------------------------

type flipsArgs struct {
	GrpIds  []string `json:"c"`
	DoClear int      `json:"clear"`
}

//
// POST /flips?c=<GrpName>&clear=<DoClear>
//
func (p *Service) WspFlips(args *flipsArgs, env *account.AdminEnv) (err error) {

	if env.Uid != p.UidMgr {
		return ErrUnacceptable
	}

	for _, id := range args.GrpIds {
		grp, ok := p.grps[id]
		if !ok {
			err = errors.Info(syscall.ENOENT, "qmset.Slave: group not found", id)
			return
		}
		if args.DoClear != 0 {
			switch g := grp.(type) {
			case *msetGroup:
				g.Clear()
			case *mbloom.Filter:
			default:
				return syscall.EINVAL
			}
		} else {
			grp.Flip()
		}
	}
	return nil
}

// ------------------------------------------------------------------------

//
// POST /badd?c=<GrpName>&v=<Value1>&v=<Value2>&v=...
//
func (p *Service) WspBadd(args *mbloom.BaddArgs, env *account.AdminEnv) (err error) {

	if env.Uid != p.UidMgr {
		return ErrUnacceptable
	}
	return mbloom.WspBadd(p.grps, args, env)
}

//
// POST /bchk?c=<GrpName>&v=<Value1>&v=<Value2>&v=...
//
func (p *Service) WspBchk(args *mbloom.BchkArgs, env *account.AdminEnv) (idxs []int, err error) {

	if env.Uid != p.UidMgr {
		return nil, ErrUnacceptable
	}
	return mbloom.WspBchk(p.grps, args, env)
}

// ------------------------------------------------------------------------
