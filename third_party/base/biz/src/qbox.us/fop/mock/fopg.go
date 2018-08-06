package mock

import (
	"github.com/qiniu/log.v1"
	"net/http"
	"qbox.us/api"
	"qbox.us/net/httputil"
	"qbox.us/servend/account"
	"strconv"
)

const utypeFopUser = account.USER_TYPE_OP

type Fopg struct {
	acc account.InterfaceEx
}

func NewFopg(acc account.InterfaceEx) *Fopg {
	return &Fopg{acc}
}

func (s *Fopg) Op(w http.ResponseWriter, req *http.Request) {
	if err := checkValidation(s.acc, req); err != nil {
		httputil.Error(w, err)
		return
	}
	body := "mockfopgate.Op"
	w.Header().Set("Content-Length", strconv.Itoa(len(body)))
	w.WriteHeader(200)
	w.Write([]byte(body))
}

func (s *Fopg) Nop(w http.ResponseWriter, req *http.Request) {
	if err := checkValidation(s.acc, req); err != nil {
		httputil.Error(w, err)
		return
	}
	body := "mockfopgate.Nop"
	w.Header().Set("Content-Length", strconv.Itoa(len(body)))
	w.WriteHeader(200)
	w.Write([]byte(body))
}

func checkValidation(acc account.InterfaceEx, req *http.Request) (err error) {
	user, err := account.GetAuthExt(acc, req)
	if err != nil || user.Utype&utypeFopUser == 0 {
		log.Error(err, user)
		err = api.EBadToken
	}
	return
}

func (s *Fopg) RegisterHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/op/", func(w http.ResponseWriter, req *http.Request) { s.Op(w, req) })
	mux.HandleFunc("/nop/", func(w http.ResponseWriter, req *http.Request) { s.Nop(w, req) })
}

func (s *Fopg) Run(addr string) error {
	mux := http.NewServeMux()
	s.RegisterHandlers(mux)
	err := http.ListenAndServe(addr, mux)
	return err
}
