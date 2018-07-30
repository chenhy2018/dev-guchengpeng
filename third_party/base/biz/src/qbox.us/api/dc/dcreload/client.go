package dcreload

import (
	"io"
	"sync"

	"qbox.us/api/dc"

	"github.com/qiniu/xlog.v1"
)

type Reloader interface {
	Reload(newClient *dc.Client)
}

var (
	_ dc.DiskCache = new(Client)
)

type Client struct {
	client *dc.Client
	mu     sync.RWMutex
}

func New(dcClient *dc.Client) *Client {
	return &Client{client: dcClient}
}

func (c *Client) Reload(newClient *dc.Client) {
	c.mu.Lock()
	old := c.client
	c.client = newClient
	c.mu.Unlock()
	if old != nil && old != newClient {
		old.Close()
	}
}

func (c *Client) safeClient() *dc.Client {
	c.mu.RLock()
	client := c.client
	c.mu.RUnlock()
	return client
}

func (c *Client) Get(xl *xlog.Logger, key []byte) (r io.ReadCloser, length int64, err error) {
	return c.safeClient().Get(xl, key)
}

func (c *Client) GetHint(xl *xlog.Logger, key []byte) (r io.ReadCloser, length int64, hint bool, err error) {
	return c.safeClient().GetHint(xl, key)
}

func (c *Client) RangeGet(xl *xlog.Logger, key []byte, from, to int64) (r io.ReadCloser, length int64, err error) {
	return c.safeClient().RangeGet(xl, key, from, to)
}

func (c *Client) RangeGetHint(xl *xlog.Logger, key []byte, from, to int64) (r io.ReadCloser, length int64, hint bool, err error) {
	return c.safeClient().RangeGetHint(xl, key, from, to)
}

func (c *Client) RangeGetAndHost(xl *xlog.Logger, key []byte, from, to int64) (host string, r io.ReadCloser, length int64, err error) {
	return c.safeClient().RangeGetAndHost(xl, key, from, to)
}

func (c *Client) KeyHost(xl *xlog.Logger, key []byte) (host string, err error) {
	return c.safeClient().KeyHost(xl, key)
}

func (c *Client) Set(xl *xlog.Logger, key []byte, r io.Reader, length int64) (err error) {
	return c.safeClient().Set(xl, key, r, length)
}

func (c *Client) SetEx(xl *xlog.Logger, key []byte, r io.Reader, length int64, sha1 []byte) (err error) {
	return c.safeClient().SetEx(xl, key, r, length, sha1)
}

func (c *Client) SetWithHostRet(xl *xlog.Logger, key []byte, r io.Reader, length int64) (host string, err error) {
	return c.safeClient().SetWithHostRet(xl, key, r, length)
}

func (c *Client) SetExWithHostRet(xl *xlog.Logger, key []byte, r io.Reader, length int64, checksum []byte) (host string, err error) {
	return c.safeClient().SetExWithHostRet(xl, key, r, length, checksum)
}

func (c *Client) Delete(xl *xlog.Logger, key []byte) (host string, err error) {
	return c.safeClient().Delete(xl, key)
}
