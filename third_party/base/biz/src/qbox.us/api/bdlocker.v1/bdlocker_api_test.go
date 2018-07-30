package bdlocker

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/qiniu/log.v1"
	xlog "github.com/qiniu/xlog.v1"
	"github.com/stretchr/testify.v1/require"
)

var lock = sync.Mutex{}

var map1 = map[string]bool{}
var map2 = map[string]bool{}

var sleep1 = 0
var sleep2 = 0

func lock1(w http.ResponseWriter, r *http.Request) {
	log.Info("lock1", r.URL.Path)
	time.Sleep(time.Duration(sleep1) * time.Millisecond)
	lock.Lock()
	map1[r.URL.Path[len("/lock/"):]] = true
	lock.Unlock()
}

func lock2(w http.ResponseWriter, r *http.Request) {
	log.Info("lock2", r.URL.Path)
	time.Sleep(time.Duration(sleep2) * time.Millisecond)
	lock.Lock()
	map2[r.URL.Path[len("/lock/"):]] = true
	lock.Unlock()
}

func unlock1(w http.ResponseWriter, r *http.Request) {
	log.Info("unlock1", r.URL.Path)
	time.Sleep(time.Duration(sleep1) * time.Millisecond)
	lock.Lock()
	delete(map1, r.URL.Path[len("/unlock/"):])
	lock.Unlock()
}

func unlock2(w http.ResponseWriter, r *http.Request) {
	log.Info("unlock2", r.URL.Path)
	time.Sleep(time.Duration(sleep2) * time.Millisecond)
	lock.Lock()
	delete(map2, r.URL.Path[len("/unlock/"):])
	lock.Unlock()
}

func exist1(w http.ResponseWriter, r *http.Request) {
	log.Info("exist1", r.URL.Path)
	time.Sleep(time.Duration(sleep1) * time.Millisecond)
	lock.Lock()
	_, ok := map1[r.URL.Path[len("/exist/"):]]
	lock.Unlock()
	if ok {
		w.WriteHeader(200)
	} else {
		w.WriteHeader(612)
	}
}

func exist2(w http.ResponseWriter, r *http.Request) {
	log.Info("exist2", r.URL.Path)
	time.Sleep(time.Duration(sleep2) * time.Millisecond)
	lock.Lock()
	_, ok := map2[r.URL.Path[len("/exist/"):]]
	lock.Unlock()
	if ok {
		w.WriteHeader(200)
	} else {
		w.WriteHeader(612)
	}
}

func TestLockerApi(t *testing.T) {
	log.SetOutputLevel(0)
	var mux1 = http.NewServeMux()
	var mux2 = http.NewServeMux()
	mux1.HandleFunc("/lock/", lock1)
	mux1.HandleFunc("/unlock/", unlock1)
	mux1.HandleFunc("/exist/", exist1)
	mux2.HandleFunc("/lock/", lock2)
	mux2.HandleFunc("/unlock/", unlock2)
	mux2.HandleFunc("/exist/", exist2)
	s1 := httptest.NewServer(mux1)
	s2 := httptest.NewServer(mux2)
	log.Info("s1:", s1.URL)
	log.Info("s2:", s2.URL)

	xl := xlog.NewDummy()
	var fh = func(i int) []byte { return []byte(fmt.Sprint("fh", i)) }
	fh1 := fh(1)

	cli := NewClient(Config{
		Hosts:              []string{s1.URL, s2.URL},
		DialTimeoutMs:      100,
		RespTimeoutMs:      100,
		RetryTimeoutMs:     100,
		MaxIdleConnPerHost: 5,
	})
	for i := 2; i < 30; i++ {
		fh1 = fh(i)
		err := cli.Lock(xl, fh1)
		require.NoError(t, err)

		exist, err := cli.Exist(xl, fh(i))
		require.NoError(t, err)
		require.True(t, exist)
		// UnLock
		err = cli.UnLock(xl, fh1)
		require.NoError(t, err)
		exist, err = cli.Exist(xl, fh(i))
		require.NoError(t, err)
		require.False(t, exist)
	}

	cli = NewClient(Config{
		Hosts:              []string{s1.URL, s2.URL},
		DialTimeoutMs:      40,
		RespTimeoutMs:      40,
		RetryTimeoutMs:     40,
		MaxIdleConnPerHost: 5,
	})

	err := cli.Lock(xl, fh1)
	require.NoError(t, err)

	exist, err := cli.Exist(xl, fh1)
	require.NoError(t, err)
	require.True(t, exist)

	sleep1 = 60

	for i := 2; i < 30; i++ {
		fh1 = fh(i)
		err = cli.Lock(xl, fh1)
		require.NoError(t, err)

		exist, err = cli.Exist(xl, fh(i))
		lock.Lock()
		_, ok := map2[base64.URLEncoding.EncodeToString(fh(i))]
		lock.Unlock()
		if ok {
			require.NoError(t, err)
			require.True(t, exist)
		} else {
			require.Error(t, err)
		}
	}

	for i := 100; i < 130; i++ {
		exist, err = cli.Exist(xl, fh(i))
		require.Error(t, err)
	}

	s1.Close()

	for i := 202; i < 230; i++ {
		fh1 = fh(i)
		err = cli.Lock(xl, fh1)
		require.NoError(t, err)

		exist, err = cli.Exist(xl, fh(i))
		lock.Lock()
		_, ok := map2[base64.URLEncoding.EncodeToString(fh(i))]
		lock.Unlock()
		if ok {
			require.NoError(t, err)
			require.True(t, exist)
		} else {
			require.Error(t, err)
		}
	}

	for i := 300; i < 330; i++ {
		exist, err = cli.Exist(xl, fh(i))
		require.Error(t, err)
	}

	sleep2 = 100
	err = cli.Lock(xl, fh(1000))
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "timeout"))
}
