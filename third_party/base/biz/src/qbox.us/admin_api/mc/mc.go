package mc

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"errors"
	"net/http"

	"github.com/qiniu/rpc.v1"
	"github.com/qiniu/rpc.v1/lb.v2.1"
)

// ----------------------------------------------------------------------------

var (
	ErrCacheMiss = errors.New("mc: cache miss")
)

type Service struct {
	conn *lb.Client
}

func New(host string, t http.RoundTripper) Service {
	cfg := &lb.Config{
		Hosts:    []string{host},
		TryTimes: 1,
	}
	client := lb.New(cfg, t)
	return Service{client}
}

func NewWithFailover(client, failover lb.Config, clientTr, failoverTr http.RoundTripper) Service {
	cli := lb.NewWithFailover(&client, &failover, clientTr, failoverTr, nil)
	return Service{cli}
}

// ----------------------------------------------------------------------------

func (r *Service) BatchDel(l rpc.Logger, keys []string) error {

	wr := &bytes.Buffer{}
	gob.NewEncoder(wr).Encode(keys)

	req, err := lb.NewRequest("POST", "/admin-batchdel", bytes.NewReader(wr.Bytes()))
	if err != nil {
		return err
	}
	resp, err := r.conn.Do(l, req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return rpc.ResponseError(resp)
	}
	return nil
}

// ----------------------------------------------------------------------------

func (r Service) Del(l rpc.Logger, key string) error {

	url := "/admin-del/" + base64.URLEncoding.EncodeToString([]byte(key))
	return transErr(r.conn.Call(l, nil, url))
}

func transErr(err error) error {

	if err1, ok := err.(*rpc.ErrorInfo); ok && err1.Code == 404 {
		return ErrCacheMiss
	}
	return err
}
