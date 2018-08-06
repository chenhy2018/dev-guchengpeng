package autoct

import (
	"net/http"
	"github.com/qiniu/log.v1"
)

// ----------------------------------------------------------

var formCt = []string{"application/x-www-form-urlencoded"}

func FormHandler(w http.ResponseWriter, req *http.Request, f func(w http.ResponseWriter, req *http.Request)) {

	if _, ok := req.Header["Content-Type"]; !ok {
		log.Info("Auto-Content-Type:", req.URL.Path, formCt)
		req.Header["Content-Type"] = formCt
	}
	f(w, req)
}

// ----------------------------------------------------------
