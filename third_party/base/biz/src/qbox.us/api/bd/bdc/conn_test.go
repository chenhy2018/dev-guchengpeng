package bdc

import (
	"bytes"
	"crypto/rand"
	"crypto/sha1"
	"fmt"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"qbox.us/auditlog2"
	"qbox.us/cc"
	"qbox.us/mockbd"

	"github.com/qiniu/xlog.v1"

	qbytes "github.com/qiniu/bytes"
)

func setupSimpleBd() (url string) {
	cfg := &mockbd.Config{
		Addr: "0.0.0.0",
		Dirs: []string{"/tmp/bd_test"},
		LogCfg: &auditlog2.Config{
			LogFile: "/tmp/audit_sbd",
		},
	}
	mux, err := mockbd.Init(cfg)
	if err != nil {
		return
	}

	ts := httptest.NewServer(mux)
	url = ts.URL
	port := strings.Split(url, ":")[2]
	cfg.Addr = "0.0.0.0:" + port
	tmp := os.TempDir()
	d := tmp + "/" + port
	os.Mkdir(d, 0755)
	fmt.Println("dir:", d)
	return
}

func genKey(d []byte) []byte {
	h := sha1.New()
	h.Write(d)
	return h.Sum([]byte{})
}

func TestBdConns(t *testing.T) {
	fmt.Print("")
	url := setupSimpleBd()
	if url == "" {
		t.Fatal("setup simple bd failed: empty url")
	}
	client := NewConn(url, nil)
	// simple test
	{
		d1 := []byte{'a', 'b', 'c'}

		key, err := doTestPut(client, d1, t, nil)
		if err != nil {
			t.Fatal(err)
		}

		d2, err := doTestGet(client, key, 0, len(d1), t)
		if err != nil {
			t.Fatal(err)
		}
		if !checkData(d1, d2) {
			t.Fatal("data check faild!")
		}
	}

	// 4M test
	{
		d1 := make([]byte, 1024*1024*4)
		rand.Read(d1)

		key, err := doTestPut(client, d1, t, nil)
		if err != nil {
			t.Fatal(err)
		}

		d2, err := doTestGet(client, key, 0, len(d1), t)
		if err != nil {
			t.Fatal(err)
		}
		if !checkData(d1, d2) {
			t.Fatal("data check faild!")
		}
	}

	// >4M test
	{
		d1 := make([]byte, 1024*1024*4+1)
		rand.Read(d1)

		_, err := doTestPut(client, d1, t, nil)
		if err == nil {
			//t.Fatal("test large then 4M faild")
			// bd可以没有这个限制
		}
	}

	// verified key test
	{
		d1 := make([]byte, 100)
		rand.Read(d1)

		key, err := doTestPut(client, d1, t, nil)
		if err != nil {
			t.Fatal(err)
		}

		_, err = doTestPut(client, d1, t, key)
		if err != nil {
			t.Fatal(err)
		}

		wrongKey := make([]byte, len(key))
		copy(wrongKey, key)
		wrongKey[0] = wrongKey[0] + 1
		_, err = doTestPut(client, d1, t, wrongKey)
		if err == nil {
			t.Fatal("test wrong key faild!")
		}
		fmt.Println(err, "(that`s OK)")

		d2, err := doTestGet(client, key, 0, len(d1), t)
		if err != nil {
			t.Fatal(err)
		}
		if !checkData(d1, d2) {
			t.Fatal("data check faild!")
		}
	}

	// "from to" test
	{
		d1 := make([]byte, 10)
		rand.Read(d1)

		key, err := doTestPut(client, d1, t, nil)
		if err != nil {
			t.Fatal(err)
		}

		d2, err := doTestGet(client, key, 2, 8, t)
		if err != nil {
			t.Fatal(err)
		}
		if !checkData(d1[2:8], d2) {
			t.Fatal("data check faild!")
		}
	}

	// out of range test
	{
		d1 := make([]byte, 8)
		rand.Read(d1)

		key, err := doTestPut(client, d1, t, nil)
		if err != nil {
			t.Fatal(err)
		}

		d2, err := doTestGet(client, key, 2, 10, t)
		if err != nil {
			t.Fatal(err)
		}
		if !checkData(d1[2:], d2) {
			t.Fatal("data check faild!")
		}
	}

	// not found test
	{
		key := make([]byte, 10)
		_, err := doTestGet(client, key, 0, 1, t)
		if err != EKeyNotFound {
			t.Fatal(err)
		}
	}
}

func doTestPut(client *Conn, data []byte, t *testing.T, verifiedKey []byte) (key []byte, err error) {
	r := cc.NewBytesReader(data)
	if verifiedKey == nil {
		verifiedKey = genKey(data)
	}
	key, err = client.Put(xlog.NewDummy(), r, len(data), verifiedKey, [3]uint16{0, 0xffff})
	return
}

func doTestGet(client *Conn, key []byte, from, to int, t *testing.T) (data []byte, err error) {
	data = make([]byte, to-from)
	w := qbytes.NewWriter(data)
	_, err = client.Get(xlog.NewDummy(), key, w, from, to, [4]uint16{0, 0xffff})
	return w.Bytes(), err
}

func checkData(d1, d2 []byte) bool {
	return bytes.Equal(d1, d2)
}
