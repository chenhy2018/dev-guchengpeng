package iam

import (
	"net/http"
)

const (
	KeyXPortalRealIP = "X-Portal-Real-Ip"
	KeyXPortalRealUA = "X-Portal-Real-User-Agent"
)

type RealInfo struct {
	IP string
	UA string
}

type RealInfoTransport struct {
	Transport http.RoundTripper
	RealInfo  *RealInfo
}

func NewRealInfoTransport(tp http.RoundTripper, info *RealInfo) http.RoundTripper {
	if tp == nil {
		tp = http.DefaultTransport
	}
	return &RealInfoTransport{
		Transport: tp,
		RealInfo:  info,
	}
}

func (t *RealInfoTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.RealInfo != nil {
		if t.RealInfo.IP != "" {
			req.Header.Set(KeyXPortalRealIP, t.RealInfo.IP)
		}
		if t.RealInfo.UA != "" {
			req.Header.Set(KeyXPortalRealUA, t.RealInfo.UA)
		}
	}
	return t.Transport.RoundTrip(req)
}
