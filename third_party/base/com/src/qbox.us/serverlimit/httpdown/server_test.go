package httpdown

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"bytes"
	"io"
	"sync"

	"github.com/facebookgo/httpdown"
	"github.com/stretchr/testify.v2/require"
)

/*
import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/facebookgo/httpdown"
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
	defer time.Sleep(1 * time.Second)
	mux := http.NewServeMux()
	mux.HandleFunc("/", handler)
	go ListenAndServe(&http.Server{Addr: "0.0.0.0:54647", Handler: mux}, &httpdown.HTTP{}, 3)
	var total uint32
	for i := 0; i < 6; i++ {
		var err error
		idx := i
		tr := &http.Transport{DisableKeepAlives: true}
		client := http.Client{Transport: tr}
		req, _ := http.NewRequest("GET", "http://0.0.0.0:54647", bytes.NewReader([]byte("aaa")))
		req.Header.Set("Host", strconv.Itoa(i))
		go func() {
			_, err = client.Do(req)
			if err == nil {
				atomic.AddUint32(&total, 1)
			}
			fmt.Println(idx, err)
		}()
	}
	time.Sleep(2 * time.Second)
	if total != 3 {
		t.Fatal("failed", total)
	}
	return
}
*/

func TestListenAndServeWithMsg(t *testing.T) {

	handler := func(w http.ResponseWriter, req *http.Request) {
		buf := make([]byte, 3)
		_, err := io.ReadFull(req.Body, buf)
		require.NoError(t, err)
		require.Equal(t, buf, []byte("aaa"))
		w.WriteHeader(206)
		return
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", handler)

	wg := sync.WaitGroup{}

	for try := 0; try < 5; try++ {
		wg.Add(1)

		done := make(chan struct{}, 1)
		notify := func() {
			done <- struct{}{}
		}

		s := httptest.NewServer(nil)
		url := s.URL
		port := strings.Split(url, ":")[2]
		s.Close()
		server := &http.Server{Addr: ":" + port, Handler: mux}

		go func() {
			<-done
			client := http.Client{}
			req, _ := http.NewRequest("GET", url, bytes.NewReader([]byte("aaa")))
			resp, err := client.Do(req)
			require.NoError(t, err)
			require.Equal(t, resp.StatusCode, 206)
			resp.Body.Close()
			wg.Done()

		}()

		go func() {
			ListenAndServeWithMsg(server, &httpdown.HTTP{}, 3, notify)
		}()
		wg.Wait()
	}
}
