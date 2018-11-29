package lbsocketproxy

import (
	"math/rand"
	"net"
	"sort"
	"sync/atomic"
	"time"

	"github.com/qiniu/log.v1"
	"qbox.us/iputil"

	"golang.org/x/net/proxy"
)

type LbSocketProxy struct {
	conf           Config
	proxies        dialers
	idx            uint32
	ShouldUseProxy ShouldUseProxy
}

type dialer struct {
	proxy.Dialer
	host                  string
	activeConnectionCount *int32
	lastFailTime          *int64
}

func (d *dialer) Dial(network, addr string) (c net.Conn, err error) {
	atomic.AddInt32(d.activeConnectionCount, 1)
	c, err = d.Dialer.Dial(network, addr)
	if err != nil {
		atomic.AddInt32(d.activeConnectionCount, -1)
		if e, ok := err.(net.Error); ok && e.Timeout() {
			atomic.StoreInt64(d.lastFailTime, time.Now().UnixNano())
		}
		return
	}
	return &conn{c, d.activeConnectionCount}, nil
}

type dialers []*dialer

func (d dialers) Len() int      { return len(d) }
func (d dialers) Swap(i, j int) { d[i], d[j] = d[j], d[i] }
func (d dialers) Less(i, j int) bool {
	return atomic.LoadInt32(d[i].activeConnectionCount) < atomic.LoadInt32(d[j].activeConnectionCount)
}

type conn struct {
	net.Conn
	activeConnectionCount *int32
}

func (c *conn) Close() error {
	atomic.AddInt32(c.activeConnectionCount, -1)
	return c.Conn.Close()
}

/*
type:
	all: 所有请求走代理
	default: 出idc请求走代理
*/

type Config struct {
	Hosts         []string          `json:"hosts"`
	DialTimeoutMs int               `json:"dial_timeout_ms"`
	TryTimes      int               `json:"try_times"`
	Auth          *proxy.Auth       `json:"auth"`
	Type          string            `json:"type"`
	Router        map[string]string `json:"router"`
	FailBanTimeMs int               `json:"fail_ban_time_ms"`
}

type ShouldUseProxy func(dstIP string) bool

var AllUseProxy = func(dstIP string) bool { return true }

func NewLbSocketProxy(conf *Config) (lbs *LbSocketProxy, err error) {
	if conf.TryTimes == 0 {
		conf.TryTimes = len(conf.Hosts)
	}
	if conf.FailBanTimeMs == 0 {
		conf.FailBanTimeMs = 5000
	}
	var proxies dialers
	for i := 0; i < len(conf.Hosts); i++ {
		forward := &net.Dialer{Timeout: time.Millisecond * time.Duration(conf.DialTimeoutMs)}
		p, err := proxy.SOCKS5("tcp", conf.Hosts[i], conf.Auth, forward)
		if err != nil {
			return nil, err
		}
		var activeConnectionCount int32
		var lastFailTime int64
		proxies = append(proxies, &dialer{
			Dialer: p,
			host:   conf.Hosts[i],
			activeConnectionCount: &activeConnectionCount,
			lastFailTime:          &lastFailTime,
		})
	}
	lbs = &LbSocketProxy{
		conf:    *conf,
		proxies: proxies,
	}
	if conf.Type == "all" {
		lbs.ShouldUseProxy = AllUseProxy
	} else if len(conf.Router) == 0 {
		lbs.ShouldUseProxy = iputil.NewDefaultIdcIpChecker().IsNotInSameIDC
	} else {
		ipt, err := iputil.NewIpCheckerWithIpMap(conf.Router)
		if err != nil {
			return nil, err
		}
		lbs.ShouldUseProxy = ipt.IsNotInSameIDC
	}
	return
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

var testCheckConn = func(c net.Conn, err error) {}

func (self *LbSocketProxy) Dial(addr net.Addr) (c net.Conn, err error) {
	direct := func() (net.Conn, error) {
		timeout := time.Millisecond * time.Duration(self.conf.DialTimeoutMs)
		return (&net.Dialer{Timeout: timeout}).Dial("tcp", addr.String())
	}
	if !self.ShouldUseProxy(addr.(*net.TCPAddr).IP.String()) {
		return direct()
	}
	var copyedProxies = make(dialers, len(self.proxies))
	copy(copyedProxies, self.proxies)
	sort.Sort(copyedProxies)
	var nextProxy int
	for i := 0; i < self.conf.TryTimes; i++ {
		var proxy *dialer
		if nextProxy == len(copyedProxies) {
			proxy = copyedProxies[rand.Intn(len(copyedProxies))]
		} else {
			lastFailTime := atomic.LoadInt64(copyedProxies[nextProxy].lastFailTime)
			if time.Since(time.Unix(lastFailTime/1e9, lastFailTime%1e9)) < time.Millisecond*time.Duration(self.conf.FailBanTimeMs) {
				nextProxy++
				i--
				continue
			}
			proxy = copyedProxies[nextProxy]
			nextProxy++
		}
		c, err = proxy.Dial("tcp", addr.String())
		if err == nil {
			log.Debugf("connect to %s use proxy %v success with local addr %v", addr.String(), c.RemoteAddr(), c.LocalAddr())
			return
		}
		log.Warnf("connect to %v with proxy %v failed: %v", addr.String(), proxy.host, err)
		if _, ok := err.(net.Error); !ok {
			break
		}
	}
	testCheckConn(c, err)
	if err != nil {
		return direct()
	}
	return
}
