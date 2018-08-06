package main

import (
	"net/http"
	"runtime"

	"qbox.us/cc/config"
	"qbox.us/qmq/mmq"

	"github.com/qiniu/errors"
	"github.com/qiniu/http/servestk.v1"
	"github.com/qiniu/http/webroute.v1"
	"github.com/qiniu/http/wsrpc.v1"
	"github.com/qiniu/log.v1"
	"github.com/qiniu/osl/signal"

	jsonlog "qbox.us/http/audit/jsonlog.v2"

	_ "github.com/qiniu/version"
	_ "qbox.us/autoua"
	_ "qbox.us/profile"
)

type Config struct {
	Mqs []mmq.MqConf `json:"mqs"`

	BindHost string `json:"bind_host"`

	AuditLog jsonlog.Config `json:"auditlog"`

	MaxProcs   int `json:"max_procs"`
	DebugLevel int `json:"debug_level"`
}

// ---------------------------------------------------------------

func main() {

	// Load Config

	config.Init("f", "qbox", "qboxmmq.conf")

	var conf Config
	if err := config.Load(&conf); err != nil {
		log.Fatal("config.Load failed:", err)
	}

	// General Settings

	runtime.GOMAXPROCS(conf.MaxProcs)
	log.SetOutputLevel(conf.DebugLevel)

	// new Service

	cfg := &mmq.Config{
		Mqs: conf.Mqs,
	}
	service := mmq.New(cfg)
	go signal.WaitForInterrupt(service.Quit)

	// run Service

	al, logf, err := jsonlog.Open("MMQ", &conf.AuditLog, nil)
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

// ---------------------------------------------------------------
