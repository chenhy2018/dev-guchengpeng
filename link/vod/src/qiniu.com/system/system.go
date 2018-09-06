package system

import (
	"qbox.us/cc/config"
	"qbox.us/qconf/qconfapi"
	log "qiniupkg.com/x/log.v7"
)

type DBConfiguration struct {
	Host     string `json:"host"`
	Db       string `json:"db"`
	Mode     string `json:"mode"`
	TimeOut  int    `json:"timeout"`
	User     string `json:"user"`
	Password string `json:"password"`
}
type RedisConf struct {
	Addr     string `json:"addr"`
	Password string `json:"password"`
	DB       int    `json:"db"`
}
type Configuration struct {
	Bind      string          `json:"bind"`
	DbConf    DBConfiguration `json:"db_conf"`
	Qconf     qconfapi.Config `json:"qconfg"`
	RedisConf RedisConf       `json:"redis_conf"`
}

func LoadConf(file string) (conf *Configuration, err error) {
	config.Init("f", "qbox", "linking_vod.conf")
	err = config.Load(&conf)
	if err != nil {
		log.Error("Load conf fail", err)
		return

	}
	return conf, nil
}
