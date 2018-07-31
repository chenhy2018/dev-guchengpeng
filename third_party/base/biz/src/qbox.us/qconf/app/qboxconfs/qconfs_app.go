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
	qconfs "qbox.us/qconf/slave"

	_ "github.com/qiniu/version"
	_ "qbox.us/autoua"
	_ "qbox.us/profile"
)

type Config struct {
	McHosts  []string `json:"mc_hosts"` // 互为镜像的Memcache服务
	BindHost string   `json:"bind_host"`

	AuditLog jsonlog.Config `json:"auditlog"`

	AuthConf digest_auth.Config `json:"auth"`

	MaxProcs   int `json:"max_procs"`
	DebugLevel int `json:"debug_level"`

	UidMgr uint32 `json:"uid_mgr"` // 只接受这个管理员发过来的请求

	IsProxy     bool               `json:"is_proxy"` //如果为true表示当前服务端作为一个代理节点使用
	ProxyConfig qconfs.ProxyConfig `json:"proxy_config"`
}

// ------------------------------------------------------------------------------------------

func main() {

	// Load Config

	config.Init("f", "qbox", "qboxconfs.conf")

	var conf Config
	if err := config.Load(&conf); err != nil {
		log.Fatal("config.Load failed:", err)
	}
	conf.AuditLog.AuthProxy = 1

	// General Settings

	runtime.GOMAXPROCS(conf.MaxProcs)
	log.SetOutputLevel(conf.DebugLevel)

	// new Service

	cfg := &qconfs.Config{
		McHosts:     conf.McHosts,
		UidMgr:      conf.UidMgr,
		IsProxy:     conf.IsProxy,
		ProxyConfig: conf.ProxyConfig,
	}
	service, err := qconfs.New(cfg)
	if err != nil {
		log.Fatal("qconfs.New failed:", errors.Detail(err))
	}
	if cfg.IsProxy {
		service.LoopRefreshFailedItems()
	}

	// run Service

	authp := digest_auth.NewParser(&conf.AuthConf)
	al, logf, err := jsonlog.Open("CFGS", &conf.AuditLog, authp)
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
