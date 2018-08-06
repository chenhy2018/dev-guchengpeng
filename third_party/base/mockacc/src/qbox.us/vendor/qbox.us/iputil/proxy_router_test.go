package iputil

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"testing"

	"github.com/stretchr/testify.v1/require"
)

type M map[string]IPNet

var conf = `
[
	{
		"desc": "nb",
		"ip_net": "192.168.0.0/16",
		"proxies": ["http://192.168.1.100","http://192.168.33.1:8080"]
	},
	{
		"desc": "xs",
		"ip_net": "10.34.0.0/15",
		"proxies": ["socks5://192.168.1.100:1234"]
	}
]
`

var confWrong = `
[
	{
		"desc": "nb",
		"ip_net": "192.168.0.0/16",
		"proxies": ["http://192.168.1.100","http://192.168.33.1:8080"]
	},
	{
		"desc": "xs",
		"ip_net": "10.34.0.0",
		"proxies": ["socks://192.168.1.100:1234"]
	}
]
`

func TestProxyRouter(t *testing.T) {
	r := require.New(t)
	var pi ProxyInfos
	err := json.Unmarshal([]byte(conf), &pi)
	r.NoError(err)
	{
		u := pi.GetProxysFromIP(net.ParseIP("192.168.33.34"))
		r.Equal(2, len(u))
		r.Equal("http://192.168.1.100", u[0].String())
		r.Equal("http://192.168.33.1:8080", u[1].String())
	}
	{
		u := pi.GetProxysFromIP(net.ParseIP("10.34.10.22"))
		r.Equal(1, len(u))
		r.Equal("socks5://192.168.1.100:1234", u[0].String())
	}
	{
		u := pi.GetProxysFromIP(net.ParseIP("10.36.10.22"))
		r.Equal(0, len(u))
	}
	{
		_ = &http.Transport{
			Proxy: pi.Proxy,
		}
	}
	{
		req, err := http.NewRequest("POST", "http://10.34.23.45:3456", nil)
		r.NoError(err)
		u, err := pi.Proxy(req)
		r.NoError(err)
		r.Equal("socks5://192.168.1.100:1234", u.String())
	}
	{
		req, err := http.NewRequest("POST", "http://10.36.23.45:3456", nil)
		r.NoError(err)
		u, err := pi.Proxy(req)
		r.NoError(err)
		r.Nil(u)
	}
	{
		req, err := http.NewRequest("POST", "http://baidu.com:3456", nil)
		r.NoError(err)
		u, err := pi.Proxy(req)
		r.NoError(err)
		r.Nil(u)
	}
	{
		var pi ProxyInfos
		err := json.Unmarshal([]byte(confWrong), &pi)
		log.Println(err)
		r.Error(err)
	}
}
