package cdnmgr

import (
	"net/http"
	. "qbox.us/api/conf"
	"qbox.us/net/httputil"
	"strconv"
)

type Service struct {
	Conn httputil.Client
}

func NewService(t http.RoundTripper) *Service {
	return &Service{httputil.Client{&http.Client{Transport: t}}}
}

func (cdnmgr Service) Refresh(urls []string, dirs []string, delay int, preFetch bool) (code int, err error) {
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

	return cdnmgr.Conn.CallWithForm(nil, CDNMGR_HOST+"/refresh/", param)
}
