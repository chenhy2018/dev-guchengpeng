package conn

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"io"

	"bytes"

	"github.com/stretchr/testify.v1/require"
)

func TestTimeoutClientConn(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/test1", func(w http.ResponseWriter, req *http.Request) {
		time.Sleep(500 * time.Millisecond)
		return
	})
	mux.HandleFunc("/test2", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("haha"))
		w.(http.Flusher).Flush()
		time.Sleep(500 * time.Millisecond)
		w.Write([]byte("haha"))
		return
	})
	mux.HandleFunc("/test3", func(w http.ResponseWriter, req *http.Request) {
		p := make([]byte, 512<<13)
		var err error
		var total, n int
		for {
			n, err = req.Body.Read(p)
			total = total + n
			if err != nil {
				break
			}
		}
		require.Equal(t, err, io.EOF)
		require.Equal(t, total, 512<<13)
		w.Write(p)
		return

	})
	svr := httptest.NewServer(mux)
	defer svr.Close()

	tr := &http.Transport{
		Dial: NewDailer(&net.Dialer{
			Timeout: 200 * time.Millisecond,
		}, 400*time.Millisecond).Dial,
	}

	req, err := http.NewRequest("POST", svr.URL+"/test1", nil)
	require.NoError(t, err)
	_, err = tr.RoundTrip(req)
	require.Error(t, err)
	operr, ok := err.(*net.OpError)
	require.True(t, ok)
	require.True(t, operr.Timeout())

	req, err = http.NewRequest("POST", svr.URL+"/test2", nil)
	require.NoError(t, err)
	resp, err := tr.RoundTrip(req)
	require.NoError(t, err)
	for {
		data := make([]byte, 4096)
		_, err = resp.Body.Read(data)
		if err != nil {
			break
		}
		require.Equal(t, string(data[:4]), "haha")
	}
	require.Error(t, err)
	operr, ok = err.(*net.OpError)
	require.True(t, ok)
	require.True(t, operr.Timeout())

	tr2 := &http.Transport{
		Dial: (&testDial{t, NewDailer(&net.Dialer{
			Timeout: 200 * time.Millisecond,
		}, 400*time.Millisecond)}).Dial,
	}
	data := make([]byte, 512<<13)
	req, err = http.NewRequest("POST", svr.URL+"/test3", bytes.NewReader(data))
	require.NoError(t, err)
	resp, err = tr2.RoundTrip(req)
	require.NoError(t, err)
	resp.Body.Close()
}

type testDial struct {
	t *testing.T
	Dialer
}

func (d *testDial) Dial(network, address string) (net.Conn, error) {
	c, err := d.Dialer.Dial(network, address)
	return &testConn{d.t, c}, err
}

type testConn struct {
	t *testing.T
	net.Conn
}

// golang 客户端每次发送4096bytes以内的数据
func (c *testConn) Read(b []byte) (count int, e error) {
	count, e = c.Conn.Read(b)
	if e != nil {
		require.True(c.t, count < 4096)
	} else {
		require.True(c.t, count == 4096)
	}
	return
}

// 服务端每次返回512bytes的数据, 然后用chunkedReader作拼接, 如果有数据,必然返回512的整数倍。如果没有数据,返回512以内。
func (c *testConn) Write(b []byte) (count int, e error) {
	count, e = c.Conn.Write(b)
	if e != nil && count > 512 {
		require.True(c.t, count%512 == 0)
	}
	return
}

type slowReader struct {
	times int
}

func (r *slowReader) Read(p []byte) (n int, error error) {
	r.times++
	if r.times == 4 {
		time.Sleep(500 * time.Millisecond)
	}
	data := make([]byte, len(p))
	p = data
	return len(p), nil
}

func TestTimeoutListener(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/test1", func(w http.ResponseWriter, req *http.Request) {
		p := make([]byte, 512<<13)
		var err error
		for {
			_, err = req.Body.Read(p)
			if err != nil {
				break
			}
		}
		require.Error(t, err)
		operr, ok := err.(*net.OpError)
		require.True(t, ok)
		require.True(t, operr.Timeout())
		w.WriteHeader(200)
		return
	})
	mux.HandleFunc("/test2", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("haha"))
		w.(http.Flusher).Flush()
		time.Sleep(500 * time.Millisecond)
		w.Write([]byte("haha"))
		return
	})

	svr := httptest.NewUnstartedServer(mux)
	svr.Listener = &TimeoutListener{svr.Listener, 400 * time.Millisecond}
	svr.Start()
	defer svr.Close()
	req, err := http.NewRequest("POST", svr.URL+"/test1", &slowReader{})
	req.ContentLength = 32768 << 3
	require.NoError(t, err)
	_, err = http.DefaultTransport.RoundTrip(req)
	require.Error(t, err)
	require.Equal(t, err, io.EOF)

	req, err = http.NewRequest("POST", svr.URL+"/test2", nil)
	require.NoError(t, err)
	resp, err := http.DefaultTransport.RoundTrip(req)
	require.NoError(t, err)
	for {
		data := make([]byte, 4096)
		_, err = resp.Body.Read(data)
		if err != nil {
			break
		}
		require.Equal(t, string(data[:4]), "haha")
	}
	require.Error(t, err)
	require.Equal(t, err, io.ErrUnexpectedEOF)
}
