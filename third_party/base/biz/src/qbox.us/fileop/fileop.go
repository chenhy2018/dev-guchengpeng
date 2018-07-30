package fileop

import (
	"encoding/base64"
	"io"
	"log"
	"net/http"
	"strings"

	"code.google.com/p/go.net/context"
	"github.com/qiniu/xlog.v1"

	"qbox.us/api"
	"qbox.us/etag"
	"qbox.us/fh/proto"
	"qbox.us/rpc"
	"qbox.us/servend/account"
	"qbox.us/sstore"
)

// ----------------------------------------------------------------------------

type Reader struct {
	proto.Source
	*sstore.FhandleInfo
	Bucket string
	Key    string
	*http.Request
	Env
	Query      []string
	StyleParam string
	Token      string
	URL        string
}

type Cache interface {
	Get(xl *xlog.Logger, key []byte) (rc io.ReadCloser, n int, err error)
	RangeGet(xl *xlog.Logger, key []byte, from, to int64) (rc io.ReadCloser, n int, err error)
	RangeGetAndHost(xl *xlog.Logger, key []byte, from, to int64) (host string, rc io.ReadCloser, n int, err error)
	Put(xl *xlog.Logger, key []byte, r io.Reader, n int) error
	PutWithHostRet(xl *xlog.Logger, key []byte, r io.Reader, n int) (string, error)
}

type Env interface {
	GetAccount() account.Interface
	GetIoHost() string
	GetCache() Cache
	FileOp(tpCtx context.Context, w Writer, r Reader)
	RFopg
}

type RFopg interface {
	Fopproxy(tpCtx context.Context, w http.ResponseWriter, r Reader)
}

// TODO: remove
func (r *Reader) QueryFhandle() (fh []byte, err error) {
	if r.FhandleInfo != nil {
		fh = r.Fhandle
		if fh[0] != 0 {
			return
		}
	}
	return r.Source.QueryFhandle()
}

// TODO: remove
func (r *Reader) QueryHash() (hash []byte, err error) {
	var fh []byte
	if r.FhandleInfo != nil {
		if r.Fhandle[0] != 0 {
			fh = r.Fhandle
		}
	}
	if fh == nil {
		if fh, err = r.Source.QueryFhandle(); err != nil {
			return
		}
	}

	return etag.Gen(fh), nil
}

type Writer struct {
	rpc.ResponseWriter
}

// ----------------------------------------------------------------------------

type Operations map[string]func(ctx context.Context, w Writer, r Reader)

func NewOperations() Operations {
	return make(Operations)
}

func (ops Operations) Register(op string, method func(context.Context, Writer, Reader)) {
	ops[op] = method
}

func (ops Operations) Do(ctx context.Context, w Writer, r Reader) {
	if method, ok := ops[r.Query[0]]; ok {
		method(ctx, w, r)
	} else {
		w.ReplyError(400, "Bad method")
	}
}

// ----------------------------------------------------------------------------

func Cat(w Writer, r Reader) {

	hash, err := r.QueryHash()
	if err != nil {
		log.Println("QueryHash failed:", err)
		w.ReplyWithError(api.FunctionFail, err)
		return
	}

	meta := &rpc.Metas{
		ETag:            base64.URLEncoding.EncodeToString(hash),
		MimeType:        r.MimeType,
		DispositionType: "inline",
		FileName:        r.AttName,
		CacheControl:    "public, max-age=31536000",
	}
	w.ReplyRange(r, r.Fsize, meta, r.Request)
}

// ----------------------------------------------------------------------------

func InputURL(req *http.Request) string {
	rawUrl := req.URL.Path
	pos1 := strings.Index(rawUrl[1:], "/")
	pos2 := strings.Index(rawUrl[2+pos1:], "/")
	return "http://" + req.Host + rawUrl[:2+pos1+pos2]
}

// ----------------------------------------------------------------------------
