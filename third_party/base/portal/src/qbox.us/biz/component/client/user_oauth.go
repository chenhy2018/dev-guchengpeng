package client

import (
	"net/http"

	"qbox.us/oauth"
)

func NewUserOAuth(host string, tr http.RoundTripper) *oauth.Transport {
	transport := &oauth.Transport{
		Config: &oauth.Config{
			Scope:    "Scope",
			TokenURL: host + "/oauth2/token",
		},
		Transport: tr,
	}
	return transport
}

func NewGenericOAuth(oauthConfig *oauth.Config, tr http.RoundTripper) *oauth.Transport {
	transport := &oauth.Transport{
		Config:    oauthConfig,
		Transport: tr,
	}
	return transport
}
