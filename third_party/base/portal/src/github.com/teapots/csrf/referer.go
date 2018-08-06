package csrf

import (
	"net/http"
	"net/url"

	"github.com/teapots/inject"
	"github.com/teapots/teapot"
)

const (
	CsrfStatusMessage = "CSRF Attack Detected"
)

func CsrfRefererFilter(handler ...teapot.Handler) inject.Provider {
	return func(ctx teapot.Context, rw http.ResponseWriter, req *http.Request, log teapot.Logger) {
		if validRequest(req, log) {
			return
		}

		if len(handler) > 0 {
			_, err := ctx.Invoke(handler[0])
			if err != nil {
				log.Error("csrf filter invoke handler err:", err)
			}

		} else {
			log.Notice(CsrfStatusMessage)

			rw.WriteHeader(http.StatusForbidden)
			rw.Write([]byte(CsrfStatusMessage))
		}
	}
}

func validRequest(req *http.Request, log teapot.Logger) bool {
	switch req.Method {
	case "GET", "HEAD":
		return true
	}

	referer := req.Referer()
	if referer == "" {
		log.Info("CSRF empty referer")
		return false
	}

	u, err := url.Parse(referer)
	if err != nil {
		log.Info("CSRF referer parse err", err)
		return false
	}

	if u.Host != req.Host {
		log.Info("CSRF referer not match", req.Host, u.Host)
		return false
	}
	return true
}
