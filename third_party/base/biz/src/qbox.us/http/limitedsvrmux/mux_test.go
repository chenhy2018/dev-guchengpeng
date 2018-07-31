package limitedsvrmux

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"qiniupkg.com/x/rpc.v7"
)

var handler = func(w http.ResponseWriter, r *http.Request) {
	time.Sleep(1 * time.Second)
	return
}

func TestMux(t *testing.T) {
	var apiHandleLimitMap = map[string]int{
		"/a|/|/b": 2,
		"/c":      1,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/a", handler)
	mux.HandleFunc("/b", handler)
	mux.HandleFunc("/c", handler)
	mux.HandleFunc("/", handler)
	limitedMux := NewServeMux(mux, apiHandleLimitMap)
	svr := httptest.NewServer(limitedMux)
	url := svr.URL
	client := rpc.Client{http.DefaultClient}
	go func() {
		resp, err := client.Get(url + "/a")
		assert.NoError(t, err)
		assert.Equal(t, resp.StatusCode, 200)
	}()
	go func() {
		resp, err := client.Get(url + "/")
		assert.NoError(t, err)
		assert.Equal(t, resp.StatusCode, 200)
	}()
	time.Sleep(100 * time.Millisecond)
	go func() {
		resp, _ := client.Get(url + "/b")
		assert.Equal(t, resp.StatusCode, 503)
	}()
	go func() {
		resp, err := client.Get(url + "/c")
		assert.NoError(t, err)
		assert.Equal(t, resp.StatusCode, 200)
	}()
	time.Sleep(100 * time.Millisecond)
	resp, _ := client.Get(url + "/c")
	assert.Equal(t, resp.StatusCode, 503)
}
