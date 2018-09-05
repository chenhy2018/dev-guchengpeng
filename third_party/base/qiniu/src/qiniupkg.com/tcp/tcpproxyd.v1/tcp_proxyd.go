package tcpproxyd

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"

	"qiniupkg.com/x/log.v7"
)

// -----------------------------------------------------------------------------

const (
	defaultProxyHost = ":12306"
)

type Proxier struct {
	Addr     string
	Listened chan bool
}

func (p *Proxier) ListenAndServe() (err error) {

	addr := p.Addr
	if addr == "" {
		addr = defaultProxyHost
	}
	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("ListenAndServe(tcp proxy) %s failed: %v\n", addr, err)
		return
	}
	if p.Listened != nil {
		p.Listened <- true
	}
	err = p.Serve(l)
	if err != nil {
		log.Fatalf("ListenAndServe(tcp proxy) %s failed: %v\n", addr, err)
	}
	return
}

var (
	cmdOk   = []byte{'o', 'k', '\n'}
	cmdDial = []byte{'d', 'i', 'a', 'l', ' '}
)

func replyErr(c *net.TCPConn, err error) {

	fmt.Fprintf(c, "error unknown: %s\n", err.Error())
	c.Close()
}

func (p *Proxier) Serve(l net.Listener) (err error) {

	defer l.Close()

	for {
		c1, err1 := l.Accept()
		if err1 != nil {
			return err1
		}
		c := c1.(*net.TCPConn)
		go func() {
			r := bufio.NewReader(c)
			line, err2 := r.ReadSlice('\n')
			if !bytes.HasPrefix(line, cmdDial) {
				c.Close()
				log.Errorf("invalid tcpproxy command: %s", line)
				return
			}

			// dail <TargetAddress>\n
			//
			address := string(line[5 : len(line)-1])
			raddr, err2 := net.ResolveTCPAddr("tcp", address)
			if err2 != nil {
				replyErr(c, err2)
				return
			}

			c2, err2 := net.DialTCP("tcp", nil, raddr)
			if err2 != nil {
				replyErr(c, err2)
				return
			}

			_, err2 = c.Write(cmdOk)
			if err2 != nil {
				c.Close()
				c2.Close()
				log.Error("tcpproxy failed:", err2)
				return
			}

			go func() {
				n3, err3 := io.Copy(c, c2)
				if err3 != nil {
					log.Info("tcpproxy (response):", n3, err3)
				}
				c.CloseWrite()
				c2.CloseRead()
			}()

			n2, err2 := r.WriteTo(c2)
			if err2 != nil {
				log.Info("tcpproxy (request):", n2, err2)
			}
			c.CloseRead()
			c2.CloseWrite()
		}()
	}
}

// -----------------------------------------------------------------------------

func ListenAndServe(host string, listened chan bool) (err error) {

	proxier := &Proxier{
		Addr:     host,
		Listened: listened,
	}
	return proxier.ListenAndServe()
}

// -----------------------------------------------------------------------------
