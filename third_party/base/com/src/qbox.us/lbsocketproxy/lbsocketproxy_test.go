package lbsocketproxy

import (
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/qiniu/log.v1"

	"github.com/stretchr/testify.v1/require"
)

func TestLBSocketProxy(t *testing.T) {
	log.SetOutputLevel(0)
	end, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err, "listen")
	log.Info("end", end.Addr())
	gate1, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err, "listen")
	log.Info("gate1", gate1.Addr())
	gate3, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err, "listen")
	log.Info("gate3", gate3.Addr())
	wg := &sync.WaitGroup{}
	wg.Add(2)
	go socks5Gateway(t, gate1, end, socks5IP4, wg, true)
	// go socks5Gateway(t, gate2, end, socks5IP4, wg)
	go socks5Gateway(t, gate3, end, socks5IP4, wg, true)
	lbs, err := NewLbSocketProxy(&Config{
		Hosts:         []string{gate1.Addr().String(), "1.1.1.1:123", gate3.Addr().String()},
		DialTimeoutMs: 100,
		TryTimes:      3,
		Type:          "",
		Router: map[string]string{
			"local": "127.0.0.0/8",
		},
	})
	require.NoError(t, err, "listen")
	addr, err := net.ResolveTCPAddr("tcp", end.Addr().String())
	require.NoError(t, err, "listen")
	log.Info("test1")
	c, err := lbs.Dial(addr)
	require.NoError(t, err, "listen")
	c.Close()
	log.Info("test2")
	c, err = lbs.Dial(addr)
	require.NoError(t, err, "listen")
	c.Close()
	wg.Wait()
}

const socks5Version = 5

const (
	socks5AuthNone     = 0
	socks5AuthPassword = 2
)

const socks5Connect = 1

const (
	socks5IP4    = 1
	socks5Domain = 3
	socks5IP6    = 4
)

var trytime = 0

func socks5Gateway(t *testing.T, gateway, endSystem net.Listener, typ byte, wg *sync.WaitGroup, support bool) {
	defer wg.Done()

	log.Info("begin accept")
	c, err := gateway.Accept()
	if err != nil {
		t.Errorf("net.Listener.Accept failed: %v", err)
		return
	}
	log.Info("connected", c.RemoteAddr(), c.LocalAddr())
	defer c.Close()
	defer gateway.Close()

	b := make([]byte, 32)
	var n int
	if typ == socks5Domain {
		n = 4
	} else {
		n = 3
	}
	if _, err := io.ReadFull(c, b[:n]); err != nil {
		t.Errorf("io.ReadFull failed: %v", err)
		return
	}
	if !support {
		trytime++
		c.Write([]byte{0x05, 0xff})
		return
	}
	if _, err := c.Write([]byte{socks5Version, socks5AuthNone}); err != nil {
		t.Errorf("net.Conn.Write failed: %v", err)
		return
	}
	if typ == socks5Domain {
		n = 16
	} else {
		n = 10
	}
	if _, err := io.ReadFull(c, b[:n]); err != nil {
		t.Errorf("io.ReadFull failed: %v", err)
		return
	}
	if b[0] != socks5Version || b[1] != socks5Connect || b[2] != 0x00 || b[3] != typ {
		t.Errorf("got an unexpected packet: %#02x %#02x %#02x %#02x", b[0], b[1], b[2], b[3])
		return
	}
	if typ == socks5Domain {
		copy(b[:5], []byte{socks5Version, 0x00, 0x00, socks5Domain, 9})
		b = append(b, []byte("localhost")...)
	} else {
		copy(b[:4], []byte{socks5Version, 0x00, 0x00, socks5IP4})
	}
	host, port, err := net.SplitHostPort(endSystem.Addr().String())
	if err != nil {
		t.Errorf("net.SplitHostPort failed: %v", err)
		return
	}
	b = append(b, []byte(net.ParseIP(host).To4())...)
	p, err := strconv.Atoi(port)
	if err != nil {
		t.Errorf("strconv.Atoi failed: %v", err)
		return
	}
	b = append(b, []byte{byte(p >> 8), byte(p)}...)
	if _, err := c.Write(b); err != nil {
		t.Errorf("net.Conn.Write failed: %v", err)
		return
	}
}

func TestLBSocketProxyNotRetryNetError(t *testing.T) {
	log.SetOutputLevel(0)
	end, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err, "listen")
	log.Info("end", end.Addr())
	gate1, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err, "listen")
	log.Info("gate1", gate1.Addr())
	gate2, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err, "listen")
	log.Info("gate2", gate2.Addr())
	gate3, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err, "listen")
	log.Info("gate3", gate3.Addr())
	wg := &sync.WaitGroup{}
	wg.Add(2)
	gate2.Close()
	go socks5Gateway(t, gate1, end, socks5IP4, wg, false)
	// go socks5Gateway(t, gate2, end, socks5IP4, wg)
	go socks5Gateway(t, gate3, end, socks5IP4, wg, false)
	lbs, err := NewLbSocketProxy(&Config{
		Hosts:         []string{gate1.Addr().String(), gate2.Addr().String(), gate3.Addr().String()},
		DialTimeoutMs: 100,
		TryTimes:      3,
		Type:          "all",
	})
	require.NoError(t, err, "listen")
	addr, err := net.ResolveTCPAddr("tcp", end.Addr().String())
	require.NoError(t, err, "listen")
	log.Info("test1")
	_, err = lbs.Dial(addr)
	require.NoError(t, err, "retry")
	log.Info("test2")
	_, err = lbs.Dial(addr)
	require.NoError(t, err, "retry 2")
	wg.Wait()
	require.Equal(t, trytime, 2)
}

func TestLBSocketLbConn(t *testing.T) {
	log.SetOutputLevel(0)
	end, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err, "listen")
	log.Info("end", end.Addr())
	gate1, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err, "listen")
	log.Info("gate1", gate1.Addr())
	wg := &sync.WaitGroup{}
	wg.Add(2)
	var gate1Count, gate3Count int
	go func() {
		for {
			conn, err := gate1.Accept()
			require.NoError(t, err)
			go io.Copy(ioutil.Discard, conn)
			gate1Count++
		}
	}()
	lbs, err := NewLbSocketProxy(&Config{
		Hosts:         []string{gate1.Addr().String(), "1.1.1.1:123", "127.0.0.1:4456"},
		DialTimeoutMs: 50,
		TryTimes:      3,
		Type:          "all",
		FailBanTimeMs: 100,
	})
	addr, err := net.ResolveTCPAddr("tcp", end.Addr().String())
	require.NoError(t, err, "listen")
	for i := 0; i < 100; i++ {
		go func(i int) {
			_, err := lbs.Dial(addr)
			require.NoError(t, err)
		}(i)
	}
	time.Sleep(time.Millisecond * 300)
	require.Equal(t, 100, gate1Count)
	gate3, err := net.Listen("tcp", "127.0.0.1:4456")
	require.NoError(t, err, "listen")
	log.Info("gate3", gate3.Addr())
	go func() {
		for {
			conn, err := gate3.Accept()
			require.NoError(t, err)
			go io.Copy(ioutil.Discard, conn)
			gate3Count++
		}
	}()
	time.Sleep(time.Millisecond * 200)
	for i := 0; i < 100; i++ {
		go func(i int) {
			time.Sleep(time.Millisecond * time.Duration(rand.Intn(100)))
			_, err := lbs.Dial(addr)
			require.NoError(t, err)
		}(i)
	}
	time.Sleep(time.Millisecond * 300)
	require.Equal(t, 100, gate3Count, fmt.Sprint(gate1Count, gate3Count))
}

func TestFail(t *testing.T) {
	testCheckConn = func(c net.Conn, err error) {
		if err == nil {
			require.NotNil(t, c)
		}
	}
	lbs, err := NewLbSocketProxy(&Config{
		Hosts:         []string{"1.1.1.1:123", "2.2.2.2:123"},
		DialTimeoutMs: 50,
		TryTimes:      2,
		Type:          "all",
		FailBanTimeMs: 1000,
	})
	addr, err := net.ResolveTCPAddr("tcp", "3.3.3.3:123")
	require.NoError(t, err, "ResolveTCPAddr")
	for i := 0; i < 5; i++ {
		log.Println("===", i)
		c, err := lbs.Dial(addr)
		require.Error(t, err)
		require.Nil(t, c)
	}
}
