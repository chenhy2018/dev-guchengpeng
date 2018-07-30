// 整个客户端只能处理服务端返回为bson的情况(changed at 2016/11/28)
package qconfapi

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/golang/groupcache/singleflight"
	"github.com/qiniu/api/auth/digest"
	"github.com/qiniu/rpc.v1"
	brpc "github.com/qiniu/rpc.v1/brpc/lb.v2.1"
	"github.com/qiniu/rpc.v1/lb.v2.1"
	"github.com/qiniu/xlog.v1"
	"gopkg.in/redis.v5"
	"labix.org/v2/mgo/bson"
	"qbox.us/errors"
	"qbox.us/redisutilv5"
)

// ------------------------------------------------------------------------

const (
	Cache_Normal      = 0
	Cache_NoSuchEntry = 1
)

var ErrRedisNoServers = errors.New("redis: no servers configured or available")

type LBConfig struct {
	brpc.Config
	Transport lb.TransportConfig `json:"transport"`
}

type MasterConfig struct {
	Default  LBConfig `json:"default"`
	Failover LBConfig `json:"failover"`
}

type RedisCfg struct {
	Hosts   []string `json:"hosts"`
	DB      int      `json:"db"`
	MaxIdle int      `json:"max_idle"`
}

type Config struct {
	RedisCfg RedisCfg `json:"redis_cfg"` //redis配置
	// 兼容老的配置
	MasterHosts []string `json:"master_hosts"`

	// 推荐使用新的配置
	Master MasterConfig `json:"master"`

	AccessKey string `json:"access_key"` // 如果不需要授权，可以没有这个值
	SecretKey string `json:"secret_key"`

	RedisCacheExpires int `json:"redis_cache_expire_ms"`
	RedisChanBufSize  int `json:"redis_chan_buf_size"` // 异步消息队列缓冲区大小。
	LcacheExpires     int `json:"lc_expires_ms"`       // 如果缓存超过这个时间，则强制刷新(防止取到太旧的值)，以毫秒为单位。
	LcacheDuration    int `json:"lc_duration_ms"`      // 如果缓存没有超过这个时间，不要去刷新(防止刷新过于频繁)，以毫秒为单位。
	LcacheChanBufSize int `json:"lc_chan_bufsize"`     // 异步消息队列缓冲区大小。
}

type Client struct {
	Lcache *localCache

	rediss            []Redis
	redisIdxBegin     uint32
	redisCacheExpires int64

	Conn *brpc.Client

	updateChan chan UpdateItem
	updateFn   func(log *xlog.Logger, id string, cacheFlags int) (v []byte, err error)

	Group singleflight.Group
}

func New(cfg *Config) *Client {

	rediss := make([]Redis, len(cfg.RedisCfg.Hosts))
	for i, host := range cfg.RedisCfg.Hosts {
		confg := redisutilv5.RedisCfg{
			Host:    host,
			DB:      cfg.RedisCfg.DB,
			MaxIdle: cfg.RedisCfg.MaxIdle,
		}
		client := NewRedis(confg)
		rediss[i] = client
	}

	p := &Client{
		rediss:            rediss,
		redisCacheExpires: int64(cfg.RedisCacheExpires) * 1e6,
		updateChan:        make(chan UpdateItem, cfg.RedisChanBufSize),
	}

	master := &cfg.Master

	if len(master.Default.Hosts) == 0 {
		master.Default.Hosts = cfg.MasterHosts
	}
	setMasterDefaultConfig(&master.Default)
	setMasterDefaultConfig(&master.Failover)

	defaultTr := lb.NewTransport(&master.Default.Transport)
	failoverTr := lb.NewTransport(&master.Failover.Transport)

	if cfg.AccessKey != "" {
		mac := &digest.Mac{
			AccessKey: cfg.AccessKey,
			SecretKey: []byte(cfg.SecretKey),
		}
		defaultTr = digest.NewTransport(mac, defaultTr)
		failoverTr = digest.NewTransport(mac, failoverTr)
	}

	if len(master.Failover.Hosts) > 0 {
		p.Conn = brpc.NewWithFailover(&master.Default.Config, &master.Failover.Config, defaultTr, failoverTr, nil)
	} else {
		p.Conn = brpc.New(&master.Default.Config, defaultTr)
	}

	if cfg.LcacheExpires > 0 {
		p.Lcache = newLocalCache(
			int64(cfg.LcacheExpires)*1e6, int64(cfg.LcacheDuration)*1e6,
			cfg.LcacheChanBufSize, p.getBytesNolc)
	}
	go p.routine()
	return p
}

func setMasterDefaultConfig(cfg *LBConfig) {
	if cfg.TryTimes == 0 {
		cfg.TryTimes = uint32(len(cfg.Hosts))
	}
	if cfg.FailRetryIntervalS == 0 {
		cfg.FailRetryIntervalS = -1
	}

	if cfg.Transport.DialTimeoutMS == 0 {
		cfg.Transport.DialTimeoutMS = 1000
	}
	if cfg.Transport.TryTimes == 0 {
		cfg.Transport.TryTimes = uint32(len(cfg.Transport.Proxys))
	}
	if cfg.Transport.FailRetryIntervalS == 0 {
		cfg.Transport.FailRetryIntervalS = -1
	}
}

// ------------------------------------------------------------------------

type insArgs struct {
	Doc interface{} `bson:"doc"`
}

func (p *Client) Insert(l rpc.Logger, doc interface{}) (err error) {
	return p.Conn.CallWithBson(l, nil, "/insb", insArgs{doc})
}

// ------------------------------------------------------------------------

type putArgs struct {
	Id     string      `bson:"id"`
	Change interface{} `bson:"chg"`
}

type M map[string]interface{}

func (p *Client) Modify(l rpc.Logger, id string, change interface{}) (err error) {
	return p.Conn.CallWithBson(l, nil, "/putb", putArgs{id, change})
}

func (p *Client) SetProp(l rpc.Logger, id, prop string, val interface{}) (err error) {

	return p.Modify(l, id, M{"$set": M{prop: val}})
}

func (p *Client) DeleteProp(l rpc.Logger, id, prop string) (err error) {

	return p.Modify(l, id, M{"$unset": M{prop: 1}})
}

// ------------------------------------------------------------------------

func (p *Client) Delete(l rpc.Logger, id string) (err error) {
	return p.Conn.CallWithForm(l, nil, "/rm", map[string][]string{
		"id": {id},
	})
}

// ------------------------------------------------------------------------

func (p *Client) Refresh(l rpc.Logger, id string) (err error) {
	return p.Conn.CallWithForm(l, nil, "/refresh", map[string][]string{
		"id": {id},
	})
}

// ------------------------------------------------------------------------

func (p *Client) GetFromLc(l rpc.Logger, ret interface{}, id string, cacheFlags int) (exist bool, err error) {

	log := xlog.NewWith(l)
	exist, err = p.getFromLc(log, ret, id, cacheFlags)
	return
}

func (p *Client) GetFromMaster(l rpc.Logger, ret interface{}, id string, cacheFlags int) (err error) {

	log := xlog.NewWith(l)
	if p.Lcache != nil {
		p.Lcache.deleteItemSafe(id)
	}
	return p.Get(log, ret, id, cacheFlags)
}

func (p *Client) Get(l rpc.Logger, ret interface{}, id string, cacheFlags int) (err error) {

	log := xlog.NewWith(l)
	err = p.get(log, ret, id, cacheFlags)
	return
}

func (p *Client) getFromLc(log *xlog.Logger, ret interface{}, id string, cacheFlags int) (exist bool, err error) {

	var exp bool
	if p.Lcache != nil {
		exist, exp, err = p.Lcache.getFromLc(log, ret, id, cacheFlags, time.Now().UnixNano())
		if err != nil {
			return
		}
		if exp {
			p.Lcache.updateItem(log, id, cacheFlags)
		}
	}
	return
}

func (p *Client) get(log *xlog.Logger, ret interface{}, id string, cacheFlags int) (err error) {
	if p.Lcache != nil {
		err = p.Lcache.get(log, ret, id, cacheFlags, time.Now().UnixNano())
		return
	}

	val, _, err := p.getBytesNolc(log, id, cacheFlags)
	if err != nil {
		return
	}
	err = bson.Unmarshal(val, ret)
	if err != nil {
		log.Error("qconf.Get: bson.Unmarshal failed -", err)
	}
	return
}

func (p *Client) getBytesNolc(log *xlog.Logger, id string, cacheFlags int) (val []byte, exp bool, err error) {
	var code int

	code, val, exp, err = p.getFromRedis(log, id, cacheFlags, time.Now().UnixNano())
	if err == nil {
		goto done
	}

	code, val, err = p.getFromMaster(log, "/v2/getb", id, cacheFlags)
	if err == nil {
		goto done
	}
	return

done:
	if code != 200 {
		err = &rpc.ErrorInfo{
			Code: code,
			Err:  string(val),
		}
	}
	return
}

func (p *Client) getFromRedis(log *xlog.Logger, id string, cacheFlags int, now int64) (code int, b []byte, exp bool, err error) {
	N := uint32(len(p.rediss))
	idxBegin := atomic.LoadUint32(&p.redisIdxBegin)
	for idx := uint32(0); idx < N; idx++ {
		var item CacheItem
		i := (idxBegin + idx) % N
		rs := p.rediss[i]
		err = rs.Get(log, id, &item)
		if err == nil {
			code = item.Code
			b = item.Val

			//过期异步更新,不写LC
			d := now - item.T
			if d > p.redisCacheExpires {
				log.Debug("cacheItem expires need update")
				exp = true
				p.updateItem(log, id, cacheFlags)
			}
			return code, b, exp, nil
		}
		if err == redis.Nil {
			atomic.AddUint32(&p.redisIdxBegin, 1)
			return 0, nil, exp, err
		}
		log.Warn("qconf.getFromRedis failed:", i, id, err)
	}

	return 0, nil, exp, ErrRedisNoServers
}

func (p *Client) updateItem(log *xlog.Logger, id string, cacheFlags int) {

	select {
	case p.updateChan <- UpdateItem{id, cacheFlags}:
	default:
		log.Warn("confg.redis: updateChan full, skipped -", id, cacheFlags)
	}
}

func (p *Client) routine() {
	for req := range p.updateChan {
		N := uint32(len(p.rediss))
		if N == uint32(0) {
			continue
		}
		xl := xlog.NewDummy()
		now := time.Now().UnixNano()
		var item CacheItem
		idxBegin := atomic.LoadUint32(&p.redisIdxBegin)
		err := p.rediss[(idxBegin)%N].Get(xl, req.Id, &item)

		if err == nil && now-item.T < p.redisCacheExpires {
			continue
		}
		p.getFromMaster(xl, "/v2/getb", req.Id, req.CacheFlags)
	}

}

func (p *Client) getFromMaster(log *xlog.Logger, url string, id string, cacheFlags int) (code int, b []byte, err error) {
	log.Debug("getFromMaster", id, cacheFlags)

	p.Group.Do(id, func() (interface{}, error) {
		var resp *http.Response
		resp, err = p.Conn.PostWithForm(log, url, map[string][]string{
			"id":         {id},
			"cache_flag": {fmt.Sprint(cacheFlags)},
		})
		if err != nil {
			log.Warn("qconf.getFromMaster: post form failed -", url, id, err)
			return nil, nil
		}
		defer resp.Body.Close()

		code = resp.StatusCode
		if code != 200 {
			b = []byte(rpc.ResponseError(resp).Error())
			switch cacheFlags {
			case Cache_NoSuchEntry:
				if code != 612 && code != 404 {
					return nil, nil
				}
			default:
				return nil, nil
			}
		} else {
			b, err = ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Warn("qconf.getFromMaster: read resp.Body failed -", url, id, err)
				return nil, nil
			}
		}

		return nil, nil

	})

	return
}

// ------------------------------------------------------------------------
