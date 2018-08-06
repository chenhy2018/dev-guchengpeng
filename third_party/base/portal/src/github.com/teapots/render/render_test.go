package render

import (
	"encoding/xml"
	"html/template"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"

	"github.com/teapots/teapot"
)

type Greeting struct {
	One string `json:"one"`
	Two string `json:"two"`
}

type GreetingXML struct {
	XMLName xml.Name `xml:"greeting"`
	One     string   `xml:"one,attr"`
	Two     string   `xml:"two,attr"`
}

func Test_Render_JSON(t *testing.T) {
	tea := teapot.New()
	tea.Provide(Renderer(Options{
	// nothing here to configure
	}))

	// routing
	tea.Routers(
		teapot.Router("/foobar",
			teapot.Get(func(r Render) {
				r.JSON(Greeting{"hello", "world"}, 300)
			}),
		),
	)

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foobar", nil)

	tea.ServeHTTP(res, req)

	expect(t, res.Code, 300)
	expect(t, res.Header().Get(ContentType), ContentJSON+"; charset=UTF-8")
	expect(t, res.Body.String(), `{"one":"hello","two":"world"}`)
}

func Test_Render_JSON_Prefix(t *testing.T) {
	tea := teapot.New()
	prefix := ")]}',\n"
	tea.Provide(Renderer(Options{
		PrefixJSON: []byte(prefix),
	}))

	// routing
	tea.Routers(
		teapot.Router("/foobar",
			teapot.Get(func(r Render) {
				r.JSON(Greeting{"hello", "world"}, 300)
			}),
		),
	)

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foobar", nil)

	tea.ServeHTTP(res, req)

	expect(t, res.Code, 300)
	expect(t, res.Header().Get(ContentType), ContentJSON+"; charset=UTF-8")
	expect(t, res.Body.String(), prefix+`{"one":"hello","two":"world"}`)
}

func Test_Render_Indented_JSON(t *testing.T) {
	tea := teapot.New()
	tea.Provide(Renderer(Options{
		IndentJSON: true,
	}))

	// routing
	tea.Routers(
		teapot.Router("/foobar",
			teapot.Get(func(r Render) {
				r.JSON(Greeting{"hello", "world"}, 300)
			}),
		),
	)

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foobar", nil)

	tea.ServeHTTP(res, req)

	expect(t, res.Code, 300)
	expect(t, res.Header().Get(ContentType), ContentJSON+"; charset=UTF-8")
	expect(t, res.Body.String(), `{
  "one": "hello",
  "two": "world"
}`)
}

func Test_Render_XML(t *testing.T) {
	tea := teapot.New()
	tea.Provide(Renderer(Options{
	// nothing here to configure
	}))

	// routing
	tea.Routers(
		teapot.Router("/foobar",
			teapot.Get(func(r Render) {
				r.XML(GreetingXML{One: "hello", Two: "world"}, 300)
			}),
		),
	)

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foobar", nil)

	tea.ServeHTTP(res, req)

	expect(t, res.Code, 300)
	expect(t, res.Header().Get(ContentType), ContentXML+"; charset=UTF-8")
	expect(t, res.Body.String(), `<greeting one="hello" two="world"></greeting>`)
}

func Test_Render_XML_Prefix(t *testing.T) {
	tea := teapot.New()
	prefix := ")]}',\n"
	tea.Provide(Renderer(Options{
		PrefixXML: []byte(prefix),
	}))

	// routing
	tea.Routers(
		teapot.Router("/foobar",
			teapot.Get(func(r Render) {
				r.XML(GreetingXML{One: "hello", Two: "world"}, 300)
			}),
		),
	)

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foobar", nil)

	tea.ServeHTTP(res, req)

	expect(t, res.Code, 300)
	expect(t, res.Header().Get(ContentType), ContentXML+"; charset=UTF-8")
	expect(t, res.Body.String(), prefix+`<greeting one="hello" two="world"></greeting>`)
}

func Test_Render_Indented_XML(t *testing.T) {
	tea := teapot.New()
	tea.Provide(Renderer(Options{
		IndentXML: true,
	}))

	// routing
	tea.Routers(
		teapot.Router("/foobar",
			teapot.Get(func(r Render) {
				r.XML(GreetingXML{One: "hello", Two: "world"}, 300)
			}),
		),
	)

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foobar", nil)

	tea.ServeHTTP(res, req)

	expect(t, res.Code, 300)
	expect(t, res.Header().Get(ContentType), ContentXML+"; charset=UTF-8")
	expect(t, res.Body.String(), `<greeting one="hello" two="world"></greeting>`)
}

func Test_Render_Bad_HTML(t *testing.T) {
	tea := teapot.New()
	tea.Provide(Renderer(Options{
		Directory: "testdata/basic",
	}))

	var err error
	// routing
	tea.Routers(
		teapot.Router("/foobar",
			teapot.Get(func(r Render) {
				err = r.HTML("nope", nil)
			}),
		),
	)

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foobar", nil)

	tea.ServeHTTP(res, req)

	expect(t, res.Code, 200)
	expect(t, err != nil, true)
	expect(t, err.Error(), "html/template: \"nope\" is undefined")
}

func Test_Render_HTML(t *testing.T) {
	tea := teapot.New()
	tea.Provide(Renderer(Options{
		Directory: "testdata/basic",
	}))

	// routing
	tea.Routers(
		teapot.Router("/foobar",
			teapot.Get(func(r Render) {
				r.HTML("hello", "jeremy")
			}),
		),
	)

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foobar", nil)

	tea.ServeHTTP(res, req)

	expect(t, res.Code, 200)
	expect(t, res.Header().Get(ContentType), ContentHTML+"; charset=UTF-8")
	expect(t, res.Body.String(), "<h1>Hello jeremy</h1>\n")
}

func Test_Render_XHTML(t *testing.T) {
	tea := teapot.New()
	tea.Provide(Renderer(Options{
		Directory:       "testdata/basic",
		HTMLContentType: ContentXHTML,
	}))

	tea.Routers(
		teapot.Router("/foobar",
			teapot.Get(func(r Render) {
				r.HTML("hello", "jeremy")
			}),
		),
	)

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foobar", nil)

	tea.ServeHTTP(res, req)

	expect(t, res.Code, 200)
	expect(t, res.Header().Get(ContentType), ContentXHTML+"; charset=UTF-8")
	expect(t, res.Body.String(), "<h1>Hello jeremy</h1>\n")
}

func Test_Render_Extensions(t *testing.T) {
	tea := teapot.New()
	tea.Provide(Renderer(Options{
		Directory:  "testdata/basic",
		Extensions: []string{".tmpl", ".html"},
	}))

	// routing
	tea.Routers(
		teapot.Router("/foobar",
			teapot.Get(func(r Render) {
				r.HTML("hypertext", nil)
			}),
		),
	)

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foobar", nil)

	tea.ServeHTTP(res, req)

	expect(t, res.Code, 200)
	expect(t, res.Header().Get(ContentType), ContentHTML+"; charset=UTF-8")
	expect(t, res.Body.String(), "Hypertext!\n")
}

func Test_Render_Funcs(t *testing.T) {

	tea := teapot.New()
	tea.Provide(Renderer(Options{
		Directory: "testdata/custom_funcs",
		Funcs: []template.FuncMap{
			{
				"myCustomFunc": func() string {
					return "My custom function"
				},
			},
		},
	}))

	// routing
	tea.Routers(
		teapot.Router("/foobar",
			teapot.Get(func(r Render) {
				r.HTML("index", "jeremy")
			}),
		),
	)

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foobar", nil)

	tea.ServeHTTP(res, req)

	expect(t, res.Body.String(), "My custom function\n")
}

func Test_Render_Layout(t *testing.T) {
	tea := teapot.New()
	tea.Provide(Renderer(Options{
		Directory: "testdata/basic",
		Layout:    "layout",
	}))

	// routing
	tea.Routers(
		teapot.Router("/foobar",
			teapot.Get(func(r Render) {
				r.HTML("content", "jeremy")
			}),
		),
	)

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foobar", nil)

	tea.ServeHTTP(res, req)

	expect(t, res.Body.String(), "head\n<h1>jeremy</h1>\n\nfoot\n")
}

func Test_Render_Layout_Current(t *testing.T) {
	tea := teapot.New()
	tea.Provide(Renderer(Options{
		Directory: "testdata/basic",
		Layout:    "current_layout",
	}))

	// routing
	tea.Routers(
		teapot.Router("/foobar",
			teapot.Get(func(r Render) {
				r.HTML("content", "jeremy")
			}),
		),
	)

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foobar", nil)

	tea.ServeHTTP(res, req)

	expect(t, res.Body.String(), "content head\n<h1>jeremy</h1>\n\ncontent foot\n")
}

func Test_Render_Nested_HTML(t *testing.T) {
	tea := teapot.New()
	tea.Provide(Renderer(Options{
		Directory: "testdata/basic",
	}))

	// routing
	tea.Routers(
		teapot.Router("/foobar",
			teapot.Get(func(r Render) {
				r.HTML("admin/index", "jeremy")
			}),
		),
	)

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foobar", nil)

	tea.ServeHTTP(res, req)

	expect(t, res.Code, 200)
	expect(t, res.Header().Get(ContentType), ContentHTML+"; charset=UTF-8")
	expect(t, res.Body.String(), "<h1>Admin jeremy</h1>\n")
}

func Test_Render_Delimiters(t *testing.T) {
	tea := teapot.New()
	tea.Provide(Renderer(Options{
		Delims:    Delims{"{[{", "}]}"},
		Directory: "testdata/basic",
	}))

	// routing
	tea.Routers(
		teapot.Router("/foobar",
			teapot.Get(func(r Render) {
				r.HTML("delims", "jeremy")
			}),
		),
	)

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foobar", nil)

	tea.ServeHTTP(res, req)

	expect(t, res.Code, 200)
	expect(t, res.Header().Get(ContentType), ContentHTML+"; charset=UTF-8")
	expect(t, res.Body.String(), "<h1>Hello jeremy</h1>")
}

func Test_Render_BinaryData(t *testing.T) {
	tea := teapot.New()
	tea.Provide(Renderer(Options{
	// nothing here to configure
	}))

	// routing
	tea.Routers(
		teapot.Router("/foobar",
			teapot.Get(func(r Render) {
				r.Data([]byte("hello there"))
			}),
		),
	)

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foobar", nil)

	tea.ServeHTTP(res, req)

	expect(t, res.Code, 200)
	expect(t, res.Header().Get(ContentType), ContentBinary)
	expect(t, res.Body.String(), "hello there")
}

func Test_Render_BinaryData_CustomMimeType(t *testing.T) {
	tea := teapot.New()
	tea.Provide(Renderer(Options{
	// nothing here to configure
	}))

	// routing
	tea.Routers(
		teapot.Router("/foobar",
			teapot.Get(func(r Render) {
				r.Header().Set(ContentType, "image/jpeg")
				r.Data([]byte("..jpeg data.."))
			}),
		),
	)

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foobar", nil)

	tea.ServeHTTP(res, req)

	expect(t, res.Code, 200)
	expect(t, res.Header().Get(ContentType), "image/jpeg")
	expect(t, res.Body.String(), "..jpeg data..")
}

func Test_Render_Status204(t *testing.T) {
	res := httptest.NewRecorder()
	r := renderer{res, nil, Options{}, "", false, nil, nil, nil}
	r.Status(204)
	expect(t, res.Code, 204)
}

func Test_Render_Error404(t *testing.T) {
	res := httptest.NewRecorder()
	r := renderer{res, nil, Options{}, "", false, nil, nil, nil}
	r.Status(404)
	expect(t, res.Code, 404)
}

func Test_Render_Error500(t *testing.T) {
	res := httptest.NewRecorder()
	r := renderer{res, nil, Options{}, "", false, nil, nil, nil}
	r.Status(500)
	expect(t, res.Code, 500)
}

func Test_Render_Redirect_Default(t *testing.T) {
	url, _ := url.Parse("http://localhost/path/one")
	req := http.Request{
		Method: "GET",
		URL:    url,
	}
	res := httptest.NewRecorder()

	r := renderer{res, &req, Options{}, "", false, nil, nil, nil}
	r.Redirect("two")

	expect(t, res.Code, 302)
	expect(t, res.HeaderMap["Location"][0], "/path/two")
}

func Test_Render_Redirect_Code(t *testing.T) {
	url, _ := url.Parse("http://localhost/path/one")
	req := http.Request{
		Method: "GET",
		URL:    url,
	}
	res := httptest.NewRecorder()

	r := renderer{res, &req, Options{}, "", false, nil, nil, nil}
	r.Redirect("two", 307)

	expect(t, res.Code, 307)
	expect(t, res.HeaderMap["Location"][0], "/path/two")
}

func Test_Render_Charset_JSON(t *testing.T) {
	tea := teapot.New()
	tea.Provide(Renderer(Options{
		Charset: "foobar",
	}))

	// routing
	tea.Routers(
		teapot.Router("/foobar",
			teapot.Get(func(r Render) {
				r.JSON(Greeting{"hello", "world"}, 300)
			}),
		),
	)

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foobar", nil)

	tea.ServeHTTP(res, req)

	expect(t, res.Code, 300)
	expect(t, res.Header().Get(ContentType), ContentJSON+"; charset=foobar")
	expect(t, res.Body.String(), `{"one":"hello","two":"world"}`)
}

func Test_Render_Default_Charset_HTML(t *testing.T) {
	tea := teapot.New()
	tea.Provide(Renderer(Options{
		Directory: "testdata/basic",
	}))

	// routing
	tea.Routers(
		teapot.Router("/foobar",
			teapot.Get(func(r Render) {
				r.HTML("hello", "jeremy")
			}),
		),
	)

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foobar", nil)

	tea.ServeHTTP(res, req)

	expect(t, res.Code, 200)
	expect(t, res.Header().Get(ContentType), ContentHTML+"; charset=UTF-8")
	// ContentLength should be deferred to the ResponseWriter and not Render
	expect(t, res.Header().Get(ContentLength), "")
	expect(t, res.Body.String(), "<h1>Hello jeremy</h1>\n")
}

func Test_Render_Override_Layout(t *testing.T) {
	tea := teapot.New()
	tea.Provide(Renderer(Options{
		Directory: "testdata/basic",
		Layout:    "layout",
	}))

	// routing
	tea.Routers(
		teapot.Router("/foobar",
			teapot.Get(func(r Render) {
				r.HTML("content", "jeremy", HTMLOptions{
					Layout: "another_layout",
				})
			}),
		),
	)

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/foobar", nil)

	tea.ServeHTTP(res, req)

	expect(t, res.Code, 200)
	expect(t, res.Header().Get(ContentType), ContentHTML+"; charset=UTF-8")
	expect(t, res.Body.String(), "another head\n<h1>jeremy</h1>\n\nanother foot\n")
}

func Test_Render_NoRace(t *testing.T) {
	// This test used to fail if run with -race
	tea := teapot.New()
	tea.Provide(Renderer(Options{
		Directory: "testdata/basic",
	}))

	// routing
	tea.Routers(
		teapot.Router("/foobar",
			teapot.Get(func(r Render) {
				r.HTML("hello", "world")
			}),
		),
	)

	done := make(chan bool)
	doreq := func() {
		res := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/foobar", nil)

		tea.ServeHTTP(res, req)

		expect(t, res.Code, 200)
		expect(t, res.Header().Get(ContentType), ContentHTML+"; charset=UTF-8")
		// ContentLength should be deferred to the ResponseWriter and not Render
		expect(t, res.Header().Get(ContentLength), "")
		expect(t, res.Body.String(), "<h1>Hello world</h1>\n")
		done <- true
	}
	// Run two requests to check there is no race condition
	go doreq()
	go doreq()
	<-done
	<-done
}

func Test_GetExt(t *testing.T) {
	expect(t, getExt("test"), "")
	expect(t, getExt("test.tmpl"), ".tmpl")
	expect(t, getExt("test.go.html"), ".go.html")
}

/* Test Helpers */
func expect(t *testing.T, a interface{}, b interface{}) {
	if a != b {
		t.Errorf("Expected %v (type %v) - Got %v (type %v)", b, reflect.TypeOf(b), a, reflect.TypeOf(a))
	}
}

func refute(t *testing.T, a interface{}, b interface{}) {
	if a == b {
		t.Errorf("Did not expect %v (type %v) - Got %v (type %v)", b, reflect.TypeOf(b), a, reflect.TypeOf(a))
	}
}
