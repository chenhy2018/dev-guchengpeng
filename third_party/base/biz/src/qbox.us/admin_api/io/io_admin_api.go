package io

import (
	"net/http"
	"qbox.us/rpc"
)

// -----------------------------------------------------------

type Service struct {
	Host string
	Conn rpc.Client
}

func New(host string, t http.RoundTripper) *Service {
	client := &http.Client{Transport: t}
	return &Service{host, rpc.Client{client}}
}

// -----------------------------------------------------------

type CacheInfo struct {
	Missing int64 `json:"missing"`
	Total   int64 `json:"total"`
	Wtotal  int64 `json:"wtotal"`
}

type CacheInfoEx struct {
	Host    string `json:"host"`
	Missing int64  `json:"missing"`
	Total   int64  `json:"total"`
	Wtotal  int64  `json:"wtotal"`
}

type StatInfo struct {
	MemCache CacheInfo      `json:"memcache"`
	BdCache  []*CacheInfoEx `json:"bdcache"`
}

func (r *Service) Stat() (info StatInfo, code int, err error) {

	code, err = r.Conn.Call(&info, r.Host+"/admin/service-stat")
	return
}

// -----------------------------------------------------------
