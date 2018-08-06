package v1

import (
	"net/http"
	"strings"
	"time"

	"qbox.us/servend/account"
	"qbox.us/servend/proxy_auth"

	"github.com/qiniu/errors"
	"github.com/qiniu/http/audit/impl"
	"github.com/qiniu/http/audit/probe"
	"github.com/qiniu/http/supervisor"
	"github.com/qiniu/largefile/log"

	"qiniu.com/probe/collector"

	. "github.com/qiniu/http/audit/proto"
)

// ----------------------------------------------------------

type iHandlerDecoder struct {
	IHandler
	impl.BaseDecoder
}

func (r *iHandlerDecoder) DecodeRequestEx(req *http.Request) (api, url string, header, params M) {

	_, api = r.IHandler.Handler(req)
	if len(api) > 0 && api[0] != '/' {
		index := strings.Index(api, "/")
		if index == -1 {
			api = "/"
		} else {
			api = api[index:]
		}
	}

	url, header, params = r.BaseDecoder.DecodeRequest(req)

	return
}

type decoder struct {
	*iHandlerDecoder
	account.AuthParser
}

func (r *decoder) DecodeRequestEx(req *http.Request) (api, url string, header, params M) {

	api, url, header, params = r.iHandlerDecoder.DecodeRequestEx(req)
	user, err := r.ParseAuth(req)
	if err != nil {
		return
	}

	token := M{
		"uid":   user.Uid,
		"utype": user.Utype,
	}
	if user.UtypeSu != 0 {
		token["sudoer"] = user.Sudoer
		token["utypesu"] = user.UtypeSu
	}
	if user.Appid != 0 {
		token["appid"] = user.Appid
	}
	if user.Devid != 0 {
		token["devid"] = user.Devid
	}
	header["Token"] = token
	return
}

// --------------------------------------------------------------------

type authProxy struct {
	p account.AuthParser
}

func (r authProxy) ParseAuth(req *http.Request) (user account.UserInfo, err error) {

	user, err = r.p.ParseAuth(req)
	if err == nil {
		req.Header.Set("Authorization", proxy_auth.MakeAuth(user))
	} else {
		req.Header.Del("Authorization") // 很重要：避免外界也可以发 proxy auth
	}
	return
}

// --------------------------------------------------------------------

type IHandler interface {
	Handler(*http.Request) (http.Handler, string)
}

type Config struct {
	Supervisor supervisor.Config `json:"supervisor"`
	LogFile    string            `json:"logdir"`
	ChunkBits  uint              `json:"chunkbits"`
	NoXlog     uint              `json:"noxlog"`
	BodyLimit  int               `json:"bodylimit"`
	AuthProxy  int               `json:"-"`
}

type Marker struct {
	*log.Logger
}

func (r Marker) Mark(measurement string,
	tim time.Time,
	tagK []string, tagV []string,
	fieldK []string, fieldV []interface{}) error {

	collector.MarkDirect(measurement, tim, tagK, tagV, fieldK, fieldV)
	return nil
}

func Open(mux IHandler, module string, cfg *Config, acc account.AuthParser) (al *probe.Marker, logf *log.Logger, err error) {

	if cfg.BodyLimit == 0 {
		cfg.BodyLimit = 1024
	}

	logf, err = log.Open(cfg.LogFile, cfg.ChunkBits)
	if err != nil {
		err = errors.Info(err, "jsonlog.Open: largefile/log.Open failed").Detail(err)
		return
	}

	var (
		dec  impl.DecoderEx
		iDec *iHandlerDecoder = &iHandlerDecoder{IHandler: mux}
	)
	if acc != nil {
		if cfg.AuthProxy != 0 {
			acc = authProxy{acc}
		}
		dec = &decoder{iHandlerDecoder: iDec, AuthParser: acc}
	} else {
		dec = iDec
	}
	al = probe.NewEx(module, Marker{logf}, dec, cfg.BodyLimit, cfg.NoXlog == 0)
	if cfg.Supervisor.BindHost != "" {
		spv := supervisor.New()
		al.SetEvent(spv)
		go spv.ListenAndServe(&cfg.Supervisor)
	}
	return
}

// ----------------------------------------------------------
