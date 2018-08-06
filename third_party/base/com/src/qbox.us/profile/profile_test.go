package profile

import (
	"io/ioutil"
	"net"
	"net/http"
	"testing"

	"github.com/stretchr/testify.v1/require"
	"qbox.us/profile/expvar"
)

func TestProfile(t *testing.T) {
	i := expvar.NewInt("test_key")
	i.Set(100)
	var defaultHandleAddr string
	initDone := make(chan bool)
	go func() {
		ln, err := net.Listen("tcp", "127.0.0.1:0")
		require.NoError(t, err)
		defaultHandleAddr = "http://" + ln.Addr().String()
		close(initDone)
		err = http.Serve(ln, nil)
		require.NoError(t, err)
	}()
	<-initDone
	// default handle 404
	resp, err := http.Get(defaultHandleAddr + "/debug/vars")
	require.NoError(t, err)
	require.Equal(t, 404, resp.StatusCode)
	resp, err = http.Get(defaultHandleAddr + "/debug/pprof/")
	require.NoError(t, err)
	require.Equal(t, 404, resp.StatusCode)
	// profile handle 200
	resp, err = http.Get(GetProfileAddr() + "/debug/vars")
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)
	resp, err = http.Get(GetProfileAddr() + "/debug/pprof/")
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)
	// get one expvar
	resp, err = http.Get(GetProfileAddr() + "/debug/var/not_exist_key")
	require.NoError(t, err)
	require.Equal(t, 612, resp.StatusCode)
	resp, err = http.Get(GetProfileAddr() + "/debug/var/test_key")
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)
	data, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, "100", string(data))
}
