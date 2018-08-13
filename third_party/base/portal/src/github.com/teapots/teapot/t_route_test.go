package teapot

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var nopFunc = func() {}

func Test_NoRoute(t *testing.T) {
	assert := &Assert{T: t}

	tea := New().Routers()
	assert.True(routeNotFound(tea, "GET", "/"))
}

func Test_RootRoute(t *testing.T) {
	assert := &Assert{T: t}

	tea := New().Routers(
		Get(nopFunc),
	)
	assert.True(routeFound(tea, "GET", "/"))
	assert.True(routeFound(tea, "HEAD", "/"))
	assert.True(routeNotFound(tea, "POST", "/"))

	tea = New().Routers(
		Post(nopFunc),
	)
	assert.True(routeNotFound(tea, "GET", "/"))
	assert.True(routeNotFound(tea, "HEAD", "/"))
	assert.True(routeFound(tea, "POST", "/"))

}

func Test_UrlRoute(t *testing.T) {
	assert := &Assert{T: t}

	path := ""
	pathFunc := func(req *http.Request) {
		path = req.URL.Path
	}

	tea := New().Routers(
		// root
		Get(nopFunc),

		Router("/user",
			Get(nopFunc),
			Put(nopFunc),

			Router("/dashboard", Get(pathFunc)),

			Router("/:uid",
				Get(pathFunc),

				Router("/dashboard", Get(pathFunc)),
			),

			Router("/:uid/name/:name", Get(nopFunc)),
			Router("/:uid/test/:nopanic", Get(nopFunc)),
		),

		Router("/user/dashboard", Put(pathFunc)),
		Router("/user/:uid/dashboard", Put(pathFunc)),
	)

	assert.True(routeFound(tea, "GET", "/"))

	assert.True(routeFound(tea, "GET", "/user"))
	assert.True(routeFound(tea, "PUT", "/user"))

	assert.True(routeFound(tea, "GET", "/user/dashboard"))
	assert.True(path == "/user/dashboard")

	assert.True(routeFound(tea, "GET", "/user/10"))
	assert.True(path == "/user/10")

	assert.True(routeFound(tea, "GET", "/user/10/dashboard"))
	assert.True(path == "/user/10/dashboard")

	assert.True(routeFound(tea, "PUT", "/user/dashboard"))
	assert.True(path == "/user/dashboard")

	assert.True(routeFound(tea, "PUT", "/user/10/dashboard"))
	assert.True(path == "/user/10/dashboard")

	assert.True(routeNotFound(tea, "POST", "/user"))
	assert.True(routeNotFound(tea, "POST", "/user/dashboard"))
}

func Test_RouteInfo(t *testing.T) {
	assert := &Assert{T: t}

	var info *RouteInfo
	infoFunc := func(i *RouteInfo) {
		info = i
	}

	tea := New().Routers(
		// root
		Get(infoFunc),

		Router("/user",
			Get(infoFunc),

			Router("/:uid",
				Get(infoFunc),
			),

			Router("/:uid/name/:name",
				Get(infoFunc),
			),
		),
	)

	assert.True(routeFound(tea, "GET", "/user/101"))
	assert.True(info.Get("uid") == "101")

	assert.True(routeFound(tea, "GET", "/user/101/name/slene"))
	assert.True(info.Get("uid") == "101")
	assert.True(info.Get("name") == "slene")

	assert.True(routeFound(tea, "GET", "/"))
	t.Log(info.Path)
	assert.True(info.Path == "/")

	assert.True(routeFound(tea, "GET", "/user"))
	t.Log(info.Path)
	assert.True(info.Path == "/user")

	assert.True(routeFound(tea, "GET", "/user/101"))
	t.Log(info.Path)
	assert.True(info.Path == "/user/:uid")

	assert.True(routeFound(tea, "GET", "/user/101/name/slene"))
	t.Log(info.Path)
	assert.True(info.Path == "/user/:uid/name/:name")
}

func Test_RouteWild(t *testing.T) {
	assert := &Assert{T: t}

	pathFunc := func(rw http.ResponseWriter, req *http.Request, i *RouteInfo) {
		path := i.Get("splat")
		rw.Write([]byte(req.Method + ":" + i.Path + ":" + path))
	}

	tea := New().Routers(
		Get(nopFunc),
		Router("/route/wild",
			Get(pathFunc),
			Router("/*:splat",
				Get(pathFunc),
			),
		),
		Router("/route/wild2",
			Router("/*:splat",
				Get(pathFunc),
			),
		),
		Router("/route/wild2",
			Router("/*:splat",
				Get(pathFunc),
			),
		),
		Router("/route/wild3",
			Router("/user",
				Get(pathFunc),
			),
			Router("/:uid",
				Get(pathFunc),
			),
			Router("/:uid/order",
				Get(pathFunc),
			),
			Get(nopFunc),
			Router("/*:splat",
				Post(pathFunc),
				Get(pathFunc),
			),
		),
	)

	assert.True(responseEqual(tea, "GET", "/route/wild", "GET:/route/wild:"))
	assert.True(responseEqual(tea, "GET", "/route/wild/", "GET:/route/wild:"))
	assert.True(responseEqual(tea, "GET", "/route/wild/1/2",
		"GET:/route/wild/*:splat:1/2",
	))
	assert.True(responseEqual(tea, "GET", "/route/wild2/1/2",
		"GET:/route/wild2/*:splat:1/2",
	))
	assert.True(responseEqual(tea, "POST", "/route/wild3/1/2",
		"POST:/route/wild3/*:splat:1/2",
	))
	assert.True(responseEqual(tea, "GET", "/route/wild3/user",
		"GET:/route/wild3/user:",
	))
	assert.True(responseEqual(tea, "GET", "/route/wild3/1",
		"GET:/route/wild3/:uid:",
	))
	assert.True(responseEqual(tea, "GET", "/route/wild3/1/order",
		"GET:/route/wild3/:uid/order:",
	))

	// wild route can not defined in middle of route path
	func() {
		defer func() {
			err := recover()
			assert.NotNil(err)
		}()

		New().Routers(
			Router("/route/wild/*:splat/name",
				Get(nopFunc),
			),
		)
	}()
}

func Test_ConflictParamRoute1(t *testing.T) {
	assert := &Assert{T: t}
	defer func() {
		err := recover()
		if err != nil {
			errStr, _ := err.(string)
			assert.True(strings.Contains(errStr, "conflict"))
			assert.True(strings.Contains(errStr, "`:panic` to `:uid`"))
		}
		assert.NotNil(err)
	}()
	New().Routers(
		Router("/:uid/name/:name", Get(nopFunc)),
		Router("/:panic/name", Get(nopFunc)),
	)
}

func Test_ConflictParamRoute2(t *testing.T) {
	assert := &Assert{T: t}
	defer func() {
		err := recover()
		if err != nil {
			errStr, _ := err.(string)
			assert.True(strings.Contains(errStr, "conflict"))
			assert.True(strings.Contains(errStr, "`:panic` to `:name`"))
		}
		assert.NotNil(err)
	}()
	New().Routers(
		Router("/:uid/name/:name", Get(nopFunc)),
		Router("/:uid/name/:panic/test", Get(nopFunc)),
	)
}

func routeFound(tea *Teapot, method, urlStr string) bool {
	req, _ := http.NewRequest(method, urlStr, nil)
	rec := httptest.NewRecorder()
	tea.ServeHTTP(rec, req)
	return rec.Code == http.StatusOK
}

func routeNotFound(tea *Teapot, method, urlStr string) bool {
	req, _ := http.NewRequest(method, urlStr, nil)
	rec := httptest.NewRecorder()
	tea.ServeHTTP(rec, req)
	return rec.Code == http.StatusNotFound
}

func responseEqual(tea *Teapot, method, urlStr string, resp string) bool {
	req, _ := http.NewRequest(method, urlStr, nil)
	rec := httptest.NewRecorder()
	tea.ServeHTTP(rec, req)
	return rec.Body.String() == resp
}

func justATeapot(tea *Teapot, method, urlStr string) bool {
	req, _ := http.NewRequest(method, urlStr, nil)
	rec := httptest.NewRecorder()
	tea.ServeHTTP(rec, req)
	return rec.Code == http.StatusTeapot
}