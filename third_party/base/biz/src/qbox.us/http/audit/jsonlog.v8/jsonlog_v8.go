package v8

import (
	"net/http"
	"qbox.us/servend/account"
	"qbox.us/servend/proxy_auth"

	"github.com/qiniu/errors"
	"github.com/qiniu/filelog/log"
	"github.com/qiniu/http/audit/jsonlog"
	. "github.com/qiniu/http/audit/proto"
	"github.com/qiniu/http/supervisor"
)

// ----------------------------------------------------------

type decoder struct {
	jsonlog.BaseDecoder
	account.AuthParser
}

func (r *decoder) DecodeRequest(req *http.Request) (url string, header, params M) {

	url, header, params = r.BaseDecoder.DecodeRequest(req)
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

type Config struct {
	Supervisor    supervisor.Config `json:"supervisor"`
	LogDir        string            `json:"logdir"`
	LogNamePrefix string            `json:"lognameprefix"`
	TimeMode      int64             `json:"timemode"` // log按时间切分模式, 单位:秒, 最小1s, 最大1天(24*3600), 且能被24*3600整除, 默认10分钟(600)
	ChunkBits     uint              `json:"chunkbits"`
	NoXlog        uint              `json:"noxlog"`
	BodyLimit     int               `json:"bodylimit"`
	AuthProxy     int               `json:"-"`
}

func Open(module string, cfg *Config, acc account.AuthParser) (al *jsonlog.Logger, logf *log.Logger, err error) {

	if cfg.BodyLimit == 0 {
		cfg.BodyLimit = 1024
	}

	logf, err = log.NewLogger(cfg.LogDir, cfg.LogNamePrefix, cfg.TimeMode, cfg.ChunkBits)
	if err != nil {
		err = errors.Info(err, "jsonlog.Open: filelog/NewWriter failed").Detail(err)
		return
	}

	var dec jsonlog.Decoder
	if acc != nil {
		if cfg.AuthProxy != 0 {
			acc = authProxy{acc}
		}
		dec = &decoder{AuthParser: acc}
	}
	al = jsonlog.NewEx(module, logf, dec, cfg.BodyLimit, cfg.NoXlog == 0)
	if cfg.Supervisor.BindHost != "" {
		spv := supervisor.New()
		al.SetEvent(spv)
		go spv.ListenAndServe(&cfg.Supervisor)
	}
	return
}

// ----------------------------------------------------------
