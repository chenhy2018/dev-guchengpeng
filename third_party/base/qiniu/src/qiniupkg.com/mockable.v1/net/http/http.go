package http

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"qiniupkg.com/mockable.v1/net"
	"qiniupkg.com/x/log.v7"
)

// ---------------------------------------------------------------------------

var DefaultTransport = http.DefaultTransport
var DefaultTransportRestorer = http.DefaultTransport
var ListenAndServe = http.ListenAndServe
var ListenAndServeTLS = http.ListenAndServeTLS

// 负责：laddr(logicIP:logicPort)地址转换
// 当 logicIP = "0.0.0.0" 时，需要将 logicIP 改为 MockingIPs 后进行地址转换（需要Listen多个端口）
//
func MockListenAndServe(laddr string, handler http.Handler) (err error) {

	pos := strings.Index(laddr, ":")
	if pos < 0 {
		log.Fatalln("invalid logic address: no port -", laddr)
	}

	ip := laddr[:pos]
	if ip == "" || ip == "0.0.0.0" {
		port := laddr[pos+1:]
		return broadcastListenAndServe(port, handler)
	}

	log.Info("http.ListenAndServe:", laddr)
	return http.ListenAndServe(net.LogicToPhy(laddr), handler)
}

func MockListenAndServeTLS(laddr, certFile, keyFile string, handler http.Handler) (err error) {

	log.Fatal("not impl")
	return
}

// ---------------------------------------------------------------------------

var errPanic = errors.New("panic")

func broadcastListenAndServe(port string, handler http.Handler) (err error) {

	ch := make(chan error, len(net.MockingIPs))

	for _, ip := range net.MockingIPs {
		laddr := ip + ":" + port
		go func() {
			err1 := errPanic
			defer func() {
				ch <- err1
			}()
			log.Info("http.ListenAndServe:", laddr)
			err1 = http.ListenAndServe(net.LogicToPhy(laddr), handler)
		}()
	}

	for range net.MockingIPs {
		err1 := <-ch
		if err1 != nil {
			err = err1
		}
	}
	return
}

// ---------------------------------------------------------------------------

func doInit() {

	ListenAndServe = MockListenAndServe
	ListenAndServeTLS = MockListenAndServeTLS
	DefaultTransport = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		Dial: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 10 * time.Second,
	}
	http.DefaultTransport = DefaultTransport
}

// if any external program want to use defaultTransport, we need to restore it
func RestoreTransport() {
	http.DefaultTransport = DefaultTransportRestorer
}

func init() {

	net.RegisterInit(doInit)
}

// ---------------------------------------------------------------------------
