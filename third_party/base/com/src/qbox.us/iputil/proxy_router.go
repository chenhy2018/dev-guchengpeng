package iputil

import (
	"encoding/json"
	"math/rand"
	"net"
	"net/http"
	"net/url"
)

type IPNet struct {
	net.IPNet
}

func (i *IPNet) UnmarshalJSON(b []byte) error {
	var s string
	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}
	_, ipnet, err := net.ParseCIDR(s)
	if err != nil {
		return err
	}
	*i = IPNet{*ipnet}
	return nil
}

type URL struct {
	url.URL
}

func (u *URL) UnmarshalJSON(b []byte) error {
	var s string
	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}
	u1, err := url.Parse(s)
	if err != nil {
		return err
	}
	*u = URL{*u1}
	return nil
}

type ProxyInfo struct {
	Desc    string `json:"desc"`
	IPNet   *IPNet `json:"ip_net"`
	Proxies []*URL `json:"proxies"`
}

type ProxyInfos []ProxyInfo

func (ps ProxyInfos) GetProxysFromIP(ip net.IP) []*URL {
	for _, p := range ps {
		if p.IPNet.Contains(ip) {
			return p.Proxies
		}
	}
	return nil
}

func getHost(address string) (host string) {
	host = address
	if host1, _, err := net.SplitHostPort(address); err == nil {
		host = host1
	}
	return
}

func (ps ProxyInfos) Proxy(req *http.Request) (*url.URL, error) {
	addrs, err := net.LookupHost(getHost(req.URL.Host))
	if err != nil || len(addrs) == 0 {
		return nil, nil
	}
	ip := net.ParseIP(addrs[0]).To4()
	if ip == nil {
		return nil, nil
	}
	urls := ps.GetProxysFromIP(ip)
	if len(urls) == 0 {
		return nil, nil
	}
	return &urls[rand.Intn(len(urls))].URL, nil
}
