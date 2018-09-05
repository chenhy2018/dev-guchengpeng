package redisutilv5

import (
	"reflect"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/qiniu/log.v1"
	"github.com/qiniu/xlog.v1"
	"gopkg.in/redis.v5"
	"qbox.us/lbsocketproxy"
)

type RedisCfg struct {
	MasterName    string   `json:"master_name"`
	SentinelAddrs []string `json:"sentinel_addrs"`

	Host     string `json:"host"`
	DB       int    `json:"db"`
	MaxIdle  int    `json:"max_idle"`
	Password string `json:"password"`

	TrackThresholdMs int64                    `json:"trackthreshold_ms"`
	Proxies          *lbsocketproxy.Config    `json:"proxies"`
	Histogram        *prometheus.HistogramVec `json:"-"`
}

type RdsClient struct {
	RedisCfg
	*redis.Client
	TrackThresholdMs int64
}

//redis nil 需要确认
func NewRdsClient(cfg RedisCfg) (client *RdsClient, err error) {

	client = &RdsClient{
		RedisCfg:         cfg,
		TrackThresholdMs: cfg.TrackThresholdMs,
	}
	if cfg.MaxIdle <= 0 {
		cfg.MaxIdle = 30
	}

	var proxy *lbsocketproxy.LbSocketProxy
	if cfg.Proxies != nil {
		proxy, err = lbsocketproxy.NewLbSocketProxy(cfg.Proxies)
		if err != nil {
			log.Panic("lbsocketproxy.NewLbSocketProxy failed", err, cfg.Proxies)
		}
	}

	if len(cfg.SentinelAddrs) != 0 {
		client.Client = redis.NewFailoverClientWithProxy(&redis.FailoverOptions{
			MasterName:    cfg.MasterName,
			SentinelAddrs: cfg.SentinelAddrs,
			Password:      cfg.Password,
			DB:            cfg.DB,
			PoolSize:      cfg.MaxIdle,
		}, proxy)
	} else {
		client.Client = redis.NewClientWithProxy(&redis.Options{
			Addr:     cfg.Host,
			Password: cfg.Password,
			DB:       cfg.DB,
			PoolSize: cfg.MaxIdle,
		}, proxy)
	}

	err = client.Ping().Err()
	return
}

//这个函数中判断redis.Nil可能是有问题的
func (client *RdsClient) track(xl *xlog.Logger, method string, from time.Time, err *error) {

	errInter := *err
	if reflect.TypeOf(errInter) != reflect.TypeOf(redis.Nil) {
		errInter = nil //代表是业务层的错误,不需要在这个函数中做任何的处理
	}

	dur := time.Since(from)
	if client.Histogram != nil {
		costs := float64(dur) / float64(time.Millisecond)
		var errStr string
		if errInter != nil {
			errStr = (errInter).Error()
		}
		client.Histogram.With(map[string]string{
			"method": method,
			"error":  errStr,
		}).Observe(costs)
	}

	if int64(dur/time.Millisecond) > client.TrackThresholdMs || (errInter != nil && errInter != redis.Nil) {
		xl.Xprof2(method, dur, errInter)
	}
}

func (client *RdsClient) GetWithTrack(xl *xlog.Logger, name, key string) (cmd *redis.StringCmd) {
	var reterr error
	defer client.track(xl, name+".get", time.Now(), &reterr)
	cmd = client.Get(key)
	reterr = cmd.Err()
	return
}

func (client *RdsClient) SetWithTrack(xl *xlog.Logger, name, key string, value interface{}, expiration time.Duration) (cmd *redis.StatusCmd) {
	var reterr error
	defer client.track(xl, name+".set", time.Now(), &reterr)
	cmd = client.Set(key, value, expiration)
	reterr = cmd.Err()
	return
}

func (client *RdsClient) DelWithTrack(xl *xlog.Logger, name string, keys ...string) (cmd *redis.IntCmd) {
	var reterr error
	defer client.track(xl, name+".del", time.Now(), &reterr)
	cmd = client.Del(keys...)
	reterr = cmd.Err()
	return
}

func (client *RdsClient) HDelWithTrack(xl *xlog.Logger, name, key string, fields ...string) (cmd *redis.IntCmd) {
	var reterr error
	defer client.track(xl, name+".hdel", time.Now(), &reterr)
	cmd = client.HDel(key, fields...)
	reterr = cmd.Err()
	return
}

func (client *RdsClient) HGetWithTrack(xl *xlog.Logger, name, key, field string) (cmd *redis.StringCmd) {
	var reterr error
	defer client.track(xl, name+".hget", time.Now(), &reterr)
	cmd = client.HGet(key, field)
	reterr = cmd.Err()
	return
}

/*
func (client *RdsClient) HGetAllWithTrack(xl *xlog.Logger, name, key string) (cmd *redis.StringSliceCmd) {
	var reterr error
	defer client.track(xl, name+".hget", time.Now(), &reterr)
	cmd = client.HGetAll(key)
	return
}
*/
func (client *RdsClient) HGetAllMapWithTrack(xl *xlog.Logger, name, key string) (cmd *redis.StringStringMapCmd) {
	var reterr error
	defer client.track(xl, name+".hgetall", time.Now(), &reterr)
	cmd = client.HGetAll(key)
	reterr = cmd.Err()
	return
}

func (client *RdsClient) HSetWithTrack(xl *xlog.Logger, name, key, field, value string) (cmd *redis.BoolCmd) {
	var reterr error
	defer client.track(xl, name+".hset", time.Now(), &reterr)
	cmd = client.HSet(key, field, value)
	reterr = cmd.Err()
	return
}

func (client *RdsClient) HSetNXWithTrack(xl *xlog.Logger, name, key, field, value string) (cmd *redis.BoolCmd) {
	var reterr error
	defer client.track(xl, name+".hsetNX", time.Now(), &reterr)
	cmd = client.HSetNX(key, field, value)
	reterr = cmd.Err()
	return
}

func (client *RdsClient) HScanWithTrack(xl *xlog.Logger, name, key string, cursor uint64, match string, count int64) (cmd *redis.ScanCmd) {
	var reterr error
	defer client.track(xl, name+".hscan", time.Now(), &reterr)
	cmd = client.HScan(key, cursor, match, count)
	reterr = cmd.Err()
	return
}

func (client *RdsClient) HMSetWithTrack(xl *xlog.Logger, name, key string, fields map[string]string) (cmd *redis.StatusCmd) {
	var reterr error
	defer client.track(xl, name+".hmset", time.Now(), &reterr)
	cmd = client.HMSet(key, fields)
	reterr = cmd.Err()
	return
}

func (client *RdsClient) WatchWithTrack(xl *xlog.Logger, name string, fn func(*redis.Tx) error, keys ...string) error {
	var reterr error
	defer client.track(xl, name+".watch", time.Now(), &reterr)
	reterr = client.Watch(fn, keys...) //reterr 可能有三种含义:1.redis的watch命令的错误 2.fn函数返回的错误 3.redis的unwatch命令的错误
	return reterr
}

type RdsPipeline struct {
	*redis.Pipeline
	client *RdsClient
}

func (client *RdsClient) PipelineWithTrack() (pipe *RdsPipeline) {
	pipe = &RdsPipeline{
		Pipeline: client.Pipeline(),
		client:   client,
	}
	return
}

func (pipe *RdsPipeline) ExecWithTrack(xl *xlog.Logger, name string) (cmds []redis.Cmder, retErr error) {
	defer pipe.client.track(xl, name+".exec", time.Now(), &retErr)
	cmds, retErr = pipe.Exec()

	//同步v3与v5的行为
	if retErr != nil && retErr.Error() == "redis: pipeline is empty" {
		retErr = nil
	}
	return
}
