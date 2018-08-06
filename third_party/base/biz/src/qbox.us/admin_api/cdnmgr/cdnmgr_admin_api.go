package cdnmgr

import (
	"net/http"
	. "qbox.us/api/conf"
	"qbox.us/net/httputil"
)

// ----------------------------------------------------------

type Service struct {
	Conn httputil.Client
}

func New(t http.RoundTripper) Service {
	client := &http.Client{Transport: t}
	return Service{httputil.Client{client}}
}

// ----------------------------------------------------------

func (cdnmgr Service) Refresh(urls, dirs []string) (code int, err error) {

	param := map[string][]string{
		"urls": urls,
		"dirs": dirs,
	}
	return cdnmgr.Conn.CallWithForm(nil, CDNMGR_HOST+"/refresh/", param)
}

// ----------------------------------------------------------
