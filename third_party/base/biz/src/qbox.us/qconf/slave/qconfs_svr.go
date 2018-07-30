package slave

import (
	"sync"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/qiniu/http/httputil.v1"
	"github.com/qiniu/log.v1"
	"github.com/qiniu/rpc.v1/lb.v2.1"
	"github.com/qiniu/xlog.v1"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"qbox.us/digest_auth"
	"qbox.us/errors"
	account "qbox.us/http/account.v2"
	"qbox.us/mgo2"
)

var (
	ErrUnacceptable  = httputil.NewError(401, "bad token: unacceptable")
	ErrSrvNotSupport = httputil.NewError(400, "api not supported on this server")
	ErrNegativeHours = httputil.NewError(400, "hours_before should not be negative")
)

// ------------------------------------------------------------------------

type keyArg struct {
	Id []string `json:"id"`
}

type Config struct {
	McHosts     []string // 互为镜像的Memcache服务
	AuthParser  account.AuthParser
	UidMgr      uint32 // 只接受这个管理员发过来的请求
	IsProxy     bool   //如果为true表示当前服务端作为一个代理节点使用
	ProxyConfig ProxyConfig
}

type ConfsSrvCfg struct {
	Client     lb.Config          `json:"client"`
	Failover   lb.Config          `json:"failover"`
	ClientTr   lb.TransportConfig `json:"client_tr"`
	FailoverTr lb.TransportConfig `json:"failover_tr"`
}

//IsProxy为true才生效
type ProxyConfig struct {
	AsyncRetryFailedIntervalSec int                    `json:"async_retry_failed_interval_sec"`
	ConfsSrv                    map[string]ConfsSrvCfg `json:"confs_srv"`
	AdminAk                     string                 `json:"admin_ak"`
	AdminSk                     string                 `json:"admin_sk"`
	Mgo                         mgo2.Config            `json:"mgo"`
}

type Service struct {
	account.Manager
	mcaches    []*memcache.Client
	Refreshers map[string]*lb.Client
	Coll       *mgo.Collection
	Config
}

// ------------------------------------------------------------------------

func New(cfg *Config) (p *Service, err error) {

	mcaches := make([]*memcache.Client, len(cfg.McHosts))
	for i, host := range cfg.McHosts {
		mcaches[i] = memcache.New(host)
	}

	if cfg.ProxyConfig.AsyncRetryFailedIntervalSec < 1 {
		cfg.ProxyConfig.AsyncRetryFailedIntervalSec = 60
	}

	refreshers := map[string]*lb.Client{}
	var coll *mgo.Collection
	if cfg.IsProxy {
		for k, lbCfg := range cfg.ProxyConfig.ConfsSrv {
			clientTr := digest_auth.NewTransport(cfg.ProxyConfig.AdminAk, cfg.ProxyConfig.AdminSk,
				lb.NewTransport(&lbCfg.ClientTr))
			lbCfg.Client.ShouldRetry = lb.ShouldRetry
			failoverTr := digest_auth.NewTransport(cfg.ProxyConfig.AdminAk, cfg.ProxyConfig.AdminSk,
				lb.NewTransport(&lbCfg.FailoverTr))
			lbCfg.Failover.ShouldRetry = lb.ShouldRetry
			refreshers[k] = lb.NewWithFailover(&lbCfg.Client, &lbCfg.Failover, clientTr, failoverTr, nil)
		}
		coll = mgo2.Open(&cfg.ProxyConfig.Mgo).Coll
		err = coll.EnsureIndex(mgo.Index{Key: []string{"create_time"}})
		if err != nil {
			return
		}
	}
	p = &Service{Config: *cfg, mcaches: mcaches, Refreshers: refreshers, Coll: coll}
	p.InitAccount(cfg.AuthParser)
	return
}

// ------------------------------------------------------------------------
type Item struct {
	Id         []string      `json:"id" bson:"id"`
	Idc        string        `json:"idc" bson:"idc"`
	CreateTime time.Time     `json:"create_time" bson:"create_time"`
	BsonId     bson.ObjectId `bson:"_id,omitempty"`
}

type FailedCountRet struct {
	Count int `json:"count"`
}

type FailedCountArgs struct {
	HoursBefore int `json:"hours_before"`
}

func (p *Service) WsFailedcount(args *FailedCountArgs, env *account.AdminEnv) (ret FailedCountRet, err error) {
	xl := xlog.New(env.W, env.Req)
	if env.Uid != p.UidMgr {
		err = ErrUnacceptable
		return
	}
	if args.HoursBefore < 0 {
		err = ErrNegativeHours
		return
	}
	if p.IsProxy {
		ret.Count, err = p.getFailedCount(xl, args.HoursBefore)
	} else {
		err = ErrSrvNotSupport
	}
	return
}

func (p *Service) WspRefresh(args *keyArg, env *account.AdminEnv) (err error) {
	xl := xlog.New(env.W, env.Req)
	if env.Uid != p.UidMgr {
		return ErrUnacceptable
	}
	if p.IsProxy {
		N := len(p.ProxyConfig.ConfsSrv)
		var wg sync.WaitGroup
		var errLck sync.Mutex
		wg.Add(N)
		for idc, cfg := range p.ProxyConfig.ConfsSrv {
			go func(idc string, proxyAddr []string) {
				defer wg.Done()
				xlCp := xl.Spawn()
				err1 := p.refreshConfs(idc, args.Id, xlCp)
				if err1 != nil {
					errLck.Lock()
					err = err1
					errLck.Unlock()
					xlCp.Error("proxy failed:", idc, proxyAddr, args.Id)
					err1 = p.insertToDb(idc, args.Id, xlCp)
					if err1 != nil {
						xlCp.Error("insertToDb failed:", idc, args.Id, err)
					}
				}
			}(idc, cfg.Client.Hosts)
		}
		wg.Wait()
	} else {
		err = p.refreshMc(args.Id, xl)
	}
	return err
}

func (p *Service) insertToDb(idc string, id []string, xl *xlog.Logger) (err error) {
	createTime := time.Now().UTC()
	for j := 0; j < 3; j++ {
		err = p.Coll.Insert(Item{Idc: idc, Id: id, CreateTime: createTime})
		if err != nil && !mgo.IsDup(err) {
			xl.Error("insertToDb failed:", id, idc, err)
			p.Coll.Database.Session.Refresh()
			time.Sleep(time.Millisecond * 100)
		} else {
			err = nil
			break
		}
	}
	return
}

func (p *Service) getFailedCount(xl *xlog.Logger, hoursBefore int) (count int, err error) {
	endTime := time.Now().UTC().Add(-time.Hour * time.Duration(hoursBefore))
	count, err = p.Coll.Find(bson.M{"create_time": bson.M{"$lt": endTime}}).Count()
	if err != nil {
		if err == mgo.ErrNotFound {
			return 0, nil
		}
		xl.Error("getFailedCount failed:", err)
		p.Coll.Database.Session.Refresh()
		time.Sleep(time.Millisecond * 100)
	}
	return
}

func (p *Service) refreshConfs(idc string, id []string, xl *xlog.Logger) (err error) {
	if refresher, ok := p.Refreshers[idc]; ok {
		params := map[string][]string{"id": id}
		err := refresher.CallWithForm(xl, nil, "/refresh", params)
		if err != nil {
			xl.Error("refreshConfs failed:", idc, id, err)
		}
		return err
	} else {
		err = errors.New("idc not found")
		xl.Error(err, idc)
	}
	return
}

func (p *Service) refreshMc(id []string, xl *xlog.Logger) (err error) {
	for i, mc := range p.mcaches {
		for _, each_id := range id {
			err1 := mc.Delete(each_id)
			if err1 != nil && err1 != memcache.ErrCacheMiss {
				xl.Warn("refresh mc failed:", p.McHosts[i], err1)
				err = err1
			}
		}
	}
	return
}

func (p *Service) LoopRefreshFailedItems() {
	go func() {
		for {
			log.Info("Start retry failed items...")
			iter := p.Coll.Find(nil).Iter()
			var item = Item{}
			for iter.Next(&item) {
				xl := xlog.NewDummy()
				err := p.refreshConfs(item.Idc, item.Id, xl)
				if err == nil {
					err = p.deleteFromDb(item, xl)
					if err != nil {
						xl.Error("deleteFromDb failed:", err)
					}
				} else {
					xl.Error("retry refresh failed:", err)
				}
			}
			iter.Close()
			time.Sleep(time.Second * time.Duration(p.ProxyConfig.AsyncRetryFailedIntervalSec))
		}
	}()
}

func (p *Service) deleteFromDb(item interface{}, xl *xlog.Logger) (err error) {
	err = p.Coll.Remove(item)
	if err != nil && err != mgo.ErrNotFound {
		xl.Error("deleteFromDb failed:", err, item)
		p.Coll.Database.Session.Refresh()
		time.Sleep(time.Millisecond * 100)
	} else {
		err = nil
		xl.Info("deleted:", item)
	}
	return
}
