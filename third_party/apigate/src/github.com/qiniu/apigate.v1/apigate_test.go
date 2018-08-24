package apigate

import (
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"testing"

	"github.com/qiniu/apigate.v1/proto"
	"github.com/qiniu/http/restrpc.v1"
	"github.com/qiniu/mockhttp.v2"
	"github.com/stretchr/testify.v1/assert"
)

// --------------------------------------------------------------------

type mockOs struct{}

func (p *mockOs) PostV1Auths_Rules(env *restrpc.Env) {
	io.WriteString(env.W, "os put auth rules")
}

func (p *mockOs) GetV1Auths_(env *restrpc.Env) {
	io.WriteString(env.W, "os get auth")
}

// --------------------------------------------------------------------

type mockOsFwd struct{}

func (p *mockOsFwd) PostFwdBV1Auths_Rules(env *restrpc.Env) {
	io.WriteString(env.W, "os put auth rules")
}

func (p *mockOsFwd) GetFwdBV1Auths_(env *restrpc.Env) {
	io.WriteString(env.W, "os get auth")
}

// --------------------------------------------------------------------

var (
	mds      = make([]string, 0)
	patterns = make([]string, 0)
)

func resetMdsPts() {
	mds = make([]string, 0)
	patterns = make([]string, 0)
}

type mockMetric struct {
}

func (m mockMetric) Register(moduleName, confStr string) error {

	return nil
}

func (m mockMetric) OpenRequest(ctx context.Context, w *http.ResponseWriter, req *http.Request) proto.RequestEvent {

	mds = append(mds, proto.ModFromContextSafe(ctx))
	patterns = append(patterns, proto.PatternFromContextSafe(ctx))
	return proto.NilReqEvent
}

// --------------------------------------------------------------------

type mockRs struct{}

func (p *mockRs) PostStat_(env *restrpc.Env) {
	io.WriteString(env.W, "rs stat")
}

func (p *mockRs) PostMkbucket_(env *restrpc.Env) {
	io.WriteString(env.W, "rs mkbucket")
}

func (p *mockRs) GetBucketinfo_(env *restrpc.Env) {
	io.WriteString(env.W, "rs bucketinfo")
}

func (p *mockRs) PostBucketinfo_(env *restrpc.Env) {
	io.WriteString(env.W, "rs bucketinfo")
}

// --------------------------------------------------------------------

type mockUp struct {
	t *testing.T
}

func (p *mockUp) PostUpload(env *restrpc.Env) {
	assert.Equal(p.t, "up.qiniu.com", env.Req.Host)
	io.WriteString(env.W, "upload")
}

func (p *mockUp) PostMkblk_(env *restrpc.Env) {
	assert.Equal(p.t, "up.qiniu.com", env.Req.Host)
	io.WriteString(env.W, "up mkblk")
}

// --------------------------------------------------------------------

func getText(t *testing.T, r io.Reader) string {

	b, err := ioutil.ReadAll(r)
	if err != nil {
		t.Fatal("ioutil.ReadAll failed:", err)
	}
	return string(b)
}

func runService(transport *mockhttp.Transport, host string, svr interface{}) {

	router := restrpc.Router{
		Mux: restrpc.NewServeMux(),
	}
	transport.ListenAndServe(host, router.Register(svr))
}

// --------------------------------------------------------------------

type mockAuthStuber uint64

func (p mockAuthStuber) AuthStub(req *http.Request) (ai AuthInfo, ok bool, err error) {

	su, _ := strconv.ParseBool(req.URL.Query().Get("su"))
	return AuthInfo{Utype: uint64(p), Su: su}, true, nil
}

var authStuber = mockAuthStuber(1)
var userAccess = AccessInfo{Allow: 1}
var adminAccess = AccessInfo{Allow: 0x80}
var publicAccess = AccessInfo{Allow: Access_Public}
var suUserAccess = AccessInfo{Allow: 1, SuOnly: true}
var suAdminAccess = AccessInfo{Allow: 0x80, SuOnly: true}

func runApigate(transport *mockhttp.Transport, mode int, t *testing.T) {

	var gate *Service

	switch mode {
	case 0:
		gate = New()
		metric := mockMetric{}

		rs := gate.Service("RS", 50, "rs.qiniuapi.com", "api.qiniu.com:8888/rs", ":6166").
			Forward("localhost:8888").Auths(authStuber)
		metric.Register("RS", "")
		rs.HandleNotFound()

		rs.Api("POST /stat/**", 50).Access(userAccess)
		rs.Api("POST /mkbucket/**", 50).Access(adminAccess)
		rs.Api("* /bucketinfo/**", 50).Access(publicAccess)

		up := gate.Service("UP", 50, "up.qiniuapi.com", "api.qiniu.com:8888/up").
			Forward("localhost:9999").ForwardHost("up.qiniu.com").Auths(authStuber)
		metric.Register("UP", "")
		up.HandleNotFound()

		up.Api("POST /mkblk/**", 50).Access(userAccess)
		up.Api("POST /**", 50).Access(userAccess).Forward("/upload")

		os := gate.Service("OS", 50, "os.qiniuapi.com", "api.qiniu.com:8888/os").
			Forward("localhost:8899").Auths(authStuber)
		metric.Register("OS", "")
		os.HandleNotFound()

		os.Api("POST /v1/auths/*/rules", 50).Access(userAccess)
		os.Api("GET /v1/auths/*", 50).Access(userAccess)

		osfwd := gate.Service("OSFWD", 50, "osfwd.qiniuapi.com", "api.qiniu.com:8888/osfwd").
			Forward("localhost:7777/fwd/b").Auths(authStuber)
		metric.Register("OSFWD", "")
		osfwd.HandleNotFound()

		osfwd.Api("POST /v1/auths/*/rules", 50).Access(userAccess)
		osfwd.Api("GET /v1/auths/*", 50).Access(userAccess)

		suOsfwd := gate.Service("SUOSFWD", 50, "su.osfwd.qiniuapi.com").
			Forward("localhost:7777/fwd/b").Auths(authStuber)
		metric.Register("SUOSFWD", "")
		suOsfwd.HandleNotFound()

		suOsfwd.Api("GET /v1/auths/*", 50).Access(suUserAccess)

		gate.Sink(metric)

	default:
		config := `
{
  "services": [
    {
      "module": "RS",
      "routes": ["rs.qiniuapi.com", "api.qiniu.com:8888/rs", ":6166"],
      "auths": ["mockauth"],
      "forward": "localhost:8888",
      "apis": [
        {
          "pattern": "POST /stat/**",
          "allow": "user"
        },
        {
          "pattern": "POST /mkbucket/**",
          "allow": "admin"
        },
        {
          "pattern": "* /bucketinfo/**",
          "allow": "public"
        }
      ]
    },
    {
      "module": "UP",
      "routes": ["up.qiniuapi.com", "api.qiniu.com:8888/up"],
      "auths": ["mockauth"],
      "forward": "localhost:9999",
      "forward_host": "up.qiniu.com",
      "apis": [
        {
          "pattern": "POST /mkblk/**",
          "allow": "user",
          "notallow": "expuser"
        },
        {
          "pattern": "POST /**",
          "forward": "/upload",
          "allow": "user"
        }
      ]
    },
    {
      "module": "OS",
      "routes": ["os.qiniuapi.com", "api.qiniu.com:8888/os"],
      "auths": ["mockauth"],
      "forward": "localhost:8899",
      "apis": [
        {
          "patterns": ["POST /v1/auths/*/rules",
                       "GET /v1/auths/*"],
          "allow": "user"
        }
      ]
    },
    {
      "module": "OSFWD",
      "routes": ["osfwd.qiniuapi.com", "api.qiniu.com:8888/osfwd"],
      "auths": ["mockauth"],
      "forward": "localhost:7777/fwd/b",
      "apis": [
        {
          "pattern": "POST /v1/auths/*/rules",
          "allow": "user"
	},
        {
          "pattern": "GET /v1/auths/*",
          "allow": "user"
        }
      ]
    },
    {
      "module": "SUOSFWD",
      "routes": ["su.osfwd.qiniuapi.com"],
      "auths": ["mockauth"],
      "forward": "localhost:7777/fwd/b",
      "apis": [
        {
          "pattern": "POST /v1/auths/*/rules",
          "allow": "user"
	},
        {
          "pattern": "GET /v1/auths/*",
          "allow": "user",
          "suonly": true
        }
      ]
    }
  ]
}
`
		var err error

		RegisterAccessInfo("user", userAccess.Allow)
		RegisterAccessInfo("expuser", 2)
		RegisterAccessInfo("admin", adminAccess.Allow)
		RegisterAuthStuber("mockauth", authStuber, authStuber)

		metric := mockMetric{}
		gate, err = NewFromString(config, metric)
		if err != nil {
			t.Fatal("apigate.NewFromString failed:", err)
		}
		resetMdsPts()
		gate.Sink(metric)
	}

	SetDefaultProxyTransport(transport)

	hosts := []string{
		"rs.qiniuapi.com",
		"rs2.qiniuapi.com:6166",
		"up.qiniuapi.com",
		"os.qiniuapi.com",
		"osfwd.qiniuapi.com",
		"su.osfwd.qiniuapi.com",
		"api.qiniu.com:8888",
	}
	for _, host := range hosts {
		transport.ListenAndServe(host, gate)
	}
}

func doTestApigate(t *testing.T, mode int) {

	transport := mockhttp.NewTransport()

	runService(transport, "localhost:8888", new(mockRs))
	runService(transport, "localhost:9999", &mockUp{t: t})
	runService(transport, "localhost:8899", new(mockOs))
	runService(transport, "localhost:7777", new(mockOsFwd))
	runApigate(transport, mode, t)

	cases := []struct {
		method string
		req    string
		resp   string
		code   int
	}{
		{
			method: "POST",
			req:    "http://api.qiniu.com:8888/rs/stat/123",
			resp:   "rs stat",
			code:   200,
		},
		{
			method: "POST",
			req:    "http://rs2.qiniuapi.com:6166/stat/123",
			resp:   "rs stat",
			code:   200,
		},
		{
			method: "GET",
			req:    "http://api.qiniu.com:8888/rs/stat/123",
			resp:   "",
			code:   404,
		},
		{
			method: "POST",
			req:    "http://rs.qiniuapi.com/mkbucket/123",
			resp:   `{"error":"unauthorized"}`,
			code:   401,
		},
		{
			method: "POST",
			req:    "http://rs.qiniuapi.com/bucketinfo/123",
			resp:   "rs bucketinfo",
			code:   200,
		},
		{
			method: "GET",
			req:    "http://rs.qiniuapi.com/bucketinfo/123",
			resp:   "rs bucketinfo",
			code:   200,
		},
		{
			method: "POST",
			req:    "http://api.qiniu.com:8888/up/mkblk/123",
			resp:   "up mkblk",
			code:   200,
		},
		{
			method: "POST",
			req:    "http://api.qiniu.com:8888/up/mkblk",
			resp:   "upload",
			code:   200,
		},
		{
			method: "GET",
			req:    "http://api.qiniu.com:8888/os/v1/auths/123",
			resp:   "os get auth",
			code:   200,
		},
		{
			method: "POST",
			req:    "http://os.qiniuapi.com/v1/auths/123/rules",
			resp:   "os put auth rules",
			code:   200,
		},
		{
			method: "GET",
			req:    "http://api.qiniu.com:8888/osfwd/v1/auths/123",
			resp:   "os get auth",
			code:   200,
		},
		{
			method: "POST",
			req:    "http://osfwd.qiniuapi.com/v1/auths/123/rules",
			resp:   "os put auth rules",
			code:   200,
		},
		{
			method: "GET",
			req:    "http://su.osfwd.qiniuapi.com/v1/auths/123",
			resp:   "{\"error\":\"unauthorized\"}",
			code:   401,
		},
		{
			method: "GET",
			req:    "http://su.osfwd.qiniuapi.com/v1/auths/123?su=true",
			resp:   "os get auth",
			code:   200,
		},
	}

	client := &http.Client{
		Transport: transport,
	}
	for i, c := range cases {
		req, _ := http.NewRequest(c.method, c.req, nil)
		w, err := client.Do(req)
		if err != nil {
			t.Fatal("Request failed:", i, err)
		}
		if w.StatusCode != c.code || (c.resp != "" && getText(t, w.Body) != c.resp) {
			t.Fatal("TestApigate failed:", i, c, w)
		}
		w.Body.Close()
	}
	assert.Equal(t, []string{"RS", "RS", "RS", "RS", "RS", "RS", "UP", "UP", "OS", "OS", "OSFWD", "OSFWD", "SUOSFWD", "SUOSFWD"}, mds)
	assert.Equal(t, []string{"POST /stat/**", "POST /stat/**", "404", "POST /mkbucket/**", "* /bucketinfo/**", "* /bucketinfo/**", "POST /mkblk/**", "POST /**", "GET /v1/auths/*", "POST /v1/auths/*/rules", "GET /v1/auths/*", "POST /v1/auths/*/rules", "GET /v1/auths/*", "GET /v1/auths/*"}, patterns)
}

func TestApigate(t *testing.T) {

	doTestApigate(t, 0)
	doTestApigate(t, 1)
}

func TestAuthStub(t *testing.T) {

	RegisterAuthStuber("mockauth", authStuber, authStuber)

	stubers, err := GetAuthStubers([]string{"mockauth"}, true)
	if err != nil {
		t.Fatal("GetAuthStubers failed:", err)
	}
	if len(stubers) != 1 || stubers[0] != authStuber {
		t.Fatal("GetAuthStubers unexpected")
	}
}

func TestAccessInfo(t *testing.T) {

	RegisterAccessInfo("user", userAccess.Allow)
	RegisterAccessInfo("admin", adminAccess.Allow)

	ai, err := ParseAccessInfo("user", "", false)
	if err != nil || ai != userAccess {
		t.Fatal("ParseAccessInfo(user) failed:", ai, err)
	}

	ai, err = ParseAccessInfo("admin", "", false)
	if err != nil || ai != adminAccess {
		t.Fatal("ParseAccessInfo(admin) failed:", ai, err)
	}

	ai, err = ParseAccessInfo(" user | admin ", "", false)
	if err != nil || ai != (AccessInfo{0x81, 0, false}) {
		t.Fatal("ParseAccessInfo(user|admin) failed:", ai, err)
	}

	ai, err = ParseAccessInfo("user", "", true)
	if err != nil || ai != suUserAccess {
		t.Fatal("ParseAccessInfo(user) failed:", ai, err)
	}

	ai, err = ParseAccessInfo("admin", "", true)
	if err != nil || ai != suAdminAccess {
		t.Fatal("ParseAccessInfo(admin) failed:", ai, err)
	}

	ai, err = ParseAccessInfo(" user | admin ", "", true)
	if err != nil || ai != (AccessInfo{0x81, 0, true}) {
		t.Fatal("ParseAccessInfo(user|admin) failed:", ai, err)
	}

}

// --------------------------------------------------------------------
