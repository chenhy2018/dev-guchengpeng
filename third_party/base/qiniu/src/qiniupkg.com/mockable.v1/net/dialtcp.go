package net

import (
	"io"
	"net"

	"qiniupkg.com/x/log.v7"
)

// 负责: 1) raddr 的地址转换; 2) laddr 的地址转换 2) 限速
func MockDialTCP(network string, laddr, raddr *net.TCPAddr) (conn *net.TCPConn, err error) {

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return
	}
	defer func() {
		if err != nil {
			listener.Close()
		}
	}()

	go func() {
		conn, err := listener.Accept()
		if err != nil {
			listener.Close()
			log.Error("accept failed:", err)
			return
		}
		listener.Close()

		defer conn.Close()

		dialer := Dialer{
			LocalAddr: laddr,
		}
		forwardConn, err := dialer.Dial(network, raddr.String())
		if err != nil {
			log.Error("mockDialEx:", laddr.String(), raddr.String(), err)
			return
		}

		wait := make(chan bool, 2)

		go func() {
			n, err := io.Copy(conn, forwardConn)
			if err != nil {
				log.Error("read failed:", n, err)
			}
			wait <- true
		}()
		go func() {
			n, err := io.Copy(forwardConn, conn)
			if err != nil {
				log.Error("write failed:", n, err)
			}

			wait <- true
		}()

		<-wait
		<-wait
	}()

	raddress, err := net.ResolveTCPAddr("tcp", listener.Addr().String())
	if err != nil {
		return
	}
	conn, err = net.DialTCP("tcp", nil, raddress)
	if err != nil {
		return
	}

	return
}
