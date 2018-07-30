package clients

import (
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/qiniu/log.v1"
	"github.com/qiniu/xlog.v1"
	"qbox.us/fop"
	"qbox.us/fop/mock"
)

var (
	ts1 *httptest.Server
	ts2 *httptest.Server
	ts3 *httptest.Server
	ts4 *httptest.Server
)

func init() {
	log.SetOutputLevel(0)
}

func TestShuffleFopds(t *testing.T) {
	conns := []*FopdConn{
		&FopdConn{host: "localhost:12306"},
		&FopdConn{host: "localhost"},
		&FopdConn{host: "xxxx:12306"},
		&FopdConn{host: "yyyy:12306"},
		&FopdConn{host: "localhost:12306"},
		&FopdConn{host: "yyyy:12306"},
		&FopdConn{host: "xxxx:12306"},
		&FopdConn{host: "localhost:12306"},
		&FopdConn{host: "localhost:12306"},
	}
	shuffleFopds(conns)
	for _, c := range conns {
		log.Info(c.host)
	}
	for i := 1; i <= 6; i++ {
		if getIp(conns[i-1].host) == getIp(conns[i].host) {
			t.Error(i-1, i)
		}
	}
}

const FailRetryInterval = 2

func TestFopdClient(t *testing.T) {
	ts1 = httptest.NewServer(mock.Fopd{}.Mux())
	defer ts1.Close()
	ts2 = httptest.NewServer(mock.Fopd{}.Mux())
	defer ts2.Close()
	ts3 = httptest.NewServer(mock.Fopd{}.Mux())
	defer ts3.Close()
	ts4 = httptest.NewServer(mock.Fopd{}.Mux())
	defer ts4.Close()
	var fopdServer = []FopdServer{
		FopdServer{Host: "http://localhost:20040", Cmds: []string{"lb0", "lb1"}},
		FopdServer{Host: ts1.URL, Cmds: []string{"lb0", "lb1"}},
		FopdServer{Host: ts2.URL, Cmds: []string{"lb0", "lb1"}},
		FopdServer{Host: ts3.URL, Cmds: []string{"lb0", "lb1"}},
		FopdServer{Host: ts4.URL, Cmds: []string{"lb0", "lb1"}},
	}

	cfg := &FopdConfig{
		Servers: fopdServer,
		LoadBalanceMode: map[string]int{
			"lb1": 1,
		},
		FailRetryInterval: FailRetryInterval,
	}
	client := NewFopd(cfg, nil)

	log.Info("\nTest LB0")
	fopCtx := &fop.FopCtx{RawQuery: "lb0/foo/bar"}
	for i := 0; i < 10; i++ {
		go func() {
			xl := xlog.NewDummy()
			if _, err := client.Op(xl, strings.NewReader("test_lb0"), 8, fopCtx); err != nil {
				xl.Warn("lb0 op failed", err)
			}
		}()
		time.Sleep(time.Duration(100) * time.Millisecond)
	}
	fmt.Println("\nStatus:\n", string(client.Status()))

	keys := []string{"lb0|http://localhost:20040"}
	fmt.Printf("\nBegin Clear Status\n\n")
	client.ClearStatus(keys)

	fmt.Println("\nStatus:\n", string(client.Status()))

	log.Info("\nTest LB1")
	fopCtx = &fop.FopCtx{RawQuery: "lb1/foo/bar"}
	for i := 0; i < 20; i++ {
		go func() {
			xl := xlog.NewDummy()
			if _, err := client.Op(xl, strings.NewReader("test_lb1"), 8, fopCtx); err != nil {
				xl.Warn("lb1 op failed", err)
			}
		}()
		time.Sleep(time.Duration(200) * time.Millisecond)
		if i == 10 {
			fmt.Println("\nStatus:\n", string(client.Status()))
		}
	}
	fmt.Println("\nStatus:\n", string(client.Status()))
}

func TestMinLoadIndex(t *testing.T) {
	times := map[int]int64{}
	cases := []int64{1, 2, 1, 2, 1}
	for i := 0; i < 1000; i++ {
		index := minLoadIndex(cases)
		times[index]++
	}

	// index only can be 0,2,4
	// index value must > 250
	for index, v := range times {
		if index == 1 || index == 3 {
			t.Fatalf("index got %v, expect 0,2,4\n", index)
		}
		if v < 250 {
			t.Fatalf("not random selection, got %v, expect v > 250\n", v)
		}
	}
}
