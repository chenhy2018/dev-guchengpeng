package discoverd

import (
	"fmt"
	"math"
	"regexp"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"

	"github.com/qiniu/errors"
	"qbox.us/mgo3"
	"qbox.us/qdiscover/discover"

	"github.com/qiniu/http/httputil.v1"
	"github.com/qiniu/log.v1"
)

const (
	DefaultHeartbeatMissSecs = 10

	collSessionNum = 100
)

var (
	ErrNoSuchEntry       = httputil.NewError(612, "no such entry")
	ErrMinEnabledProtect = httputil.NewError(403, "operation is denied, please check min_enabled_ratio")
	ErrUnregisterEnabled = httputil.NewError(403, "operation is denied, can not unregister instance in the enabled state")
)

func newError(err error) error {
	if err == mgo.ErrNotFound {
		return ErrNoSuchEntry
	}
	return err
}

type M map[string]interface{}

type Config struct {
	ServiceColl                 mgo3.Config        `json:"service_coll"`
	HeartbeatMissSecs           int                `json:"heartbeat_miss_secs"`
	MinEnabledRatio             map[string]float64 `json:"min_enabled_ratio"`
	MinOnlineRatios             map[string]float64 `json:"min_online_ratio"`
	MinOnlineRatioDefault       float64            `json:"min_online_ratio_default"`
	OnlineCacheReloadIntervalMs int                `json:"online_cache_reload_interval_ms"`
}

type ServiceManager struct {
	session *mgo3.Session

	colls   []*mgo.Collection
	collIdx uint32

	/*
		{
			"<Node>": {
				"<Name>": [<ServiceInfo>, ...],
				...
			},
			...
		}
	*/
	onlineCache     map[string]map[string][]*discover.ServiceInfo
	onlineCacheLock sync.RWMutex

	Config
}

func NewServiceManager(cfg *Config) (s *ServiceManager, err error) {
	if cfg.HeartbeatMissSecs == 0 {
		cfg.HeartbeatMissSecs = DefaultHeartbeatMissSecs
	}

	session := mgo3.Open(&cfg.ServiceColl)
	colls := make([]*mgo.Collection, collSessionNum)
	for i := 0; i < collSessionNum; i++ {
		colls[i] = session.Coll.Copy().Collection
	}

	s = &ServiceManager{session: session, colls: colls, Config: *cfg}
	err = s.ensureIndexs()
	if err != nil {
		return
	}
	err = s.ReloadOnlineCache()
	if err != nil {
		return
	}
	go func() {
		interval := time.Duration(s.OnlineCacheReloadIntervalMs) * time.Millisecond
		if interval == 0 {
			interval = 5 * time.Second
		}
		for {
			time.Sleep(interval)
			err := s.ReloadOnlineCache()
			if err != nil {
				log.Error("ServiceManager: reloadOnlineCache failed", errors.Detail(err))
				continue
			}
			log.Debug("ServiceManager: reloadOnlineCache done.")
		}
	}()
	return
}

func (s *ServiceManager) ensureIndexs() error {
	err := s.session.Coll.EnsureIndex(mgo.Index{Key: []string{"addr"}, Unique: true})
	if err != nil {
		return err
	}
	return s.session.Coll.EnsureIndex(mgo.Index{Key: []string{"state", "name"}})
}

func (s *ServiceManager) getColl() *mgo.Collection {
	idx := atomic.AddUint32(&s.collIdx, 1)
	return s.colls[idx%uint32(len(s.colls))]
}

func (s *ServiceManager) Register(addr, name string, attrs discover.Attrs) (err error) {
	coll := s.getColl()
	change := mgo.Change{
		Update: M{"$set": M{"name": name, "lastUpdate": bson.Now(), "attrs": attrs}},
		Upsert: true,
	}
	var oldInfo discover.ServiceInfo
	if _, err = coll.Find(M{"addr": addr}).Apply(change, &oldInfo); err != nil {
		log.Warn("Register: update serviceInfo failed:", err)
		return
	}
	if oldInfo.Addr == "" { // 第一次心跳，增加 state 字段为 pending。
		err = s.SetLastChange(addr, "first register")
		if err != nil {
			return
		}
		err = coll.Update(
			M{"addr": addr, "state": M{"$exists": false}},
			M{"$set": M{"state": discover.StatePending}})
		if err != nil {
			log.Warn("Register: update state to pending failed:", err)
		}
	}
	return
}

func (s *ServiceManager) Unregister(addr string) (err error) {
	coll := s.getColl()
	var info discover.ServiceInfo
	if err = coll.Find(M{"addr": addr}).One(&info); err != nil {
		return newError(err)
	}
	if info.State == discover.StateEnabled {
		return ErrUnregisterEnabled
	}
	if err = coll.Remove(M{"addr": addr}); err != nil {
		err = newError(err)
	}
	return
}

func (s *ServiceManager) checkEnabledRatio(name string) (err error) {
	ratio := s.MinEnabledRatio[name]
	if ratio == 0.0 {
		return
	}
	coll := s.getColl()
	var infos []*discover.ServiceInfo
	if err = coll.Find(M{"name": name}).All(&infos); err != nil {
		if err != mgo.ErrNotFound {
			return newError(err)
		}
		return nil
	}

	var enabled int
	for _, info := range infos {
		if info.State == discover.StateEnabled {
			enabled++
		}
	}
	if enabled != 0 && float64(enabled-1)/float64(len(infos)) < ratio {
		err = ErrMinEnabledProtect
	}
	return
}

func (s *ServiceManager) changeState(addr string, state discover.State) (err error) {
	coll := s.getColl()
	if state == discover.StateDisabled {
		var info discover.ServiceInfo
		if err = coll.Find(M{"addr": addr}).One(&info); err != nil {
			return newError(err)
		}
		if err = s.checkEnabledRatio(info.Name); err != nil {
			return newError(err)
		}
	}

	if err = coll.Update(M{"addr": addr}, M{"$set": M{"state": state}}); err != nil {
		err = newError(err)
	}
	return
}

func (s *ServiceManager) Enable(addr string) (err error) {
	return s.changeState(addr, discover.StateEnabled)
}

func (s *ServiceManager) Disable(addr string) (err error) {
	return s.changeState(addr, discover.StateDisabled)
}

func (s *ServiceManager) Get(addr string) (*discover.ServiceInfo, error) {
	coll := s.getColl()
	var info discover.ServiceInfo
	if err := coll.Find(M{"addr": addr}).One(&info); err != nil {
		return nil, newError(err)
	}
	return &info, nil
}

func (s *ServiceManager) SetCfg(addr string, args *discover.CfgArgs) (err error) {
	coll := s.getColl()
	cfgKey := "cfg" + "." + args.Key
	if err = coll.Update(M{"addr": addr}, M{"$set": M{cfgKey: args.Value}}); err != nil {
		return newError(err)
	}
	return nil
}

func (s *ServiceManager) SetLastChange(addr, op string) (err error) {
	now := time.Now().Format("2006-01-02/15:04") //format for jumpbox bash.log time
	value := fmt.Sprintf("%s->%s", op, now)
	var info = discover.CfgArgs{Key: "last_change", Value: value}
	log.Debugf("%s setcfg - addr:%s, key:%s, value:%#v", op, addr, info.Key, info.Value)
	return s.SetCfg(addr, &info)
}

func (s *ServiceManager) newQueryCondWith(args *QueryArgs) M {
	var cond = M{}
	if args.Node != "" {
		cond["addr"] = M{"$regex": "^" + regexp.QuoteMeta(args.Node+":")}
	}
	if len(args.Name) > 0 {
		cond["name"] = M{"$in": args.Name}
	}
	switch discover.State(args.State) {
	case discover.StatePending, discover.StateEnabled, discover.StateDisabled:
		cond["state"] = args.State
	case discover.StateOnline:
		cond["state"] = discover.StateEnabled
		cond["lastUpdate"] = M{"$gt": bson.Now().Add(-time.Duration(s.HeartbeatMissSecs) * time.Second)}
	case discover.StateOffline:
		cond["$or"] = [2]M{
			M{"state": discover.StateDisabled},
			M{"state": discover.StateEnabled, "lastUpdate": M{"$lte": bson.Now().Add(-time.Duration(s.HeartbeatMissSecs) * time.Second)}},
		}
	}
	return cond
}

func (s *ServiceManager) Count(args *QueryArgs) (count int, err error) {
	coll := s.getColl()
	cond := s.newQueryCondWith(args)
	return coll.Find(cond).Count()
}

func (s *ServiceManager) ListAll(args *QueryArgs) (infos []*discover.ServiceInfo, err error) {
	if discover.State(args.State) == discover.StateOnline {
		infos, err = s.ListOnline(args.Node, args.Name)
		return
	}
	coll := s.getColl()
	cond := s.newQueryCondWith(args)
	err = coll.Find(cond).All(&infos)
	infos = SortUniq(infos)
	return
}

func (s *ServiceManager) getMinOnlineRatio(name string) float64 {

	ratio, ok := s.MinOnlineRatios[name]
	if !ok {
		return s.MinOnlineRatioDefault
	}
	return ratio
}

func (s *ServiceManager) ReloadOnlineCache() error {
	var infos []*discover.ServiceInfo
	err := s.getColl().Find(M{"state": discover.StateEnabled}).All(&infos)
	if err != nil {
		return errors.Info(err, "coll find all failed")
	}

	services := make(map[string][]*discover.ServiceInfo)
	for _, info := range infos {
		services[info.Name] = append(services[info.Name], info)
	}

	minLastUpdate := time.Now().Add(-time.Duration(s.HeartbeatMissSecs) * time.Second)
	for name, infos := range services {
		sort.Sort(byLastUpdates(infos))
		min := int(math.Ceil((s.getMinOnlineRatio(name) * float64(len(infos)))))
		var keep int
		for keep = len(infos) - 1; keep >= min; keep-- {
			if infos[keep].LastUpdate.After(minLastUpdate) {
				break
			}
		}
		services[name] = infos[:keep+1]
	}

	cache := make(map[string]map[string][]*discover.ServiceInfo)
	for _, infos := range services {
		for _, info := range infos {
			node := strings.SplitN(info.Addr, ":", 2)[0]
			if node == "" {
				continue
			}
			nodeServices, ok := cache[node]
			if !ok {
				nodeServices = make(map[string][]*discover.ServiceInfo)
				cache[node] = nodeServices
			}
			nodeServices[info.Name] = append(nodeServices[info.Name], info)
		}
	}
	cache[""] = services // for query without node

	s.onlineCacheLock.Lock()
	s.onlineCache = cache
	s.onlineCacheLock.Unlock()
	return nil
}

func (s *ServiceManager) setOnlineCache(cache map[string]map[string][]*discover.ServiceInfo) {

	s.onlineCacheLock.Lock()
	s.onlineCache = cache
	s.onlineCacheLock.Unlock()
}

func (s *ServiceManager) getOnlineCache() map[string]map[string][]*discover.ServiceInfo {

	s.onlineCacheLock.RLock()
	cache := s.onlineCache
	s.onlineCacheLock.RUnlock()
	return cache
}

func (s *ServiceManager) ListOnline(node string, names []string) (infos []*discover.ServiceInfo, err error) {

	cache := s.getOnlineCache()
	services, ok := cache[node]
	if !ok {
		return
	}
	if len(names) > 0 {
		for _, name := range names {
			if v, ok := services[name]; ok {
				infos = append(infos, v...)
			}
		}
	} else {
		for _, v := range services {
			infos = append(infos, v...)
		}
	}
	infos = SortUniq(infos)
	return
}

func (s *ServiceManager) List(args *QueryArgs, marker string, limit int) (infos []*discover.ServiceInfo, marker2 string, err error) {
	if limit <= 0 || limit > 1000 {
		limit = 1000
	}

	coll := s.getColl()
	cond := s.newQueryCondWith(args)
	if marker != "" {
		cond["addr"] = M{"$gte": marker}
	}
	err = coll.Find(cond).Sort("addr").Limit(limit + 1).All(&infos) // limit + 1 for marker
	if err != nil {
		return
	}

	if len(infos) == limit+1 {
		marker2 = infos[limit].Addr
		infos = infos[:limit]
	}
	return
}

func SortUniq(infos []*discover.ServiceInfo) []*discover.ServiceInfo {
	if len(infos) == 0 {
		return infos
	}
	sort.Sort(byAddrs(infos))
	n := 0
	prev := new(discover.ServiceInfo)
	for _, info := range infos {
		if info.Addr == prev.Addr {
			continue
		}
		prev = info
		infos[n] = info
		n++
	}
	return infos[:n]
}

type byAddrs []*discover.ServiceInfo

func (a byAddrs) Len() int           { return len(a) }
func (a byAddrs) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byAddrs) Less(i, j int) bool { return a[i].Addr < a[j].Addr }

type byLastUpdates []*discover.ServiceInfo

func (a byLastUpdates) Len() int           { return len(a) }
func (a byLastUpdates) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byLastUpdates) Less(i, j int) bool { return a[i].LastUpdate.After(a[j].LastUpdate) }
