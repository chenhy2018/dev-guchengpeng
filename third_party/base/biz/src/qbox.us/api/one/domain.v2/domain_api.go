package domain

import (
	"net/http"
	"strconv"

	"github.com/qiniu/rpc.v1"
	"qbox.us/api/one/domain"
)

type Client struct {
	domain.Client
}

func New(host string, t http.RoundTripper) Client {
	return Client{domain.New(host, t)}
}

func NewWithMultiHosts(hosts []string, t http.RoundTripper) Client {
	return Client{domain.NewWithMultiHosts(hosts, t)}
}

// 列出所有域名， 包括PublishedDomains和所有有通道权限的IDomains
func (p Client) Domains(l rpc.Logger, owner uint32, tbl string) (domains []string, err error) {

	err = p.Conn.CallWithForm(l, &domains, "/v2/domains", map[string][]string{
		"owner": {strconv.FormatUint(uint64(owner), 10)},
		"tbl":   {tbl},
	})
	return
}
