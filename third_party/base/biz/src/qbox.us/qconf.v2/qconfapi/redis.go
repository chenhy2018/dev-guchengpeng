package qconfapi

import (
	"errors"
	"time"

	"github.com/qiniu/http/httputil.v1"
	"github.com/qiniu/rpc.v1"
	"github.com/qiniu/xlog.v1"
	"gopkg.in/mgo.v2/bson"
	"gopkg.in/redis.v5"
	"qbox.us/redisutilv5"
)

const (
	DefaultBufSize = 1024 * 16
)

var ErrCacheMiss = redis.Nil

type CacheItem struct {
	Val  []byte `bson:"val" json:"val"`
	T    int64  `bson:"t" json:"t"`
	Code int    `bson:"code" json:"code"`
}

type UpdateItem struct {
	Id         string
	CacheFlags int
}
type Redis interface {
	Get(xl *xlog.Logger, key string, item interface{}) (err error)
	Set(xl *xlog.Logger, key string, value interface{}) (err error)
	Del(xl *xlog.Logger, key string) (err error)
}

type RedisClient struct {
	*redisutilv5.RdsClient
}

var errMc404 = errors.New("404")
var errMc500 = errors.New("500")
var errMcBad = errors.New("bad")

func NewRedis(cfg redisutilv5.RedisCfg) Redis {
	cli, err := redisutilv5.NewRdsClient(cfg)
	if err != nil {
		panic("NewRedis" + err.Error())
	}
	return &RedisClient{
		RdsClient: cli,
	}
}

func (redis *RedisClient) Get(xl *xlog.Logger, key string, ret interface{}) (err error) {
	var xerr error
	defer xl.Xtrack("redis.g", time.Now(), &xerr)
	value, err := redis.Client.Get(key).Bytes()
	if err != nil {
		if err == ErrCacheMiss {
			xerr = errMc404
			return err
		}
		xerr = errMc500
		xl.Warn("qconf Redis.Get: call failed =>", err)
		return err
	}

	err = bson.Unmarshal(value, ret)
	if err != nil {
		xerr = errMcBad
		xl.Error("qconf redis.Get: bad value =>", err)
		return
	}
	return nil
}

func (redis *RedisClient) Set(xl *xlog.Logger, key string, value interface{}) (err error) {
	var xerr error
	defer xl.Xtrack("redis.s", time.Now(), &xerr)

	val, err := bson.Marshal(value)
	if err != nil {
		xerr = errMcBad
		xl.Warn("redis.Set: bad value =>", err)
		return err
	}

	err = redis.Client.Set(key, val, 0).Err()
	if err != nil {
		xerr = errMc500
		xl.Warn("redis.Set: call failed =>", err)
		return err
	}
	return nil
}

func (redis *RedisClient) Del(xl *xlog.Logger, key string) (err error) {
	var xerr error
	defer xl.Xtrack("reids.d", time.Now(), &xerr)

	err = redis.Client.Del(key).Err()
	if err != nil {
		xerr = errMc500
		xl.Warn("redis.Del: call failed =>", err)
		return err
	}
	return nil
}

type RdsClient struct {
	Client Redis

	updateChan chan UpdateItem
	updateFn   func(xl *xlog.Logger, id string) (doc interface{}, err error)
	expires    int64
}

type RedisCacheConf struct {
	RedisOptions redisutilv5.RedisCfg `json:"redis_options"`
	ExpireS      int                  `json:"expire_s"`
	ChanBufSize  int                  `json:"chan_buf_size"`
}

func NewRdsClient(cfg *RedisCacheConf, updateFn func(xl *xlog.Logger, id string) (doc interface{}, err error)) (client *RdsClient) {

	cli := NewRedis(cfg.RedisOptions)
	if cfg.ChanBufSize == 0 {
		cfg.ChanBufSize = DefaultBufSize
	}

	redis := &RdsClient{
		Client:     cli,
		expires:    int64(cfg.ExpireS) * 1e6,
		updateChan: make(chan UpdateItem, cfg.ChanBufSize),
		updateFn:   updateFn,
	}
	go redis.routine()
	return redis
}

func (p *RdsClient) GetFromRedis(log *xlog.Logger, ret interface{}, id string, cacheFlags int, now int64) (errRes, err error) {
	var val []byte
	var code int
	code, val, errRes = p.getWithCacheFlags(log, id, cacheFlags, time.Now().UnixNano())
	if errRes == nil {
		if code != 200 {
			err = &rpc.ErrorInfo{
				Code: code,
				Err:  string(val),
			}
			return
		}

		errRes = bson.Unmarshal(val, ret)
		if errRes != nil {
			return
		}
	}
	return
}

func (p *RdsClient) get(log *xlog.Logger, id string) (item CacheItem, err error) {
	err = p.Client.Get(log, id, &item)
	if err != nil {
		return
	}
	return item, nil
}

func (p *RdsClient) getWithCacheFlags(log *xlog.Logger, id string, cacheFlags int, now int64) (code int, value []byte, err error) {
	var item CacheItem
	item, err = p.get(log, id)
	if err != nil {
		return
	}
	code = item.Code
	value = item.Val

	log.Debug("getWithCacheFlags id", item.T)

	//过期异步更新
	d := now - item.T
	if d > p.expires {
		log.Debug("cacheItem expires need update")
		p.updateItem(log, id, cacheFlags)
	}
	return
}

func (p *RdsClient) set(log *xlog.Logger, item CacheItem, id string) error {
	return p.Client.Set(log, id, item)
}

func (p *RdsClient) del(log *xlog.Logger, id string) error {
	return p.Client.Del(log, id)
}

func (p *RdsClient) updateRedis(log *xlog.Logger, cacheFlags int, id string) {
	info, err := p.updateFn(log, id)

	p.UpdateRedis(log, id, info, cacheFlags, err)
}

func (p *RdsClient) UpdateRedis(log *xlog.Logger, id string, info interface{}, cacheFlags int, err error) {
	var b []byte
	var code int
	var err2 error
	if err != nil {
		code = httputil.DetectCode(err)
		b = []byte(err.Error())
		switch cacheFlags {
		case Cache_NoSuchEntry:
			if code != 612 && code != 404 {
				return
			}
		case Cache_Normal:
			if code == 612 || code == 404 {
				log.Debug("need delete redis cache")
				err = p.del(log, id)
				if err != nil {
					p.updateItem(log, id, cacheFlags)
				}
				return

			} else {
				return
			}
		default:
			return
		}

	} else {
		code = 200
		b, err2 = bson.Marshal(info)
		if err2 != nil {
			return
		}
	}

	item := CacheItem{
		Val:  b,
		Code: code,
		T:    time.Now().UnixNano(),
	}
	log.Debug("setRedis", id, item)
	errRet := p.set(log, item, id)
	if errRet != nil {
		log.Error("qconfg set redis err: ", errRet)
	}
	return

}

func (p *RdsClient) updateItem(log *xlog.Logger, id string, cacheFlags int) {

	select {
	case p.updateChan <- UpdateItem{id, cacheFlags}:
	default:
		log.Warn("confg.redis: updateChan full, skipped -", id, cacheFlags)
	}
}

func (p *RdsClient) routine() {
	for req := range p.updateChan {
		xl := xlog.NewDummy()
		now := time.Now().UnixNano()
		item, err := p.get(xl, req.Id)
		xl.Debug("routine item", item)
		if err == nil && now-item.T < p.expires {
			continue
		}
		p.updateRedis(xl, req.CacheFlags, req.Id)
	}
}
