package tcpproxy

import (
	"errors"
	"io"
	"io/ioutil"
	"net"
	mocknet "qiniupkg.com/mockable.v1/net"
	"strings"
	"syscall"
)

// -----------------------------------------------------------------------------

var (
	errInvalidProxyData = errors.New("tcpproxy: invalid proxy data")
)

type ProxyError struct {
	Detail string
}

func (p *ProxyError) Error() string {
	return p.Detail
}

func DialFrom(laddr, raddr string) (c net.Conn, err error) {
	var l *net.TCPAddr
	if laddr != "" {
		l, err = net.ResolveTCPAddr("tcp", laddr)
		if err != nil {
			return
		}
	}
	return dial(raddr, &mocknet.Dialer{
		LocalAddr: l,
	})
}

func Dial(host string) (c net.Conn, err error) {
	return dial(host, &mocknet.Dialer{})
}

// host = "-X <ProxyAddress> <TargetAddress>"
// eg. host = "-X 192.168.1.10:3245 223.72.136.2:8888"
//
func dial(host string, dialer *mocknet.Dialer) (c net.Conn, err error) {
	if strings.HasPrefix(host, "-X ") {
		host = host[3:]
		pos := strings.Index(host, " ")
		if pos <= 0 {
			return nil, syscall.EINVAL
		}
		c, err = dialer.Dial("tcp", host[:pos])
		if err != nil {
			return
		}
		host = host[pos+1:]
		// dial <TargetAddress>\n
		b := make([]byte, 0, len(host)+6)
		b = append(b, "dial "...)
		b = append(b, host...)
		b = append(b, '\n')
		_, err = c.Write(b)
		if err != nil {
			c.Close()
			return
		}
		// ok\n
		_, err = io.ReadFull(c, b[:3])
		if err != nil {
			c.Close()
			return
		}
		if b[0] == 'o' && b[1] == 'k' && b[2] == '\n' {
			return
		}
		if b[0] == 'e' && b[1] == 'r' && b[2] == 'r' {
			// error <ErrorType>: <Key1>=<Val1>&<Key2>=<Val2>&...\n
			b, err = ioutil.ReadAll(c)
			c.Close()
			if err != nil {
				return
			}
			if len(b) > 3 && b[0] == 'o' && b[1] == 'r' && b[2] == ' ' && b[len(b)-1] == '\n' {
				return nil, &ProxyError{string(b[3 : len(b)-1])}
			}
		} else {
			c.Close()
		}
		return nil, errInvalidProxyData
	} else {
		return dialer.Dial("tcp", host)
	}
}

// -----------------------------------------------------------------------------
