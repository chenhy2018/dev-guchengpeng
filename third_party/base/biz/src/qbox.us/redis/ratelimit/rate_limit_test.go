package ratelimit

import (
	"testing"
	"time"

	"github.com/qiniu/xlog.v1"
	"github.com/stretchr/testify.v2/require"
	"gopkg.in/redis.v3"
	"qbox.us/redisutilv5"
)

var skip bool

func init() {
	client := redis.NewClient(&redis.Options{
		Addr:         ":6379",
		DialTimeout:  10 * time.Second,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		PoolSize:     10,
		PoolTimeout:  30 * time.Second,
	})
	client.FlushDb()
	err := client.Ping().Err()
	if err != nil {
		skip = true
	}
}

func TestClient_Limit(t *testing.T) {
	if skip {
		return
	}
	xl := xlog.NewDummy()
	rate := NewClient(
		&Config{
			RedisOptions: redisutilv5.RedisCfg{
				MasterName: "master",
				Host:       ":6379",
			},
			ExpireS: 1,
		})
	keys := []string{"key1", "key2", "keys3"}
	for i := 0; i < rate.ExpireLimit; i++ {
		islimit, err := rate.Limit(xl, keys)
		require.NoError(t, err)
		require.False(t, islimit)
	}
	islimit, err := rate.Limit(xl, keys[0:1])
	require.True(t, islimit)
	require.NoError(t, err)
	islimit, err = rate.Limit(xl, []string{"key4", "key5"})
	require.NoError(t, err)
	require.False(t, islimit)
	time.Sleep(time.Second)
	islimit, err = rate.Limit(xl, keys[0:1])
	require.False(t, islimit)
	require.NoError(t, err)
	err = rate.Release(xl, keys[0:1])
	require.NoError(t, err)
	islimit, err = rate.Limit(xl, keys[0:1])
	require.False(t, islimit)
	require.NoError(t, err)
}
