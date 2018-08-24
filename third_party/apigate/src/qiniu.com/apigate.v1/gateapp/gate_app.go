package gateapp

import (
	"net/http"
	"runtime"

	"qbox.us/cc/config"

	"github.com/qiniu/apigate.v1"
	"github.com/qiniu/apigate.v1/proto"
	"github.com/qiniu/http/graceful.v1"
	"github.com/qiniu/log.v1"
	"github.com/qiniu/reqid.v1"
	"github.com/qiniu/xlog.v1"
	"qiniu.com/apigate.v1/audit/jsonlog"

	qconf "qbox.us/qconf/qconfapi"
	auth "qiniu.com/apigate.v1/auth"
	"qiniu.com/apigate.v1/metric/prometheus"
	"qiniu.com/apigate.v1/proxy"
	"qiniu.com/auth/account.v1"
	"qiniu.com/auth/account.v1/static.v1"
)

const (
	Apigate_Default = 0
	Apigate_MockApp = 1
)

// --------------------------------------------------------------------

type Config struct {
	Qconfg     qconf.Config       `json:"qconfg"`
	Audit      AuditConfig        `json:"audit"`
	Prometheus *prometheus.Config `json:"prometheus"`
	ApigateCfg string             `json:"apigate_conf"`
	StaticAuth static.Config      `json:"static_auth"`

	QuitWaitTimeoutMs int64 `json:"quit_wait_timeout_ms"`

	BindHost   string   `json:"bind_host"`
	BindHosts  []string `json:"bind_hosts"`
	MaxProcs   int      `json:"max_procs"`
	DebugLevel int      `json:"debug_level"`
	Reqid      string   `json:"reqid"`
}

func Main(mode int) {

	// Load Config
	switch mode {
	case Apigate_Default:
		config.Init("f", "qiniu", "qiniugate.conf")
	default:
		config.Init("f", "qiniu", "qiniumockgate.conf")
	}

	var conf Config
	if err := config.Load(&conf); err != nil {
		log.Fatalln("config.Load failed:", err)
	}

	// General Settings
	runtime.GOMAXPROCS(conf.MaxProcs)
	log.SetOutputLevel(conf.DebugLevel)

	if conf.Reqid == "v1" {
		xlog.SetGenReqId(reqid.Gen)
	}

	acccfg := &account.Config{
		Qconfg: conf.Qconfg,
	}
	acc := account.New(acccfg)
	staticAcc := static.New(&conf.StaticAuth)

	// Init up proxy
	proxy.InitUpProxy(http.DefaultTransport, acc)

	// Init proto-switch proxy
	proxy.InitProtoSwitchProxy(http.DefaultTransport, acc)

	// Init auths
	switch mode {
	case Apigate_Default:
		err := auth.Init(acc, staticAcc)
		if err != nil {
			log.Fatalln("auth.Init failed:", err)
		}
	default:
		auth.InitMock()
	}

	// audit log
	al, closer, err := initAuditLog(&conf.Audit)
	if err != nil {
		log.Fatalln("initAuditLog failed:", err)
	}
	defer closer.Close()

	// init metric
	var prom *prometheus.Prometheus = nil
	if conf.Prometheus != nil {
		prom, err = prometheus.New(*conf.Prometheus)
		if err != nil {
			log.Fatalln("new metric failed")
		}
	}

	// graceful service
	svrCreator := func() http.Handler {
		service, err := newApiSvr(conf.ApigateCfg, al, prom)
		if err != nil {
			log.Fatalln("New apigate failed:", err)
		}
		return service
	}
	svr := graceful.New(svrCreator)

	if prom != nil {
		go prom.Run()
	}

	// process signal
	go svr.ProcessSignals(conf.QuitWaitTimeoutMs)

	// run Service
	log.Info("Starting apigate ...")
	if conf.BindHost != "" {
		go listenAndServe(conf.BindHost, svr)
	}
	for _, bindHost := range conf.BindHosts {
		go listenAndServe(bindHost, svr)
	}
	ch := make(chan struct{})
	<-ch
}

func listenAndServe(bindHost string, handler http.Handler) {
	log.Info("listening", bindHost)
	err := http.ListenAndServe(bindHost, handler)
	if err != nil {
		log.Fatalln("http.ListenAndServe(apigate):", bindHost, err)
	}
}

// --------------------------------------------------------------------

func newApiSvr(confFile string, al *jsonlog.Logger, p *prometheus.Prometheus) (*apigate.Service, error) {

	var metricR proto.Metric = p
	if p == nil {
		metricR = proto.NilMetric
	}
	service, err := apigate.NewFromFile(confFile, metricR)
	if err != nil {
		return nil, err
	}

	service.Sink(al)
	if p != nil {
		service.Sink(p)
	}
	return service, nil
}

// --------------------------------------------------------------------
