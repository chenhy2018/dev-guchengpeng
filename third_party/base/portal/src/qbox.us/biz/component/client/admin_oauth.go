package client

import (
	"net/http"

	"qbox.us/oauth"
)

func NewAdminOAuth(host string, tr http.RoundTripper) *oauth.Transport {
	transport := &oauth.Transport{
		Config: &oauth.Config{
			Scope:    "Scope",
			TokenURL: host + "/oauth2/token",
		},
		Transport: tr,
	}
	return transport
}
