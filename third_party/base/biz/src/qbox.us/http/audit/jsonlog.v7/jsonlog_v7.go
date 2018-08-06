package v3

import (
	"net/http"
	"qbox.us/servend/proxy_auth"

	"github.com/qiniu/errors"
	"github.com/qiniu/http/audit/jsonlog"
	"github.com/qiniu/http/supervisor"
	"github.com/qiniu/largefile/log"

	. "github.com/qiniu/http/audit/proto"
)

// ----------------------------------------------------------

type decoder struct {
	jsonlog.BaseDecoder
}

func (r decoder) DecodeRequest(req *http.Request) (url string, header, params M) {

	url, header, params = r.BaseDecoder.DecodeRequest(req)
	user, err := proxy_auth.ParseAuth(req)
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

type Config struct {
	Supervisor supervisor.Config `json:"supervisor"`
	LogFile    string            `json:"logdir"`
	ChunkBits  uint              `json:"chunkbits"`
	NoXlog     uint              `json:"noxlog"`
	BodyLimit  int               `json:"bodylimit"`
}

func Open(module string, cfg *Config) (al *jsonlog.Logger, logf *log.Logger, err error) {

	if cfg.BodyLimit == 0 {
		cfg.BodyLimit = 1024
	}

	logf, err = log.Open(cfg.LogFile, cfg.ChunkBits)
	if err != nil {
		err = errors.Info(err, "jsonlog.Open: largefile/log.Open failed").Detail(err)
		return
	}

	var dec decoder
	al = jsonlog.NewEx(module, logf, dec, cfg.BodyLimit, cfg.NoXlog == 0)
	if cfg.Supervisor.BindHost != "" {
		spv := supervisor.New()
		al.SetEvent(spv)
		go spv.ListenAndServe(&cfg.Supervisor)
	}
	return
}

// ----------------------------------------------------------
