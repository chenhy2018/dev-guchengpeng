package method

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/teapots/teapot"
)

func Test_OverrideFilter(t *testing.T) {
	assert := &teapot.Assert{T: t}

	tea := teapot.New()
	tea.ProvideAs(teapot.NewLogger(log.New(ioutil.Discard, "", 0)), (*teapot.Logger)(nil))

	tea.Filter(OverrideFilter())

	var endReq *http.Request
	tea.Routers(
		// for root router
		teapot.All(func(req *http.Request) {
			endReq = req
		}),
	)

	for _, method := range []string{"PUT", "DELETE", "PATCH", "put", "delete", "patch"} {
		req, _ := http.NewRequest("POST", "http://example.com/", nil)
		req.PostForm = url.Values{}
		req.PostForm.Set(ParamHTTPMethodOverride, method)
		rec := httptest.NewRecorder()
		tea.ServeHTTP(rec, req)

		t.Log("test method:", method)
		assert.True(rec.Code == http.StatusOK)
		assert.NotNil(endReq)
		assert.True(endReq.Method == strings.ToUpper(method))
	}
}

func Test_OverrideRequestMethod(t *testing.T) {
	assert := &teapot.Assert{T: t}

	req, _ := http.NewRequest("POST", "http://example.com/", nil)

	for _, method := range []string{"PUT", "DELETE", "PATCH", "put", "delete", "patch"} {
		t.Log("test override method:", method)
		assert.NoError(OverrideRequestMethod(req, method))

		method = strings.ToUpper(method)
		assert.True(req.Method == method)
		assert.True(req.Header.Get(HeaderHTTPMethodOverride) == method)
	}
}

func Test_IsValidOverrideMethod(t *testing.T) {
	assert := &teapot.Assert{T: t}

	assert.True(isValidOverrideMethod("") == false)
	assert.True(isValidOverrideMethod("POST") == false)

	for _, method := range []string{"PUT", "DELETE", "PATCH", "put", "delete", "patch"} {
		assert.True(isValidOverrideMethod(method))
	}
}
