package tcputil

import (
	"net"

	"qiniupkg.com/x/log.v7"
)

// -----------------------------------------------------------------------------

type Server interface {
	Serve(l net.Listener) (err error)
}

func ListenAndServe(addr string, server Server, listened chan bool) (err error) {

	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("ListenAndServe(tcp) %s failed: %v\n", addr, err)
		return
	}
	if listened != nil {
		listened <- true
	}
	err = server.Serve(l)
	if err != nil {
		log.Fatalf("ListenAndServe(tcp) %s failed: %v\n", addr, err)
	}
	return
}

// -----------------------------------------------------------------------------

