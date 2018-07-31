package api

import (
	"bytes"
	"io"
	"net/http"

	"github.com/qiniu/encoding/binary"
	"github.com/qiniu/rpc.v1"
	"github.com/qiniu/rpc.v1/lb.v2.1"
	"github.com/qiniu/xlog.v1"
	. "qbox.us/ebd/api/types"
)

type StripeRgetInfo struct { // stripe rget info
	Soff    uint32
	Bsize   uint32
	BadSuid uint64
	Psects  [N + M]uint64
}

func ReadStripeRgetInfo(r io.Reader) (srgi *StripeRgetInfo, err error) {
	srgi = new(StripeRgetInfo)
	err = binary.Read(r, binary.LittleEndian, srgi)
	return
}

func EncodeStripeRgetInfo(srgi *StripeRgetInfo) (b []byte, err error) {
	w := new(bytes.Buffer)
	err = binary.Write(w, binary.LittleEndian, srgi)
	b = w.Bytes()
	return
}

type Client struct {
	conn *lb.Client
}

func shouldRetry(code int, err error) bool {
	if code == http.StatusServiceUnavailable {
		return true
	}
	return lb.ShouldRetry(code, err)
}

func New(hosts []string, tr http.RoundTripper) (c Client, err error) {
	cfg := &lb.Config{
		Hosts:              hosts,
		ShouldRetry:        shouldRetry,
		FailRetryIntervalS: -1,
		TryTimes:           uint32(len(hosts)),
	}

	conn := lb.New(cfg, tr)
	if err != nil {
		return
	}
	return Client{conn: conn}, nil
}

func (self *Client) Rget(l rpc.Logger, srgi *StripeRgetInfo) (body io.ReadCloser, err error) {
	b, err := EncodeStripeRgetInfo(srgi)
	if err != nil {
		return
	}

	host, resp, err := self.conn.PostWithHostRet(l, "/rget", "application/octet-stream", bytes.NewReader(b), len(b))
	xlog.Info(l.ReqId(), "Rget to ", host)
	if err != nil {
		return
	}
	defer func() {
		if err != nil {
			resp.Body.Close()
		}
	}()

	if resp.StatusCode != 200 {
		err = rpc.ResponseError(resp)
		return
	}

	return resp.Body, nil
}

// ---------------------------------------------------------
