package redirect

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func handler(w http.ResponseWriter, req *http.Request) {
	req.URL, _ = url.Parse("http://localhost:12306/file/aaaaaaaaaaaa")
	w.WriteHeader(http.StatusNotFound)
}

func TestNotFoundRedirectHandler(t *testing.T) {
	req := &http.Request{
		Method: "GET",
		URL:    &url.URL{},
	}
	{
		w := httptest.NewRecorder()
		req.URL.Path = "//a/b///c"

		NotFoundRedirectHandler(w, req, handler)
		if w.Code != http.StatusMovedPermanently {
			t.Fatal("w.Code is not http.StatusMovedPermanently:", w.Code)
		}
		if location := w.Header().Get("Location"); location != "/a/b/c" {
			t.Fatal("Location Header is wrong:", location)
		}
	}

	{
		w := httptest.NewRecorder()
		req.URL.Path = "/a/b/c"
		NotFoundRedirectHandler(w, req, handler)
		if w.Code != http.StatusNotFound {
			t.Fatal("w.Code is not http.StatusNotFound:", w.Code)
		}
	}
}
