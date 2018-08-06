package ustack

import (
	"net/http"
	"net/url"

	"github.com/qiniu/rpc.v2"
)

// --------------------------------------------------

type Config struct {
	Hosts      []string
	TenantName string
	ProjectId  string
	User       string
	Password   string
	NewConn    func(hosts []string, rt http.RoundTripper) Conn
	Transport  http.RoundTripper
}

type Services map[string]Conn

func (p Services) Find(serviceName string) (conn Conn, ok bool) {

	conn, ok = p[serviceName]
	return
}

// --------------------------------------------------

type endpoint struct {
	AdminURL    string `json:"adminURL"`
	Region      string `json:"region"`
	InternalURL string `json:"internalURL"`
	PublicURL   string `json:"publicURL"`
	Id          string `json:"id"`
}

type serviceCatalog struct {
	Endpoints []*endpoint `json:"endpoints"`
	Type      string      `json:"type"`
	Name      string      `json:"name"`
}

type tokensRet struct {
	Access struct {
		Token          Token             `json:"token"`
		ServiceCatalog []*serviceCatalog `json:"serviceCatalog"`
	} `json:"access"`
}

func getHost(rawurl string) (host string, err error) {

	u, err := url.Parse(rawurl)
	if err != nil {
		return
	}
	host = u.Scheme + "://" + u.Host
	return
}

func (p *serviceCatalog) register(
	services Services, newConn func(hosts []string, rt http.RoundTripper) Conn, rt http.RoundTripper) {

	hosts := make([]string, len(p.Endpoints))
	for i, ep := range p.Endpoints {
		// 由于历史遗留问题，keystone 返回的 endpoint url 格式各不相同，
		// 因此这里统一只取 http://host，忽略 path
		h, err := getHost(ep.PublicURL)
		if err != nil {
			panic("invald endpoint url format: " + ep.PublicURL)
		}
		hosts[i] = h
	}
	services[p.Name] = newConn(hosts, rt)
}

func New(l rpc.Logger, cfg *Config) (p Services, err error) {

	newConn := cfg.NewConn
	if newConn == nil {
		newConn = defaultNewConn
	}

	conn := newConn(cfg.Hosts, cfg.Transport)

	var ret tokensRet
	var args = new(tokensArgs)
	args.Auth.TenantName = cfg.TenantName
	args.Auth.PasswordCredentials.Username = cfg.User
	args.Auth.PasswordCredentials.Password = cfg.Password
	err = conn.CallWithJson(l, &ret, "POST", "/v2.0/tokens", args)
	if err != nil && rpc.HttpCodeOf(err)/100 != 2 {
		return
	}
	token := new(Token)
	*token = ret.Access.Token

	userTransport := newTransport(token, args, conn, cfg.Transport)

	services := make(Services)
	for _, sc := range ret.Access.ServiceCatalog {
		sc.register(services, newConn, userTransport)
	}
	hosts := make([]string, len(cfg.Hosts))
	for i, host := range cfg.Hosts {
		hosts[i] = host + "/v3/US-INTERNAL"
	}
	services["keystone.v3.us-internal"] = newConn(hosts, userTransport)

	return services, nil
}

// --------------------------------------------------
