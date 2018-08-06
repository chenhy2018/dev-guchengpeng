package fopd

import (
	"net/http"
	"strconv"

	"code.google.com/p/go.net/context"
	"github.com/qiniu/errors"
	"github.com/qiniu/http/httputil.v1"
	"github.com/qiniu/xlog.v1"
	"qbox.us/fop"
	"qbox.us/limit"
	"qbox.us/limit/keycount"
	"qbox.us/limit/null"
)

const xQiniuFop = "X-Qiniu-Fop-Stats"

var (
	ErrOpsPerUidCmdOutOfLimit = httputil.NewError(573, "ops per uid+cmd is out of limit")
)

type PipeConfig struct {
	Fopd            Fopd
	MaxOpsPerUidCmd int
}

type Pipe struct {
	*PipeConfig
	uidCmdLimit limit.Limit
}

func NewPipe(cfg *PipeConfig) *Pipe {
	var lim limit.Limit
	if cfg.MaxOpsPerUidCmd <= 0 {
		lim = null.New()
	} else {
		lim = keycount.New(cfg.MaxOpsPerUidCmd)
	}
	return &Pipe{cfg, lim}
}

func getUidCmdKey(ctx *fop.FopCtx) string {
	return strconv.FormatUint(uint64(ctx.Uid), 36) + ":" + ctx.CmdName
}

func (p *Pipe) Exec(tpCtx context.Context, fh []byte, fsize int64, tasks []*fop.FopCtx, out fop.Out) (resp *http.Response, err error) {
	xl := xlog.FromContextSafe(tpCtx)

	ntask := len(tasks)
	if ntask == 0 {
		return nil, errors.New("fop: no task to exec")
	}

	first := tasks[0]
	if ntask == 1 { // first is end
		copyOut(first, &out)
	}

	uidCmdKey := getUidCmdKey(first)
	err = p.uidCmdLimit.Acquire([]byte(uidCmdKey))
	if err != nil {
		err = errors.Info(ErrOpsPerUidCmdOutOfLimit, "pipe.Exec").Detail(err)
		xl.Warn(uidCmdKey, err)
		return
	}
	resp, err = p.Fopd.Op2(tpCtx, fh, fsize, first)
	p.uidCmdLimit.Release([]byte(uidCmdKey))

	if err != nil {
		return
	}

	if ntask == 1 { // 单命令
		return
	}
	combinedXStats := resp.Header.Get(xQiniuFop)

	// 管道后命令
	r := resp.Body
	fsize = resp.ContentLength
	mimeType := resp.Header.Get("Content-Type")
	tasks = tasks[1:]
	for i, ctx := range tasks {
		defer r.Close()
		ctx.MimeType = mimeType
		// 此处只让fopagent在缓存管道命令的最终结果，中间结果不缓存
		if i == len(tasks)-1 { // end
			copyOut(ctx, &out)
		}

		uidCmdKey := getUidCmdKey(ctx)
		err = p.uidCmdLimit.Acquire([]byte(uidCmdKey))
		if err != nil {
			err = errors.Info(ErrOpsPerUidCmdOutOfLimit, "pipe.Exec").Detail(err)
			xl.Warn(uidCmdKey, err)
			return
		}

		// 1. 对于向 IO 流式返回数据的情况，在 fopg 处向 dc 写完整的计费信息
		// 2. 对于向 IO 返回地址、由 IO 向 dc/fopagent 拉结果数据的情况，在 fopagent server 处向 dc 写完整的计费信息
		// 对于第二种情况，需要把combinedXStats 通过op命令传到fopagent，这是通过设置ctx.PreviousXstats来实现的
		if combinedXStats != "" {
			ctx.PreviousXstats = combinedXStats
		}
		resp, err = p.Fopd.Op(tpCtx, r, fsize, ctx)
		p.uidCmdLimit.Release([]byte(uidCmdKey))
		if err != nil {
			xl.Warnf("Exec: pipe[%d] %s error:%v", i, ctx.RawQuery, err)
			return
		}

		combinedXStats = resp.Header.Get(xQiniuFop)
		r = resp.Body
		fsize = resp.ContentLength
		mimeType = resp.Header.Get("Content-Type")
	}
	resp.Header.Set(xQiniuFop, combinedXStats)
	return
}

func copyOut(dst *fop.FopCtx, src *fop.Out) {
	dst.OutType = src.Type
	dst.OutRSBucket = src.RSBucket
	dst.OutRSKey = src.RSKey
	dst.OutRSDeleteAfterDays = src.RSDeleteAfterDays
	dst.OutDCKey = src.DCKey
}
