package net

import (
	"net"
	"strings"
	"time"

	"qiniupkg.com/x/log.v7"
)

type Dialer net.Dialer

func (d *Dialer) Dial(network, raddr string) (conn net.Conn, err error) {

	if !Mocking {
		return ((*net.Dialer)(d)).Dial(network, raddr)
	}

	var idx int = 0
	var pDialer net.Dialer = (net.Dialer)(*d)
	var localIP, remoteIP string

	// convert raddr, which may be: 1. mocked addr 0.0.0.x; 2. mocked localhost; 3. real phy addr
	remoteAddr, err := net.ResolveTCPAddr(network, raddr)
	if err != nil {
		log.Fatalln("invalid remote address: ", raddr, err)
	}

	if ip := remoteAddr.IP.String(); ip == "127.0.0.1" && remoteAddr.Port < 256 {
		ip = MockingIPs[0]
		remoteAddr.IP = net.ParseIP(ip).To4()
	}

	remoteIP = remoteAddr.IP.String()

	raddr = remoteAddr.String()
	if strings.HasPrefix(raddr, "0.0.0.0") {
		log.Fatalln("not impl")
	}
	if strings.HasPrefix(raddr, "0.0.0.") {
		raddr = LogicToPhy(raddr)
	}

	// convert laddr, which may be: 1. mocked addr 0.0.0.x; 2. mocked localhost
	if d.LocalAddr != nil {
		localAddr, err := net.ResolveTCPAddr(d.LocalAddr.Network(), d.LocalAddr.String())
		if err != nil {
			log.Fatalln("invalid local address: ", d.LocalAddr.String(), err)
		}

		if localAddr.IP != nil {
			localIP = localAddr.IP.String()

			match := false
			if localIP == "127.0.0.1" {
				idx = 0
				match = true
			} else {
				for i, ip := range MockingIPs {
					if ip == localIP {
						idx = i
						match = true
						break
					}
				}
			}

			if !match {
				log.Fatalln("invalid local address: not found -", localIP)
			}

			paddr := LogicToPhy(localAddr.String())
			localAddr, err = net.ResolveTCPAddr(d.LocalAddr.Network(), paddr)
			if err != nil {
				log.Fatalln("invalid local address: ", paddr, err)
			}

			pDialer.LocalAddr = localAddr
		}
	}

	// dail mocked
	if speed, ok := speeds[remoteIP]; ok {
		if len(speed) == 0 {
			return nil, net.ErrWriteToConnected
		}
		c1, err1 := pDialer.Dial(network, raddr)
		if err1 != nil {
			log.Warn("Dial raddr:", raddr, err1)
			return nil, err1
		}
		return newMockConn(idx, c1, speed[idx]), nil
	}

	// dail real
	return pDialer.Dial(network, raddr)
}

// ---------------------------------------------------------------------------

type mockConn struct {
	ipIdx int
	net.Conn
	fcr *FlowControl
	fcw *FlowControl
}

func newMockConn(ipIdx int, conn net.Conn, speed Speed) net.Conn {

	if len(speed) == 1 && speed[0].Bps == BpsNotLimit {
		return conn
	}
	return &mockConn{
		ipIdx: ipIdx,
		Conn:  conn,
		fcr:   NewFlowControl(speed),
		fcw:   NewFlowControl(speed),
	}
}

func (p *mockConn) LocalAddr() (address net.Addr) {

	log.Fatal("not impl")
	return
}

func (p *mockConn) RemoteAddr() (address net.Addr) {

	log.Fatal("not impl")
	return
}

func (p *mockConn) Read(b []byte) (n int, err error) {

	var limit int
	for {
		limit = p.fcr.Require(len(b))
		if limit != 0 {
			break
		}
		time.Sleep(FlowControlWindow)
	}
	n, err = p.Conn.Read(b[:limit])
	p.fcr.Consume(n)
	MockingIPInfos[p.ipIdx].AddInBps(n)
	return
}

func (p *mockConn) Write(b []byte) (n int, err error) {

	base, left := 0, len(b)
	for {
		limit := p.fcw.Require(left)
		n1, err1 := p.Conn.Write(b[base : base+limit])
		p.fcw.Consume(n1)
		MockingIPInfos[p.ipIdx].AddOutBps(n1)
		n += n1
		if err1 != nil {
			return n, err1
		}
		if n == len(b) {
			return
		}
		base += n1
		left -= n1
		time.Sleep(FlowControlWindow)
	}
}

// ---------------------------------------------------------------------------
