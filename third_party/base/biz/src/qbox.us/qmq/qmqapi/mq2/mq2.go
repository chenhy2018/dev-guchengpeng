package mq2

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync/atomic"

	"github.com/qiniu/rpc.v1"
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

func (p *Client) Put(l rpc.Logger, mqId, msgId string, msg []byte) error {

	return p.PutEx(l, mqId, msgId, bytes.NewReader(msg), len(msg))
}

func (p *Client) PutString(l rpc.Logger, mqId, msgId string, msg string) error {

	return p.PutEx(l, mqId, msgId, strings.NewReader(msg), len(msg))
}

func (p *Client) PutEx(l rpc.Logger, mqId, msgId string, msgr io.Reader, bytes int) error {

	idx := int(atomic.AddInt32(&p.idxW, 1)) % len(p.mqHosts)
	err := p.put(l, idx, mqId, msgId, msgr, bytes)
	if err != nil {
		idx = int(idx+1) % len(p.mqHosts)
		err = p.put(l, idx, mqId, msgId, msgr, bytes)
	}
	return err
}

func (p *Client) put(l rpc.Logger, idx int, mqId, msgId string, msgr io.Reader, bytes int) error {

	host := p.mqHosts[idx]
	url := host + "/put/" + mqId + "/id/" + msgId
	return p.Conn.CallWith(l, nil, url, "application/octet-stream", msgr, bytes)
}

func (p *Client) Get(l rpc.Logger, idx int, mqId string) (msg []byte, msgId string, err error) {

	host := p.mqHosts[idx]
	url := host + "/get/" + mqId

	resp, err := p.Conn.PostEx(l, url)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		err = rpc.ResponseError(resp)
		return
	}

	msgId = resp.Header.Get("X-Id")

	length := resp.ContentLength
	if length <= 0 || length > maxMsgLen {
		err = ErrInvalidContentLength
		return
	}
	msg = make([]byte, length)
	_, err = io.ReadFull(resp.Body, msg)
	return
}

func (p *Client) Delete(l rpc.Logger, idx int, mqId string, msgId string) (err error) {

	host := p.mqHosts[idx]
	url := host + "/delete/" + mqId + "/id/" + msgId
	return p.Conn.Call(l, nil, url)
}

// ------------------------------------------------------------------
