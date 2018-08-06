package dcreload

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"qbox.us/api/dc"
	"qbox.us/mockdc"

	"github.com/qiniu/xlog.v1"
	"github.com/stretchr/testify/assert"
)

func runDC(addr string) (svr *httptest.Server) {
	dc, _ := mockdc.New(nil)
	dc.Dir = os.TempDir()
	mux := http.NewServeMux()
	dc.RegisterHandlers(mux)
	svr = httptest.NewServer(mux)
	return svr
}

func TestClient(t *testing.T) {
	// run dc
	s1 := runDC(":33334")
	s2 := runDC(":33332")
	s3 := runDC(":33333")
	defer s1.Close()
	defer s2.Close()
	defer s3.Close()

	dcHostsCfg1 := []dc.DCConn{
		{Keys: []string{"host1"}, Host: s1.URL},
	}

	// cfg
	dcclient := dc.NewClient(dcHostsCfg1, 1, nil)
	client := New(dcclient)

	xl := xlog.NewWith("TestClient")

	host, err := client.SetWithHostRet(xl, []byte("key"), strings.NewReader("value"), 5)
	assert.NoError(t, err)
	assert.Equal(t, host, s1.URL)

	host, err = client.KeyHost(xl, []byte("key"))
	assert.NoError(t, err)
	assert.Equal(t, host, s1.URL)

	var dcHostsCfg2 = []dc.DCConn{
		{Keys: []string{"host1"}, Host: s1.URL},
		{Keys: []string{"host2"}, Host: s2.URL},
		{Keys: []string{"host3"}, Host: s3.URL},
	}

	dcclient = dc.NewClient(dcHostsCfg2, 1, nil)
	client = New(dcclient)

	host, err = client.SetWithHostRet(xl, []byte("key"), strings.NewReader("value"), 5)
	assert.NoError(t, err)
	assert.Equal(t, host, s3.URL)

	host, err = client.KeyHost(xl, []byte("key"))
	assert.NoError(t, err)
	assert.Equal(t, host, s3.URL)

	dcclient = dc.NewClient(dcHostsCfg1, 1, nil)
	client = New(dcclient)

	host, err = client.SetWithHostRet(xl, []byte("key"), strings.NewReader("value"), 5)
	assert.NoError(t, err)
	assert.Equal(t, host, s1.URL)

	host, err = client.KeyHost(xl, []byte("key"))
	assert.NoError(t, err)
	assert.Equal(t, host, s1.URL)
}
