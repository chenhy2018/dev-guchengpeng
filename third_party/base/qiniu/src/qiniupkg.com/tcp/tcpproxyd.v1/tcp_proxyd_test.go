package tcpproxyd

import (
	"io"
	"io/ioutil"
	"net"
	"testing"

	"qiniupkg.com/tcp/tcpproxy.v1"
)

// -----------------------------------------------------------------------------

func TestProxy(t *testing.T) {

	done := make(chan bool, 2)
	var l net.Listener
	go func() {
		err := ListenAndServe(":12306", done)
		if err != nil {
			t.Fatal("proxy ListenAndServe failed:", err)
		}
	}()
	go func() {
		var err error
		l, err = net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatal("net.Listen failed:", err)
		}
		done <- true
		for {
			c, err2 := l.Accept()
			if err2 != nil {
				t.Fatal("Accept failed:", err2)
			}
			go func() {
				defer c.Close()
				io.Copy(c, c)
			}()
		}
	}()
	<-done
	<-done

	c1, err := tcpproxy.Dial("-X 127.0.0.1:12306 " + l.Addr().String())
	if err != nil {
		t.Fatal("tcpproxy.Dial failed:", err)
	}
	c := c1.(*net.TCPConn)
	defer c.CloseRead()

	_, err = c.Write([]byte("Hello"))
	c.CloseWrite()
	if err != nil {
		t.Fatal("Write failed:", err)
	}

	b, err := ioutil.ReadAll(c)
	if err != nil {
		t.Fatal("ReadAll failed:", err)
	}
	if string(b) != "Hello" {
		t.Fatal("b != \"Hello\"")
	}

	c2, err := tcpproxy.Dial("-X 127.0.0.1:12306 127.0.0.1:7777")
	if err == nil {
		t.Fatal("tcpproxy.Dial:", c2)
	}
}

// -----------------------------------------------------------------------------
