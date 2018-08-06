package mc

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/qiniu/rpc.v1"
	"net/http"
)

// ----------------------------------------------------------------------------

var (
	ErrCacheMiss = errors.New("mc: cache miss")
)

type Service struct {
	host string
	conn rpc.Client
}

func New(host string, t http.RoundTripper) Service {

	client := &http.Client{Transport: t}
	return Service{host, rpc.Client{client}}
}

// ----------------------------------------------------------------------------

func (r Service) Set(l rpc.Logger, key string, value interface{}) error {

	b, err := json.Marshal(value)
	if err != nil {
		return err
	}
	body := bytes.NewReader(b)
	url := r.host + "/set/" + base64.URLEncoding.EncodeToString([]byte(key))
	return r.conn.CallWith(l, nil, url, "application/json", body, body.Len())
}

func (r Service) Get(l rpc.Logger, key string, ret interface{}) error {

	url := r.host + "/get/" + base64.URLEncoding.EncodeToString([]byte(key))
	return transErr(r.conn.Call(l, ret, url))
}

func (r Service) Del(l rpc.Logger, key string) error {

	url := r.host + "/del/" + base64.URLEncoding.EncodeToString([]byte(key))
	return transErr(r.conn.Call(l, nil, url))
}

func transErr(err error) error {

	if err1, ok := err.(*rpc.ErrorInfo); ok && err1.Code == 404 {
		return ErrCacheMiss
	}
	return err
}
