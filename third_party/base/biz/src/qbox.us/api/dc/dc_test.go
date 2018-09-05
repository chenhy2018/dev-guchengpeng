package dc

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/qiniu/xlog.v1"
	"github.com/stretchr/testify/assert"

	"qbox.us/mockdc"
)

func doTestClient(t *testing.T) {

	fmt.Println("Begin doTestClient")

	conns := []DCConn{
		DCConn{Keys: []string{"host0_0", "host0_1"}, Host: "http://127.0.0.1:4460"},
		DCConn{Keys: []string{"host1"}, Host: "http://127.0.0.1:4461"},
		DCConn{Keys: []string{"host2"}, Host: "http://127.0.0.1:4462"},
		DCConn{Keys: []string{"host3"}, Host: "http://127.0.0.1:4463"},
	}
	client := NewClient(conns, 4, nil)
	defer client.Close()

	xl := xlog.NewDummy()
	_, _, err := client.Get(xl, []byte("key"))
	if err == nil {
		t.Fatal(err)
	}

	err = client.Set(xl, []byte("key"), bytes.NewReader([]byte("body")), 4)
	if err != nil {
		t.Fatal(err)
	}

	_, _, err = client.RangeGet(xl, []byte("key"), 5, 7)
	if err == nil {
		t.Fatal(err)
	}

	_, _, err = client.RangeGet(xl, []byte("key"), 0, 4)
	if err != nil {
		t.Fatal(err)
	}

	conns = []DCConn{
		DCConn{Keys: []string{"host0_0", "host0_1"}, Host: "http://127.0.0.1:14460"},
		DCConn{Keys: []string{"host1_0", "host1_1"}, Host: "http://127.0.0.1:14461"},
	}
	client = NewClient(conns, 4, nil)

	err = client.Set(xl, []byte("key"), bytes.NewReader([]byte("body")), 4)
	if err == nil {
		t.Fatal(err)
	}
}

var dcTestDir = os.Getenv("HOME") + "/dcclientTest"

func dcRun(addr string, id int, t *testing.T) {

	root := dcTestDir + strconv.Itoa(id)
	os.RemoveAll(root)
	os.MkdirAll(root, 0777)
	var cfg = &mockdc.Config{
		Key: []byte("mockdc"),
	}

	fmt.Println("mockdc run")
	svr, err := mockdc.New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	go svr.Run(addr, root)
}

func TestClient(t *testing.T) {

	go dcRun(":4461", 1, t)
	go dcRun(":4462", 2, t)
	time.Sleep(1e9)

	doTestClient(t)
	os.RemoveAll(dcTestDir + "1")
	os.RemoveAll(dcTestDir + "2")
}

func TestTimeout(t *testing.T) {

	xl := xlog.NewDummy()
	var sleepMs int
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		time.Sleep(time.Duration(sleepMs) * time.Millisecond)
	}))
	defer ts.Close()

	options := &TimeoutOptions{
		ClientMs: 200,
	}
	conns := []DCConn{{Keys: []string{"a"}, Host: ts.URL}}
	p := NewWithTimeout(conns, options)

	sleepMs = 0
	rc, _, err := p.Get(xl, []byte("test"))
	assert.NoError(t, err)
	_, err = io.Copy(ioutil.Discard, rc)
	assert.NoError(t, err)
	rc.Close()

	sleepMs = 210
	_, _, err = p.Get(xl, []byte("test"))
	assert.Error(t, err)

	options = &TimeoutOptions{
		DialMs: 200,
	}
	p = NewWithTimeout(conns, options)
	rc, _, err = p.Get(xl, []byte("test"))
	assert.NoError(t, err)
	_, err = io.Copy(ioutil.Discard, rc)
	assert.NoError(t, err)
	rc.Close()

	options = &TimeoutOptions{
		RespMs: 200,
	}
	p = NewWithTimeout(conns, options)
	_, _, err = p.Get(xl, []byte("test"))
	assert.Error(t, err)

	sleepMs = 0
	options = &TimeoutOptions{
		DialMs:   200,
		RespMs:   200,
		ClientMs: 200,
	}
	p = NewWithTimeout(conns, options)
	rc, _, err = p.Get(xl, []byte("test"))
	assert.NoError(t, err)
	_, err = io.Copy(ioutil.Discard, rc)
	assert.NoError(t, err)
	rc.Close()

	sleepMs = 1000
	options = &TimeoutOptions{
		RangegetSpeed: 1000,
		SetSpeed:      1000,
	}
	p = NewWithTimeout(conns, options)
	rc, _, err = p.RangeGet(xl, []byte("test"), 0, 10000)
	assert.NoError(t, err)
	_, err = io.Copy(ioutil.Discard, rc)
	assert.NoError(t, err)

	_, _, err = p.RangeGet(xl, []byte("test"), 0, 1)
	assert.Error(t, err) // timeout

	err = p.Set(xl, []byte("test"), io.LimitReader(rand.Reader, 10000), 10000)
	assert.NoError(t, err)

	err = p.Set(xl, []byte("test"), io.LimitReader(rand.Reader, 1), 1)
	assert.Error(t, err) // timeout
}
