package master

import (
	"strings"
	"syscall"
	"time"

	"github.com/qiniu/errors"
	"github.com/qiniu/log.v1"
	"github.com/qiniu/xlog.v1"
	account "qbox.us/http/account.v2"
	"qbox.us/qmset/master/mbloom"
	. "qbox.us/qmset/proto"
)

// ------------------------------------------------------------------------

type FlipsNotifier interface {
	FlipsNotify(grpIds []string, clear bool)
}

type Config struct {
	FlipsNotifier // 向 slaves 发送通知
	FlipCfgs      []*FlipConfig
	AuthParser    account.AuthParser
	UidMgr        uint32 // 只接受这个管理员发过来的请求
	LogSetFull    bool
}

type Service struct {
	account.Manager
	grps map[string]Flipper
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
			err = errors.Info(err, "mset/master.New: initFlip failed", *flipCfg).Detail(err)
			return
		}
	}
	return
}

func (p *Service) initFlip(flipCfg *FlipConfig) (err error) {

	grpIds := make([]string, 0, len(flipCfg.Msets))

	for _, cfg := range flipCfg.Msets {
		id := cfg.Id
		if _, ok := p.grps[id]; ok {
			return syscall.EEXIST
		}
		grp := newMsetGroup(cfg.Max)
		p.grps[id] = grp
		grpIds = append(grpIds, id)
	}

	for _, cfg := range flipCfg.Mblooms {
		id := cfg.Id
		if _, ok := p.grps[id]; ok {
			return syscall.EEXIST
		}
		grp := mbloom.New(cfg.Max, cfg.Fp)
		p.grps[id] = grp
		grpIds = append(grpIds, id)
	}

	go p.flipRoutine(grpIds, flipCfg.Expires)

	return nil
}

func (p *Service) flipRoutine(grpIds []string, expires int) {

	notifier := p.FlipsNotifier
	grps := p.grps

	notifier.FlipsNotify(grpIds, true)

	c := time.Tick(time.Duration(expires) * time.Second)
	for _ = range c {
		for _, id := range grpIds {
			if grp, ok := grps[id]; ok {
				grp.Flip()
			} else {
				log.Warn("flipRoutine: group not found -", id)
			}
		}
		notifier.FlipsNotify(grpIds, false)
	}
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
		err = errors.Info(syscall.ENOENT, "qmset.Master: group not found", args.Id)
		return
	}

	msetg, ok := grp.(*msetGroup)
	if !ok {
		return ErrNotMsetGroup
	}

	log := xlog.New(env.W, env.Req)
	for _, kv := range args.KeyValues {
		pos := strings.Index(kv, ":")
		if pos < 0 {
			err = errors.Info(ErrInvalidKeyValues, "qset.Master.Adds: invalid key-values", kv)
			return
		}
		err2 := msetg.Add(kv[:pos], kv[pos+1:])
		if err2 != nil { // exists or full?
			if err2 == ErrSetFull && !p.LogSetFull {
				continue
			}
			if err2 != syscall.EEXIST {
				log.Warnf("qmset.Master.Adds key %v value %v failed %v", kv[:pos], kv[pos+1:], err2)
			}
			continue
		}
	}
	return
}

// ------------------------------------------------------------------------

type getArgs struct {
	Id  string `json:"c"`
	Key string `json:"k"`
}

//
// POST /get?c=<GrpName>&k=<SetId>
//
func (p *Service) WspGet(args *getArgs, env *account.AdminEnv) (values []string, err error) {

	if env.Uid != p.UidMgr {
		return nil, ErrUnacceptable
	}

	if grp, ok := p.grps[args.Id]; ok {
		if msetg, ok := grp.(*msetGroup); ok {
			return msetg.Get(args.Key), nil
		}
		err = ErrNotMsetGroup
		return
	}
	err = syscall.ENOENT
	return
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
