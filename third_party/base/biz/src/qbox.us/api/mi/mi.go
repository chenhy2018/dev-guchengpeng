package mi

import (
	"net/http"
	"qbox.us/rpc"
)

var Agent string // eg. qbox-ios-0.5.11-dev

// ----------------------------------------------------------

type Service struct {
	Host string
	Conn rpc.Client
}

func New(host string, t http.RoundTripper) *Service {
	client := &http.Client{Transport: t}
	return &Service{host, rpc.Client{client}}
}

func agentParam() map[string][]string {
	return map[string][]string{"agent": {Agent}}
}

func (p *Service) IoHosts() (ioHosts []string, code int, err error) {
	code, err = p.Conn.CallWithForm(&ioHosts, p.Host+"/iohosts", agentParam())
	return
}

type Host struct {
	FS string `json:"fs"`
	ES string `json:"es"`
	RS string `json:"rs"`
	IO string `json:"io"`
}

func (p *Service) Hosts() (host Host, code int, err error) {
	code, err = p.Conn.CallWithForm(&host, p.Host+"/hosts", agentParam())
	return
}

type accHostRet struct {
	Host string `json:"account"`
}

func (p *Service) Connect() (account string, code int, err error) {
	var ret accHostRet
	code, err = rpc.DefaultClient.CallWithForm(&ret, p.Host+"/connect", agentParam())
	account = ret.Host
	return
}

// ----------------------------------------------------------
