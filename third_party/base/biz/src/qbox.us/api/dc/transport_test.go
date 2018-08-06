package dc

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/qiniu/log.v1"
	"github.com/stretchr/testify.v1/require"
)

func TestTransportClose(t *testing.T) {

	testTransportClose(t, func(t *Transport) { t.Transport.CloseIdleConnections() }, true)
	testTransportClose(t, func(t *Transport) { t.Close() }, false)
}

func testTransportClose(t *testing.T, closeFn func(*Transport), shouldFdLeak bool) {
	ts := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
	}))

	var lock sync.Mutex
	tr := &Transport{Transport: &http.Transport{}}
	client := &http.Client{Transport: tr}

	doGet := make(chan struct{})
	go func() {
		for {
			lock.Lock()
			c := client
			lock.Unlock()
			<-doGet
			resp, err := c.Get(ts.URL)
			require.NoError(t, err)
			io.Copy(ioutil.Discard, resp.Body)
			resp.Body.Close()
		}
	}()

	var oldFds int
	var newFds int
	for i := 0; i < 10; i++ {
		newTr := &Transport{Transport: &http.Transport{}}

		lock.Lock()
		client = &http.Client{Transport: newTr}
		lock.Unlock()

		closeFn(tr)
		doGet <- struct{}{}

		time.Sleep(20 * time.Millisecond)
		out, err := exec.Command("/bin/sh", "-c", fmt.Sprintf("lsof -n -p %v", os.Getpid())).Output()
		require.NoError(t, err)

		newFds = len(strings.Split(string(out), "\n")) - 1
		if oldFds == 0 {
			oldFds = newFds
		}

		tr = newTr
	}
	if !shouldFdLeak {
		// 2 is for accident fd recycle delay.
		require.True(t, oldFds+2 >= newFds, "old %v new %v", oldFds, newFds)
	} else {
		if oldFds+2 >= newFds {
			log.Warnf("authrpc: fds old %v new %v, it seems no fd leak now!", oldFds, newFds)
		}
	}
}
