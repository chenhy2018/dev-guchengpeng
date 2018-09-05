package account

import (
	"qbox.us/biz/component/client"
	"qbox.us/oauth"
)

func EmptyUserOAuth(host string) interface{} {
	return func(tr *client.TransportWithReqLogger) func() *oauth.Transport {
		return func() *oauth.Transport {
			return client.NewUserOAuth(host, tr)
		}
	}
}
