package qauthgate

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"qbox.us/mgo2"
	"qbox.us/mockacc"
	"qbox.us/qauthgate/agmapi"

	"github.com/qiniu/api/auth/digest"
	"github.com/qiniu/errors"
	"github.com/qiniu/http/httputil.v1"
	"github.com/qiniu/http/webroute.v1"
	"github.com/qiniu/http/wsrpc.v1"
	"github.com/qiniu/log.v1"
	"github.com/qiniu/mockhttp.v1"
	"github.com/qiniu/rpc.v1"
)

var sa = mockacc.GetSa()
var user = sa[0]

// ------------------------------------------------------------------------

type aService struct {
	Host   string `json:"host"`
	Server string `json:"server"`
}

type retService struct {
	Host     string `json:"host"`
	Server   string `json:"server"`
	RealHost string `json:"real"`
	Path     string `json:"path"`
}

func (p *aService) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.URL.Path == "/slow" {
		time.Sleep(5e9)
	}
	httputil.Reply(w, 200, &retService{p.Host, p.Server, req.Host, req.URL.Path})
}

// ------------------------------------------------------------------------

func init() {
	http.DefaultClient = &http.Client{Transport: mockhttp.Transport}
	http.DefaultTransport = mockhttp.Transport
	rpc.DefaultClient = mockhttp.Client
	log.SetOutputLevel(0)
}

func runGate(t *testing.T) {

	mockhttp.Bind("localhost:1234", &aService{"a.qiniu.com", "localhost:1234"})
	mockhttp.Bind("localhost:1235", &aService{"a.qiniu.com", "localhost:1235"})

	mockhttp.Bind("localhost:2234", &aService{"b.qiniu.com", "localhost:2234"})
	mockhttp.Bind("localhost:2235", &aService{"b.qiniu.com", "localhost:2235"})

	c := mgo2.Open(&mgo2.Config{
		Host: "localhost",
		DB:   "qbox_authgate_test",
		Coll: "authgate",
		Mode: "strong",
	})
	defer c.Close()

	coll := c.Coll
	coll.RemoveAll(all)

	coll.Insert(&entry{
		Host:  "a.qiniu.com",
		Alias: []string{"a.qbox.me"},
		Items: []*routeEntry{
			&routeEntry{Server: "localhost:1234"},
			&routeEntry{Server: "localhost:1235"},
		},
	}, &entry{
		Host: "b.qiniu.com",
		Items: []*routeEntry{
			&routeEntry{Server: "localhost:2234"},
			&routeEntry{Server: "localhost:2235"},
		},
	})

	conf := &Config{
		Coll:       coll,
		AuthParser: mockacc.NewParser(sa),
	}

	service, err := New(conf)
	if err != nil {
		t.Fatal("qconfm.New failed:", errors.Detail(err))
	}

	mux := http.NewServeMux()
	router := &webroute.Router{Factory: wsrpc.Factory, Mux: mux}
	mockhttp.Bind("agm.qiniu.com", router.Register(service))
	mockhttp.Bind("a.qiniu.com", service)
	mockhttp.Bind("b.qiniu.com", service)
	mockhttp.Bind("a.qbox.me", service)
	mockhttp.Bind("a.qbox.us", service)

	err = rpc.DefaultClient.Call(nil, nil, "http://a.qiniu.com/")
	if err == nil {
		t.Fatal("DefaultClient.Call: ok?")
	}
	e, ok := err.(*rpc.ErrorInfo)
	if !ok {
		t.Fatal("DefaultClient.Call: NOT rpc.ErrorInfo -", err)
	}
	if e.Code != 401 || e.Err != "bad token" {
		t.Fatal("DefaultClient.Call failed:", *e)
	}

	mac := &digest.Mac{
		AccessKey: user.AccessKey,
		SecretKey: []byte(user.SecretKey),
	}
	conn := rpc.Client{digest.NewClient(mac, nil)}

	var ret retService
	err = conn.Call(nil, &ret, "http://a.qiniu.com/")
	if err != nil {
		t.Fatal("conn.Call failed:", err)
	}
	fmt.Println(ret)
	if ret.Host != "a.qiniu.com" || ret.Server != "localhost:1234" || ret.RealHost != "a.qiniu.com" {
		t.Fatal("conn.Call failed:", ret, err)
	}

	err = conn.Call(nil, &ret, "http://a.qbox.me/hello")
	if err != nil {
		t.Fatal("conn.Call failed:", err)
	}
	fmt.Println(ret)
	if ret.Host != "a.qiniu.com" || ret.Server != "localhost:1234" || ret.RealHost != "a.qbox.me" {
		t.Fatal("conn.Call failed:", ret, err)
	}

	err = agmapi.Enable("http://agm.qiniu.com", nil, "a.qiniu.com", ret.Server, 0)
	if err != nil {
		t.Fatal("agmapi.Enable failed:", err)
	}

	err = conn.Call(nil, &ret, "http://a.qiniu.com/abc")
	if err != nil {
		t.Fatal("conn.Call failed:", err)
	}
	fmt.Println(ret)
	if ret.Host != "a.qiniu.com" || ret.Server != "localhost:1235" {
		t.Fatal("conn.Call failed:", ret, err)
	}

	info, err := agmapi.Query("http://agm.qiniu.com", nil, "a.qiniu.com", ret.Server)
	if err != nil || info.Active != 0 {
		t.Fatal("agmapi.Query failed:", info, err)
	}
	fmt.Println("agmapi.Query:", info)

	go conn.Call(nil, nil, "http://a.qiniu.com/slow")
	time.Sleep(3e8)
	info, err = agmapi.Query("http://agm.qiniu.com", nil, "a.qiniu.com", ret.Server)
	if err != nil || info.Active != 1 {
		t.Fatal("agmapi.Query failed:", info, err)
	}
	fmt.Println("agmapi.Query:", info)

	err = conn.Call(nil, &ret, "http://b.qiniu.com/world")
	if err != nil {
		t.Fatal("conn.Call failed:", err)
	}
	fmt.Println(ret)
	if ret.Host != "b.qiniu.com" || ret.Server != "localhost:2234" {
		t.Fatal("conn.Call failed:", ret, err)
	}

	err = coll.UpdateId("a.qiniu.com", &entry{
		Host:  "a.qiniu.com",
		Alias: []string{"a.qbox.us"},
		Items: []*routeEntry{
			&routeEntry{Server: "localhost:1235"},
		},
	})
	if err != nil {
		t.Fatal("coll.Update failed:", err)
	}

	err = agmapi.Reload("http://agm.qiniu.com", nil, "a.qiniu.com")
	if err != nil {
		t.Fatal("agmapi.Reload failed:", err)
	}

	err = conn.Call(nil, &ret, "http://a.qiniu.com/abc")
	if err != nil {
		t.Fatal("conn.Call failed:", err)
	}
	fmt.Println(ret)
	if ret.Host != "a.qiniu.com" || ret.Server != "localhost:1235" {
		t.Fatal("conn.Call failed:", ret, err)
	}

	err = conn.Call(nil, &ret, "http://a.qbox.us/hello")
	if err != nil {
		t.Fatal("conn.Call failed:", err)
	}
	fmt.Println(ret)
	if ret.Host != "a.qiniu.com" || ret.Server != "localhost:1235" || ret.RealHost != "a.qbox.us" {
		t.Fatal("conn.Call failed:", ret, err)
	}

	err = conn.Call(nil, &ret, "http://a.qbox.me/hello")
	if err == nil {
		t.Fatal("conn.Call: ok?")
	}
	fmt.Println(err)
	e, ok = err.(*rpc.ErrorInfo)
	if !ok {
		t.Fatal("conn.Call: NOT rpc.ErrorInfo - ", err)
	}
	if e.Code != 612 || e.Err != "service not found" {
		t.Fatal("conn.Call: ", err)
	}
}

func Test(t *testing.T) {

	runGate(t)
}

// ------------------------------------------------------------------------
