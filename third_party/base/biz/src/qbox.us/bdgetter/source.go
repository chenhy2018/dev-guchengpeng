package bdgetter

import (
	"io"
	"io/ioutil"

	"github.com/qiniu/xlog.v1"
	"qbox.us/bdgetter/internal/bufpool"
	bdfh "qbox.us/fh"
	"qbox.us/fh/fhver"
	"qbox.us/fh/proto"
	"qbox.us/fh/urlbd"

	"github.com/qiniu/http/httputil.v1"
	qioutil "github.com/qiniu/io/ioutil"
	"github.com/qiniu/log.v1"
	"github.com/qiniu/rpc.v1"
)

type Sourcer interface {
	io.ReadCloser
	RangeRead(w io.Writer, from, to int64) (err error)
}

func Source(l rpc.Logger, getter *proto.Getter, fh []byte, fsize int64) (Sourcer, error) {
	ver := fhver.FhVer(fh)
	if ver == fhver.FhUrlbd {
		return urlbdSource(l, fh, fsize)
	}
	src, err := bdfh.Source(xlog.NewWith(l), getter, fh, fsize)
	return bdSource{src}, err
}

func urlbdSource(l rpc.Logger, fh []byte, fsize int64) (Sourcer, error) {
	url := urlbd.UrlOf(fh)
	resp, err := rpc.DefaultClient.Get(l, url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		resp.Body.Close()
		return nil, httputil.NewError(resp.StatusCode, "urlbd.Get failed")
	}
	return urlBdSource{resp.Body}, nil
}

type urlBdSource struct {
	io.ReadCloser
}

func (bd urlBdSource) RangeRead(w io.Writer, from, to int64) (err error) {
	sw := qioutil.SeqWriter(ioutil.Discard, from, w, to-from)
	_, err = bufpool.Copy(sw, bd)
	return
}

type bdSource struct {
	proto.Source
}

func (bd bdSource) WriteTo(w io.Writer) (n int64, err error) {
	return bd.Source.(io.WriterTo).WriteTo(w)
}

func (bd bdSource) Read(p []byte) (n int, err error) {
	log.Warn("Use Source.WriteTo!")
	return bd.Source.Read(p)
}

func (bd bdSource) Close() error {
	return nil
}
