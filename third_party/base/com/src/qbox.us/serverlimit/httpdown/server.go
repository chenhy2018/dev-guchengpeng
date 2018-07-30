package httpdown

import (
	"crypto/tls"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"qbox.us/net/netutil"

	"github.com/facebookgo/httpdown"
	"github.com/facebookgo/stats"
)

func ListenAndServe(s *http.Server, hd *httpdown.HTTP, limit int) {
	ListenAndServeWithMsg(s, hd, limit, nil)
}

func ListenAndServeWithMsg(s *http.Server, hd *httpdown.HTTP, limit int, notify func()) error {
	if hd == nil {
		hd = &httpdown.HTTP{}
	}
	addr := s.Addr
	if addr == "" {
		if s.TLSConfig == nil {
			addr = ":http"
		} else {
			addr = ":https"
		}
	}
	l, err := net.Listen("tcp", addr)
	if err != nil {
		stats.BumpSum(hd.Stats, "listen.error", 1)
		return err
	}
	l = tcpKeepAliveListener{l.(*net.TCPListener)}
	if notify != nil {
		notify()
	}
	if s.TLSConfig != nil {
		l = tls.NewListener(l, s.TLSConfig)
	}
	listener := l
	if limit > 0 {
		listener = netutil.LimitListener(l, limit)
	}
	hs := hd.Serve(s, listener)
	log.Printf("serving on http://%s/ with pid %d\n", s.Addr, os.Getpid())
	waiterr := make(chan error, 1)
	go func() {
		defer close(waiterr)
		waiterr <- hs.Wait()
	}()

	signals := make(chan os.Signal, 10)
	signal.Notify(signals, syscall.SIGTERM, syscall.SIGINT)

	select {
	case err := <-waiterr:
		if err != nil {
			return err
		}
	case s := <-signals:
		signal.Stop(signals)
		log.Printf("signal received: %s\n", s)
		if err := hs.Stop(); err != nil {
			return err
		}
		if err := <-waiterr; err != nil {
			return err
		}
	}
	log.Println("exiting")
	return nil
}

// from "net/http"
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
