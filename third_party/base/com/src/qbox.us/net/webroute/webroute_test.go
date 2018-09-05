package webroute

import (
	"io"
	"io/ioutil"
	"net/http"
	"github.com/qiniu/log.v1"
	"qbox.us/net/seshcookie"
	"github.com/qiniu/ts"
	"github.com/qiniu/xlog.v1"
	"testing"
	"time"
)

func init() {
	log.SetOutputLevel(0)
}

// ---------------------------------------------------------------------------

type Service struct {
}

func (r *Service) Do_(w http.ResponseWriter, req *http.Request) {
	log := xlog.New(w, req)
	log.Info("xlog:", req.URL.Path)
	log.Xlog("xlog:", req.URL.Path)
	io.WriteString(w, "Do_: "+req.URL.String())
}

func (r *Service) DoFoo_bar(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "DoFoo_bar: "+req.URL.String())
}

func (r *Service) DoFoo_bar_(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "DoFoo_bar_: "+req.URL.String())
}

func (r *Service) DoPage(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "DoPage: "+req.URL.String())
}

func (r *Service) DoPageAction1(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "DoPageAction1: "+req.URL.String())
}

// ---------------------------------------------------------------------------

func (r *Service) DoLogin(w http.ResponseWriter, req *http.Request, sessions map[string]interface{}) {
	sessions["userId"] = "<UserId>"
	io.WriteString(w, "DoLogin: "+req.URL.String())
}

func (r *Service) DoSth(w http.ResponseWriter, req *http.Request, sessions map[string]interface{}) {
	userId, _ := sessions["userId"].(string)
	io.WriteString(w, "DoSth: "+userId)
}

// ---------------------------------------------------------------------------

var routeCases = [][2]string{
	{"http://localhost:2357/page?a=1&b=2", "DoPage: /page?a=1&b=2"},
	{"http://localhost:2357/page/action1?a=2&b=3", "DoPageAction1: /page/action1?a=2&b=3"},
	{"http://localhost:2357/abc?a=3", "Do_: /abc?a=3"},
	{"http://localhost:2357/foo-bar?c=3", "DoFoo_bar: /foo-bar?c=3"},
	{"http://localhost:2357/foo-bar/?c=3", "DoFoo_bar_: /foo-bar/?c=3"},
	{"http://localhost:2357/login?user=123", "DoLogin: /login?user=123"},
	{"http://localhost:2357/sth?abc=3", "DoSth: <UserId>"},
}

func TestRoute(t *testing.T) {

	go func() {
		service := new(Service)
		sessions := seshcookie.NewSessionManager("_test_session", "", "fweaf5e9aef9a3c6da5dwcb1e2b601f3251a536d")
		ts.Fatal(t, ListenAndServe(":2357", service, sessions))
	}()
	time.Sleep(1e9)

	var err error
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

	if details, ok := resp.Header["X-Log"]; ok {
		for i, detail := range details {
			log.Info("Detail:", i, detail)
		}
	}

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
