package connectpoints

import (
	"fmt"
	"github.com/qiniu/log.v1"
	"github.com/qiniu/ts"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

// -------------------------------------------------------

const (
	QFoo  = 0x0001
	QBar1 = 0x0002
	QBar2 = 0x0004
)

type Foo struct {
	A int
	B string
}

// -------------------------------------------------------

var cps = NewServer()

func worker(t *testing.T) {

	cps.Fire(QFoo, &Foo{123, "foo"})
	time.Sleep(1e9)

	cps.Fire(QBar2, &Foo{456, "bar2"})
	time.Sleep(1e9)

	cps.Fire(QBar1, nil)
}

var serverAddr string

func server() *httptest.Server {

	mux := http.NewServeMux()
	mux.Handle("/longpoll/", cps)
	svr := httptest.NewServer(mux)
	return svr
}

// -------------------------------------------------------

func onFoo(req Request) {
	var foo Foo
	err := req.Decode(&foo)
	if err != nil {
		fmt.Println("req.Decode:", err)
		return
	}
	fmt.Println("onFoo:", foo.A, foo.B)
}

func onBar1(req Request) {
	fmt.Println("onBar1")
	os.Exit(0)
}

func client(t *testing.T, eventTypes int64) {

	mux := NewServeMux(nil)
	mux.HandleFunc(QFoo, onFoo)
	mux.HandleFunc(QBar2, onFoo)
	mux.HandleFunc(QBar1, onBar1)

	c, err := Connect(serverAddr+"/longpoll/", eventTypes)
	if err != nil {
		ts.Fatal(t, "Connect:", err)
	}
	defer c.Close()

	err = c.RunLoop(mux)
	if err != nil {
		ts.Fatal(t, "RunLoop failed:", err)
	}
}

// -------------------------------------------------------

func TestConnectPoints(t *testing.T) {

	log.SetFlags(log.Llongfile)
	log.Println("go server")

	svr := server()
	serverAddr = svr.URL
	defer svr.Close()

	log.Println("go client")

	go client(t, QFoo|QBar1)
	go client(t, QFoo|QBar2)
	time.Sleep(1e9)

	log.Println("go worker")

	go worker(t)
	time.Sleep(5e9)
}

// -------------------------------------------------------
