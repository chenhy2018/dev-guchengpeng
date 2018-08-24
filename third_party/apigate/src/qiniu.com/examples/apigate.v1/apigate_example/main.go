package main

import (
	"net/http"
	"runtime"

	"qbox.us/cc/config"

	"github.com/qiniu/http/restrpc.v1"
	"github.com/qiniu/log.v1"

	apigate_example "qiniu.com/examples/apigate.v1"
)

// ---------------------------------------------------------------------------

type Config struct {
	ApigateExample apigate_example.Config `json:"apigate_example"`

	BindHost   string `json:"bind_host"`
	MaxProcs   int    `json:"max_procs"`
	DebugLevel int    `json:"debug_level"`
}

func main() {

	// Load Config

	config.Init("f", "qiniu", "apigate_example.conf")

	var conf Config
	if err := config.Load(&conf); err != nil {
		log.Fatal("config.Load failed:", err)
	}

	// General Settings

	runtime.GOMAXPROCS(conf.MaxProcs)
	log.SetOutputLevel(conf.DebugLevel)

	// new Service

	service, err := apigate_example.New(&conf.ApigateExample)
	if err != nil {
		log.Fatal("apigate_example.New failed:", err)
	}

	// run Service

	router := restrpc.Router{
		PatternPrefix: "v1",
	}
	log.Info("Starting apigate_example ...")
	err = http.ListenAndServe(conf.BindHost, router.Register(service))
	log.Fatal("http.ListenAndServe(apigate_example):", err)
}

// ---------------------------------------------------------------------------

