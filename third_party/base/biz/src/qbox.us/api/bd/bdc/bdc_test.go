package bdc

import (
	"bytes"
	"crypto/rand"
	"crypto/sha1"
	"fmt"
	"net/http"
	"net/http/httptest"
	"runtime"
	"strconv"
	"testing"

	"github.com/qiniu/http/formutil.v1"
	"github.com/qiniu/xlog.v1"

	qbytes "github.com/qiniu/bytes"
)

var testUrl string

func _TestPickClient(t *testing.T) {
	fmt.Print("")
	bdc := getBdc()
	conn, _ := bdc.pickClient(xlog.NewDummy(), 0)
	if conn == nil {
		t.Fatal("pick client nil")
	}
}

type BdcForm struct {
	From int `json:"from"`
	To   int `json:"to"`
}

func getRetryBdc(url string) *BdClient {
	clients := [...]*Conn{
		NewConn(url, nil),
		NewConn(url, nil),
	}
	return NewBdClient(clients[:], 10, 3)
}

func TestGetFailed(t *testing.T) {
	// go1.1.2 has a bug: https://code.google.com/p/go/issues/detail?id=5738
	if runtime.Version() == "go1.1.2" {
		return
	}
	getFunc := func(w http.ResponseWriter, r *http.Request) {
		var ret BdcForm
		r.ParseForm()
		err := formutil.Parse(&ret, r.Form)
		if err != nil {
			t.Fatal(err)
		}
		length := ret.To - ret.From
		w.Header().Set("Content-Length", strconv.Itoa(length))
		if length > 1 {
			length -= 1
		}
		w.WriteHeader(200)
		w.Write(bytes.Repeat([]byte("A"), length))
	}
	ts := httptest.NewServer(http.HandlerFunc(getFunc))
	defer ts.Close()

	length := 10
	client := getRetryBdc(ts.URL)
	data, err := doTestGet2(client, []byte("xxx"), 0, length, t)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) != length {
		t.Fatalf("length should be %d, but get %d", length, len(data))
	}
}

func TestBdc(t *testing.T) {
	testUrl = setupSimpleBd()
	client := getBdc()

	// simple test
	{
		d1 := []byte{'a', 'b', 'c'}

		key, err := doTestRetryPut2(client, d1, t)
		if err != nil {
			t.Fatal(err)
		}

		d2, err := doTestRetryGet2(client, key, 0, len(d1), t)
		if err != nil {
			t.Fatal(err)
		}
		if !checkData(d1, d2) {
			t.Fatal("[simple test]data check faild!")
		}
	}

	// 4M test
	{
		d1 := make([]byte, 1024*1024*4)
		rand.Read(d1)

		key, err := doTestRetryPut2(client, d1, t)
		if err != nil {
			t.Fatal(err)
		}

		d2, err := doTestRetryGet2(client, key, 0, len(d1), t)
		if err != nil {
			t.Fatal(err)
		}
		if !checkData(d1, d2) {
			t.Fatal("[4M test]data check faild!")
		}
	}

	// >4M test
	// bd可以没有这个限制
	if false {
		d1 := make([]byte, 1024*1024*4+1)
		rand.Read(d1)

		_, err := doTestRetryPut2(client, d1, t)
		if err == nil {
			t.Fatal("test large then 4M faild")
		}
	}

	// verified key test
	{
		d1 := make([]byte, 100)
		rand.Read(d1)

		key, err := doTestRetryPut2(client, d1, t)
		if err != nil {
			t.Fatal(err)
		}

		_, err = doTestVerifiedRetryPut(client, d1, t, key)
		if err != nil {
			t.Fatal(err)
		}

		wrongKey := make([]byte, len(key))
		copy(wrongKey, key)
		wrongKey[0] = wrongKey[0] + 1
		_, err = doTestVerifiedRetryPut(client, d1, t, wrongKey)
		if err == nil {
			t.Fatal("test wrong key faild!")
		}
		fmt.Println(err, "(that`s OK)")

		d2, err := doTestRetryGet2(client, key, 0, len(d1), t)
		if err != nil {
			t.Fatal(err)
		}
		if !checkData(d1, d2) {
			t.Fatal("[verified key test]data check faild!")
		}
	}

	// "from to" test
	{
		d1 := make([]byte, 10)
		rand.Read(d1)

		key, err := doTestRetryPut2(client, d1, t)
		if err != nil {
			t.Fatal(err)
		}

		d2, err := doTestRetryGet2(client, key, 2, 8, t)
		if err != nil {
			t.Fatal(err)
		}
		if !checkData(d1[2:8], d2) {
			t.Fatal("[from to test]data check faild!")
		}
	}

	// can not retry test
	for i := 0; i < 3; i++ {
		d1 := []byte{'a', 'b', 'c'}

		key, err := doTestPut2(client, d1, t)
		if err != nil {
			fmt.Println(err)
			continue
		}

		_, err = doTestGet2(client, key, 0, len(d1), t)
		if err != nil {
			fmt.Println(err)
			continue
		}
	}
}

// it`s slow!
func _TestSchedule(t *testing.T) {
	clients := [...]*Conn{
		NewConn(testUrl, nil),
		//NewConn("http://localhost:8007", nil),
		//NewConn("http://localhost:8008", nil),
	}

	f := func(c chan int) {
		client := NewBdClient(clients[:], 10, 3)
		for i := 0; i < 1000000; i++ {
			d1 := make([]byte, 3)
			rand.Read(d1)

			key, err := doTestPut2(client, d1, t)
			if err != nil {
				t.Fatal(err)
			}

			d2, err := doTestGet2(client, key, 0, len(d1), t)
			if err != nil {
				t.Fatal(err)
			}
			if !checkData(d1, d2) {
				t.Fatal("data check faild!")
			}
		}
		c <- 0
	}

	l := make([]chan int, 10)
	for i := 0; i < len(l); i++ {
		l[i] = make(chan int, 0)
		go f(l[i])
	}
	for i := 0; i < len(l); i++ {
		<-l[i]
	}
}

func getBdc() *BdClient {
	clients := [...]*Conn{
		NewConn(testUrl, nil),
		//NewConn("http://localhost:8007", nil),
		//NewConn("http://localhost:8008", nil),
	}
	return NewBdClient(clients[:], 10, 3)
}

func doTestPut2(client *BdClient, data []byte, t *testing.T) (key []byte, err error) {
	key2 := sha1.New()
	key2.Write(data)
	err = client.PutEx(xlog.NewDummy(), key2.Sum(nil), bytes.NewReader(data), len(data), [3]uint16{0, 0xffff})
	key = key2.Sum(nil)
	return
}

func doTestVerifiedPut(client *BdClient, data []byte, t *testing.T, verifiedKey []byte) (key []byte, err error) {
	key2 := sha1.New()
	key2.Write(data)
	err = client.PutEx(xlog.NewDummy(), key2.Sum(nil), bytes.NewReader(data), len(data), [3]uint16{0, 0xffff})
	key = key2.Sum(nil)
	return
}

func doTestGet2(client *BdClient, key []byte, from, to int, t *testing.T) (data []byte, err error) {
	buff := make([]byte, to-from)
	w := qbytes.NewWriter(buff)
	err = client.Get(xlog.NewDummy(), key, w, from, to, [4]uint16{0, 0xffff})
	data = w.Bytes()
	return
}

func doTestRetryPut2(client *BdClient, data []byte, t *testing.T) (key []byte, err error) {
	key2 := sha1.New()
	key2.Write(data)
	err = client.PutEx(xlog.NewDummy(), key2.Sum(nil), bytes.NewReader(data), len(data), [3]uint16{0, 0xffff})
	key = key2.Sum(nil)
	return
}

func doTestVerifiedRetryPut(client *BdClient, data []byte, t *testing.T, verifiedKey []byte) (key []byte, err error) {
	err = client.PutEx(xlog.NewDummy(), verifiedKey, bytes.NewReader(data), len(data), [3]uint16{0, 0xffff})
	return
}

func doTestRetryGet2(client *BdClient, key []byte, from, to int, t *testing.T) (data []byte, err error) {
	buff := make([]byte, to-from)
	w := qbytes.NewWriter(buff)
	err = client.Get(xlog.NewDummy(), key, w, from, to, [4]uint16{0, 0xffff})
	data = w.Bytes()
	return
}
