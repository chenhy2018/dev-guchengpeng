package utils

import (
	"net/http"
	"strings"
)

func RealIp(req *http.Request) string {
	if ip := req.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	res := strings.Split(req.RemoteAddr, ":")
	if len(res) > 0 {
		return res[0]
	}
	return ""
}
