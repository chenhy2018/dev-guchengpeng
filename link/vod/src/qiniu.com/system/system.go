package system

import (
	"qbox.us/cc/config"
	"qbox.us/qconf/qconfapi"
	log "qiniupkg.com/x/log.v7"
)

var (
	haveDbConf = true
	haveQconf  = true
)

type DBConfiguration struct {
	Host     string `json:"host"`
	Db       string `json:"db"`
	Mode     string `json:"mode"`
	TimeOut  int    `json:"timeout"`
	User     string `json:"user"`
	Password string `json:"password"`
}
type GrpcConf struct {
	Addr string `json:"addr"`
}
type Configuration struct {
	Bind     string          `json:"bind"`
	DbConf   DBConfiguration `json:"db_conf"`
	Qconf    qconfapi.Config `json:"qconfg"`
	GrpcConf GrpcConf        `json:"grpc_conf"`
}

func LoadConf(dir, file string) (conf *Configuration, err error) {
	config.Init("f", dir, file)
	err = config.Load(&conf)
	if err != nil {
		log.Error("Load conf fail", err)
		return
	}
	if conf.DbConf.Host == "" {
		haveDbConf = false
	}
	if len(conf.Qconf.McHosts) == 0 {
		haveQconf = false
	}
	return conf, nil
}

func HaveDb() bool {
	// cause golang runtime will inlined only one
	// leaf functions, but in test case we need mock those function
	// so we just disable this inline
	defer func() {}()
	return haveDbConf
}

func HaveQconf() bool {
	defer func() {}()
	return haveQconf
}
