package su

import (
	"net/http"
	"qbox.us/cc/time"
	"qbox.us/rpc"
	"strconv"
)

type Service struct {
	Host string
	Conn rpc.Client
}

func New(host string, t http.RoundTripper) *Service {
	client := &http.Client{Transport: t}
	return &Service{host, rpc.Client{client}}
}

type SuRet struct {
	AccessToken string `json:"access_token"`
	TokenExpiry int64  `json:"expires_in"`
}

/*
	POST /su?id=<UserName>&utype=<Utype>
*/
func (r *Service) SuAs(name string, utype uint32) (ret SuRet, code int, err error) {

	code, err = r.Conn.CallWithForm(&ret, r.Host+"/su", map[string][]string{
		"id":    {name},
		"utype": {strconv.FormatUint(uint64(utype), 10)},
	})
	if err == nil {
		ret.TokenExpiry += time.Seconds()
	}
	return
}

func (r *Service) Su(name string) (ret SuRet, code int, err error) {

	code, err = r.Conn.CallWithForm(&ret, r.Host+"/su", map[string][]string{
		"id": {name},
	})
	if err == nil {
		ret.TokenExpiry += time.Seconds()
	}
	return
}
