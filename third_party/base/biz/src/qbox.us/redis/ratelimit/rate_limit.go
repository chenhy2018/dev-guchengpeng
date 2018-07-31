package ratelimit

import (
	"errors"
	"time"

	"sync"

	"github.com/qiniu/xlog.v1"
	"gopkg.in/redis.v5"
	"qbox.us/redisutilv5"
)

var (
	DefaultExpireS            = 24 * 3600
	DefaultExpireLimit        = 10
	DefaultReloadingIntervals = 300
)

type Config struct {
	RedisOptions       redisutilv5.RedisCfg `json:"redis_options"`
	ExpireS            int                  `json:"expire_s"`
	ExpireLimit        int                  `json:"expire_limit"`
	ReloadingIntervals int                  `json:"reloading_interval_s"`
}

type Client struct {
	*Config
	mutex sync.RWMutex
	redis *redisutilv5.RdsClient
}

func NewClient(cfg *Config) (client *Client) {

	client = new(Client)
	if cfg == nil {
		return
	}
	if cfg.ExpireS == 0 {
		cfg.ExpireS = DefaultExpireS
	}
	if cfg.ExpireLimit == 0 {
		cfg.ExpireLimit = DefaultExpireLimit
	}
	if cfg.ReloadingIntervals == 0 {
		cfg.ReloadingIntervals = DefaultReloadingIntervals
	}
	client.Config = cfg
	redis, err := redisutilv5.NewRdsClient(cfg.RedisOptions)
	if err == nil {
		client.redis = redis
	} else {
		go client.reloadingRedis(cfg)
	}
	return

}

func (cli *Client) reloadingRedis(cfg *Config) {
	for {
		redis, err := redisutilv5.NewRdsClient(cfg.RedisOptions)
		if err == nil {
			cli.mutex.Lock()
			cli.redis = redis
			cli.mutex.Unlock()
			break
		}
		time.Sleep(time.Second * time.Duration(cli.ReloadingIntervals))
	}
}

func (cli *Client) Limit(xl *xlog.Logger, keys []string) (islimit bool, err error) {

	cli.mutex.RLock()
	redisCli := cli.redis
	cli.mutex.RUnlock()
	if redisCli == nil {
		err = errors.New("redis cli fails")
		return
	}
	pipe := redisCli.PipelineWithTrack()
	defer pipe.Close()
	for _, key := range keys {
		pipe.Incr(key)
	}
	cmds, err := pipe.ExecWithTrack(xl, "incr eblocks")
	if err != nil {
		xl.Error("pipe.Exec.Limit failed: %v", err)
		return
	}

	pipe2 := redisCli.PipelineWithTrack()
	defer pipe2.Close()
	for i, cmd := range cmds {
		c := cmd.(*redis.IntCmd)
		val := c.Val()
		if val == 1 {
			pipe2.Expire(keys[i], time.Duration(cli.ExpireS)*time.Second)
		} else if val > int64(cli.ExpireLimit) {
			return true, nil
		}
	}
	_, err = pipe2.ExecWithTrack(xl, "expire eblocks")
	if err != nil {
		xl.Error("pipe2.Exec.Limit failed: %v", err)
	}

	return
}

func (cli *Client) Release(xl *xlog.Logger, keys []string) (err error) {
	cli.mutex.RLock()
	redisCli := cli.redis
	cli.mutex.RUnlock()
	if redisCli == nil {
		err = errors.New("redis cli fails")
		return
	}
	pipe := redisCli.PipelineWithTrack()
	defer pipe.Close()
	for _, key := range keys {
		pipe.Decr(key)
	}
	_, err = pipe.ExecWithTrack(xl, "decr eblocks")
	if err != nil {
		xl.Error("pipe.Exec.Release failed: %v", err)
		return
	}
	return
}
