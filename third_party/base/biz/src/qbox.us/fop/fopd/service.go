package fopd

import (
	"encoding/base64"
	"net/http"
	"runtime/debug"
	"strings"

	"code.google.com/p/go.net/context"
	"github.com/qiniu/http/httputil.v1"
	"github.com/qiniu/log.v1"
	"github.com/qiniu/xlog.v1"

	"qbox.us/api"
	"qbox.us/auditlog2"
	"qbox.us/dcutil"
	"qbox.us/fop"
	"qbox.us/fop/clients/fopd.v2"
	"qbox.us/fop/localcache"
	"qbox.us/qdiscover/discover"
	"qbox.us/servend/account"
	"qbox.us/servestk"
	"qbox.us/servestk/metrics"

	dcapi "qbox.us/api/dc"
	qtime "qbox.us/cc/time"
)

type Config struct {
	Fops            map[string]func(w fop.Writer, r fop.Reader)
	Cache           dcapi.DiskCache
	LogCfg          auditlog2.Config
	TempDir         string
	LocalCache      localcache.LocalCacheConfig
	Account         account.InterfaceEx
	HeartbeatConfig discover.HeartbeatConfig
	AccessKey       string
	SecretKey       string
	PfopHosts       []string
	RsHosts         []string
}

type Service struct {
	Config
	lc    *localcache.LocalCache
	dcExt dcutil.Interface

	metrics   *metrics.Status
	heartbeat *discover.Heartbeat
}

var (
	AdminAK   = ""
	AdminSK   = ""
	PfopHosts []string
	RsHosts   []string
)

func New(cfg *Config) *Service {
	log.Info("supported fop:")
	var cmds []string
	for k := range cfg.Fops {
		cmds = append(cmds, k)
		log.Info(k)
	}

	AdminAK = cfg.AccessKey
	AdminSK = cfg.SecretKey
	PfopHosts = cfg.PfopHosts
	RsHosts = cfg.RsHosts
	log.Info("has set ak/sk, pfopHosts, rsHosts")

	if cfg.LocalCache.DiskDir == "" {
		cfg.LocalCache.DiskDir = cfg.TempDir
	}

	lc, err := localcache.NewLocalCache(&cfg.LocalCache)
	if err != nil {
		log.Fatal("NewLocalCache failed:", err)
	}

	var dcExt dcutil.Interface
	if cfg.Cache != nil {
		dcExt = dcutil.NewExt(dcapi.NewDiskCacheExt(cfg.Cache))
	}

	s := &Service{Config: *cfg, lc: lc, dcExt: dcExt, metrics: metrics.NewStatus()}

	cfg.HeartbeatConfig.GetAttrs = func() discover.Attrs {
		return discover.Attrs{
			"processing": s.metrics.Counter.Count(),
			"cmds":       cmds,
		}
	}

	if len(cfg.HeartbeatConfig.DiscoverHosts) > 0 {
		s.heartbeat, err = discover.NewHeartbeat(&cfg.HeartbeatConfig)
		if err != nil {
			log.Fatal("NewHeartbeat failed:", err)
		}
	}
	return s
}

// ----------------------------------------------------------------

type fopEnv struct {
	xl         *xlog.Logger
	tempDir    string
	localCache *localcache.LocalCache
	dc         dcutil.Interface
	acc        account.InterfaceEx
}

func (e *fopEnv) Xlogger() *xlog.Logger      { return e.xl }
func (e *fopEnv) TempDir() string            { return e.tempDir }
func (e *fopEnv) Xdc() dcutil.Interface      { return e.dc }
func (e *fopEnv) Acc() account.InterfaceEx   { return e.acc }
func (e *fopEnv) LocalCache() fop.LocalCache { return e.localCache }

// ----------------------------------------------------------------

//
// POST /op?src=<EncodedSrcURL>&fh=<Fhandle>&fsize=<Fsize>&cmd=<FopCmd>&sp=<StyleParam>&url=<EncodedIOURL>&token=<Token>
//          [mode=<Mode>&uid=<Uid>&bucket=<Bucket>&key=<Key>]
// Body: [<FileData>]
//
// 当有 src 时，HTTP Body 为空，源数据通过 HTTP GET SrcURL 获取，如果没有 src，源数据则在 HTTP Body 中。
//
func (p *Service) Do(w http.ResponseWriter, req *http.Request) {
	xl := xlog.New(w, req)
	totalTm := qtime.NewMeter()
	defer func() {
		xl.Info("Fopd.Do: total(100ns):", totalTm.Elapsed()/100)
	}()

	xl.Info("url:", req.URL)

	fh, fsize, ctx, err := fopd.DecodeQuery(req.URL.RawQuery)
	if err != nil {
		err = httputil.NewError(api.InvalidArgs, err.Error())
		httputil.Error(w, err)
		return
	}
	// 对空文件做 fop 对于我们内置数据处理来说暂无意义，所以直接在 fopd 框架里面返回错误。
	if fsize == 0 {
		err = httputil.NewError(api.InvalidArgs, "empty file")
		httputil.Error(w, err)
		return
	}

	// source
	var (
		src       fop.Source
		decodeURL string
	)
	if ctx.SourceURL != "" {
		rc := NewUrlReader(xl, ctx.SourceURL)
		defer rc.Close()
		src = rc
	} else {
		defer req.Body.Close()
		src = req.Body
	}
	//URL
	bytes, err := base64.URLEncoding.DecodeString(ctx.URL)
	if err != nil {
		err = httputil.NewError(api.InvalidArgs, "invalid ctx.URL")
		httputil.Error(w, err)
		return
	} else {
		decodeURL = string(bytes)
	}
	// cmd
	cmdQuery := strings.Split(ctx.RawQuery, "/")
	if ctx.CmdName == "" {
		ctx.CmdName = cmdQuery[0]
	}
	f, ok := p.Fops[ctx.CmdName]
	if !ok {
		err = httputil.NewError(api.InvalidArgs, "unsupported cmd "+ctx.CmdName)
		httputil.Error(w, err)
		return
	}

	// mimeType
	if ctx.MimeType == "" {
		ctx.MimeType = req.Header.Get("Content-Type")
	}
	if ctx.MimeType == "application/octet-stream" {
		ctx.MimeType = ""
	}

	tpCtx, cancel := context.WithCancel(context.Background())
	tpCtx = xlog.NewContext(tpCtx, xl)

	fopReq := &fop.Request{
		Source:               src,
		SourceURL:            ctx.SourceURL,
		Query:                cmdQuery,
		Fsize:                fsize,
		Env:                  &fopEnv{xl, p.TempDir, p.lc, p.dcExt, p.Account},
		IsGlobal:             ctx.IsGlobal,
		RawQuery:             ctx.RawQuery,
		StyleParam:           ctx.StyleParam,
		MimeType:             ctx.MimeType,
		URL:                  decodeURL,
		Token:                ctx.Token,
		Mode:                 ctx.Mode,
		Uid:                  ctx.Uid,
		Bucket:               ctx.Bucket,
		Key:                  ctx.Key,
		Fh:                   fh,
		OutRSBucket:          ctx.OutRSBucket,
		OutRSDeleteAfterDays: ctx.OutRSDeleteAfterDays,
		Ctx:                  tpCtx,
	}

	reqDone := make(chan bool, 1)
	go func() {
		defer func() {
			p := recover()
			if p != nil {
				w.WriteHeader(597)
				xl.Errorf("WARN: panic fired in %v.panic - %v\n", f, p)
				xl.Println(string(debug.Stack()))
				reqDone <- true
			}
		}()
		f(w, fop.Reader(fopReq))
		reqDone <- true
	}()
	select {
	case <-httputil.GetCloseNotifierSafe(w).CloseNotify():
		cancel()
		xl.Info("request was closed by peer:", tpCtx.Err())
		<-reqDone
	case <-reqDone:
	}
}

func (p *Service) RegisterHandlers(mux1 *http.ServeMux, mod string) (lh auditlog2.Instance, err error) {
	if mod == "" {
		mod = "FOPD"
	}
	lh, err = auditlog2.Open(mod, &p.LogCfg, nil)
	if err != nil {
		return
	}
	mux := servestk.New(mux1, p.metrics.Handler(), lh.Handler())
	mux.HandleFunc("/op", func(w http.ResponseWriter, req *http.Request) { p.Do(w, req) })
	mux.HandleFunc("/op2", func(w http.ResponseWriter, req *http.Request) { p.Do(w, req) })
	return
}

func (p *Service) Run(addr string, mod string) (err error) {
	mux := http.NewServeMux()
	lh, err := p.RegisterHandlers(mux, mod)
	if err != nil {
		return
	}
	defer lh.Close()
	err = http.ListenAndServe(addr, mux)
	return
}

func (p *Service) RunEx(addr string, mux1 servestk.Mux) (err error) {

	mux := servestk.New(mux1, p.metrics.Handler())
	mux.HandleFunc("/op", func(w http.ResponseWriter, req *http.Request) { p.Do(w, req) })
	mux.HandleFunc("/op2", func(w http.ResponseWriter, req *http.Request) { p.Do(w, req) })
	return http.ListenAndServe(addr, mux)
}
