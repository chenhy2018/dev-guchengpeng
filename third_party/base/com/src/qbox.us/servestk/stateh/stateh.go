package stateh

import (
	"net/http"
	"qbox.us/servestk"
	"qbox.us/state"
	"strings"
)

//---------------------------------------------------------------------------//

var disabledURLs []string

func DisableURL(url string) {
	disabledURLs = append(disabledURLs, url)
}

func IsURLDisabled(url string) bool {
	for _, m := range disabledURLs {
		if m == url {
			return true
		}
	}
	return false
}

//---------------------------------------------------------------------------//

func Instance(w http.ResponseWriter, req *http.Request,
	f func(w http.ResponseWriter, req *http.Request)) {

	if IsURLDisabled(req.URL.Path) {
		servestk.SafeHandler(w, req, f)
		return
	}

	method := ""
	if req.URL.Path == "/" {
		method = "/"
	} else {
		method = "/" + strings.SplitN(req.URL.Path, "/", -1)[1]
	}
	unit := state.Enter("http://" + method)
	defer unit.Leave(nil)

	servestk.SafeHandler(w, req, f)
}
