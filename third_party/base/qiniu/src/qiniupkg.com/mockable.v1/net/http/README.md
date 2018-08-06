HTTP based on Mocking Network
===========

## 规格

```go
import (
	"net/http"
	"time"

	"qiniupkg.com/mocking/net"
)

type Transport http.Transport

func (p *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	if !net.Mocking {
		return ((*Transport)(p)).RoundTrip(req)
	}
	transport := *(*Transport)(p)
	dial := transport.Dial
	transport.Dial = func(network, address string) (net.Conn, error) {
		// 1) address转换; 2) 限速
	}
	return transport.RoundTrip(req)
}

var DefaultTransport http.RoundTripper = &http.Transport{
	Proxy: http.ProxyFromEnvironment,
	Dial: (&net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}).Dial,
	TLSHandshakeTimeout:   10 * time.Second,
	ExpectContinueTimeout: 1 * time.Second,
}

var DefaultClient = &http.Client{
	Transport: DefaultTransport,
}

type Server http.Server

func (p *Server) ListenAndServe() error {
	if !net.Mocking {
		return ((*http.Server)(p)).ListenAndServe()
	}
	// ...
}

func (p *Server) ListenAndServeTLS() error {
	if !net.Mocking {
		return ((*http.Server)(p)).ListenAndServeTLS()
	}
	// ...
}

var ListenAndServe = http.ListenAndServe
var ListenAndServeTLS = http.ListenAndServeTLS

// 负责：laddr(logicIP:logicPort)地址转换
// 当 logicIP = "0.0.0.0" 时，需要将 logicIP 改为 MockingIPs 后进行地址转换（需要Listen多个端口）
//
func MockListenAndServe(laddr string, handler http.Handler) error {}
func MockListenAndServeTLS(laddr, certFile, keyFile string, handler http.Handler) error {}

// 更换 DefaultTransport 以便支持address转换及限速
// 但是有的库有可能在 DefaultTransport 被替换前获取了 DefaultTransport，需要自行调整
// 例如：qiniupkg.com/x/rpc.v7，需要客户主动执行一次赋值：rpc.DefaultClient.Transport = DefaultTransport
//
func Init() {
	http.DefaultTransport = DefaultTransport
}
```
