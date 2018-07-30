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
	"qbox.us/http/account.v2.1/digest_auth"
	qmsetm "qbox.us/qmset/master"
	. "qbox.us/qmset/proto"

	_ "github.com/qiniu/version"
)

type Config struct {
	FlipCfgs []*FlipConfig `json:"flips"`

	MgrAccessKey string   `json:"access_key"`
	MgrSecretKey string   `json:"secret_key"`  // 向 slave 发送指令时的帐号
	SlaveHosts   []string `json:"slave_hosts"` // [idc1_slave, idc2_slave, ...]

	BindHost string `json:"bind_host"`

	AuthConf digest_auth.Config `json:"auth"`

	MaxProcs   int `json:"max_procs"`
	DebugLevel int `json:"debug_level"`

	UidMgr uint32 `json:"uid_mgr"`

	LogSetFull bool `json:"log_set_full"`
}

// ------------------------------------------------------------------------------------------

func main() {

	// Load Config

	config.Init("f", "qbox", "qboxmsetm.conf")

	var conf Config
	if err := config.Load(&conf); err != nil {
		log.Fatal("config.Load failed:", err)
	}
	// General Settings

	runtime.GOMAXPROCS(conf.MaxProcs)
	log.SetOutputLevel(conf.DebugLevel)

	// new Service

	flipNotifier := qmsetm.NewFlipsNotifier(conf.MgrAccessKey, conf.MgrSecretKey, conf.SlaveHosts)
	authp := digest_auth.NewParser(&conf.AuthConf)
	cfg := &qmsetm.Config{
		FlipCfgs:      conf.FlipCfgs,
		FlipsNotifier: flipNotifier,
		UidMgr:        conf.UidMgr,
		AuthParser:    authp,
		LogSetFull:    conf.LogSetFull,
	}
	service, err := qmsetm.New(cfg)
	if err != nil {
		log.Fatal("qmsetm.New failed:", errors.Detail(err))
	}

	// run Service

	mux := servestk.New(nil)
	router := &webroute.Router{Factory: wsrpc.Factory, Mux: mux}
	router.Register(service)

	err = http.ListenAndServe(conf.BindHost, mux.Mux)
	log.Fatal("http.ListenAndServe:", err)
}

// ------------------------------------------------------------------------------------------
