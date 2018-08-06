package main

import (
	"net/http"
	"runtime"

	"qbox.us/cc/config"
	"qbox.us/example/v7/foosvr"
	"qbox.us/mgo2"

	"github.com/qiniu/errors"
	"github.com/qiniu/http/servestk.v1"
	"github.com/qiniu/http/webroute.v1"
	"github.com/qiniu/http/wsrpc.v1"
	"github.com/qiniu/log.v1"

	jsonlog "qbox.us/http/audit/jsonlog.v7"
)

type Config struct {
	Mgo mgo2.Config `json:"mgo"`

	BindHost string `json:"bind_host"`

	AuditLog jsonlog.Config `json:"auditlog"`

	MaxProcs   int `json:"max_procs"`
	DebugLevel int `json:"debug_level"`
}

// ------------------------------------------------------------------------------------------

func main() {

	// Load Config

	config.Init("f", "qbox", "qfoosvr.conf")

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

	cfg := &foosvr.Config{
		Coll: c.Coll,
	}
	service, err := foosvr.New(cfg)
	if err != nil {
		log.Fatal("foosvr.New failed:", errors.Detail(err))
	}

	// run Service

	al, logf, err := jsonlog.Open("FOO", &conf.AuditLog)
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
