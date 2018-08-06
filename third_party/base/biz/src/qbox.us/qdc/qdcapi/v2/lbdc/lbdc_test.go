package lbdc

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"runtime"
	"strconv"
	"testing"

	"qbox.us/dht"

	"github.com/qiniu/http/formutil.v1"
	"github.com/qiniu/xlog.v1"

	qbytes "github.com/qiniu/bytes"
)

func TestRetryGet(t *testing.T) {
	// go1.1.2 has a bug: https://code.google.com/p/go/issues/detail?id=5738
	if runtime.Version() == "go1.1.2" {
		return
	}
	tse := getRetrySev(t)
	hosts := []string{tse.URL, tse.URL}
	client, err := getLBdClient(hosts, 0, 2)
	if err != nil {
		t.Fatal(err)
	}

	length := 10
	data, err := doLBdcGet(client, []byte("xxx"), 0, length, t)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) != length {
		t.Fatalf("should reveive %d bytes, but get %d", length, len(data))
	}
}

func doLBdcGet(client *LBdClient, key []byte, from, to int, t *testing.T) (data []byte, err error) {
	buff := make([]byte, to-from)
	w := qbytes.NewWriter(buff)
	err = client.Get(xlog.NewDummy(), key, w, from, to, [4]uint16{0, 0xffff})
	data = w.Bytes()
	return
}

func getLBdClient(hosts []string, retryInterval int64, retryTimes int) (stg *LBdClient, err error) {
	nodes := make([]dht.NodeInfo, 0)
	bds := make(map[string]*Conn)
	for _, host := range hosts {
		bds[host] = NewConn(host, nil)
		nodes = append(nodes, dht.NodeInfo{host, []byte("key")})
	}
	d := dht.NewCarp(nodes)
	acl := NewAcl(&AclConfig{
		MaxBdCount:  1,
		MaxLbdCount: 1,
		MaxIpCount:  1,
		MaxIdcCount: 1,
	})
	return New(d, bds, retryInterval, retryTimes, acl), nil
}

type BdcForm struct {
	From int `json:"from"`
	To   int `json:"to"`
}

func getRetrySev(t *testing.T) *httptest.Server {
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
		t.Log(ret, r.Form)
		w.WriteHeader(200)
		w.Write(bytes.Repeat([]byte("A"), length))
	}
	return httptest.NewServer(http.HandlerFunc(getFunc))
}
