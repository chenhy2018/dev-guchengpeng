package cached

import (
	"crypto/sha1"
	"encoding/binary"
	"io"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/qiniu/http/httputil.v1"
	qio "github.com/qiniu/io"
	qioutil "github.com/qiniu/io/ioutil"
	"github.com/qiniu/rpc.v1"
	"github.com/qiniu/xlog.v1"
	"qbox.us/bdgetter/internal/bufpool"
	"qbox.us/bdgetter/retrys"
	qfh "qbox.us/fh"
	"qbox.us/fh/proto"
)

// fh must be qbox.us/fh/proto.Fsizer.

type Cached interface {
	Set(xl *xlog.Logger, key []byte, r io.Reader, length int64) error
	RangeGetHint(xl *xlog.Logger, key []byte, from, to int64) (io.ReadCloser, int64, bool, error)
}

// 描述缓存策略，对于一个fh，给出缓存使用的key、缓存实例，和是否需要分块缓存。
type CachedController interface {
	Cached(l rpc.Logger, fh []byte) (key []byte, cached Cached, blocks bool)
}

// -----------------------------------------------------------------------------

type cacheGetter struct {
	getter proto.ReaderGetter
	cc     CachedController
	retrys proto.CommonGetter
}

func New(g proto.ReaderGetter, cc CachedController, maxTrys int) proto.CommonGetter {

	return &cacheGetter{getter: g, cc: cc, retrys: retrys.NewRetrys(g, maxTrys)}
}

func (p *cacheGetter) getFromCached(xl *xlog.Logger, cached Cached, key []byte, w io.Writer, from, to int64) (n int64, hint bool, err error) {

	rc, _, hint, err := cached.RangeGetHint(xl, key, from, to)
	if err != nil {
		if httputil.DetectCode(err) == 404 {
			return
		}
		xl.Warn("cacheGetter.Get: getCached failed =>", err, n)
		return
	}
	defer rc.Close()
	n, err = bufpool.Copy(w, rc)
	return
}

func streamWriter(fw io.Writer, ow io.Writer, from, to int64) *qio.OptimisticMultiWriter {
	fw2 := qioutil.SeqWriter(ioutil.Discard, from, ow, to-from)
	return &qio.OptimisticMultiWriter{
		Writers: []qio.OptimisticWriter{
			{Writer: fw},
			{Writer: fw2},
		},
	}
}

func isWriterErr(err error) bool {

	return err == io.ErrShortWrite || strings.Contains(err.Error(), "write tcp")
}

var (
	chunkBits uint  = 18
	chunkSize int64 = 1 << chunkBits
)

func Init(bit uint) {
	chunkBits = bit
	chunkSize = 1 << chunkBits
}

func genRangeCacheKey(key []byte, from, to int64) []byte {
	b2 := make([]byte, 16)
	binary.LittleEndian.PutUint64(b2, uint64(from))
	binary.LittleEndian.PutUint64(b2[8:], uint64(to))
	data := append(key, b2...)
	key2 := sha1.Sum(data)
	return key2[:]
}

func (p *cacheGetter) Get(xl1 rpc.Logger, fh []byte, w io.Writer, from, to int64) (n int64, err error) {

	if p.cc == nil {
		return p.retrys.Get(xl1, fh, w, from, to)
	}
	key, cached, blocks := p.cc.Cached(xl1, fh)
	if cached == nil {
		return p.retrys.Get(xl1, fh, w, from, to)
	}
	xl := xlog.NewWith(xl1)

	fsize := qfh.Fsize(fh)
	if blocks {
		// 分块读写dc
		xl0 := xl.Spawn()
		ctx := xl.Context()
		for probe := from >> chunkBits << chunkBits; probe < to; probe += chunkSize {
			var n1 int64
			afrom, ato := probe, probe+chunkSize
			if ato > fsize {
				ato = fsize
			}
			key := genRangeCacheKey(key, afrom, ato)

			bfrom, bto := int64(0), ato-afrom
			if afrom < from {
				bfrom = from - afrom
			}
			if ato > to {
				bto = to - afrom
			}

			// see: https://jira.qiniu.io/browse/KODO-3218
			// 如果要缓存的终点（ato）在用户实际请求的终点（bto + chunkSize * chunkIndex）之后，
			// 那么DC 缓存完成的时间就要落后于用户获取到所请求数据的时间。
			// 用户在获取到请求的所有数据后，可能会主动关闭连接。如果请求 EBD 的连接继承了用户请求
			// 的 context，那么这时候连接就会关闭，DC 设置缓存就会失败。所以，这里传入的 xl0 不能
			// 继承用户的 context。
			n1, err = p.rangeGet(xl0, fh, afrom, ato, key, cached, w, bfrom, bto)
			n += n1
			if err != nil {
				break
			}
			select {
			case <-ctx.Done():
				xl0.Info("context is done")
				err = rpc.NewHttpError(499, ctx.Err())
				break // 用户关闭连接，处理结束
			default:
			}
		}
		merge(xl, xl0)
	} else {
		xl.Debug("not blocks")
		n, err = p.rangeGet(xl, fh, 0, fsize, key, cached, w, from, to)
	}
	return
}

func decode_modFn(modFn string) (service string, dur int64, msg string, err error) {
	idx := strings.LastIndex(modFn, "/")
	if idx == -1 {
		msg = ""
	} else {
		msg = modFn[idx+1:]
		modFn = modFn[:idx]
	}

	idx = strings.LastIndex(modFn, ":")
	if idx == -1 {
		dur = 0
		service = modFn
	} else {
		dur, err = strconv.ParseInt(modFn[idx+1:], 10, 64)
		service = modFn[:idx]
	}
	return
}

func merge(xl, xl0 *xlog.Logger) {
	x_log := xl0.Header()["X-Log"]

	type svrMsg struct {
		svr string
		msg string
	}
	type item struct {
		dur   int64
		count int
	}
	set := make(map[svrMsg]item)
	if len(x_log) < 30 {
		xl.Xput(x_log)
		return
	}
	for _, modFns := range x_log {
		for _, modFn := range strings.Split(modFns, ";") {
			svr, dur, msg, err := decode_modFn(modFn)
			if err != nil {
				xl.Xput([]string{modFn})
				xl.Info(err)
				continue
			}

			svr_msg := svrMsg{svr, msg}
			if item0, ok := set[svr_msg]; ok {
				set[svr_msg] = item{item0.dur + dur, item0.count + 1}
			} else {
				set[svr_msg] = item{dur, 1}
			}
		}
	}
	for svr_msg, item := range set {
		modFn := svr_msg.svr + "*" + strconv.Itoa(item.count)
		if item.dur > 0 {
			modFn += ":" + strconv.FormatInt(item.dur, 10)
		}
		if svr_msg.msg != "" {
			modFn += "/" + svr_msg.msg
		}
		xl.Xlog(modFn)
	}

}

/*
        bfrom         bto             "bfrom", "bto" are relivate to "from"
	|___________________________|
   from                         to    "from", "to" are absolute to fh
*/
func (p *cacheGetter) rangeGet(xl *xlog.Logger,
	fh []byte, from, to int64,
	key []byte, cached Cached,
	w io.Writer, bfrom, bto int64) (n int64, err error) {

	n, hint, err := p.getFromCached(xl, cached, key, w, bfrom, bto)
	if err == nil {
		return
	}

	xl.Debug("rangeGet", from, to, bfrom, bto)
	if code := httputil.DetectCode(err); code != 404 {
		xl.Warnf("getFromCached failed: n: %v, err: %v", n, err)
		if bfrom+n >= bto || isWriterErr(err) {
			return
		}
		var n1 int64
		n1, err = p.retrys.Get(xl, fh, w, from+bfrom+n, from+bto)
		return n + n1, err
	}

	rc, _, err := p.getter.Get(xl, fh, from, to)
	if err != nil {
		xl.Warn("cachedGetter.Get: stg.Get failed =>", err)
		return
	}
	defer rc.Close()

	pr, pw := io.Pipe()
	go func(xl *xlog.Logger) {
		var err error
		if hint {
			err = cached.Set(xl, key, pr, to-from)
			if err != nil {
				xl.Warn("cachedGetter.Get: dc.Set failed =>", err)
			}
		} else {
			// No need to use buffer pool here
			// because ioutil.Discard implements io.ReadFrom,
			_, err = io.Copy(ioutil.Discard, pr)
		}
		pr.CloseWithError(err)
	}(xlog.NewWith(xl.ReqId()))

	mw := streamWriter(pw, w, bfrom, bto)
	n, err = bufpool.Copy(mw, rc)
	pw.CloseWithError(err)
	if err != nil {
		xl.Warn("cachedGetter.Get: Copy failed =>", err, n)
	}

	written := mw.Writers[1].Written
	if written >= bto {
		// w is full written.
		return bto - bfrom, nil
	}

	n = 0
	if written > bfrom {
		n = written - bfrom
	}

	if err = mw.Writers[1].Err; err != nil {
		// w is failed.
		xl.Warn("cachedGetter.Get: Copy w failed =>", err, n)
		return
	}

	// w is ok, but not full written.
	n1, err := p.retrys.Get(xl, fh, w, from+bfrom+n, from+bto)
	return n + n1, err
}
