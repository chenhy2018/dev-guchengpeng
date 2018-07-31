package main

import (
	"net/http"
	"runtime"

	"github.com/qiniu/http/restrpc.v1"
	"qiniupkg.com/x/config.v7"
	"qiniupkg.com/x/log.v7"

	"qiniu.com/kodo/foo.v1"
)

// ---------------------------------------------------------------------------

type Config struct {
	Foo foo.Config `json:"foo"`

	BindHost   string `json:"bind_host"`
	MaxProcs   int    `json:"max_procs"`
	DebugLevel int    `json:"debug_level"`
}

func main() {

	// Load Config

	config.Init("f", "qiniu", "qfoo.conf")

	var conf Config
	if err := config.Load(&conf); err != nil {
		log.Fatal("config.Load failed:", err)
	}

	// General Settings

	runtime.GOMAXPROCS(conf.MaxProcs)
	log.SetOutputLevel(conf.DebugLevel)

	// new Service

	service, err := foo.New(&conf.Foo)
	if err != nil {
		log.Fatal("foo.New failed:", err)
	}

	// run Service

	router := restrpc.Router{
		PatternPrefix: "v1",
	}
	log.Info("Starting qfoo ...")
	err = http.ListenAndServe(conf.BindHost, router.Register(service))
	log.Fatal("http.ListenAndServe(qfoo):", err)
}

// ---------------------------------------------------------------------------

