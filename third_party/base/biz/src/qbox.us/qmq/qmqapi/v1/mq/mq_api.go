package mq

import (
	"bytes"
	"errors"
	digest_auth "github.com/qiniu/api/auth/digest"
	"github.com/qiniu/rpc.v1"
	"io"
	"net/http"
	. "qbox.us/api/conf"
	"strconv"
	"strings"
)

const (
	NoSuchEntry = 612 // 指定的 Entry 不存在或已经 Deleted
)

// ----------------------------------------------------------

type Service struct {
	Host string
	Conn rpc.Client
}

func New() Service {
	t := digest_auth.NewTransport(nil, rpc.DefaultTransport)
	client := &http.Client{Transport: t}
	return Service{MQ_HOST, rpc.Client{client}}
}

func NewEx(t http.RoundTripper) Service {
	client := &http.Client{Transport: t}
	return Service{MQ_HOST, rpc.Client{client}}
}

func NewWithHost(t http.RoundTripper, host string) Service {
	client := &http.Client{Transport: t}
	return Service{host, rpc.Client{client}}
}

// ----------------------------------------------------------

func (r Service) Make(l rpc.Logger, mqId string, expires int) (err error) {

	url := r.Host + "/make/" + mqId + "/expires/" + strconv.Itoa(expires)
	return r.Conn.Call(l, nil, url)
}

func (r Service) Put(l rpc.Logger, mqId string, msg []byte) (msgId string, err error) {

	return r.PutEx(l, mqId, bytes.NewReader(msg), len(msg))
}

func (r Service) PutString(l rpc.Logger, mqId string, msg string) (msgId string, err error) {

	return r.PutEx(l, mqId, strings.NewReader(msg), len(msg))
}

func (r Service) PutEx(l rpc.Logger, mqId string, msgr io.Reader, bytes int) (msgId string, err error) {

	resp, err := r.Conn.PostWith(l, r.Host+"/put/"+mqId, "application/octet-stream", msgr, bytes)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		err = rpc.ResponseError(resp)
		return
	}
	msgId = resp.Header.Get("X-Id")
	return
}

func (r Service) Get(l rpc.Logger, mqId string) (msg []byte, msgId string, err error) {

	resp, err := r.Conn.PostWith64(l, r.Host+"/get/"+mqId, "application/x-www-form-urlencoded", nil, 0)
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
	if length <= 0 || length > (1<<22) {
		err = errors.New("invalid ContentLength")
		return
	}
	msg = make([]byte, length)
	_, err = io.ReadFull(resp.Body, msg)
	return
}

func (r Service) Delete(l rpc.Logger, mqId string, msgId string) (err error) {

	req, err := http.NewRequest("POST", r.Host+"/delete/"+mqId, nil)
	if err != nil {
		return
	}
	req.Header.Set("X-Id", msgId)
	resp, err := r.Conn.Do(l, req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		err = rpc.ResponseError(resp)
	}
	return
}

// ----------------------------------------------------------
