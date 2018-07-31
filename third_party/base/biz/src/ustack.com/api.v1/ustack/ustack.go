package ustack

import (
	"net/http"
	"net/url"
	"time"

	"github.com/qiniu/log.v1"
	"github.com/qiniu/rpc.v2"
)

// --------------------------------------------------

type Config struct {
	Hosts     []string
	ProjectId string
	User      string
	Password  string
	NewConn   func(hosts []string, rt http.RoundTripper) Conn
	Transport http.RoundTripper
}

type Services map[string]Conn

func (p Services) Find(serviceName string) (conn Conn, ok bool) {

	conn, ok = p[serviceName]
	return
}

// --------------------------------------------------

type endpoint struct {
	Id        string `json:"id"`
	Interface string `json:"interface"`
	Region    string `json:"region"`
	Url       string `json:"url"`
}

type catalog struct {
	Endpoints []*endpoint `json:"endpoints"`
	Id        string      `json:"id"`
	Type      string      `json:"type"`
}

type Token struct {
	Id        string     `json:"-"`
	Catalog   []*catalog `json:"catalog"`
	ExpiresAt string     `json:"expires_at"`
	IssuedAt  string     `json:"issued_at"`
	//...
}

func (t *Token) expired() bool {
	tim, err := time.Parse(time.RFC3339, t.ExpiresAt)
	if err != nil {
		log.Error("parse time failed:", t.ExpiresAt, err)
		return false
	}
	return tim.Before(time.Now())
}

type tokensRet struct {
	Token Token `json:"token"`
}

func getHost(rawurl string) (host string, err error) {

	u, err := url.Parse(rawurl)
	if err != nil {
		return
	}
	host = u.Scheme + "://" + u.Host
	return
}

func (p *catalog) register(
	services Services, newConn func(hosts []string, rt http.RoundTripper) Conn, rt http.RoundTripper) {

	hosts := make([]string, len(p.Endpoints))
	for i, ep := range p.Endpoints {
		// 由于历史遗留问题，keystone 返回的 endpoint url 格式各不相同，
		// 因此这里统一只取 http://host，忽略 path
		h, err := getHost(ep.Url)
		if err != nil {
			panic("invald endpoint url format: " + ep.Url)
		}
		hosts[i] = h
	}
	services[p.Type] = newConn(hosts, rt)
}

func New(l rpc.Logger, cfg *Config) (p Services, err error) {

	newConn := cfg.NewConn
	if newConn == nil {
		newConn = defaultNewConn
	}

	conn := newConn(cfg.Hosts, cfg.Transport)

	var ret tokensRet
	var args = new(tokensArgs)

	args.Auth.Identity.Methods = []string{"password"}
	args.Auth.Identity.Password.User.Id = cfg.User
	args.Auth.Identity.Password.User.Password = cfg.Password
	args.Auth.Scope.Project.Id = cfg.ProjectId

	resp, err := conn.DoRequestWithJson(l, "POST", "/v3/auth/tokens", args)
	if err != nil {
		return
	}
	err = rpc.CallRet(l, &ret, resp)
	if err != nil && rpc.HttpCodeOf(err)/100 != 2 {
		return
	}

	token := new(Token)
	*token = ret.Token
	token.Id = resp.Header.Get("X-Subject-Token")

	userTransport := newTransport(token, args, conn, cfg.Transport)

	services := make(Services)
	for _, cl := range ret.Token.Catalog {
		cl.register(services, newConn, userTransport)
	}
	return services, nil
}

// --------------------------------------------------
