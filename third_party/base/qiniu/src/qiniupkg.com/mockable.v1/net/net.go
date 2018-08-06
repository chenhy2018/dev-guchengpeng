package net

import (
	"net"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"qiniupkg.com/x/log.v7"
)

// ---------------------------------------------------------------------------

var Mocking bool        // 网络是否处于mockable环境下
var MockingIPs []string // 在mockable环境下，本Node的IP列表
var MockingIPInfos []*IPInfo

type IPInfo struct {
	rwlock  sync.RWMutex
	secInfo [2]SecInfo
}

type SecInfo struct {
	time int64
	bps  [2]int // 0: in 1: out
}

func (info *IPInfo) AddInBps(b int) {
	info.add(b, 0)
}

func (info *IPInfo) AddOutBps(b int) {
	info.add(b, 1)
}

func (info *IPInfo) GetBps() (inBps, outBps int) {
	now := time.Now().Unix()
	info.rwlock.RLock()
	defer info.rwlock.RUnlock()

	if info.secInfo[0].time == now {
		return info.secInfo[0].bps[0], info.secInfo[0].bps[1]
	}
	return 0, 0
}

func (info *IPInfo) add(b, idx int) {

	now := time.Now().Unix()
	info.rwlock.Lock()
	defer info.rwlock.Unlock()

	if now == info.secInfo[1].time {
		info.secInfo[1].bps[idx] += b
	} else {
		info.secInfo[0] = info.secInfo[1]

		var bps [2]int
		bps[idx] = b
		info.secInfo[1] = SecInfo{time: now, bps: bps}
	}
}

const (
	BpsNotLimit = -1
)

type SpeedItem struct {
	Duration int `json:"duration"` // Duration in millisecond, 0 表示永久
	Bps      int `json:"Bps"`      // Bytes per second
}

type Speed []SpeedItem // 在len(Speed)为0的情况下表示两个节点不连通

var (
	SpeedNotLimit = Speed{{Bps: BpsNotLimit}}
)

// localIPs => "0.0.0.1" "0.0.0.2"
//  {
//    "0.0.0.3": [ # 访问0.0.0.3的时候
//      [ # 首个IP(0.0.0.1)
//        {
//          "Bps": 1024 # 速度为1K/s
//        },
//        {
//          "duration": 1000, # 1s后速度变成2K/s
//          "Bps": 2048
//        }
//      ],
//      [ # 第二个IP(0.0.0.2)
//        ...
//      ]
//    ],
//    "0.0.0.4": [ # 访问0.0.0.4的时候
//      ...
//    ]
//  }
var speeds map[string][]Speed

// ---------------------------------------------------------------------------

type Conn net.Conn
type Listener net.Listener
type TCPAddr net.TCPAddr
type TCPConn net.TCPConn

var Dial = net.Dial
var Listen = net.Listen

// ---------------------------------------------------------------------------

// 逻辑地址 logicIP:logicPort => 物理地址 phyIP:phyPort 算法：
//	1) logicIP 必须是 "0.0.0.xxx"，且 xxx 在 1..255 之间
//	2) logicPort 必须在 0..255 之间
//	3) phyIP = "127.0.0.1"
//	4) phyPort = xxx * 256 + logicPort
//	5) logicIP = "0.0.0.0" 时，需要将 logicIP 改为 MockingIPs 后进行地址转换（需要Listen多个端口）
//
func LogicToPhy(address string) string {

	pos := strings.Index(address, ":")
	if pos < 0 {
		log.Fatalln("invalid logic address: no port -", address)
	}
	ip := address[:pos]
	port := address[pos+1:]
	if ip == "127.0.0.1" {
		ip = MockingIPs[0]
	}
	if !strings.HasPrefix(ip, "0.0.0.") {
		log.Fatalln("invalid logic address: not 0.0.0.xxx -", address)
	}
	ipPart, err := strconv.Atoi(ip[6:])
	if err != nil || ipPart < 1 || ipPart > 255 {
		log.Fatalln("invalid logic address: not 0.0.0.(1-255) -", address)
	}
	localPort, err := strconv.Atoi(port)
	if err != nil || localPort < 0 || localPort > 255 {
		log.Fatalln("invalid logic address: invalid port (must 0-255) -", address)
	}
	return "127.0.0.1:" + strconv.Itoa(ipPart*256+localPort)
}

func LogicToPhyUrl(rawUrl string) string {
	toUrl, err := url.Parse(rawUrl)
	if err != nil {
		log.Fatalln("invalid rawUrl", rawUrl)
	}

	toUrl.Host = LogicToPhy(toUrl.Host)
	return toUrl.String()
}

// ---------------------------------------------------------------------------

var (
	defaultDialer Dialer
)

// 负责: 1) remote address的地址转换; 2) 限速
//
func MockDial(network, address string) (conn net.Conn, err error) {

	return defaultDialer.Dial(network, address)
}

// 负责：laddr(logicIP:logicPort)地址转换
// 当 logicIP = "0.0.0.0" 时，需要将 logicIP 改为 MockingIPs 后进行地址转换（需要Listen多个端口）
//
func MockListen(nett, laddr string) (listener net.Listener, err error) {

	pos := strings.Index(laddr, ":")
	if pos < 0 {
		log.Fatalln("invalid logic address: no port -", laddr)
	}

	ip := laddr[:pos]
	if ip == "" || ip == "0.0.0.0" {
		port := laddr[pos+1:]
		return broadcastListen(nett, port)
	}
	log.Info("net.Listen:", laddr)
	return net.Listen(nett, LogicToPhy(laddr))
}

// ---------------------------------------------------------------------------

type broadcastListener struct {
	accepts   chan net.Conn
	listeners map[string]net.Listener
}

func (p *broadcastListener) Accept() (conn net.Conn, err error) {

	c1 := <-p.accepts
	return c1, nil
}

func (p *broadcastListener) Close() (err error) {

	for _, l := range p.listeners {
		l.Close()
	}
	return nil
}

func (p *broadcastListener) Addr() (address net.Addr) {

	log.Fatal("not impl")
	return
}

func broadcastListen(nett, port string) (listener net.Listener, err error) {

	accepts := make(chan net.Conn)
	listeners := make(map[string]net.Listener)

	for _, ip := range MockingIPs {
		laddr := ip + ":" + port
		log.Info("net.Listen:", laddr)
		l1, err1 := net.Listen(nett, LogicToPhy(laddr))
		if err1 != nil {
			log.Fatalln("net.Listen failed:", err1)
		}
		listeners[laddr] = l1
		go func() {
			for {
				c1, err1 := l1.Accept()
				if err1 != nil {
					log.Errorf("Accept from `%s` failed: %v\n", laddr, err1)
				}
				accepts <- c1
			}
		}()
	}
	return &broadcastListener{accepts: accepts, listeners: listeners}, nil
}

// ---------------------------------------------------------------------------

// 初始化整个Mocking Network系统
//
type InitConfig struct {
	MockingIPs []string
	Speeds     map[string][]Speed
}

var (
	initChain []func()
)

func RegisterInit(initFn func()) {

	initChain = append(initChain, initFn)
}

func Init(cfg *InitConfig) {

	Mocking = true
	MockingIPs = cfg.MockingIPs
	for _ = range MockingIPs {
		MockingIPInfos = append(MockingIPInfos, &IPInfo{})
	}
	speeds = cfg.Speeds
	Dial = MockDial
	Listen = MockListen

	for _, initFn := range initChain {
		initFn()
	}
}

// ---------------------------------------------------------------------------
