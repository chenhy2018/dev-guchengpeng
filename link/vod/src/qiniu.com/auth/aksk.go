package auth

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	xlog "github.com/qiniu/xlog.v1"
	"gopkg.in/redis.v5"
	"qbox.us/qconf/qconfapi"
	proto "qiniu.com/auth/proto.v1"
	"qiniu.com/system"
)

const (
	AK_PREFIX = "ak:"
)

var QConfClient *qconfapi.Client
var RedisClint *redis.Client

func Init(conf *system.Configuration) {
	QConfClient = qconfapi.New(&conf.Qconf)
	xl := xlog.NewDummy()
	if QConfClient == nil {
		xl.Error("init qconf client failed")
		os.Exit(3)
	}
	ret, err := getSKByAK("754zGDRRNFrtTQxjk3HXSpcttYqU-Unu5zKPp8fh")
	fmt.Println(ret, err)
	RedisClint = redis.NewClient(&redis.Options{
		Addr: conf.RedisConf.Addr,
		DB:   conf.RedisConf.DB})
	if RedisClint == nil {
		xl.Error("init reis falied")
		os.Exit(3)
	}
	pong, err := RedisClint.Ping().Result()
	fmt.Println(pong, err)
}
func getSKByAK(accessKey string) (string, error) {
	resp := proto.AccessInfo{}
	if QConfClient == nil {
		return "", errors.New("qconf client has not been initialized")

	}
	err := QConfClient.Get(nil, &resp, AK_PREFIX+accessKey, qconfapi.Cache_NoSuchEntry)
	if err != nil {
		xl := xlog.NewDummy()
		xl.Errorf("get account info failed, ak = %v", accessKey)
		return "", errors.New("get account info failed")

	}
	return string(resp.Secret[:]), nil
}
func getAKSKByUid(xl *xlog.Logger, uid uint32, ak string) (newAk, sk string, err error) {
	// 1. get from redis
	// 2. if not get from qconf and update to redis
	ret, err := RedisClint.Get(strconv.FormatUint(uint64(uid), 10)).Result()
	if err == redis.Nil {
		xl.Info("key doesn't exist, query for qconf")
	} else if err != nil {
		xl.Errorf("get aksk from redis failed err = %v", err)
	}
	aksk := strings.Split(ret, ":")
	if len(aksk) == 2 && aksk[0] == ak {
		return ak, aksk[1], nil
	}
	sk, err = getSKByAK(ak)
	if err != nil {
		xl.Errorf("get sk err = %v", err)
		return "", "", errors.New("get sk failed")
	}
	RedisClint.Set(strconv.FormatUint(uint64(uid), 10), ak+":"+sk, 0)
	return ak, sk, nil
}
