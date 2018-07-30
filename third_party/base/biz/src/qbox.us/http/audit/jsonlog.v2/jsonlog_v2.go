package v2

import (
	"net/http"
	"qbox.us/servend/account"
	"qbox.us/servend/proxy_auth"

	"github.com/qiniu/errors"
	"github.com/qiniu/http/audit/jsonlog"
	"github.com/qiniu/largefile"
	"github.com/qiniu/reliable/log"
	"github.com/qiniu/reliable/osl"

	glog "github.com/qiniu/log.v1"
	. "github.com/qiniu/http/audit/proto"
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

func openLargeFiles(names []string, chunkBits uint, allowfails int) (files []osl.File, err error) {

	fails := 0
	files = make([]osl.File, len(names))
	for i, name := range names {
		f2, err2 := largefile.Open(name, chunkBits)
		if err2 != nil {
			glog.Warn("reliable.Open: largefile.Open failed -", errors.Detail(err2))
			fails++
			if fails > allowfails {
				for j := 0; j < i; j++ {
					if files[j] != nil {
						files[j].Close()
						files[j] = nil
					}
				}
				return nil, osl.ErrTooManyFails
			}
			continue
		}
		files[i] = f2
	}
	return
}

// ----------------------------------------------------------

type Config struct {
	LogFiles   []string `json:"logdirs"`
	AllowFails int      `json:"allowfails"`
	LineMax    int      `json:"linemax"`
	ChunkBits  uint     `json:"chunkbits"`
	NoXlog     uint     `json:"noxlog"`
	BodyLimit  int      `json:"bodylimit"`
	AuthProxy  int      `json:"-"`
}

func Open(module string, cfg *Config, acc account.AuthParser) (al *jsonlog.Logger, logf *log.Logger, err error) {

	if cfg.BodyLimit == 0 {
		cfg.BodyLimit = 1024
	}

	files, err := openLargeFiles(cfg.LogFiles, cfg.ChunkBits, cfg.AllowFails)
	if err != nil {
		err = errors.Info(err, "jsonlog.Open: openLargeFiles failed").Detail(err)
		return
	}

	logf, err = log.OpenEx(files, cfg.LineMax, cfg.AllowFails)
	if err != nil {
		err = errors.Info(err, "jsonlog.Open: log.OpenEx failed").Detail(err)
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
	return
}

// ----------------------------------------------------------
