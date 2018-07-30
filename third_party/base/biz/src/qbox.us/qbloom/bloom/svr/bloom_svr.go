package svr

import (
	"encoding/base64"
	"net/http"

	"github.com/qiniu/http/httputil.v1"

	"qbox.us/qbloom/bloom"
)

var (
	ErrValueNotSpecified        = httputil.NewError(400, "value not specified")
	ErrValueMustBeUrlsafeBase64 = httputil.NewError(400, "value must be urlsafe-base64 encoded")
)

// ------------------------------------------------------------------------

func parseVal(req *http.Request) (v []byte, err error) {

	fv := req.FormValue("v")
	if fv == "" {
		return nil, ErrValueNotSpecified
	}

	v, err = base64.URLEncoding.DecodeString(fv)
	if err != nil {
		return nil, ErrValueMustBeUrlsafeBase64
	}
	return
}

// ------------------------------------------------------------------------

type Config struct {
	File string `json:"file"`
	M    uint   `json:"m"` // m 和 k 值一旦确定，不能更改
	K    uint   `json:"k"`
}

type Service struct {
	F *bloom.Filter
}

func Open(cfg *Config) (p Service, err error) {

	f, err := bloom.Open(cfg.File, cfg.M, cfg.K)
	if err != nil {
		return
	}

	return Service{f}, nil
}

func (p Service) Close() error {

	return p.F.Close()
}

func (p Service) DoSet(w http.ResponseWriter, req *http.Request) {

	v, err := parseVal(req)
	if err != nil {
		httputil.Error(w, err)
		return
	}

	err = p.F.Add(v)
	if err != nil {
		httputil.Error(w, err)
		return
	}

	httputil.ReplyWithCode(w, 200)
}

func (p Service) DoTest(w http.ResponseWriter, req *http.Request) {

	v, err := parseVal(req)
	if err != nil {
		httputil.Error(w, err)
		return
	}

	ok, err := p.F.Test(v)
	if err != nil {
		httputil.Error(w, err)
		return
	}

	code := 612
	if ok {
		code = 200
	}
	httputil.ReplyWithCode(w, code)
}

func (p Service) DoTas(w http.ResponseWriter, req *http.Request) {

	v, err := parseVal(req)
	if err != nil {
		httputil.Error(w, err)
		return
	}

	ok, err := p.F.TestAndAdd(v)
	if err != nil {
		httputil.Error(w, err)
		return
	}

	code := 612
	if ok {
		code = 200
	}
	httputil.ReplyWithCode(w, code)
}

// ------------------------------------------------------------------------
