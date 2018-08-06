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
	jsonlog "qbox.us/http/audit/jsonlog.v3"
	. "qbox.us/qmset/proto"
	qmsets "qbox.us/qmset/slave"

	_ "github.com/qiniu/version"
)

type Config struct {
	FlipCfgs []*FlipConfig `json:"flips"`

	MgrAccessKey string `json:"access_key"`
	MgrSecretKey string `json:"secret_key"` // 向 master 发送指令时的帐号
	MasterHost   string `json:"master_host"`

	BindHost string `json:"bind_host"`

	AuditLog jsonlog.Config `json:"auditlog"`

	AuthConf digest_auth.Config `json:"auth"`

	MaxProcs   int `json:"max_procs"`
	DebugLevel int `json:"debug_level"`

	UidMgr uint32 `json:"uid_mgr"`
}

// ------------------------------------------------------------------------------------------

func main() {

	// Load Config

	config.Init("f", "qbox", "qboxmsets.conf")

	var conf Config
	if err := config.Load(&conf); err != nil {
		log.Fatal("config.Load failed:", err)
	}
	conf.AuditLog.AuthProxy = 1

	// General Settings

	runtime.GOMAXPROCS(conf.MaxProcs)
	log.SetOutputLevel(conf.DebugLevel)

	// new Service

	addNotifier := qmsets.NewAddNotifier(conf.MgrAccessKey, conf.MgrSecretKey, conf.MasterHost)
	cfg := &qmsets.Config{
		FlipCfgs:    conf.FlipCfgs,
		AddNotifier: addNotifier,
		UidMgr:      conf.UidMgr,
	}
	service, err := qmsets.New(cfg)
	if err != nil {
		log.Fatal("qmsets.New failed:", errors.Detail(err))
	}

	// run Service

	authp := digest_auth.NewParser(&conf.AuthConf)
	al, logf, err := jsonlog.Open("MSETS", &conf.AuditLog, authp)
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
