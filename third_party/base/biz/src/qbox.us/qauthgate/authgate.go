package qauthgate

import (
	"net/http"
	"sync"
	"sync/atomic"

	"labix.org/v2/mgo"

	"qbox.us/servend/account"
	"qbox.us/servend/proxy_auth"

	"github.com/qiniu/errors"
	"github.com/qiniu/http/httputil.v1"

	gohttputil "net/http/httputil"
)

var ErrBadToken = httputil.NewError(401, "bad token")
var ErrServiceBusy = httputil.NewError(612, "service busy")
var ErrServiceNotFound = httputil.NewError(612, "service not found")
var ErrServerNotFound = httputil.NewError(612, "server not found")

/* ------------------------------------------------------------------------

* Entry:

	_id: <Host> // 比如 rs.qiniu.com
	alias: [<Host1>, <Host2>, ...]	// 同一个服务有可能有多个对外 Host，比如 rs.qbox.me, rs.qbox.me:8888
	items: [<Item1>, <Item2>, ...]

* Item:

	server: <Ip:Port> // RealServer
	disabled: <Disabled>

// ----------------------------------------------------------------------*/

type Config struct {
	AuthParser account.AuthParser
	Coll       *mgo.Collection
}

// ------------------------------------------------------------------------

type routeEntry struct {
	Server   string `bson:"server"`
	Disabled int32  `bson:"disabled"` // 非0表示这个host可能在升级，暂时不可用
	Active   int32  `bson:"-"`        // 当前有多少活跃的连接
}

type routeInfo struct {
	items []*routeEntry
	mutex sync.RWMutex
}

//@@TODO: 考虑重试逻辑
//
func (p *routeInfo) selectHost() (e *routeEntry, err error) {

	activeMin := int32(0x7fffffff)

	p.mutex.RLock()
	for _, te := range p.items {
		if atomic.LoadInt32(&te.Disabled) != 0 {
			continue
		}
		active := atomic.LoadInt32(&te.Active)
		if active < activeMin {
			e, activeMin = te, active
		}
	}
	p.mutex.RUnlock()

	if e == nil {
		return nil, ErrServiceBusy
	}
	return
}

func (p *routeInfo) selectServer(server string) (e *routeEntry, err error) {

	p.mutex.RLock()
	for _, te := range p.items {
		if te.Server == server {
			p.mutex.RUnlock()
			return te, nil
		}
	}
	p.mutex.RUnlock()

	return nil, ErrServerNotFound
}

func (p *routeInfo) refresh(newEntries []*routeEntry) {

	oldEntries := p.items
	for i, ne := range newEntries {
		if oe, ok := inRoute(ne.Server, oldEntries); ok {
			atomic.StoreInt32(&oe.Disabled, ne.Disabled)
			newEntries[i] = oe
		}
	}

	p.mutex.Lock()
	p.items = newEntries
	p.mutex.Unlock()
}

func inRoute(server string, oldEntries []*routeEntry) (*routeEntry, bool) {

	for _, oe := range oldEntries {
		if oe.Server == server {
			return oe, true
		}
	}
	return nil, false
}

// ------------------------------------------------------------------------

type Service struct {
	router      map[string]*routeInfo // host => $route
	mutexRouter sync.RWMutex

	main      map[string][]string // 主host => $alias
	mutexMain sync.Mutex

	Config
}

// ------------------------------------------------------------------------

type M map[string]interface{}

type entry struct {
	Host  string        `bson:"_id"`
	Alias []string      `bson:"alias"`
	Items []*routeEntry `bson:"items"`
}

var all = M{}

func New(cfg *Config) (p *Service, err error) {

	router := make(map[string]*routeInfo) // host => $route
	main := make(map[string][]string)     // 主host => $alias

	iter := cfg.Coll.Find(all).Iter()
	for {
		var e entry
		if !iter.Next(&e) {
			break
		}
		route := &routeInfo{items: e.Items}
		router[e.Host] = route
		if len(e.Alias) > 0 {
			for _, host := range e.Alias {
				router[host] = route
			}
			main[e.Host] = e.Alias
		}
	}
	err = iter.Err()
	if err != nil {
		err = errors.Info(err, "authgate.New: coll.Load failed").Detail(err)
		return
	}

	p = &Service{
		router: router,
		main:   main,
		Config: *cfg,
	}
	return
}

// ------------------------------------------------------------------------

func (p *Service) getRoute(host string) (route *routeInfo, ok bool) {

	p.mutexRouter.RLock()
	route, ok = p.router[host]
	p.mutexRouter.RUnlock()
	return
}

func (p *Service) refreshRoute(host string) (err error) {

	var e entry
	err = p.Coll.FindId(host).One(&e)
	if err != nil {
		err = errors.Info(err, "authgate.refreshRoute: find failed", host).Detail(err)
		return
	}

	p.mutexRouter.Lock()
	route, ok := p.router[host]
	if !ok {
		route = new(routeInfo)
		p.router[host] = route
	}
	p.mutexRouter.Unlock()

	p.mutexMain.Lock()
	defer p.mutexMain.Unlock()

	if alias, ok := p.main[host]; ok {
		for _, v := range alias {
			if inSet(v, e.Alias) {
				continue
			}
			p.mutexRouter.Lock()
			delete(p.router, v)
			p.mutexRouter.Unlock()
		}
	}

	if len(e.Alias) > 0 {
		for _, ahost := range e.Alias {
			p.mutexRouter.Lock()
			p.router[ahost] = route
			p.mutexRouter.Unlock()
		}
		p.main[host] = e.Alias
	} else {
		delete(p.main, host)
	}

	route.refresh(e.Items)
	return nil
}

func inSet(v string, set []string) bool {

	for _, e := range set {
		if v == e {
			return true
		}
	}
	return false
}

// ------------------------------------------------------------------------

func (p *Service) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	user, err := p.AuthParser.ParseAuth(req)
	if err != nil {
		err = errors.Info(ErrBadToken, "ParseAuth failed").Detail(err)
		httputil.Error(w, err)
		return
	}

	auth := proxy_auth.MakeAuth(user)
	req.Header.Set("Authorization", auth)

	route, ok := p.getRoute(req.Host)
	if !ok {
		httputil.Error(w, ErrServiceNotFound)
		return
	}

	e, err := route.selectHost()
	if err != nil {
		httputil.Error(w, err)
		return
	}

	atomic.AddInt32(&e.Active, 1)
	defer atomic.AddInt32(&e.Active, -1)

	proxy(w, req, e.Server)
}

// ------------------------------------------------------------------------

func nilDirector(req *http.Request) {}

var theProxy = &gohttputil.ReverseProxy{Director: nilDirector}

func proxy(w http.ResponseWriter, req *http.Request, server string) {

	req.URL.Scheme = "http"
	req.URL.Host = server
	theProxy.ServeHTTP(w, req)
}

// ------------------------------------------------------------------------
