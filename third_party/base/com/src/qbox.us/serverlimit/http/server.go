package http

import (
	"net"
	"net/http"
	"time"

	"qbox.us/net/netutil"
)

type Server struct {
	*http.Server
	limit int
}

func ListenAndServe(addr string, handler http.Handler, limit int) error {
	server := &Server{&http.Server{Addr: addr, Handler: handler}, limit}
	return server.ListenAndServe()
}

func (s *Server) ListenAndServe() error {
	addr := s.Addr
	if addr == "" {
		addr = ":http"
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	return s.Server.Serve(netutil.LimitListener(tcpKeepAliveListener{ln.(*net.TCPListener)}, s.limit))
}

// tcpKeepAliveListener sets TCP keep-alive timeouts on accepted
// connections. It's used by ListenAndServe and ListenAndServeTLS so
// dead TCP connections (e.g. closing laptop mid-download) eventually
// go away.
type tcpKeepAliveListener struct {
	*net.TCPListener
}

func (ln tcpKeepAliveListener) Accept() (c net.Conn, err error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return
	}
	tc.SetKeepAlive(true)
	tc.SetKeepAlivePeriod(3 * time.Minute)
	return tc, nil
}
