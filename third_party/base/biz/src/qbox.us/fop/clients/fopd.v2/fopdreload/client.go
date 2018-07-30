package fopdreload

import (
	"io"
	"net/http"
	"sync"

	"code.google.com/p/go.net/context"

	"qbox.us/fop"
	"qbox.us/fop/clients/fopd.v2"
)

type Reloader interface {
	Reload(newClient fopd.Fopd)
}

type Client struct {
	client fopd.Fopd
	mu     sync.RWMutex
}

func New(client fopd.Fopd) *Client {
	return &Client{client: client}
}

func (c *Client) Reload(newClient fopd.Fopd) {
	c.mu.Lock()
	oc := c.client
	c.client = newClient
	c.mu.Unlock()
	go oc.Close()
}

func (c *Client) safeFopd() fopd.Fopd {
	c.mu.RLock()
	client := c.client
	c.mu.RUnlock()
	return client
}

func (c *Client) Op(tpCtx context.Context, f io.Reader, fsize int64, ctx *fop.FopCtx) (resp *http.Response, err error) {
	return c.safeFopd().Op(tpCtx, f, fsize, ctx)
}

func (c *Client) Op2(tpCtx context.Context, fh []byte, fsize int64, ctx *fop.FopCtx) (resp *http.Response, err error) {
	return c.safeFopd().Op2(tpCtx, fh, fsize, ctx)
}

func (c *Client) Close() error {
	return c.safeFopd().Close()
}

func (c *Client) HasCmd(cmd string) bool {
	return c.safeFopd().HasCmd(cmd)
}

func (c *Client) NeedCache(cmd string) bool {
	return c.safeFopd().NeedCache(cmd)
}

func (c *Client) NeedCdnCache(cmd string) bool {
	return c.safeFopd().NeedCdnCache(cmd)
}

func (c *Client) List() []string {
	return c.safeFopd().List()
}

func (c *Client) Stats() *fopd.Stats {
	return c.safeFopd().Stats()
}
