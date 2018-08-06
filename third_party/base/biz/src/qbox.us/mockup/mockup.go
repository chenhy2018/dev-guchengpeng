package mockup

import (
	"fmt"
	"net/http"
	"qbox.us/auditlog2"
	"qbox.us/net/httputil"
	"qbox.us/servend/account"
	"qbox.us/servestk"
)

// -----------------------------------------------------------

type Config struct {
	Account account.InterfaceEx
	LogCfg  *auditlog2.Config
	Addr    string
}

type Ret struct {
	Hash string `json:"hash"`
	Key  string `json:"key"`
}

// -----------------------------------------------------------
func uploadV2(w http.ResponseWriter, r *http.Request, cfg *Config) {

	uploadToken := r.FormValue("token")
	if uploadToken == "" {
		httputil.ReplyWithCode(w, 401)
		return
	}
	_, _, err := cfg.Account.DigestAuthEx(uploadToken)
	if err != nil {
		fmt.Println(err)
		httputil.ReplyWithCode(w, 401)
		return
	}

	f, _, err := r.FormFile("file")
	if err != nil {
		httputil.ReplyWithCode(w, 40)
		return
	}
	hash, err := GetEtag(f)
	if err != nil {
		httputil.ReplyWithCode(w, 400)
		return
	}

	ret := Ret{
		Hash: hash,
		Key:  r.FormValue("key"),
	}
	httputil.Reply(w, 200, ret)
	return
}

// -----------------------------------------------------------

func RegisterHandlers(mux1 *http.ServeMux, cfg *Config) (lh auditlog2.Instance, err error) {
	lh, err = auditlog2.Open("MOCKUP", cfg.LogCfg, nil)
	if err != nil {
		return
	}
	mux := servestk.New(mux1, lh.Handler())
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) { uploadV2(w, req, cfg) })
	return
}

func Run(cfg *Config) error {
	mux := http.NewServeMux()
	lh, err := RegisterHandlers(mux, cfg)
	if err != nil {
		return err
	}
	defer lh.Close()
	return http.ListenAndServe(cfg.Addr, mux)
}
