package fopd

import (
	"errors"
	"math"
	"math/rand"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/qiniu/log.v1"
)

const (
	LBMode0 = iota
	LBMode1
	LBMode2
	LBMode3
)

var (
	FailRecoverIntervalSecs = 0
)

var (
	ErrServiceNotAvailable = errors.New("fop: service not available")
)

type ConnSelector interface {
	PickConn() (*Conn, error)
}

// --------------------------------------------------------
// lbSelector0 用于耗时比较短的 fop 负载均衡，比如图片处理。
type lbSelector0 struct {
	Conns     []*ConnInfo
	lastIndex uint64
}

func (s *lbSelector0) PickConn() (*Conn, error) {
	n := len(s.Conns)
	for i := 0; i < n; i++ {
		index := int(atomic.AddUint64(&s.lastIndex, 1)) % n
		info := s.Conns[index]
		conn := info.Conn
		lastFailed := conn.LastFailedTime()
		if lastFailed == 0 || time.Now().Unix()-lastFailed >= int64(FailRecoverIntervalSecs) {
			log.Debugf("lbSelector0.PickConn - Index:%d, Host:%s", index, conn.Host)
			return conn, nil
		}
	}
	return nil, ErrServiceNotAvailable
}

// --------------------------------------------------------
// lbSelector1 用于耗时比较长的 fop 负载均衡，比如视频转码。
type lbSelector1 struct {
	Conns []*ConnInfo
}

func (s *lbSelector1) PickConn() (*Conn, error) {
	ns := make([]int64, len(s.Conns))
	for i, info := range s.Conns {
		conn := info.Conn
		lastFailed := conn.LastFailedTime()
		if lastFailed == 0 || time.Now().Unix()-lastFailed >= int64(FailRecoverIntervalSecs) {
			ns[i] = conn.ProcessingNum()
		} else {
			ns[i] = math.MaxInt64 // 不会被选择到
		}
	}
	index := minLoadIndex(ns)
	if ns[index] == math.MaxInt64 {
		return nil, ErrServiceNotAvailable
	}
	info := s.Conns[index]
	conn := info.Conn
	log.Debugf("lbSelector1.PickConn - Index:%d, Host:%s", index, conn.Host)
	return conn, nil
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

// --------------------------------------------------------
// lbSelector2 加权最少连接, 用于耗时比较长的 fop 负载均衡，比如视频转码。
type lbSelector2 struct {
	Conns []*ConnInfo
}

func (s *lbSelector2) PickConn() (*Conn, error) {

	ns := make([]float64, len(s.Conns))
	for i, info := range s.Conns {
		conn := info.Conn
		if info.FopWeight == 0 {
			info.FopWeight = 1
		}
		lastFailed := conn.LastFailedTime()
		if info.FopWeight < 0 || lastFailed == 0 || time.Now().Unix()-lastFailed >= int64(FailRecoverIntervalSecs) {
			ns[i] = float64(conn.ProcessingNum()) / float64(info.FopWeight)
		} else {
			ns[i] = math.MaxFloat64
		}
	}
	index := minFloatLoadIndex(ns)
	if ns[index] == math.MaxFloat64 {
		return nil, ErrServiceNotAvailable
	}
	info := s.Conns[index]
	conn := info.Conn
	log.Debugf("lbSelector2.PickConn - Index:%d, Host:%s", index, conn.Host)
	return conn, nil
}

// 从加权后的值中随机选一个，返回它在 ns 中的索引。
func minFloatLoadIndex(ns []float64) (minIndex int) {
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
		if math.Abs(n-minVal) < 0.001 {
			minIndexs = append(minIndexs, i)
		}
	}

	// 从最小值的 indexs 中随机选择一个
	randVal := rand.Intn(len(minIndexs))
	return minIndexs[randVal]
}

// --------------------------------------------------------
// lbSelector3 加权轮询, 用于耗时比较短的 fop 负载均衡，比如图片处理。
type lbSelector3 struct {
	Conns      []*ConnInfo
	lastIndex  int
	lastWeight int64
	lock       sync.Mutex
}

func (s *lbSelector3) PickConn() (*Conn, error) {

	ns := make([]int64, len(s.Conns))
	for i, info := range s.Conns {
		conn := info.Conn
		if info.FopWeight <= 0 {
			info.FopWeight = 1
		}
		lastFailed := conn.LastFailedTime()
		if lastFailed == 0 || time.Now().Unix()-lastFailed >= int64(FailRecoverIntervalSecs) {
			ns[i] = info.FopWeight
		} else {
			ns[i] = 0 // 不会被选择到
		}
	}
	index := s.pick(ns)
	if index < 0 {
		return nil, ErrServiceNotAvailable
	}
	info := s.Conns[index]
	conn := info.Conn
	log.Debugf("lbSelector3.PickConn - Index:%d, Host:%s", index, conn.Host)
	return conn, nil
}

func (s *lbSelector3) pick(ns []int64) (index int) {
	s.lock.Lock()
	index, weight := wwrs(ns, s.lastIndex, s.lastWeight)
	s.lastIndex = index
	s.lastWeight = weight
	s.lock.Unlock()
	return
}

func wwrs(ns []int64, index int, weight int64) (lastIndex int, lastWeigth int64) {
	length := len(ns)
	gcd := getGcd(ns)
	max := getMax(ns)
	for {
		index = (index + 1) % length
		if index == 0 {
			weight = weight - gcd
			if weight <= 0 {
				weight = max
				if weight == 0 {
					return -1, weight
				}
			}
		}
		if ns[index] >= weight {
			return index, weight
		}
	}
}

// --------------------------------------------------------
func shuffleConns(conns []*ConnInfo) {
	ip2conns := make(map[string][]*ConnInfo)
	ipList := make([]string, 0)
	for _, info := range conns {
		conn := info.Conn
		ip := getIp(conn.Host)
		if _, ok := ip2conns[ip]; !ok {
			ipList = append(ipList, ip)
		}
		ip2conns[ip] = append(ip2conns[ip], info)
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

func getIp(host string) string {
	ip := host
	if idx := strings.Index(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}
	return ip
}

func getGcd(ns []int64) (gcd int64) {
	length := len(ns)
	n := make([]int64, length)
	copy(n, ns[:])
	gcd = n[0]
	for i := 1; i < length; i++ {
		t := int64(0)
		for n[i] != 0 {
			t = n[i]
			n[i] = gcd % n[i]
			gcd = t
		}
	}
	return gcd
}

func getMax(n []int64) (max int64) {
	max = n[0]
	for _, v := range n {
		if v > max {
			max = v
		}
	}
	return
}
