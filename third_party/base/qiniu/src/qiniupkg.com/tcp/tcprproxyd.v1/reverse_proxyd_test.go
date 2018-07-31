package tcprproxyd

import (
	"io"
	"io/ioutil"
	"net"
	"testing"

	"qiniupkg.com/tcp/tcprproxy.v1"
)

// -----------------------------------------------------------------------------

type fooService struct {
}

func (p *fooService) ListenAndServe(done chan bool) (err error) {

	//文档约定业务服务器和额外监听的端口为+1关系,但不是必要. 此测试用例没有按照该约定
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return
	}
	done <- true
	return p.Serve(l)
}

func (p *fooService) Serve(l net.Listener) (err error) {

	defer l.Close()

	for {
		c, err2 := l.Accept()
		if err2 != nil {
			return err2
		}
		go func() {
			defer c.Close()
			raddr := c.RemoteAddr().String()
			io.WriteString(c, raddr)
		}()
	}
}

func TestProxy(t *testing.T) {

	done := make(chan bool, 3)
	go func() {
		router := SingleBackend("localhost:8889")
		err := ListenAndServe(":18888", router, done)
		if err != nil {
			t.Fatal("proxy ListenAndServe failed:", err)
		}
	}()
	go func() {
		foo := new(fooService)
		go tcprproxy.ListenAndServe(":8889", foo, done)
		t.Fatal("net.Listen failed:", foo.ListenAndServe(done))
	}()
	<-done
	<-done
	<-done

	laddr, _ := net.ResolveTCPAddr("tcp", "localhost:9999")
	dialer := &net.Dialer{LocalAddr: laddr}
	c, err := dialer.Dial("tcp", "localhost:18888")
	if err != nil {
		t.Fatal("Dial failed:", err)
	}
	defer c.Close()

	_, err = c.Write([]byte("Hello"))
	if err != nil {
		t.Fatal("Write failed:", err)
	}

	b, err := ioutil.ReadAll(c)
	if err != nil {
		t.Fatal("ReadAll failed:", err)
	}
	if string(b) != "127.0.0.1:9999" {
		t.Fatal("b:", string(b))
	}
}

// -----------------------------------------------------------------------------
