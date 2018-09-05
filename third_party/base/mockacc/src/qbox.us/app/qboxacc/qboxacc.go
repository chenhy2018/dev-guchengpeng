package main

import (
	"runtime"
)

import (
	"qbox.us/account"
	"qbox.us/cc/config"
	"github.com/qiniu/log.v1"
	"qbox.us/mgo2"
)

type AccConf struct {
	MaxProcs   int      `json:"max_procs"`
	BindHost   string   `json:"bind_host"`
	MgoHost    string   `json:"mgo_host"`
	FsHost     string   `json:"fs_host"`
	WbHost     string   `json:"www_host"`
	MaHost     string   `json:"mail_host"`
	MgoDBName  string   `json:"mgo_db_name"`
	LogHosts   []string `json:"log_hosts"`
	DebugLevel int      `json:"debug_level"`
	UseSSL     bool     `json:"use_ssl"`
	MgoMode    string   `json:"mgo_mode"`
}

func main() {

	config.Init("f", "qbox", "qboxacc.conf")

	var conf AccConf
	err := config.Load(&conf)
	if err != nil {
		return
	}

	log.SetOutputLevel(conf.DebugLevel)
	runtime.GOMAXPROCS(conf.MaxProcs)
	session := mgo2.Dail(conf.MgoHost, conf.MgoMode, 1)
	defer session.Close()

	db := session.DB(conf.MgoDBName)
	accCfg := account.Config{
		db, conf.FsHost, conf.WbHost, conf.MaHost,
		account.AuditLogConf{
			Hosts: conf.LogHosts,
		},
		nil,
	}
	log.Println("qboxacc running at ", conf.BindHost)
	err = account.Run(conf.BindHost, &accCfg, conf.UseSSL)
	log.Fatal("account.Run:", err)
}
