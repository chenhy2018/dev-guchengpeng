package tcprproxy

import (
	"errors"
	"io"
	"net"

	"qiniupkg.com/x/log.v7"
)

var (
	errInvalidProxyData = errors.New("tcprproxy: invalid reverse-proxy data")
)

// -----------------------------------------------------------------------------

type proxyConn struct {
	*net.TCPConn
	raddr *net.TCPAddr
}

func (p *proxyConn) RemoteAddr() net.Addr {

	return p.raddr
}

// -----------------------------------------------------------------------------

type proxyListener struct {
	net.Listener
}

// HeaderLen byte
// AddrLen byte
// Addr [AddrLen]byte
// Padding [HeaderLen-AddrLen-2]byte
//
func (p *proxyListener) Accept() (conn net.Conn, err error) {

	c1, err := p.Listener.Accept()
	if err != nil {
		return
	}
	c := c1.(*net.TCPConn)

	var b [32]byte
	_, err = io.ReadFull(c, b[:])
	if err != nil {
		return
	}
	if (b[0]&31) != 0 || b[1] > 30 {
		return nil, errInvalidProxyData
	}
	if b[0] > 32 {
		left := b[0] - 32
		padding := make([]byte, left)
		_, err = io.ReadFull(c, padding)
		if err != nil {
			return
		}
	}
	raddr := string(b[2:b[1]+2])
	address, err := net.ResolveTCPAddr("tcp", raddr)
	if err != nil {
		return
	}

	return &proxyConn{c, address}, nil
}

func ReverseProxy(rpl net.Listener) net.Listener {

	return &proxyListener{rpl}
}

// -----------------------------------------------------------------------------

type Server interface {
	Serve(l net.Listener) (err error)
}

func ListenAndServe(addr string, server Server, listened chan bool) (err error) {

	rpl, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("ListenAndServe(reverse proxy) %s failed: %v\n", addr, err)
		return
	}
	if listened != nil {
		listened <- true
	}
	err = server.Serve(&proxyListener{rpl})
	if err != nil {
		log.Fatalf("ListenAndServe(reverse proxy) %s failed: %v\n", addr, err)
	}
	return
}

// -----------------------------------------------------------------------------

