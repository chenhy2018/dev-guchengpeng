package up

import (
	"bytes"
	"encoding/base64"
	"github.com/qiniu/rpc.v1"
	"github.com/qiniu/rpc.v1/lb.v1"
	"io"
	"net/http"
	"qbox.us/api/up"
	"strconv"
)

type Client struct {
	conn *lb.Client
}

func New(upHosts []string, tr http.RoundTripper) (c Client, err error) {

	cfg := &lb.Config{
		Http:              &http.Client{Transport: tr},
		FailRetryInterval: -1,
		TryTimes:          uint32(len(upHosts)),
	}
	conn, err := lb.New(upHosts, cfg)
	if err != nil {
		return
	}
	return Client{
		conn: conn,
	}, nil
}

func (s Client) Put(xl rpc.Logger, fsize int64, body io.Reader) (fh []byte, err error) {

	var ret struct {
		Efh string `json:"fh"`
	}
	err = s.conn.CallWith64(xl, &ret, "/", "application/octet-stream", body, fsize)
	if err != nil {
		return
	}

	fh, err = base64.URLEncoding.DecodeString(ret.Efh)
	return
}

// -----------------------------------------------------------------------------

// POST /mkblk/<BlockSize>
// Body: <FirstChunkData>
func (s Client) Mkblk(xl rpc.Logger, blockSize uint32, r io.Reader) (ret up.PutRet, err error) {

	path := "/mkblk/" + strconv.Itoa(int(blockSize))

	err = s.conn.CallWith(xl, &ret, path, "application/octet-stream", r, -1)
	if err != nil {
		return
	}
	return
}

// POST /bput/<Ctx>/<Offset>
// Body: <NextChunkData>
func (s Client) Bput(xl rpc.Logger, ctx string, offset uint32, r io.Reader) (ret up.PutRet, err error) {

	path := "/bput/" + ctx + "/" + strconv.Itoa(int(offset))

	err = s.conn.CallWith(xl, &ret, path, "application/octet-stream", r, -1)
	if err != nil {
		return
	}
	return
}

// POST /mkfile/<Fsize>
// Body: <Ctx-Array> e.g. ctx1,ctx2,ctx3...
func (s Client) Mkfile(xl rpc.Logger, ctxs []byte, fsize int64) (fh []byte, err error) {

	body := bytes.NewReader(ctxs)
	path := "/mkfile/" + strconv.FormatInt(fsize, 10)

	var ret struct {
		Efh string `json:"fh"`
	}
	err = s.conn.CallWith(xl, &ret, path, "application/octet-stream", body, len(ctxs))
	if err != nil {
		return
	}

	fh, err = base64.URLEncoding.DecodeString(ret.Efh)
	return
}
