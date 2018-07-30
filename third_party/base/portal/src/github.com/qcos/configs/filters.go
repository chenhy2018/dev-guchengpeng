package configs

import (
	"net/http"
	"os"
	"syscall"

	"qiniu.com/auth/authstub.v1"

	"github.com/teapots/inject"
	"github.com/teapots/teapot"
)

func StubTokenRequired(ctx teapot.Context, rw http.ResponseWriter, req *http.Request, log teapot.ReqLogger) {

	auth := req.Header.Get("Authorization")

	var err error
	defer func() {
		if err != nil {
			log.Warn("error token:", auth, err)
			rw.WriteHeader(403)
		}
	}()

	sinfo, err := authstub.Parse(auth)
	if err != nil {
		return
	}
	info, err := Repo.AccessRepo(sinfo.Access)
	if err != nil {
		return
	}

	ctx.Provide(info)
}

func GracefuleWait(ch chan os.Signal, code int, sigs ...os.Signal) inject.Provider {
	if len(sigs) == 0 {
		sigs = []os.Signal{syscall.SIGTERM}
	}

	reseived := false
	go func() {
		select {
		case sig := <-ch:
			for _, s := range sigs {
				if sig == s {
					reseived = true
					return
				}
			}
		}
	}()

	return func(rw http.ResponseWriter, req *http.Request, log teapot.ReqLogger) {
		if reseived {
			log.Info("not served request", req.URL, code)
			rw.WriteHeader(code)
			rw.Write(nil)
		}
	}
}
