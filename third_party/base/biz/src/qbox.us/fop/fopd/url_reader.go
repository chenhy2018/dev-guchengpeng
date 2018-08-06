package fopd

import (
	"io"
	"time"

	"github.com/qiniu/rpc.v1"
	"github.com/qiniu/xlog.v1"
)

type UrlReader struct {
	reader io.ReadCloser
	uri    string
	opened bool
	xl     *xlog.Logger
}

func NewUrlReader(xl *xlog.Logger, uri string) *UrlReader {
	return &UrlReader{xl: xl, uri: uri}
}

func (ur *UrlReader) Read(p []byte) (n int, err error) {

	if !ur.opened {
		start := time.Now()
		resp, err := rpc.DefaultClient.Get(ur.xl, ur.uri)
		if err != nil {
			return 0, err
		}
		n := time.Now()
		if n.Add(-time.Second).After(start) {
			ur.xl.Errorf("start at %v, got resp at %v", start, n)
		}
		ur.opened = true
		ur.reader = resp.Body
	}

	return ur.reader.Read(p)
}

func (ur *UrlReader) Close() error {

	if !ur.opened {
		return nil
	}

	return ur.reader.Close()
}
