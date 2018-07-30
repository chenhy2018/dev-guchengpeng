package ratelimiter

import (
	"net/http"
	"strconv"

	"github.com/qiniu/rpc.v1"
)

type Client struct {
	Host string
	Conn rpc.Client
}

func New(host string, t http.RoundTripper) *Client {
	client := &http.Client{Transport: t}
	return &Client{
		Host: host,
		Conn: rpc.Client{client},
	}
}

func (p Client) Set(l rpc.Logger, bucket string, period uint64, quota uint64) (err error) {

	err = p.Conn.CallWithForm(l, nil, p.Host+"/ratelimiter/set", map[string][]string{
		"bucket": {bucket},
		"period": {strconv.FormatUint(period, 10)},
		"quota":  {strconv.FormatUint(quota, 10)},
	})
	return
}

func (p Client) Delete(l rpc.Logger, bucket string) (err error) {

	err = p.Conn.CallWithForm(l, nil, p.Host+"/ratelimiter/delete", map[string][]string{
		"bucket": {bucket},
	})
	return
}

func (p Client) Get(l rpc.Logger, bucket string) (period uint64, quota uint64, err error) {

	var ret struct {
		Period uint64 `json:"period" bson:"period"`
		Quota  uint64 `json:"quota" bson:"quota"`
	}
	err = p.Conn.CallWithForm(l, &ret, p.Host+"/ratelimiter/get", map[string][]string{
		"bucket": {bucket},
	})
	period, quota = ret.Period, ret.Quota
	return
}

func (p Client) Allowed(l rpc.Logger, bucket string, key string) (allowed bool, wait uint64, err error) {

	var ret struct {
		Allowed bool   `json:"allowed" bson:"allowed"`
		Wait    uint64 `json:"wait" bson:"wait"`
	}
	err = p.Conn.CallWithForm(l, &ret, p.Host+"/ratelimiter", map[string][]string{
		"bucket": {bucket},
		"key":    {key},
	})
	allowed, wait = ret.Allowed, ret.Wait
	return
}
