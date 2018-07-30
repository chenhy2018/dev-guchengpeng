package fopd

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"sync/atomic"
	"time"

	"qbox.us/fop"

	"code.google.com/p/go.net/context"
	"github.com/qiniu/http/httputil.v1"
	"github.com/qiniu/log.v1"
	"github.com/qiniu/rpc.v3"
	"github.com/qiniu/xlog.v1"
)

const (
	MaxIdleConnsPerHost = 10
	DefaultTryTimes     = 2
	//	DefaultFopTimeout   = 7200 // 7200 secs = 2 hours
	DefaultFopTimeout = 0 // 中间服务不设置超时时间，让调用端来控制超时时间（当前是io, prefop）
)

var (
	DialTimeout = 2 * time.Second
)

var (
	ErrCantRetry  = errors.New("can not retry in pipe")
	ErrInvalidUrl = errors.New("invalid url")
)

var (
	defaultTransport = newTransport(DialTimeout, time.Duration(DefaultFopTimeout)*time.Second, MaxIdleConnsPerHost)
)

func newTransport(dial, resp time.Duration, poolSize int) *http.Transport {
	t := &http.Transport{MaxIdleConnsPerHost: poolSize}
	t.Dial = func(network, addr string) (net.Conn, error) {
		return net.DialTimeout(network, addr, dial)
	}
	t.ResponseHeaderTimeout = resp
	return t
}

func addProxiesToTransport(tr *http.Transport, urls []string) error {
	if len(urls) == 0 {
		return nil
	}

	urlList := []*url.URL{}
	for _, u := range urls {
		u, err := url.Parse(u)
		if err != nil {
			return ErrInvalidUrl
		}
		urlList = append(urlList, u)
	}
	lenProxy := uint64(len(urlList))
	var idx uint64 = 0
	tr.Proxy = func(req *http.Request) (*url.URL, error) {
		if lenProxy == 0 {

			return http.ProxyFromEnvironment(req)
		}
		index := atomic.AddUint64(&idx, 1) % lenProxy
		return urlList[index], nil
	}
	return nil
}

func ResetDialTimeoutMs(timeoutMs int64) {
	DialTimeout = time.Duration(timeoutMs) * time.Millisecond
}

type Config struct {
	Servers             []HostInfo     `json:"servers"`
	NotCache            []string       `json:"not_cache"`
	NotCdnCache         []string       `json:"not_cdn_cache"`
	Timeouts            map[string]int `json:"timeouts"`
	LoadBalanceMode     map[string]int `json:"load_balance_mode"`
	DefaultTimeout      int            `json:"default_timeout"`
	TryTimes            int            `json:"try_times"`
	FailRecoverInterval int            `json:"fail_recover_interval"`
	Proxies             []string       `json:"proxies"`
}

type HostInfo struct {
	Host      string   `json:"host"`
	Cmds      []string `json:"cmds"`
	FopWeight int64    `json:"fop_weight"`
	FopMode   uint32   `json:"fop_mode"`
}

type ConnInfo struct {
	Conn      *Conn
	FopWeight int64
	FopMode   uint32
}

// Client 是管理所有 fop 的客户端
type Client struct {
	cmdClients    map[string]*CmdClient
	cmdTransports map[string]*http.Transport
	processingNum int64
	stats         *Stats

	Config
	mapNotCache    map[string]bool
	mapNotCdnCache map[string]bool
	cmds           []string
}

func NewClient(cfg *Config) (*Client, error) {
	if cfg.DefaultTimeout == 0 {
		cfg.DefaultTimeout = DefaultFopTimeout
	}
	defaultTransport.ResponseHeaderTimeout = time.Duration(cfg.DefaultTimeout) * time.Second
	if len(cfg.Proxies) > 0 {
		err := addProxiesToTransport(defaultTransport, cfg.Proxies)
		if err != nil {
			return nil, err
		}
	}

	if cfg.TryTimes == 0 {
		cfg.TryTimes = DefaultTryTimes
	}
	if cfg.FailRecoverInterval != 0 {
		FailRecoverIntervalSecs = cfg.FailRecoverInterval
	}

	mapNotCache := make(map[string]bool)
	for _, cmd := range cfg.NotCache {
		mapNotCache[cmd] = true // 如果不配置，默认值 为 false，表示需要缓存
	}

	mapNotCdnCache := make(map[string]bool)
	for _, cmd := range cfg.NotCdnCache {
		mapNotCdnCache[cmd] = true // 如果不配置，默认值 为 false，表示需要缓存
	}

	conns := make(map[string][]*ConnInfo)
	cmdTransports := make(map[string]*http.Transport)
	for _, server := range cfg.Servers {
		for _, cmd := range server.Cmds {
			transport := cmdTransports[cmd]
			timeout, ok := cfg.Timeouts[cmd]
			if ok && transport == nil {
				transport = newTransport(DialTimeout, time.Duration(timeout)*time.Second, MaxIdleConnsPerHost)
				if len(cfg.Proxies) > 0 {
					err := addProxiesToTransport(transport, cfg.Proxies)
					if err != nil {
						return nil, err
					}
				}
			}
			if transport == nil {
				transport = defaultTransport
			}
			cmdTransports[cmd] = transport

			cmdConns := conns[cmd]
			if cmdConns == nil {
				cmdConns = make([]*ConnInfo, 0)
			}
			conn := NewConn(server.Host, transport)
			connInfo := &ConnInfo{
				Conn:      conn,
				FopWeight: server.FopWeight,
				FopMode:   server.FopMode,
			}
			conns[cmd] = append(cmdConns, connInfo)
		}
	}

	stats := NewStats()
	cmds := make([]string, 0)
	cmdClients := make(map[string]*CmdClient)
	for cmd, cmdConns := range conns {
		shuffleConns(cmdConns)
		cmds = append(cmds, cmd)
		lbMode := cfg.LoadBalanceMode[cmd]
		cmdClients[cmd] = NewCmdClient(cmd, lbMode, cmdConns, cfg.TryTimes, stats)
	}

	c := &Client{
		cmdClients:     cmdClients,
		cmdTransports:  cmdTransports,
		stats:          stats,
		Config:         *cfg,
		mapNotCache:    mapNotCache,
		mapNotCdnCache: mapNotCdnCache,
		cmds:           cmds,
	}
	return c, nil
}

func (c *Client) Close() error {

	xl := xlog.NewDummy()
	for {
		nprocess := atomic.LoadInt64(&c.processingNum)
		xl.Info("remain process: ", nprocess)

		for _, transport := range c.cmdTransports {
			if transport != defaultTransport {
				transport.CloseIdleConnections()
			}
		}

		// 等待正在处理的任务完成
		if nprocess == 0 { // done
			break
		}

		time.Sleep(time.Second * 2)
	}

	return nil
}

func (c *Client) List() []string {
	return c.cmds
}

func (c *Client) HasCmd(cmd string) bool {
	_, ok := c.cmdClients[cmd]
	return ok
}

func (c *Client) NeedCache(cmd string) bool {
	notCache := c.mapNotCache[cmd]
	return !notCache
}

func (c *Client) NeedCdnCache(cmd string) bool {
	notCache := c.mapNotCdnCache[cmd]
	return !notCache
}

func (c *Client) Op2(tpCtx context.Context, fh []byte, fsize int64, ctx *fop.FopCtx) (resp *http.Response, err error) {
	atomic.AddInt64(&c.processingNum, 1)
	defer atomic.AddInt64(&c.processingNum, -1)

	cmdClient, ok := c.cmdClients[ctx.CmdName]
	if !ok {
		return nil, httputil.NewError(400, "unsupported cmd "+ctx.CmdName)
	}
	return cmdClient.Op2(tpCtx, fh, fsize, ctx)
}

func (c *Client) Op(tpCtx context.Context, f io.Reader, fsize int64, ctx *fop.FopCtx) (resp *http.Response, err error) {
	atomic.AddInt64(&c.processingNum, 1)
	defer atomic.AddInt64(&c.processingNum, -1)

	cmdClient, ok := c.cmdClients[ctx.CmdName]
	if !ok {
		return nil, httputil.NewError(400, "unsupported cmd "+ctx.CmdName)
	}
	return cmdClient.Op(tpCtx, f, fsize, ctx)
}

func (c *Client) Stats() *Stats {
	return c.stats
}

// --------------------------------------------------------
// CmdClient 是一个特定 fop 的客户端
type CmdClient struct {
	modeClients map[uint32]*ModeClient
}

func NewCmdClient(cmd string, lbMode int, conns []*ConnInfo, tryTimes int, stats *Stats) *CmdClient {
	if stats == nil {
		stats = NewStats()
	}
	modesConns := make(map[uint32][]*ConnInfo)
	for _, cmdConn := range conns {
		fopMode := cmdConn.FopMode
		modeConns, ok := modesConns[fopMode]
		if !ok {
			modeConns = make([]*ConnInfo, 0)
		}
		modesConns[fopMode] = append(modeConns, cmdConn)
	}

	modeClients := make(map[uint32]*ModeClient)
	for fopMode, modeConns := range modesConns {
		modeClients[fopMode] = NewModeClient(cmd, lbMode, fopMode, modeConns, tryTimes, stats)
	}

	return &CmdClient{modeClients: modeClients}
}

func (c *CmdClient) pick(xl *xlog.Logger, fopMode uint32) (client *ModeClient, err error) {
	xl.Debugf("mode: %d", fopMode)
	client, ok := c.modeClients[fopMode]
	if !ok {
		if fopMode == 0 {
			return nil, fmt.Errorf("no client for mode %d", fopMode)
		}
		client, ok = c.modeClients[0]
		if !ok {
			return nil, fmt.Errorf("no client for mode %d or 0", fopMode)
		}
		xl.Debugf("no client for mode %d, use 0", fopMode)
	}
	return client, err
}

func (c *CmdClient) Op2(tpCtx context.Context, fh []byte, fsize int64, ctx *fop.FopCtx) (*http.Response, error) {
	xl := xlog.FromContextSafe(tpCtx)
	client, err := c.pick(xl, ctx.Mode)
	if err != nil {
		return nil, err
	}
	return client.Op2(tpCtx, fh, fsize, ctx)
}

func (c *CmdClient) Op(tpCtx context.Context, f io.Reader, fsize int64, ctx *fop.FopCtx) (*http.Response, error) {
	xl := xlog.FromContextSafe(tpCtx)
	client, err := c.pick(xl, ctx.Mode)
	if err != nil {
		return nil, err
	}
	return client.Op(tpCtx, f, fsize, ctx)
}

// --------------------------------------------------------
// ModeClient 是一个特定 fop 特定 fopMode 的客户端
type ModeClient struct {
	Cmd      string
	FopMode  uint32
	Stats    *Stats
	selector ConnSelector
	tryTimes int
}

func NewModeClient(cmd string, lbMode int, fopMode uint32, conns []*ConnInfo, tryTimes int, stats *Stats) *ModeClient {
	if stats == nil {
		stats = NewStats()
	}
	var selector ConnSelector
	switch lbMode {
	case LBMode0:
		selector = &lbSelector0{Conns: conns}
	case LBMode1:
		selector = &lbSelector1{Conns: conns}
	case LBMode2:
		selector = &lbSelector2{Conns: conns}
	case LBMode3:
		selector = &lbSelector3{Conns: conns}
	default:
		panic("unknown lbmode")
	}
	return &ModeClient{
		Cmd:      cmd,
		FopMode:  fopMode,
		Stats:    stats,
		selector: selector,
		tryTimes: tryTimes,
	}
}

func (c *ModeClient) Op2(tpCtx context.Context, fh []byte, fsize int64, ctx *fop.FopCtx) (resp *http.Response, err error) {
	xl := xlog.FromContextSafe(tpCtx)
	for i := 0; i < c.tryTimes; i++ {
		conn, err2 := c.selector.PickConn()
		if err2 != nil {
			xl.Warn("ModeClient.Op2: PickConn failed:", err2)
			return nil, err2
		}
		statsKey := MakeStatsKey(c.Cmd, c.FopMode, conn.Host)
		resp, err = conn.Op2(tpCtx, fh, fsize, ctx)
		if err != nil {
			xl.Warnf("ModeClient.Op2:%s host:%s, tryTime:%d error:%v", ctx.RawQuery, conn.Host, i, err)
			c.Stats.IncFailed(statsKey)
			if isResponseTimeout(err) {
				c.Stats.IncTimeout(statsKey)
			}
			if ShouldRetry(err) {
				c.Stats.IncRetry(statsKey)
				continue
			}
		}
		break
	}
	return
}

func (c *ModeClient) Op(tpCtx context.Context, f io.Reader, fsize int64, ctx *fop.FopCtx) (resp *http.Response, err error) {
	xl := xlog.FromContextSafe(tpCtx)
	for i := 0; i < c.tryTimes; i++ {
		conn, err2 := c.selector.PickConn()
		if err2 != nil {
			xl.Warn("CmdClient.Op: PickConn failed:", err2)
			return nil, err2
		}
		statsKey := MakeStatsKey(c.Cmd, c.FopMode, conn.Host)
		resp, err = conn.Op(tpCtx, f, fsize, ctx)
		if err != nil {
			xl.Warnf("CmdClient.Op:%s host:%s, tryTime:%d error:%v", ctx.RawQuery, conn.Host, i, err)
			c.Stats.IncFailed(statsKey)
			if isResponseTimeout(err) {
				c.Stats.IncTimeout(statsKey)
			}
			if ShouldRetry(err) {
				c.Stats.IncRetry(statsKey)
				if seeker, ok := f.(io.Seeker); ok {
					if _, err2 = seeker.Seek(0, 0); err2 == nil {
						continue
					}
				}
				err = ErrCantRetry // 使用了管道会出现不能重试
				log.Warnf("CmdClient.Op:%s error:%v", ctx.RawQuery, err)
				break
			}
		}
		break
	}
	return
}

const StatusShouldRetry = 570

var ErrShouldRetry = httputil.NewError(StatusShouldRetry, "should retry")

func ShouldRetry(err error) bool {
	if info, ok := err.(rpc.RespError); ok { // 有 HTTP Response
		if info.HttpCode() == StatusShouldRetry {
			return true
		}
		return false
	}
	if info, ok := err.(*httputil.ErrorInfo); ok {
		if info.Code == StatusShouldRetry {
			return true
		}
		return false
	}
	if !isServerFail(err) { // 非服务端错误不重试
		return false
	}
	return true
}
