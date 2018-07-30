package csrf

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/teapots/teapot"
)

var logger = teapot.NewLogger(log.New(ioutil.Discard, "", 0))

func Test_CsrfRefererFilter(t *testing.T) {
	assert := &teapot.Assert{T: t}

	tea := teapot.New()
	tea.SetLogger(logger)

	csrfFilter := CsrfRefererFilter()

	tea.Routers(
		teapot.Filter(csrfFilter),

		// for root router
		teapot.Post(func() {}),

		teapot.Router("/user_1",
			teapot.Post(func() {}),
		),

		teapot.Router("/user_2",
			teapot.Exempt(csrfFilter),
			teapot.Post(func() {}),
		),

		teapot.Router("/user_3",
			teapot.Post(func() {}).Exempt(csrfFilter),
		),
	)

	assert.False(csrfPass(tea, "POST", "http://example.com/", ""))
	assert.True(csrfPass(tea, "POST", "http://example.com/", "http://example.com/"))

	assert.False(csrfPass(tea, "POST", "http://example.com/user_1", ""))
	assert.True(csrfPass(tea, "POST", "http://example.com/user_1", "http://example.com/"))

	assert.True(csrfPass(tea, "POST", "http://example.com/user_2", ""))
	assert.True(csrfPass(tea, "POST", "http://example.com/user_2", "http://example.com/"))

	assert.True(csrfPass(tea, "POST", "http://example.com/user_3", ""))
	assert.True(csrfPass(tea, "POST", "http://example.com/user_3", "http://example.com/"))
}

func Test_ValidRequest(t *testing.T) {
	assert := &teapot.Assert{T: t}

	var req *http.Request

	req, _ = http.NewRequest("GET", "http://example.com/", nil)
	assert.True(validRequest(req, logger))

	req, _ = http.NewRequest("HEAD", "http://example.com/", nil)
	assert.True(validRequest(req, logger))

	req, _ = http.NewRequest("POST", "http://example.com/", nil)
	assert.False(validRequest(req, logger))

	req, _ = http.NewRequest("POST", "http://example.com/", nil)
	req.Header.Set("Referer", "http://example.com:8080/")
	assert.False(validRequest(req, logger))

	req, _ = http.NewRequest("POST", "http://example.com/", nil)
	req.Header.Set("Referer", "http://example.com/dash")
	assert.True(validRequest(req, logger))
}

func csrfPass(tea *teapot.Teapot, method, urlStr, referer string) bool {
	req, _ := http.NewRequest(method, urlStr, nil)
	req.Header.Set("Referer", referer)
	rec := httptest.NewRecorder()
	tea.ServeHTTP(rec, req)
	tea.Logger().Info(method, urlStr, referer, rec.Code, rec.Body.String())
	if rec.Code == http.StatusForbidden || rec.Body.String() == CsrfStatusMessage {
		return false
	}
	return true
}
