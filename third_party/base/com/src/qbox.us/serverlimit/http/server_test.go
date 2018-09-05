package http

/*
import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func handler(w http.ResponseWriter, req *http.Request) {
	fmt.Println("one")
	time.Sleep(1 * time.Second)
	return
}

func TestListenAndServe(t *testing.T) {
	s := httptest.NewServer(nil)
	url := s.URL
	port := strings.Split(url, ":")[2]
	fmt.Println(port)
	s.Close()
	time.Sleep(1 * time.Second)
	mux := http.NewServeMux()
	mux.HandleFunc("/", handler)
	go ListenAndServe("0.0.0.0:"+port, mux, 3)
	var total uint32
	for i := 0; i < 6; i++ {
		var err error
		tr := &http.Transport{DisableKeepAlives: true}
		client := http.Client{Transport: tr}
		req, _ := http.NewRequest("GET", "http://0.0.0.0:"+port, nil)
		req.Header.Set("Host", strconv.Itoa(i))
		go func() {
			_, err = client.Do(req)
			if err == nil {
				atomic.AddUint32(&total, 1)
			}
		}()
	}
	time.Sleep(2 * time.Second)
	if total != 3 {
		t.Fatal("failed", total)
	}
	return
}
*/
