package main

import (
	"runtime"

	"qbox.us/cc/config"
	"qbox.us/qdiscover/discoverd"

	"github.com/qiniu/errors"
	"github.com/qiniu/http/servestk.v1"
	"github.com/qiniu/log.v1"

	jsonlog "qbox.us/http/audit/jsonlog.v3"
)

type Config struct {
	BindHost   string         `json:"bind_host"`
	MaxProcs   int            `json:"max_procs"`
	DebugLevel int            `json:"debug_level"`
	AuditLog   jsonlog.Config `json:"auditlog"`
	discoverd.Config
}

func main() {
	config.Init("f", "qbox", "qboxdiscover.conf")

	var conf Config
	if err := config.Load(&conf); err != nil {
		log.Fatal("config.Load failed:", err)
	}

	runtime.GOMAXPROCS(conf.MaxProcs)
	log.SetOutputLevel(conf.DebugLevel)

	al, logf, err := jsonlog.Open("DISCOVER", &conf.AuditLog, nil)
	if err != nil {
		log.Fatal("jsonlog.Open failed:", errors.Detail(err))
	}
	defer logf.Close()

	log.Info("qboxdiscover is running at", conf.BindHost)

	mux := servestk.New(nil, al.Handler)
	log.Fatal(discoverd.Run(&conf.Config, conf.BindHost, mux))
}
