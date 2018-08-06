package main

import (
	"github.com/qiniu/errors"
	"github.com/qiniu/http/bsonrpc.v1"
	"github.com/qiniu/http/jsonrpc.v1"
	"github.com/qiniu/http/servestk.v1"
	"github.com/qiniu/http/webroute.v1"
	"github.com/qiniu/http/wsrpc.v1"
	"github.com/qiniu/log.v1"
	"net/http"
	"qbox.us/cc/config"
	jsonlog "qbox.us/http/audit/jsonlog.v2"
	"qbox.us/mgo2"
	qconfm "qbox.us/qconf/master"
	"qbox.us/qconf/master/mcrefresher"
	. "qbox.us/qconf/master/qrefresher/proto"
	. "qbox.us/servend/account"
	"runtime"
)

type Config struct {
	Mgo     mgo2.Config `json:"mgo"`
	McHosts []string    `json:"mc_hosts"` // 互为镜像的Memcache服务

	BindHost string `json:"bind_host"`

	AuditLog jsonlog.Config `json:"auditlog"`

	MaxProcs   int `json:"max_procs"`
	DebugLevel int `json:"debug_level"`
}

// ------------------------------------------------------------------------------------------

type adminAuthParser struct{}

func (p adminAuthParser) ParseAuth(req *http.Request) (user UserInfo, err error) {
	user.Utype = USER_TYPE_ADMIN
	user.Uid = 1
	return
}

var authp adminAuthParser

// ------------------------------------------------------------------------------------------

func main() {

	// Load Config

	config.Init("f", "qbox", "qboxconfone.conf")

	var conf Config
	if err := config.Load(&conf); err != nil {
		log.Fatal("config.Load failed:", err)
	}

	// General Settings

	runtime.GOMAXPROCS(conf.MaxProcs)
	log.SetOutputLevel(conf.DebugLevel)

	c := mgo2.Open(&conf.Mgo)
	defer c.Close()

	// new Service

	cfg := &qconfm.Config{
		Coll:       c.Coll,
		UidMgr:     1,
		AuthParser: authp,
	}
	var refresher Refresher = NilRefresher
	if len(conf.McHosts) > 0 {
		var err error
		refresher, err = mcrefresher.New(&mcrefresher.Config{conf.McHosts})
		if err != nil {
			log.Fatal("mcrefresher.New failed:", errors.Detail(err))
		}
	}
	service, err := qconfm.NewEx(cfg, refresher)
	if err != nil {
		log.Fatal("qconfm.New failed:", errors.Detail(err))
	}

	// run Service

	al, logf, err := jsonlog.Open("CFGONE", &conf.AuditLog, authp)
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
