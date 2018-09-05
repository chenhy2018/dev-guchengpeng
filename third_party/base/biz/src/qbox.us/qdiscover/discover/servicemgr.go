package discover

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"labix.org/v2/mgo/bson"

	"github.com/qiniu/log.v1"
	"github.com/qiniu/rpc.v1"
)

const (
	DefaultPersistDir = "./discover"
)

var (
	PersistFname         = "services.local"
	DiscoverdRecoverTime = time.Minute
	DefaultTimeoutSecs   = 3
)

type Config struct {
	DiscoverHosts       []string `json:"discover_hosts"`
	DiscoverTimeoutSecs int      `json:"discover_timeout"`
	PersistDir          string   `json:"persist_dir"`         // 将从 discover 取过来的服务列表保存到本地，这里指定保存的本地目录
	Node                string   `json:"node"`                // 管理的节点名
	Watchs              []string `json:"watchs"`              // 管理的在线服务名
	FetchIntervalSecs   int      `json:"fetch_interval_secs"` // 后台到 discover 取服务列表的周期

	MinServicesNum       int     `json:"min_services_num"`        // 当实例数变化到 MinServicesNum 以下时，不触发 ChangeNotify
	MaxServicesDescRatio float64 `json:"max_services_desc_ratio"` // 当实例数下降比例大于 MaxServicesDescRatio 时，不触发 ChangeNotify

	IsServicesChanged func([]*ServiceInfo, []*ServiceInfo) bool // 如果为空, 默认判断服务地址是否改变
}

type ServiceLister interface {
	ServiceListAllEx(l rpc.Logger, ret interface{}, args *QueryArgs) error
}

type ServiceManager struct {
	services      []*ServiceInfo
	mu            sync.Mutex // guard services
	change        chan bool
	persistPath   string
	client        ServiceLister
	lastFetchFail time.Time

	Config
}

func NewServiceManager(cfg *Config) (*ServiceManager, error) {
	transport := rpc.NewTransportTimeout(time.Second, 0)
	client := New(cfg.DiscoverHosts, transport)

	// discover client timeout
	if cfg.DiscoverTimeoutSecs == 0 {
		cfg.DiscoverTimeoutSecs = DefaultTimeoutSecs
	}
	client.Conn.Client.Timeout = time.Duration(cfg.DiscoverTimeoutSecs) * time.Second

	if cfg.IsServicesChanged == nil {
		cfg.IsServicesChanged = isAddrsChanged
	}
	return newServiceManager(cfg, client)
}

func initConfig(cfg *Config) error {
	if len(cfg.DiscoverHosts) == 0 {
		return errors.New("empty discover hosts")
	}
	if cfg.PersistDir == "" {
		cfg.PersistDir = DefaultPersistDir
	}
	if err := os.MkdirAll(cfg.PersistDir, 0777); err != nil {
		return err
	}
	return nil
}

func newServiceManager(cfg *Config, client ServiceLister) (s *ServiceManager, err error) {
	if err = initConfig(cfg); err != nil {
		return nil, err
	}
	s = &ServiceManager{
		change:      make(chan bool, 1),
		persistPath: filepath.Join(cfg.PersistDir, PersistFname),
		client:      client,
		Config:      *cfg,
	}

	// 先从 qdiscover 获取服务，如果失败，从本地持久化文件获取。
	services, err := s.fetchServices()
	if err != nil {
		log.Warn("fetchServices failed:", err)
		services, err = s.loadFromPersistFile()
		if err != nil {
			log.Warn("loadFromPersistFile failed:", err)
			return nil, err
		}
		log.Info("loadFromPersistFile success")
	}
	sort.Sort(byAddrs(services))
	printServices(services)
	s.services = services
	if err = s.writeToPersistFile(); err != nil {
		log.Warn("writeToPersistFile failed:", err)
		return
	}

	if cfg.FetchIntervalSecs != 0 { // FetchIntervalSecs = 0 时，服务列表改变不能够从 AddrsChangeNotify 获取通知。
		go func() {
			for {
				time.Sleep(time.Duration(cfg.FetchIntervalSecs) * time.Second)
				s.watchServices()
			}
		}()
	}
	return s, nil
}

func (s *ServiceManager) ChangeNotify() <-chan bool {
	return s.change
}

func (s *ServiceManager) Services() []*ServiceInfo {
	s.mu.Lock()
	services := s.services
	s.mu.Unlock()
	return services
}

func (s *ServiceManager) fetchServices() ([]*ServiceInfo, error) {
	args := &QueryArgs{
		Node:  s.Node,
		Name:  s.Watchs,
		State: string(StateOnline),
	}
	var list ServiceListRet
	err := s.client.ServiceListAllEx(nil, &list, args)
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

func (s *ServiceManager) loadFromPersistFile() ([]*ServiceInfo, error) {
	b, err := ioutil.ReadFile(s.persistPath)
	if err != nil {
		return nil, err
	}
	var list ServiceListRet
	err = bson.Unmarshal(b, &list)
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

func (s *ServiceManager) writeToPersistFile() error {
	s.mu.Lock()
	services := s.services
	s.mu.Unlock()

	list := &ServiceListRet{Items: services}
	b, err := bson.Marshal(list)
	if err != nil {
		return err
	}

	newFile := fmt.Sprintf("%s.%d", s.persistPath, time.Now().UnixNano())
	if err = ioutil.WriteFile(newFile, b, 0666); err != nil {
		return err
	}
	linkFile := newFile + ".link"
	if err = os.Link(newFile, linkFile); err != nil {
		return err
	}
	return os.Rename(linkFile, s.persistPath)
}

func (s *ServiceManager) watchServices() {
	services, err := s.fetchServices()
	if err != nil {
		log.Warn("watchServices - fetchServices failed:", err)
		s.lastFetchFail = time.Now()
		return
	}
	// 有一种情况是 discover 服务挂掉一段时间，恢复后由于心跳已经都丢失，所以 fetch 到的在线服务列表很少或者是空的。
	// 避免这种情况发生的一种方法是在 fetch 失败后记录 lastFetchFail 的时间。
	// 每次获取成功，将当前时间减去 lastFetchFail 的时间，如果小于 DiscoverdRecoverTime，则认为获取到的服务列表还不可信。
	if time.Now().Sub(s.lastFetchFail) < DiscoverdRecoverTime {
		log.Info("fetchServices ok, but need wait a moment.")
		return
	}
	sort.Sort(byAddrs(services))
	s.mu.Lock()
	oservices := s.services
	s.mu.Unlock()

	// 对于实例数骤减的保护措施
	newNum := len(services)
	oldNum := len(oservices)
	if s.MinServicesNum > 0 && newNum < s.MinServicesNum {
		log.Warnf("watchServices - protect, meet min_services_num:%d, now:%d, old:%d ", s.MinServicesNum, newNum, oldNum)
		return
	}
	descNum := oldNum - newNum
	if s.MaxServicesDescRatio > 0 && float64(descNum)/float64(oldNum) > s.MaxServicesDescRatio {
		log.Warnf("watchServices - protect, meet max_services_desc_ratio:%v, now:%d, old:%d", s.MaxServicesDescRatio, newNum, oldNum)
		return
	}

	if s.IsServicesChanged(services, oservices) {
		log.Info("watchServices - changed", len(services))
		printServices(services)
		s.mu.Lock()
		s.services = services
		s.mu.Unlock()
		if err = s.writeToPersistFile(); err != nil {
			log.Warn("watchServices - writeToPersistFile failed:", err)
		}
		s.change <- true
	}
}

func printServices(services []*ServiceInfo) {
	for i, service := range services {
		log.Infof("%d\t%v", i+1, service)
	}
}

func isAddrsChanged(a, b []*ServiceInfo) bool {
	if len(a) != len(b) {
		return true
	}
	for i := 0; i < len(a); i++ {
		if a[i].Addr != b[i].Addr {
			return true
		}
	}
	return false
}

type byAddrs []*ServiceInfo

func (a byAddrs) Len() int           { return len(a) }
func (a byAddrs) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byAddrs) Less(i, j int) bool { return a[i].Addr < a[j].Addr }
