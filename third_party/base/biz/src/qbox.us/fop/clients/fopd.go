package clients

import (
	"encoding/json"
	"io"
	"math"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/qiniu/log.v1"
	"github.com/qiniu/xlog.v1"
	"qbox.us/errors"
	"qbox.us/fop"
)

const (
	DefaultRetryTimes        = 2
	DefaultFailRetryInterval = 30                  // 30s
	DefaultTimeout           = 60 * 60 * 24 * 1000 // unit:millisecond 1day
)

var (
	ErrServiceNotAvailable = errors.New("service not available")
	ErrOpTimeout           = errors.New("fopd op timeout")
	ErrCantRetry           = errors.New("fopd op can not retry")
)

type FopdConfig struct {
	Servers           []FopdServer
	WhetherCache      map[string]bool
	Timeouts          map[string]int
	LoadBalanceMode   map[string]int
	DefaultTimeout    int
	RetryTimes        int
	FailRetryInterval int
}

type FopdServer struct {
	Host string   `json:"host"`
	Cmds []string `json:"cmds"`
}

//--------------------------------------------------------------------------------
// statedConns - manage conns of one fop

type statedConns struct {
	conns             []*FopdConn
	retryTimes        int
	failRetryInterval int64
	needCache         bool
	timeouts          time.Duration
	lastIndex         uint32
	lbMode            int
}

func (sc *statedConns) pickConn() (conn *FopdConn, index int, err error) {
	n := len(sc.conns)
	switch sc.lbMode {
	case 0: // 短时间Fop，顺序选择
		index = int(atomic.AddUint32(&sc.lastIndex, 1))
		for i := 0; i < n; i++ {
			index = (index + 1) % n
			lastFailedTime := sc.conns[index].LastFailedTime()
			if lastFailedTime == 0 || time.Now().Unix()-lastFailedTime >= sc.failRetryInterval {
				log.Debugf("pickConn lb0 - index: %d, host:%s\n", index, sc.conns[index].host)
				return sc.conns[index], index, nil
			}
		}
	case 1: // 长时间Fop，选择正在处理的任务数最小的Host
		ns := make([]int64, n)
		for i, conn := range sc.conns {
			lastFailedTime := conn.LastFailedTime()
			if lastFailedTime == 0 || time.Now().Unix()-lastFailedTime >= sc.failRetryInterval {
				ns[i] = conn.ProcessingNum()
			} else {
				ns[i] = math.MaxInt64 // the conn can not be chosen
			}
		}
		log.Debug("ns:", ns)
		index = minLoadIndex(ns)
		log.Debugf("pickConn lb1 - index: %d, host:%s\n", index, sc.conns[index].host)
		return sc.conns[index], index, nil
	}

	return nil, 0, ErrServiceNotAvailable
}

// 从值最小的数中随机选一个，返回它在 ns 中的索引。
func minLoadIndex(ns []int64) int {
	if len(ns) == 0 {
		panic("minLoadIndex: ns must have elements")
	}

	// 找到最小值
	minVal := ns[0]
	for _, n := range ns {
		if n < minVal {
			minVal = n
		}
	}

	// 构造最小值对应的 indexs
	minIndexs := make([]int, 0, len(ns))
	for i, n := range ns {
		if n == minVal {
			minIndexs = append(minIndexs, i)
		}
	}

	// 从最小值的 indexs 中随机选择一个
	randVal := rand.Intn(len(minIndexs))
	return minIndexs[randVal]
}

func getIp(host string) string {
	ip := host
	if idx := strings.Index(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}
	return ip
}

func shuffleFopds(conns []*FopdConn) {

	ip2conns := make(map[string][]*FopdConn)
	ipList := make([]string, 0)
	for _, conn := range conns {
		ip := getIp(conn.host)
		if _, ok := ip2conns[ip]; !ok {
			ipList = append(ipList, ip)
		}
		ip2conns[ip] = append(ip2conns[ip], conn)
	}

	idx := 0
	for i, finish := 0, false; !finish; i++ {
		finish = true
		for _, ip := range ipList {
			ipconns, ok := ip2conns[ip]
			if !ok {
				panic("Shuffle bug: cannot reach: !ok")
			}
			if i < len(ipconns) {
				conns[idx] = ipconns[i]
				idx++
				finish = false
			}
		}
	}
	if idx != len(conns) {
		panic("Shuffle bug: cannot reach: idx != len(conns)")
	}
}

//--------------------------------------------------------------------------------
// Fopd - manage statedConns of all fop

type Fopd struct {
	mu     sync.RWMutex
	fopds  map[string]*statedConns
	status *Status
}

func NewFopd(cfg *FopdConfig, transport http.RoundTripper) *Fopd {
	fopds := newFopdConns(cfg, transport)
	return &Fopd{fopds: fopds, status: NewStatus()}
}

func newFopdConns(cfg *FopdConfig, transport http.RoundTripper) map[string]*statedConns {
	if cfg.RetryTimes == 0 {
		cfg.RetryTimes = DefaultRetryTimes
	}
	if cfg.FailRetryInterval == 0 {
		cfg.FailRetryInterval = DefaultFailRetryInterval
	}
	if cfg.DefaultTimeout == 0 {
		cfg.DefaultTimeout = DefaultTimeout
	}
	fopds := make(map[string]*statedConns)
	for _, s := range cfg.Servers {
		for _, c := range s.Cmds {
			fopd := fopds[c]
			if fopd == nil {
				to := cfg.DefaultTimeout
				if to1, ok := cfg.Timeouts[c]; ok {
					to = to1
				}
				needCache := true
				if needCache1, ok := cfg.WhetherCache[c]; ok {
					needCache = needCache1
				}
				lbMode := 0
				if lbMode1, ok := cfg.LoadBalanceMode[c]; ok {
					lbMode = lbMode1
				}
				fopd = &statedConns{
					conns:             make([]*FopdConn, 0),
					retryTimes:        cfg.RetryTimes,
					failRetryInterval: int64(cfg.FailRetryInterval),
					needCache:         needCache,
					timeouts:          time.Duration(to) * time.Millisecond,
					lastIndex:         0,
					lbMode:            lbMode,
				}
				fopds[c] = fopd
			}

			fopd.conns = append(fopd.conns, NewFopdConn(s.Host, transport))
		}
	}

	for _, st := range fopds {
		shuffleFopds(st.conns)
	}

	return fopds
}

func (p *Fopd) Reload(cfg *FopdConfig, transport http.RoundTripper) {
	p.mu.Lock()
	p.fopds = newFopdConns(cfg, transport)
	p.mu.Unlock()
	p.status.ClearAll()
}

func (p *Fopd) safeFopds() map[string]*statedConns {
	p.mu.RLock()
	fopds := p.fopds
	p.mu.RUnlock()
	return fopds
}

func (p *Fopd) List() []string {

	fopds := p.safeFopds()
	fops := make([]string, len(fopds))
	i := 0

	for fop, _ := range fopds {
		fops[i] = fop
		i++
	}
	return fops
}

func (p *Fopd) HasCmd(cmd string) (ok bool) {
	_, ok = p.safeFopds()[cmd]
	return
}

func (p *Fopd) NeedCache(cmd string) (need bool, err error) {
	if fopd, ok := p.safeFopds()[cmd]; ok {
		return fopd.needCache, nil
	}
	return false, errors.New("no sunch cmd" + cmd)
}

type opRet struct {
	xl    *xlog.Logger
	resp  *http.Response
	err   error
	retry bool
}

func (p *Fopd) Op(xl *xlog.Logger, f io.Reader, fsize int64, fopCtx *fop.FopCtx) (resp *http.Response, err error) {

	query := strings.Split(fopCtx.RawQuery, "/")
	cmd := query[0]
	st, ok := p.safeFopds()[cmd]
	if !ok || len(st.conns) == 0 {
		xl.Warn("Fopd.Op: No fopd server is available:", query)
		err = ErrServiceNotAvailable
		return
	}

	for i := 0; i < st.retryTimes; i++ {
		conn, _, err2 := st.pickConn()
		if err2 != nil {
			err = errors.Info(err2, "pickConn failed").Detail(err)
			break
		}
		xl.Debugf("begin Fopd.Op - trytime:%d, cmd: %s, host:%s\n", i, cmd, conn.host)

		statusKey := MakeStatusKey(cmd, conn.host)

		var retry bool
		c := make(chan *opRet, 1)
		go func() {
			// spawn a new logger to avoid synchronous access
			xl2 := xl.Spawn()
			respx, e, try := conn.Op(xl2, f, fsize, fopCtx)
			c <- &opRet{xl2, respx, e, try}
		}()
		select {
		case ret := <-c:
			xl.Xput(ret.xl.Xget())
			resp, err, retry = ret.resp, ret.err, ret.retry
		case <-time.After(st.timeouts):
			go func() {
				// in case conn.Op returned a valid response after timeout
				if ret := <-c; ret.err == nil {
					xl.Infof("Fopd.Op: fopd.OpWithRT finished after timeout host:%v,fsize:%v,fopCtx:%v\n",
						conn.host, fsize, fopCtx)
					ret.resp.Body.Close()
				}
			}()
			err = ErrOpTimeout
			p.status.IncTimeout(statusKey)
			xl.Xlog("FOPG.TO:", st.timeouts)
			xl.Warnf("Fopd.Op: fopd.OpWithRT timeout(ms):%v,host:%v,fsize:%v,fopCtx:%v\n",
				st.timeouts, conn.host, fsize, fopCtx)
		}

		if err != nil {
			xl.Warnf("Fopd.Op failed - tryTime:%d, cmd:%s, host:%s, err:%v", i, cmd, conn.host, err)
			p.status.IncFailed(statusKey)
			if retry {
				p.status.IncRetry(statusKey)
				if seeker, ok := f.(io.Seeker); ok {
					if _, err2 := seeker.Seek(0, 0); err2 == nil {
						continue
					}
				}
				err = ErrCantRetry // may happen in pipe cmd
				break
			}
		}
		return
	}
	return
}

func (p *Fopd) Status() []byte {
	fopds := p.safeFopds()
	for cmd, sc := range fopds {
		for _, conn := range sc.conns {
			statusKey := MakeStatusKey(cmd, conn.host)
			p.status.SetProcessing(statusKey, conn.ProcessingNum())
		}
	}
	return p.status.Dump()
}

func (p *Fopd) ClearStatus(keys []string) {
	p.status.Clear(keys)
}

//--------------------------------------------------------------------------------
// Status

type Status struct {
	items map[string]*Item // key = cmd|host
	mu    sync.Mutex
}

type Item struct {
	Failed     int64 `json:"failed"` // network error && code != 200 && timeout
	Retry      int64 `json:"retry"`  // network error
	Timeout    int64 `json:"timeout"`
	Processing int64 `json:"processing"`
}

func NewStatus() *Status {
	return &Status{items: make(map[string]*Item)}
}

func MakeStatusKey(cmd, host string) string {
	return cmd + "|" + host
}

func (s *Status) getItem(key string) *Item {
	s.mu.Lock()
	defer s.mu.Unlock()

	item := s.items[key]
	if item == nil {
		item = new(Item)
		s.items[key] = item
	}
	return item
}

func (s *Status) IncFailed(key string) {
	item := s.getItem(key)
	atomic.AddInt64(&item.Failed, 1)
}

func (s *Status) IncTimeout(key string) {
	item := s.getItem(key)
	atomic.AddInt64(&item.Timeout, 1)
}

func (s *Status) IncRetry(key string) {
	item := s.getItem(key)
	atomic.AddInt64(&item.Retry, 1)
}

func (s *Status) SetProcessing(key string, n int64) {
	item := s.getItem(key)
	atomic.StoreInt64(&item.Processing, n)
}

func (s *Status) Dump() []byte {
	s.mu.Lock()
	defer s.mu.Unlock()
	b, _ := json.MarshalIndent(s.items, "", "\t")
	return b
}

func (s *Status) Clear(keys []string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, key := range keys {
		if _, ok := s.items[key]; ok {
			s.items[key] = new(Item)
		}
	}
}

func (s *Status) ClearAll() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items = make(map[string]*Item)
}
