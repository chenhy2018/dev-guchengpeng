package httpdown

import (
	"errors"
	"net/http"
	"strings"

	"qiniupkg.com/httpdown.v1"
	"qiniupkg.com/mockable.v1/net"
	"qiniupkg.com/x/log.v7"
)

// ---------------------------------------------------------------------------

var ListenAndServe = httpdown.ListenAndServe

// 负责：laddr(logicIP:logicPort)地址转换
// 当 logicIP = "0.0.0.0" 时，需要将 logicIP 改为 MockingIPs 后进行地址转换（需要Listen多个端口）
//
func MockListenAndServe(s *http.Server, hd *httpdown.HTTP) (err error) {

	laddr := s.Addr
	pos := strings.Index(laddr, ":")
	if pos < 0 {
		log.Fatalln("invalid logic address: no port -", laddr)
	}

	ip := laddr[:pos]
	if ip == "" || ip == "0.0.0.0" {
		port := laddr[pos+1:]
		return broadcastListenAndServe(port, s, hd)
	}
	s2 := *s
	s2.Addr = net.LogicToPhy(laddr)
	return httpdown.ListenAndServe(&s2, hd)
}

// ---------------------------------------------------------------------------

var errPanic = errors.New("panic")

func broadcastListenAndServe(port string, s *http.Server, hd *httpdown.HTTP) (err error) {

	ch := make(chan error, len(net.MockingIPs))

	for _, ip := range net.MockingIPs {
		laddr := ip + ":" + port
		go func() {
			err1 := errPanic
			defer func() {
				ch <- err1
			}()
			s2 := *s
			s2.Addr = net.LogicToPhy(laddr)
			err1 = httpdown.ListenAndServe(&s2, hd)
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
}

func init() {

	net.RegisterInit(doInit)
}

// ---------------------------------------------------------------------------
