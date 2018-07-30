package main

import (
	"net/http"
	"runtime"

	"qbox.us/cc/config"

	"github.com/qiniu/errors"
	"github.com/qiniu/http/servestk.v1"
	"github.com/qiniu/http/webroute.v1"
	"github.com/qiniu/log.v1"

	jsonlog "qbox.us/http/audit/jsonlog.v3"
	bloom "qbox.us/qbloom/bloom/svr"
)

type Config struct {
	Bloom bloom.Config `json:"bloom"`

	BindHost string `json:"bind_host"`

	AuditLog jsonlog.Config `json:"auditlog"`

	MaxProcs   int `json:"max_procs"`
	DebugLevel int `json:"debug_level"`
}

// ---------------------------------------------------------------

func main() {

	// Load Config

	config.Init("f", "qbox", "qboxbloom.conf")

	var conf Config
	if err := config.Load(&conf); err != nil {
		log.Fatal("config.Load failed:", err)
	}

	// General Settings

	runtime.GOMAXPROCS(conf.MaxProcs)
	log.SetOutputLevel(conf.DebugLevel)

	// new Service

	service, err := bloom.Open(&conf.Bloom)
	if err != nil {
		log.Fatal("bloom.Open failed:", err)
	}
	defer service.Close()

	// run Service

	al, logf, err := jsonlog.Open("BF", &conf.AuditLog, nil)
	if err != nil {
		log.Fatal("jsonlog.Open failed:", errors.Detail(err))
	}
	defer logf.Close()

	mux := servestk.New(nil, al.Handler)
	router := &webroute.Router{Mux: mux}
	router.Register(service)

	err = http.ListenAndServe(conf.BindHost, mux.Mux)
	log.Fatal("http.ListenAndServe:", err)
}

// ---------------------------------------------------------------
