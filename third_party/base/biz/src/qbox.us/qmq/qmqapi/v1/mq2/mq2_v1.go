package mq2

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync/atomic"

	"github.com/qiniu/rpc.v1"
	"qbox.us/qmq/qmqapi/v1/mq"
)

var (
	ErrInvalidMsgId         = errors.New("invalid msgId")
	ErrInvalidContentLength = errors.New("invalid Content-Length")
)

const (
	maxMsgLen = (1 << 22) // 4M
)

// ------------------------------------------------------------------

type Client struct {
	Conn    rpc.Client
	mqHosts []string
	idxW    int32
	idxR    int32
}

func New(mqHosts []string, t http.RoundTripper) *Client {

	c := &http.Client{Transport: t}
	return &Client{
		Conn:    rpc.Client{c},
		mqHosts: mqHosts,
	}
}

func (p *Client) Hosts() []string {
	return p.mqHosts
}

func (p *Client) Put(l rpc.Logger, mqId string, msg []byte) (idx int, msgId string, err error) {

	return p.PutEx(l, mqId, bytes.NewReader(msg), len(msg))
}

func (p *Client) PutString(l rpc.Logger, mqId string, msg string) (idx int, msgId string, err error) {

	return p.PutEx(l, mqId, strings.NewReader(msg), len(msg))
}

func (p *Client) PutEx(l rpc.Logger, mqId string, msgr io.Reader, bytes int) (idx int, msgId string, err error) {

	idx = int(atomic.AddInt32(&p.idxW, 1)) % len(p.mqHosts)
	msgId, err = p.put(l, idx, mqId, msgr, bytes)
	if err != nil {
		idx = int(idx+1) % len(p.mqHosts)
		msgId, err = p.put(l, idx, mqId, msgr, bytes)
	}
	return
}

func (p *Client) put(l rpc.Logger, idx int, mqId string, msgr io.Reader, bytes int) (msgId string, err error) {

	host := p.mqHosts[idx]
	return mq.Service{Host: host, Conn: p.Conn}.PutEx(l, mqId, msgr, bytes)
}

func (p *Client) Get(l rpc.Logger, idx int, mqId string) (msg []byte, msgId string, err error) {

	host := p.mqHosts[idx]
	return mq.Service{Host: host, Conn: p.Conn}.Get(l, mqId)
}

func (p *Client) Delete(l rpc.Logger, idx int, mqId string, msgId string) (err error) {

	host := p.mqHosts[idx]
	return mq.Service{Host: host, Conn: p.Conn}.Delete(l, mqId, msgId)
}

// ------------------------------------------------------------------
