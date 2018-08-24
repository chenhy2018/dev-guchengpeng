package apigate

import (
	"net/http"
)

var g_proxys = make(map[string]http.Handler)

func RegisterProxy(name string, proxy http.Handler) {

	g_proxys[name] = proxy
}

func GetProxy(name string) (proxy http.Handler, ok bool) {

	proxy, ok = g_proxys[name]
	return
}
