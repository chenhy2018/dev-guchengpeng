package main

import (
	"net/http"
	"runtime"

	"github.com/qiniu/errors"
	"github.com/qiniu/http/servestk.v1"
	"github.com/qiniu/http/webroute.v1"
	"github.com/qiniu/http/wsrpc.v1"
	"github.com/qiniu/log.v1"

	"qbox.us/cc/config"
	"qbox.us/mgo2"
	"qbox.us/qauthgate"

	auth "qbox.us/http/account.v2.1/digest_auth"
	jsonlog "qbox.us/http/audit/jsonlog.v3"
)

// ------------------------------------------------------------------------

type Config struct {
	Mgo mgo2.Config `json:"mgo"`

	BindHost    string `json:"bind_host"`
	BindHostMgr string `json:"bind_host_mgr"` // 局域网端口

	AuditLogMgr jsonlog.Config `json:"auditlog_mgr"`
	AuthConf    auth.Config    `json:"auth"`

	MaxProcs   int `json:"max_procs"`
	DebugLevel int `json:"debug_level"`
}

func main() {

	// Load Config

	config.Init("f", "qbox", "qboxauthgate.conf")

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

	authp := auth.NewParser(&conf.AuthConf)
	cfg := &qauthgate.Config{
		Coll:       c.Coll,
		AuthParser: authp,
	}
	service, err := qauthgate.New(cfg)
	if err != nil {
		log.Fatal("qauthgate.New failed:", errors.Detail(err))
	}

	// run MgrService

	go func() {

		al, logf, err := jsonlog.Open("AGM", &conf.AuditLogMgr, nil)
		if err != nil {
			log.Fatal("jsonlog.Open failed:", errors.Detail(err))
		}
		defer logf.Close()

		mux := servestk.New(nil, al.Handler)
		router := &webroute.Router{Factory: wsrpc.Factory, Mux: mux}
		router.Register(service)

		err = http.ListenAndServe(conf.BindHostMgr, mux.Mux)
		log.Fatal("http.ListenAndServe(authgate mgr):", err)
	}()

	// run Service

	err = http.ListenAndServe(conf.BindHost, service)
	log.Fatal("http.ListenAndServe(authgate):", err)
}

// ------------------------------------------------------------------------
