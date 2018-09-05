package proxy

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"

	"github.com/qiniu/apigate.v1"
	qhttputil "github.com/qiniu/http/httputil.v1"
	"github.com/qiniu/xlog.v1"

	"qiniu.com/auth/account.v1"
)

type protoSwitchTransport struct {
	tp    http.RoundTripper
	acc   account.Account
	proxy *httputil.ReverseProxy
}

func NewProtoSwitchTransport(tp http.RoundTripper, acc account.Account) *protoSwitchTransport {
	proxy := &httputil.ReverseProxy{
		Director:  nilDirector,
		Transport: tp,
	}
	return &protoSwitchTransport{tp: tp, acc: acc, proxy: proxy}
}

func InitProtoSwitchProxy(tp http.RoundTripper, acc account.Account) {

	proxy := NewProtoSwitchTransport(tp, acc)

	apigate.RegisterProxy("qbox/proto-switch", proxy)
}

func (p *protoSwitchTransport) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	xl := xlog.NewWithReq(req)

	if !isHttpUpgrade(req.Header) {
		p.proxy.ServeHTTP(w, req)
		return
	}

	// TODO not support tls
	dial, err := net.DialTimeout("tcp", req.URL.Host, 3*time.Second)
	if err != nil {
		xl.Error("net.DiaTimeout, err:", err)
		qhttputil.Error(w, qhttputil.NewError(http.StatusGatewayTimeout, http.StatusText(http.StatusGatewayTimeout)))
		return
	}
	defer dial.Close()

	input, output, err := hijackServer(w)
	if err != nil {
		xl.Error("hijackServer:", err)
		qhttputil.Error(w, qhttputil.NewError(500, http.StatusText(500)))
		return
	}
	defer input.Close()

	err = req.Write(dial)
	if err != nil {
		xl.Error("try send request:", err)
		writeErrorResponse(output)
		return
	}

	buf := bufio.NewReader(dial)
	resp, err := http.ReadResponse(buf, req)
	if err != nil {
		xl.Error("try receive response:", err)
		writeErrorResponse(output)
		return
	}

	errch := make(chan error, 2)

	err = resp.Write(output)
	if err != nil {
		xl.Error("try write response:", err)
		return
	}

	if resp.StatusCode != http.StatusSwitchingProtocols || !isHttpUpgrade(resp.Header) {
		return
	}

	cp := func(dst io.Writer, src io.Reader) {
		_, err := io.Copy(dst, src)
		errch <- err
	}
	go cp(dial, input)
	go cp(output, dial)

	err = <-errch
	if err != nil {
		xl.Warn("proxy copy data:", err)
	}
	return
}

func writeErrorResponse(w io.Writer) {
	w.Write([]byte("HTTP/1.1 500 Internal Server Error\r\nConnection: close\r\nContent-Type: application/json\r\n\r\n{\"error\":\"Internal Server Error\"}"))
}

func isHttpUpgrade(headers http.Header) bool {
	return strings.ToLower(headers.Get("Connection")) == "upgrade" && headers.Get("Upgrade") != ""
}

func hijackServer(w http.ResponseWriter) (io.ReadCloser, io.Writer, error) {
	hijacker, ok := qhttputil.GetHijacker(w)
	if !ok {
		return nil, nil, fmt.Errorf("%T is not a Hijacker", w)
	}
	conn, _, err := hijacker.Hijack()
	if err != nil {
		return nil, nil, err
	}
	return conn, conn, nil
}
