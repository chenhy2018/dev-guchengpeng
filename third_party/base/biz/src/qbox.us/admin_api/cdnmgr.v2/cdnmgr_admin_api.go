package cdnmgr

import (
	"net/http"

	"qbox.us/net/httputil"
)

type Service struct {
	host   string
	client httputil.Client
}

func NewService(oauth http.RoundTripper, host string) *Service {
	return &Service{
		host:   host,
		client: httputil.Client{&http.Client{Transport: oauth}},
	}
}
func (s *Service) Refresh(urls, dirs []string) (code int, err error) {
	param := map[string][]string{
		"urls": urls,
		"dirs": dirs,
	}
	code, err = s.client.CallWithForm(nil, s.host+"/refresh/", param)
	return
}
