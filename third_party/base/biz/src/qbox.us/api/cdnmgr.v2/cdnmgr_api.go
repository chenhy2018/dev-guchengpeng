package cdnmgr

import (
	"net/http"
	"strconv"

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

func (s *Service) Refresh(urls []string, dirs []string, delay int, preFetch bool) (code int, err error) {
	param := map[string][]string{
		"urls": urls,
	}

	if len(dirs) > 0 {
		param["dirs"] = dirs
	}

	if delay > 0 {
		param["delay"] = []string{strconv.Itoa(delay)}
	}

	if preFetch {
		param["prefetch"] = []string{"1"}
	}

	return s.client.CallWithForm(nil, s.host+"/refresh/", param)
}
