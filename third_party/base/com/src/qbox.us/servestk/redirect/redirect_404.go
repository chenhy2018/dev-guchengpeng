package redirect

import (
	"net/http"
	"net/url"
	"path"
	"qbox.us/audit/logh"
)

func pathCleanPath(p string) string {
	np := path.Clean(p)
	// path.Clean removes trailing slash except for root;
	// put the trailing slash back if necessary.
	if p[len(p)-1] == '/' && np != "/" {
		np += "/"
	}
	return np
}

func NotFoundRedirectHandler(w http.ResponseWriter, r *http.Request, f func(w http.ResponseWriter, req *http.Request)) {
	var url url.URL
	if p := pathCleanPath(r.URL.Path); p != r.URL.Path {
		url = *r.URL
		url.Path = p
	} else {
		f(w, r)
		return
	}

	// originUrl := *r.URL
	redirected := false
	w1 := &hookedResponseWriter{
		ResponseWriter: w,
		hookedWriteHeader: func(code int) {
			if code == http.StatusNotFound {
				w.Header().Del("Content-Type")
				w.Header().Del("Content-Length")
				w.Header().Del("Cache-Control")
				w.Header().Del("Etag")
				w.Header().Del("Last-Modified")

				logh.Xwarn(w, "redirect")

				http.Redirect(w, r, url.String(), http.StatusMovedPermanently)
				redirected = true
				return
			}
			w.WriteHeader(code)
		},
		hookedWrite: func(p []byte) (int, error) {
			if redirected {
				// return 0, errors.New("redirected") ?
				return len(p), nil
			}
			return w.Write(p)
		},
	}
	f(w1, r)
}
