package authproxy

import (
	"net/http"
	"strings"
)

type AuthProxyTimeoutOption struct {
	DialMs         int `json:"dial_ms"`
	GetRespMs      int `json:"get_resp_ms"`
	ProxyGetRespMs int `json:"proxy_get_resp_ms"`
}

func ShouldReproxy(code int, err error) bool {
	if err != nil {
		return strings.Contains(err.Error(), "connecting to proxy") && strings.Contains(err.Error(), "dial tcp")
	}
	return code == http.StatusServiceUnavailable
}
