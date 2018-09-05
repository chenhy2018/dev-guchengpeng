package main

import (
	"net/http"
	"runtime"

	"qbox.us/cc/config"
	"qbox.us/example/v2/foosvr"
	"qbox.us/mgo2"

	"github.com/qiniu/errors"
	"github.com/qiniu/http/servestk.v1"
	"github.com/qiniu/http/webroute.v1"
	"github.com/qiniu/http/wsrpc.v1"
	"github.com/qiniu/log.v1"

	auth "qbox.us/http/account.v2.1/digest_auth"
	jsonlog "qbox.us/http/audit/jsonlog.v3"
)

type Config struct {
	Mgo mgo2.Config `json:"mgo"`

	BindHost string `json:"bind_host"`

	AuditLog jsonlog.Config `json:"auditlog"`
	AuthConf auth.Config    `json:"auth"`

	MaxProcs   int `json:"max_procs"`
	DebugLevel int `json:"debug_level"`
}

// ------------------------------------------------------------------------------------------

func main() {

	// Load Config

	config.Init("f", "qbox", "qboxfoosvr.conf")

	var conf Config
	if err := config.Load(&conf); err != nil {
		log.Fatal("config.Load failed:", err)
	}
	conf.AuditLog.AuthProxy = 1
	// AuthProxy 的含义：如果发现授权合法，则将 Authoraztion 转为 proxy auth；否则就清除 Authoraztion
	// 有了这句后，下面 foosvr.Config 的 AuthParser 就不需要初始化了（默认为 proxy_auth.Parser）

	// General Settings

	runtime.GOMAXPROCS(conf.MaxProcs)
	log.SetOutputLevel(conf.DebugLevel)

	c := mgo2.Open(&conf.Mgo)
	defer c.Close()

	// new Service

	cfg := &foosvr.Config{
		Coll: c.Coll,
	}
	service, err := foosvr.New(cfg)
	if err != nil {
		log.Fatal("foosvr.New failed:", errors.Detail(err))
	}

	// run Service

	authp := auth.NewParser(&conf.AuthConf)
	al, logf, err := jsonlog.Open("FOO", &conf.AuditLog, authp)
	if err != nil {
		log.Fatal("jsonlog.Open failed:", errors.Detail(err))
	}
	defer logf.Close()

	mux := servestk.New(nil, al.Handler)
	router := &webroute.Router{Factory: wsrpc.Factory, Mux: mux}
	router.Register(service)

	err = http.ListenAndServe(conf.BindHost, mux.Mux)
	log.Fatal("http.ListenAndServe:", err)
}

// ------------------------------------------------------------------------------------------
