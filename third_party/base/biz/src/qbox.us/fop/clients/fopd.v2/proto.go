package fopd

import (
	"io"
	"net/http"

	"code.google.com/p/go.net/context"

	"qbox.us/fop"
)

type Fopd interface {
	Op(tpCtx context.Context, f io.Reader, fsize int64, ctx *fop.FopCtx) (resp *http.Response, err error)
	Op2(tpCtx context.Context, fh []byte, fsize int64, ctx *fop.FopCtx) (resp *http.Response, err error)
	Close() error
	HasCmd(cmd string) bool
	NeedCache(cmd string) bool
	NeedCdnCache(cmd string) bool
	List() []string
	Stats() *Stats
}
