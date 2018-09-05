package tcprproxyd

import (
	"io"
	"net"
	"sync/atomic"

	"qiniupkg.com/x/log.v7"
)

// -----------------------------------------------------------------------------

type Router interface {
	Pick(raddr net.Addr) (backend string, err error)
	Unpick(backend string, raddr net.Addr)
}

type ReverseProxier struct {
	Addr     string
	Router   Router
	Listened chan bool
}

func (p *ReverseProxier) ListenAndServe() (err error) {

	addr := p.Addr
	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("ListenAndServe(tcprproxyd) %s failed: %v\n", addr, err)
		return
	}
	if p.Listened != nil {
		p.Listened <- true
	}
	err = p.Serve(l)
	if err != nil {
		log.Fatalf("ListenAndServe(tcprproxyd) %s failed: %v\n", addr, err)
	}
	return
}

func (p *ReverseProxier) Serve(l net.Listener) (err error) {

	defer l.Close()

	router := p.Router
	for {
		c1, err1 := l.Accept()
		if err1 != nil {
			return err1
		}
		c := c1.(*net.TCPConn)
		go func() {
			raddr := c.RemoteAddr()
			address, err2 := router.Pick(raddr)
			if err2 != nil {
				log.Error("tcprproxy route select:", err2)
				c.Close()
				return
			}

			ref := int32(1)
			release := func() {
				if atomic.AddInt32(&ref, -1) == 0 {
					router.Unpick(address, raddr)
				}
			}
			defer release()

			backend, err2 := net.ResolveTCPAddr("tcp", address)
			if err2 != nil {
				log.Error("tcprproxy: invalid tcp address -", address, "error:", err2)
				c.Close()
				return
			}
			c2, err2 := net.DialTCP("tcp", nil, backend)
			if err2 != nil {
				log.Error("tcprproxy: dial backend failed -", address, "error:", err2)
				c.Close()
				return
			}

			var b [32]byte
			addr := raddr.String()
			b[0] = 32
			b[1] = byte(len(addr))
			copy(b[2:], addr)
			_, err2 = c2.Write(b[:])
			if err2 != nil {
				c.Close()
				c2.Close()
				log.Error("tcprproxy write failed:", err2)
				return
			}

			ref++
			go func() {
				io.Copy(c, c2)
				c.CloseWrite()
				c2.CloseRead()
				release()
			}()

			n2, err2 := io.Copy(c2, c)
			if err2 != nil {
				log.Info("tcprproxy (request):", n2, err2)
			}
			c.CloseRead()
			c2.CloseWrite()
		}()
	}
}

// -----------------------------------------------------------------------------

func ListenAndServe(addr string, router Router, listened chan bool) (err error) {

	rp := &ReverseProxier{Addr: addr, Router: router, Listened: listened}
	return rp.ListenAndServe()
}

// -----------------------------------------------------------------------------

