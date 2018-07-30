package dc

import (
	"io"
	"net/http"
	"time"

	"qbox.us/api"
	"qbox.us/dht"

	"github.com/qiniu/errors"
	qio "github.com/qiniu/io"
	"github.com/qiniu/rpc.v1"
	"github.com/qiniu/xlog.v1"
)

type DCConn struct {
	Keys []string `json:"keys"`
	Host string   `json:"host"`
}

type Config struct {
	Servers           []DCConn `json:"servers"`
	DialTimeoutMS     int      `json:"dial_timeout_ms"`
	RespTimeoutMS     int      `json:"resp_timeout_ms"`
	TransportPoolSize int      `json:"transport_pool_size"`
	TryTimes          int      `json:"try_times"`
}

type Client struct {
	clients      map[string]*Conn
	cacheChooser dht.Interface
	tryTimes     int
}

func NewClient(conns []DCConn, tryTimes int, transport http.RoundTripper) *Client {

	nodes := make([]dht.NodeInfo, 0)
	clients := make(map[string]*Conn)
	for _, conn := range conns {
		clients[conn.Host] = NewConn(conn.Host, transport)
		for _, key := range conn.Keys {
			nodes = append(nodes, dht.NodeInfo{conn.Host, []byte(key)})
		}
	}
	if tryTimes == 0 {
		tryTimes = 1
	}
	return &Client{
		clients:      clients,
		cacheChooser: dht.NewCarp(nodes),
		tryTimes:     tryTimes,
	}
}

func NewTransport(dialTimeoutMS, respTimeoutMS, poolSize int) http.RoundTripper {
	if dialTimeoutMS == 0 {
		dialTimeoutMS = 5000
	}
	if respTimeoutMS == 0 {
		respTimeoutMS = 3000
	}
	if poolSize == 0 {
		poolSize = 5
	}
	return rpc.NewTransportTimeoutWithConnsPool(time.Duration(dialTimeoutMS)*time.Millisecond,
		time.Duration(respTimeoutMS)*time.Millisecond, poolSize)
}

type TimeoutOptions struct {
	DialMs        int   `json:"dial_ms"`
	RespMs        int   `json:"resp_ms"`
	ClientMs      int   `json:"client_ms"`
	RangegetSpeed int64 `json:"rangeget_speed"` // Bytes per second
	SetSpeed      int64 `json:"set_speed"`      // Bytes per second
}

func NewWithTimeout(conns []DCConn, options *TimeoutOptions) *Client {

	nodes := make([]dht.NodeInfo, 0)
	clients := make(map[string]*Conn)
	for _, conn := range conns {
		clients[conn.Host] = NewConnWithTimeout(conn.Host, options)
		for _, key := range conn.Keys {
			nodes = append(nodes, dht.NodeInfo{conn.Host, []byte(key)})
		}
	}
	return &Client{
		clients:      clients,
		cacheChooser: dht.NewCarp(nodes),

		// dc 业务上可以不重试。
		// 另外 dht 组件对重试不是特别友好：
		// 1. dht 存在虚拟结点，重试可能选到同一个实例。
		// 2. dht 重试是 hash 重试，可能选到同一台机器。
		tryTimes: 1,
	}
}

// -----------------------------------------------------------------------------

func (p *Client) Close() error {
	for _, c := range p.clients {
		c.Close()
	}
	return nil
}

func (p *Client) Get(xl *xlog.Logger, key []byte) (r io.ReadCloser, length int64, err error) {
	r, length, _, err = p.GetHint(xl, key)
	return
}

func (p *Client) GetHint(xl *xlog.Logger, key []byte) (r io.ReadCloser, length int64, hint bool, err error) {

	routers := p.cacheChooser.Route(key, p.tryTimes)
	if len(routers) == 0 {
		err = errors.New("No dc server is available")
		return
	}

	for _, router := range routers {
		client, ok := p.clients[router.Host]
		if !ok {
			continue
		}
		r, length, hint, err = client.GetHint(xl, key)
		if err == nil || err == api.EInvalidArgs || err == ENoSuchEntry {
			break
		}
		if r != nil {
			r.Close()
		}
	}
	if err != nil {
		xl.Warn("dc.Get error:", err)
	}
	return
}

func (p *Client) RangeGet(xl *xlog.Logger, key []byte, from int64, to int64) (r io.ReadCloser, length int64, err error) {
	r, length, _, err = p.RangeGetHint(xl, key, from, to)
	return
}

func (p *Client) RangeGetHint(xl *xlog.Logger, key []byte, from int64, to int64) (r io.ReadCloser, length int64, hint bool, err error) {

	routers := p.cacheChooser.Route(key, p.tryTimes)
	if len(routers) == 0 {
		err = errors.New("No dc server is available")
		return
	}

	for _, router := range routers {
		client, ok := p.clients[router.Host]
		if !ok {
			continue
		}
		r, length, hint, err = client.RangeGetHint(xl, key, from, to)
		if err == nil || err == api.EInvalidArgs || err == ENoSuchEntry {
			break
		}
		if r != nil {
			r.Close()
		}
	}
	if err != nil {
		xl.Warn("dc.RangeGet error:", err)
	}
	return
}

func (p *Client) RangeGetAndHost(xl *xlog.Logger, key []byte, from int64, to int64) (host string, r io.ReadCloser, length int64, err error) {

	routers := p.cacheChooser.Route(key, p.tryTimes)
	if len(routers) == 0 {
		err = errors.New("No dc server is available")
		return
	}

	for _, router := range routers {
		client, ok := p.clients[router.Host]
		if !ok {
			continue
		}
		host, r, length, err = client.RangeGetAndHost(xl, key, from, to)
		if err == nil || err == api.EInvalidArgs || err == ENoSuchEntry {
			break
		}
		if r != nil {
			r.Close()
		}
	}
	if err != nil {
		xl.Warn("dc.RangeGetAndHost error:", err)
	}
	return
}

func (p *Client) KeyHost(xl *xlog.Logger, key []byte) (host string, err error) {
	routers := p.cacheChooser.Route(key, p.tryTimes)
	if len(routers) == 0 {
		err = errors.New("dc server is unavailable")
		return
	}
	for _, router := range routers {
		client, ok := p.clients[router.Host]
		if !ok {
			continue
		}
		host, err = client.KeyHost(xl, key)
		if err == nil || !shouldRetry(err) {
			break
		}
	}
	if err != nil {
		xl.Warn("Client.KeyHost - error:", err)
	}
	return
}

func shouldRetry(err error) bool {
	if _, ok := err.(*rpc.ErrorInfo); ok { // 有 HTTP Response
		return false
	}
	return true
}

func (p *Client) set(xl *xlog.Logger, key []byte, r io.Reader, length int64, checksum []byte) (host string, err error) {

	routers := p.cacheChooser.Route(key, p.tryTimes)
	if len(routers) == 0 {
		err = errors.New("No dc server is available")
		return
	}

	retry := false
	for _, router := range routers {
		client, ok := p.clients[router.Host]
		if !ok {
			continue
		}
		var rt io.Reader
		if _, ok := r.(io.ReaderAt); ok {
			rt = &qio.Reader{r.(io.ReaderAt), 0}
			retry = true
		} else {
			rt = r
		}

		host, err = client.set(xl, key, rt, length, checksum)
		if err == nil || err == api.EInvalidArgs {
			break
		}
		xl.Warn("dc.Set error:", err)
		if !retry {
			xl.Warn("can not retry")
			break
		}
	}
	return
}

func (p *Client) Set(xl *xlog.Logger, key []byte, r io.Reader, length int64) (err error) {
	_, err = p.set(xl, key, r, length, nil)
	return
}

func (p *Client) SetEx(xl *xlog.Logger, key []byte, r io.Reader, length int64, checksum []byte) (err error) {
	_, err = p.set(xl, key, r, length, checksum)
	return
}

func (p *Client) SetWithHostRet(xl *xlog.Logger, key []byte, r io.Reader, length int64) (host string, err error) {
	return p.set(xl, key, r, length, nil)
}

func (p *Client) SetExWithHostRet(xl *xlog.Logger, key []byte, r io.Reader, length int64, checksum []byte) (host string, err error) {
	return p.set(xl, key, r, length, checksum)
}

func (p *Client) Delete(xl *xlog.Logger, key []byte) (host string, err error) {
	routers := p.cacheChooser.Route(key, p.tryTimes)
	if len(routers) == 0 {
		err = errors.New("dc server is unavailable")
		return
	}
	for _, router := range routers {
		client, ok := p.clients[router.Host]
		if !ok {
			continue
		}
		err = client.Delete(xl, key)
		if err == nil || !shouldRetry(err) {
			break
		}
	}
	if err != nil {
		xl.Warn("Client.Delete - error:", err)
	}
	return
}
