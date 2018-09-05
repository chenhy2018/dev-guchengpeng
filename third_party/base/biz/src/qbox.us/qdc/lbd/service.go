package lbd

import (
	"encoding/base64"
	"net/http"
	"strconv"
	"strings"

	"qbox.us/auditlog2"
	"qbox.us/cc/time"
	"qbox.us/net/httputil"
	"qbox.us/net/serverstat"
	"qbox.us/rpc"
	"qbox.us/servestk"
	"qbox.us/state"
	"qbox.us/store"
	"qbox.us/timeio"
	"github.com/qiniu/xlog.v1"

	ioadmin "qbox.us/admin_api/io"
	. "qbox.us/api/bd/errors"
)

type Config struct {
	LogCfg          *auditlog2.Config
	LocalFileName   string
	LocalDirName    string
	CacheSpaceLimit int64
	Duration        int64
	CacheCountLimit int
	MaxBuf          int
	ChunkBits       uint
	Storage         StgInterface
	OtherLbds       map[int]store.MultiBdInterface
	MyId            int
}

type Service struct {
	*Config
	lbd *SimpleLocalBd
}

func New(cfg *Config) (service *Service, err error) {

	if cfg.OtherLbds == nil {
		cfg.OtherLbds = make(map[int]store.MultiBdInterface)
	}

	lbd, err := NewSimpleLocalBd(cfg.LocalDirName, cfg.LocalFileName, cfg.Duration, cfg.MaxBuf,
		cfg.ChunkBits, cfg.CacheCountLimit, cfg.CacheSpaceLimit, cfg.Storage, cfg.OtherLbds, cfg.MyId)

	return &Service{cfg, lbd}, err
}

func (p *Service) put(w http.ResponseWriter, req *http.Request, local bool) {

	xl := xlog.New(w, req)
	totalTm := time.NewMeter()
	defer func() {
		xl.Info("Service.put: total(100ns):", totalTm.Elapsed()/100)
	}()

	v := req.URL.Query()
	l := v.Get("len")
	k := v.Get("key")

	/*	c := v.Get("cache")
		doCache, _ := strconv.Atoi(c)

		xl.Info("Service.put:", k, l, c)
	*/

	doCache := 0
	xl.Info("Service.put:", k, l)

	if l == "" {
		xl.Warn("Service.put: l == nil")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	length, err := strconv.Atoi(l)
	if err != nil {
		xl.Warn("Service.put: atoi l:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	key, err := base64.URLEncoding.DecodeString(k)
	if err != nil {
		xl.Warn("Service.put: decode k:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if local {
		err = p.lbd.PutLocal(xl, req.Body, length, key)
	} else {
		bds := [3]uint16{0xffff, 0xffff, 0xffff}
		if list, ok := v["bds"]; ok && len(list) <= len(bds) {
			for i, item := range list {
				bd, err := strconv.ParseUint(item, 10, 16)
				if err != nil {
					xl.Warn("Service.put: invalid bd:", item)
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				bds[i] = uint16(bd)
			}
		}
		err = p.lbd.Put(xl, req.Body, length, key, doCache != 0, bds)
	}

	if err != nil {
		xl.Warn("Service.put: put to bd:", err, key)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(key)
}

func (p *Service) get(w1 http.ResponseWriter, req *http.Request, local bool) {

	xl := xlog.New(w1, req)
	totalTm := time.NewMeter()
	defer func() {
		xl.Info("Service.get: total(100ns):", totalTm.Elapsed()/100)
	}()

	w := rpc.ResponseWriter{w1}
	v := req.URL.Query()
	k := v.Get("key")
	f := v.Get("from")
	t := v.Get("to")

	i := v.Get("idc")
	idc := -1
	idc, _ = strconv.Atoi(i)

	xl.Info("Service.get:", k, f, t, i)

	if k == "" {
		xl.Warn("Service.get: k == nil")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	key, err := base64.URLEncoding.DecodeString(k)
	if err != nil {
		xl.Warn("Service.get: decode k", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	from := 0
	to := 0
	if f != "" {
		from, err = strconv.Atoi(f)
		if err != nil {
			xl.Warn("Service.get: atoi f:", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}
	if t != "" {
		to, err = strconv.Atoi(t)
		if err != nil {
			xl.Warn("Service.get: atii t:", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	if from > to {
		xl.Warn("Service.get: from <= to:", from, to)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if local {
		getTm := time.NewMeter()
		r, n, err := p.lbd.GetLocal(xl, key)
		xl.Info("Service.get: p.lbd.GetLogcal(100ns):", getTm.Elapsed()/100)
		if err != nil {
			if err == EKeyNotFound {
				w.ReplyWithError(http.StatusNotFound, err)
			} else {
				w.ReplyWithError(http.StatusInternalServerError, err)
			}
			return
		}
		defer r.Close()
		tr := timeio.NewReader(r)
		tm := time.NewMeter()
		w.ReplyWithBinary(tr, int64(n))
		xl.Debugf("Service.get: local, %v bytes, io.read %v(100ns), reply %v(100ns)", n, tr.Time()/100, tm.Elapsed()/100)
		return
	}

	bds := [4]uint16{0, 0xffff, 0xffff, uint16(idc)}
	if list, ok := v["bds"]; ok && len(list) < len(bds) {
		for i, item := range list {
			bd, err := strconv.ParseUint(item, 10, 16)
			if err != nil {
				xl.Warn("Service.get: invalid bd:", item)
				w.ReplyWithError(http.StatusBadRequest, err)
				return
			}
			bds[i] = uint16(bd)
		}
	}

	getTm := time.NewMeter()
	err = p.lbd.Get(xl, key, w, from, to, bds)
	xl.Info("Service.get: p.lbd.Get(100ns):", getTm.Elapsed()/100)
	if err != nil {
		xl.Warn("Service.get: lbd.get err :", err)
		if err == EKeyNotFound {
			httputil.ReplyErr(w, 404, err.Error())
			return
		}
		if strings.Contains(err.Error(), "write tcp") {
			// "write tcp 192.168.0.126:18072: broken pipe"
			// "write tcp 192.168.0.126:52119: connection reset by peer"
			httputil.ReplyErr(w, 499, err.Error())
			return
		}
		w.ReplyWithError(http.StatusInternalServerError, err)
	}
}

func (p *Service) serviceStat(w http.ResponseWriter, req *http.Request) {

	missing, total, wtotal := p.lbd.cache.Stat()

	httputil.Reply(w, 200, &ioadmin.CacheInfo{
		Missing: missing,
		Total:   total,
		Wtotal:  wtotal,
	})
}

func (p *Service) ServiceStat() interface{} {
	missing, total, wtotal := p.lbd.cache.Stat()
	cacheInfo := ioadmin.CacheInfo{Missing: missing, Total: total, Wtotal: wtotal}
	return struct {
		CacheInfo ioadmin.CacheInfo
		State     map[string]*state.Info
	}{cacheInfo, state.Dump()}
}

func (p *Service) RegisterHandlers(mux1 *http.ServeMux) (lh auditlog2.Instance, err error) {

	lh, err = auditlog2.Open("LBD", p.LogCfg, nil)
	if err != nil {
		return
	}
	mux := servestk.New(mux1, servestk.DiscardHandler, lh.Handler())

	mux.HandleFuncEx("put", "/put", func(w http.ResponseWriter, req *http.Request) { p.put(w, req, false) })
	mux.HandleFuncEx("get", "/get", func(w http.ResponseWriter, req *http.Request) { p.get(w, req, false) })

	mux.HandleFuncEx("put_local", "/put_local", func(w http.ResponseWriter, req *http.Request) { p.put(w, req, true) })
	mux.HandleFuncEx("get_local", "/get_local", func(w http.ResponseWriter, req *http.Request) { p.get(w, req, true) })

	mux.HandleFunc("/service-stat", func(w http.ResponseWriter, req *http.Request) { p.serviceStat(w, req) })
	return
}

func (p *Service) RegisterPubHandlers(mux1 *http.ServeMux) error {

	mux := servestk.New(mux1, servestk.DiscardHandler, servestk.SafeHandler)
	mux.HandleFuncEx("pub_get", "/get", func(w http.ResponseWriter, req *http.Request) { p.get(w, req, false) })
	return nil
}

func (p *Service) RunMgr(addr string) (err error) {

	return serverstat.Run(addr, p)
}

func (p *Service) Run(addr string) error {

	mux := http.NewServeMux()
	lh, err := p.RegisterHandlers(mux)
	if err != nil {
		return err
	}
	defer lh.Close()
	return http.ListenAndServe(addr, mux)
}

func (p *Service) RunPub(addr string) error {

	mux := http.NewServeMux()
	p.RegisterPubHandlers(mux)
	return http.ListenAndServe(addr, mux)
}
