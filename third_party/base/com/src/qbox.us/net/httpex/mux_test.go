package httpex

import (
	"github.com/qiniu/log.v1"
	"github.com/qiniu/ts"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func init() {
	log.SetOutputLevel(0)
}

// ---------------------------------------------------------------------------

type Env struct {
}

type Service struct {
}

func (r *Service) Do_(w http.ResponseWriter, req *http.Request, env *Env) {
	io.WriteString(w, "Do_: "+req.URL.String())
}

func (r *Service) DoFoo_bar_(w http.ResponseWriter, req *http.Request, env *Env) {
	io.WriteString(w, "DoFoo_bar_: "+req.URL.String())
}

func (r *Service) DoPage(w http.ResponseWriter, req *http.Request, env *Env) {
	io.WriteString(w, "DoPage: "+req.URL.String())
}

func (r *Service) DoPageAction1(w http.ResponseWriter, req *http.Request, env *Env) {
	io.WriteString(w, "DoPageAction1: "+req.URL.String())
}

// ---------------------------------------------------------------------------

var unusedEnv *Env
var typeOfEnv = reflect.TypeOf(unusedEnv)

func TestRouter(t *testing.T) {

	service := new(Service)
	mux := NewServeMux()
	err := Register(mux, service, typeOfEnv)
	if err != nil {
		ts.Fatal(t, err)
	}
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		env := &Env{}
		mux.CallIntf(service, w, req, env)
	}))
	defer svr.Close()
	svrUrl := svr.URL

	var routeCases = [][2]string{
		{svrUrl + "/page?a=1&b=2", "DoPage: /page?a=1&b=2"},
		{svrUrl + "/page/action1?a=2&b=3", "DoPageAction1: /page/action1?a=2&b=3"},
		{svrUrl + "/abc?a=3", "Do_: /abc?a=3"},
		{svrUrl + "/foo-bar?c=3", "DoFoo_bar_: /foo-bar?c=3"},
		{svrUrl + "/foo-bar/?c=3", "DoFoo_bar_: /foo-bar/?c=3"},
	}

	var cookies []*http.Cookie
	var resp *http.Response
	for _, c := range routeCases {
		req, _ := http.NewRequest("GET", c[0], nil)
		for _, cookie := range cookies {
			req.AddCookie(cookie)
		}
		resp, err = http.DefaultClient.Do(req)
		cookies = checkResp(t, resp, err, c[1])
	}
}

func checkResp(t *testing.T, resp *http.Response, err error, respText string) (cookies []*http.Cookie) {

	if err != nil {
		ts.Fatal(t, err)
	}
	defer resp.Body.Close()

	text1, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		ts.Fatal(t, "ioutil.ReadAll failed:", err)
	}

	text := string(text1)
	if text != respText {
		ts.Fatal(t, "unexpected resp:", text, respText)
	}

	cookies = resp.Cookies()
	if len(cookies) != 0 {
		log.Info("Cookies:", cookies)
	}
	return
}

// ---------------------------------------------------------------------------
