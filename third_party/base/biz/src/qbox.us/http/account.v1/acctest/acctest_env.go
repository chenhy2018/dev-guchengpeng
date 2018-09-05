package acctest

import (
	"net/http"
	"net/http/httptest"
	account "qbox.us/http/account.v2"
	servacc "qbox.us/servend/account"
)

// ------------------------------------------------------------------------

type Config struct {
	Uid     uint32
	Sudoer  uint32
	Utype   uint32
	UtypeSu uint32
}

func NewEnv(cfg *Config) *account.Env {

	requireCfg(&cfg)

	if cfg.Utype == 0 {
		cfg.Utype = servacc.USER_TYPE_STDUSER
	}
	return newEnv(cfg)
}

func NewAdminEnv(cfg *Config) *account.AdminEnv {

	requireCfg(&cfg)

	if cfg.Utype == 0 {
		cfg.Utype = servacc.USER_TYPE_ADMIN
	}
	return &account.AdminEnv{
		Env: *newEnv(cfg),
	}
}

// ------------------------------------------------------------------------

func requireCfg(cfg **Config) {

	if *cfg == nil {
		*cfg = &Config{}
	}
}

func newEnv(cfg *Config) *account.Env {

	w := httptest.NewRecorder()
	req := &http.Request{
		Header: make(http.Header),
	}

	return &account.Env{
		W:   w,
		Req: req,
		UserInfo: servacc.UserInfo{
			Uid:     cfg.Uid,
			Sudoer:  cfg.Sudoer,
			Utype:   cfg.Utype,
			UtypeSu: cfg.UtypeSu,
		},
	}
}

// ------------------------------------------------------------------------
