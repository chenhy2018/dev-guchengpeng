package dn

import (
	"encoding/base64"
	"fmt"
	"github.com/qiniu/rpc.v1"
	"github.com/qiniu/rpc.v1/lb.v1"
	"io"
	"net/http"
)

type Client struct {
	conn *lb.Client
}

func New(dnHosts []string, tr http.RoundTripper) (c Client, err error) {

	cfg := &lb.Config{
		Http:              &http.Client{Transport: tr},
		FailRetryInterval: -1,
		TryTimes:          uint32(len(dnHosts)),
	}
	conn, err := lb.New(dnHosts, cfg)
	if err != nil {
		return
	}
	return Client{
		conn: conn,
	}, nil
}

func (s Client) Get(xl rpc.Logger, fh []byte, w io.Writer, from, to int64) (n int64, err error) {

	if from == to {
		return 0, nil
	}

	efh := base64.URLEncoding.EncodeToString(fh)
	req, err := http.NewRequest("GET", "/get/"+efh, nil)
	if err != nil {
		return
	}
	req.Header.Add("Range", fmt.Sprintf("bytes=%v-%v", from, to-1))
	resp, err := s.conn.Do(xl, req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		err = rpc.ResponseError(resp)
		return
	}
	n, err = io.Copy(w, resp.Body)
	return
}
