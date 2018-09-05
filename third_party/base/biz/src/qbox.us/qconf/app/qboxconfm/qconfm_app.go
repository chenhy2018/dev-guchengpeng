package main

import (
	"net/http"
	"runtime"

	"github.com/qiniu/errors"
	"github.com/qiniu/http/bsonrpc.v1"
	"github.com/qiniu/http/jsonrpc.v1"
	"github.com/qiniu/http/servestk.v1"
	"github.com/qiniu/http/webroute.v1"
	"github.com/qiniu/http/wsrpc.v1"
	"github.com/qiniu/log.v1"
	"qbox.us/cc/config"
	"qbox.us/http/account.v2.1/digest_auth"
	jsonlog "qbox.us/http/audit/jsonlog.v3"
	"qbox.us/mgo2"
	qconfm "qbox.us/qconf/master"

	_ "github.com/qiniu/version"
	_ "qbox.us/autoua"
	_ "qbox.us/profile"
)

type Config struct {
	Mgo          mgo2.Config `json:"mgo"`
	MgrAccessKey string      `json:"access_key"`
	MgrSecretKey string      `json:"secret_key"`  // 向 slave 发送指令时的帐号
	SlaveHosts   [][]string  `json:"slave_hosts"` // [[idc1_slave1, idc1_slave2], [idc2_slave1, idc2_slave2], ...]

	BindHost string `json:"bind_host"`

	AuditLog jsonlog.Config `json:"auditlog"`

	AuthConf digest_auth.Config `json:"auth"`

	MaxProcs   int `json:"max_procs"`
	DebugLevel int `json:"debug_level"`

	UidMgr uint32 `json:"uid_mgr"`
}

// ------------------------------------------------------------------------------------------

func main() {

	// Load Config

	config.Init("f", "qbox", "qboxconfm.conf")

	var conf Config
	if err := config.Load(&conf); err != nil {
		log.Fatal("config.Load failed:", err)
	}
	conf.AuditLog.AuthProxy = 1

	// General Settings

	runtime.GOMAXPROCS(conf.MaxProcs)
	log.SetOutputLevel(conf.DebugLevel)

	c := mgo2.Open(&conf.Mgo)
	defer c.Close()

	// new Service

	cfg := &qconfm.Config{
		Coll:         c.Coll,
		MgrAccessKey: conf.MgrAccessKey,
		MgrSecretKey: conf.MgrSecretKey,
		SlaveHosts:   conf.SlaveHosts,
		UidMgr:       conf.UidMgr,
	}
	service, err := qconfm.New(cfg)
	if err != nil {
		log.Fatal("qconfm.New failed:", errors.Detail(err))
	}

	// run Service

	authp := digest_auth.NewParser(&conf.AuthConf)
	al, logf, err := jsonlog.Open("CFGM", &conf.AuditLog, authp)
	if err != nil {
		log.Fatal("jsonlog.Open failed:", errors.Detail(err))
	}
	defer logf.Close()

	mux := servestk.New(nil, al.Handler)
	factory := wsrpc.Factory.Union(bsonrpc.Factory).Union(jsonrpc.Factory)
	router := &webroute.Router{Factory: factory, Mux: mux}
	router.Register(service)

	err = http.ListenAndServe(conf.BindHost, mux.Mux)
	log.Fatal("http.ListenAndServe:", err)
}

// ------------------------------------------------------------------------------------------
