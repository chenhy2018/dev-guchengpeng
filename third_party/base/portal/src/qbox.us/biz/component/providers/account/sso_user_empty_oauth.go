package account

import (
	"qbox.us/biz/component/client"
	"qbox.us/oauth"
)

func EmptyLoginOAuth(oauthConfig *oauth.Config) interface{} {
	return func(tr *client.TransportWithReqLogger) func() *oauth.Transport {
		return func() *oauth.Transport {
			return client.NewGenericOAuth(oauthConfig, tr)
		}
	}
}
