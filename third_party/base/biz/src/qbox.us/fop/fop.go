package fop

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"code.google.com/p/go.net/context"
	"github.com/qiniu/errors"
	"github.com/qiniu/http/httputil.v1"
	"github.com/qiniu/rpc.v1"
	"github.com/qiniu/xlog.v1"

	"qbox.us/dcutil"
	"qbox.us/etag"
	"qbox.us/servend/account"
)

// ----------------------------------------------------------------------------

var (
	ErrGlbNoRegion = httputil.NewError(400, "global api should has region prefix")
)

const (
	regionMagicPrefix = "$(region)/"
)

type Source io.Reader

type Writer http.ResponseWriter
type Reader *Request

type Request struct {
	Source
	Env
	Ctx        context.Context
	SourceURL  string
	Fsize      int64    // must. -1 if unknown
	URL        string   // must. scheme://host/path[?query]
	RawQuery   string   // must. 不带 Token 的 QueryString
	Query      []string // deprecated. use RawQuery please
	Token      string   // opt. 下载授权凭证（对于私有资源而言）
	StyleParam string   // opt.
	Etag       string   // opt. empty if unknown
	MimeType   string   // opt. empty if unknown
	IsGlobal   int
	Range      string // opt. empty if unknow

	// fop persistence
	Mode                 uint32
	Uid                  uint32
	Bucket               string
	Key                  string
	Fh                   []byte
	OutRSBucket          string
	OutRSDeleteAfterDays uint32
}

func DownloadSource(srcFilePrefix string, r Reader) (fname string, err error) {
	xl := r.Xlogger()

	start := time.Now()
	f, err := r.LocalCache().TempFile(srcFilePrefix, r.Fsize)
	if err != nil {
		return
	}
	defer f.Close()
	fname = f.Name()

	now := time.Now()
	fileCreationCost := now.Sub(start)

	start = now
	n, err := io.Copy(f, r.Source)
	if err != nil {
		xl.Errorf("copy source failed, %d bytes copied with %s", n, err)
		os.Remove(f.Name())
		return
	}
	downloadCost := time.Now().Sub(start)

	xl.Infof("dl source %s, file creation cost %v, dl cost %v", fname, fileCreationCost, downloadCost)
	return
}

// CacheKey caculate dc key of realtime fop.
func CacheKey(fh []byte, fsize int64, fopCmd, styleParam string) []byte {
	key := "?fh=" + base64.URLEncoding.EncodeToString(fh) + "&fsize=" + strconv.FormatInt(fsize, 10)
	if fopCmd != "" {
		key += "&cmd=" + fopCmd
	}
	if styleParam != "" {
		key += "&sp=" + styleParam
	}
	h := sha1.New()
	io.WriteString(h, key)
	return h.Sum(nil)
}

// PersistentKey caculate rs key of persistence fop.
// key = ${prefix}/${etag}[/${styleParam}]
func PersistentKey(fh []byte, fopCmd, styleParam string) string {
	h := sha1.New()
	io.WriteString(h, fopCmd)
	prefix := base64.URLEncoding.EncodeToString(h.Sum(nil))

	etag := etag.GenString(fh)
	key := prefix + "/" + etag
	if styleParam != "" {
		key += "/" + styleParam
	}
	return key
}

func PersistentKeyEx(fh []byte, fopCmd, styleParam string) string {
	return ".qiniu/" + PersistentKey(fh, fopCmd, styleParam)
}

type LocalCache interface {
	TempDir(fsize int64) (dir string, err error)
	TempFile(prefix string, fsize int64) (f *os.File, err error)
}

type Env interface {
	Xlogger() *xlog.Logger // must.
	Xdc() dcutil.Interface // opt. maybe is nil
	TempDir() string       // 推荐使用 LocalCache()，防止磁盘 io 跟不上卡住请求
	LocalCache() LocalCache
	Acc() account.InterfaceEx
}

type FopCtx struct {
	CmdName    string // 处理命令，可以是 avthumb, imageView2, ufop 等等
	RawQuery   string // 处理命令
	RawQueries string // 对于管道来说记录了从第一个处理命令到当前处理命令的信息，该记录最后一个命令为 RawQuery
	SourceURL  string // 待处理资源的 URL
	MimeType   string // 待处理资源的 MimeType
	URL        string // 从 IO 实时访问 Fop 的 URL
	Token      string // deprecated
	StyleParam string // deprecated

	IsGlobal int // 判断这个fop是否global的请求，后端需要对这个做对应的事情或者不做

	// 持久化相关字段
	Mode    uint32 // 处理模式，1 表示来自 IO 的实时持久化请求，2 表示是来自 PFOP 的异步持久化请求
	Uid     uint32 // 待处理资源所属用户
	Bucket  string // 待处理资源所在的 Bucket
	Key     string // 待处理资源的 Key
	Version string // 待处理资源的 Version
	Fh      []byte // 待处理资源的 Fh
	Force   uint32 // 0 表示会判断处理结果是否已经存在，如果存在则不进行处理；1 表示不做判断直接做处理

	// IO 向 FOPG 发起的一个请求参数，表明接收 FOPG 返回处理结果的地址
	AcceptAddrOut string

	// FOPAGENT 处理输出
	globalParsed         bool
	OutType              string // 处理后保存到哪里，可以是 dc || rs || tmp
	OutRSBucket          string // 处理结果 RS 的 Bucket
	OutRSKey             string // 处理结果 RS 的 Key
	OutRSDeleteAfterDays uint32 // 处理结果 RS 的 deleteAfterDays参数
	OutDCKey             []byte // 处理结果 DC 的 Key

	// 计费字段 对于异步任务，这里纪录了使用的队列id，在计费时，仅对客户私有队列收费，公有队列不收费
	PipelineId string

	// 记录了pipe.Exec中当前命令之前的命令产生的收费字断，需要通过设置http header 中的X-Qiniu-Fop-Stats，来传递给fopagent
	PreviousXstats string
}

func (ctx *FopCtx) ParseGlobalKey(regionPrefix string) (err error) {

	if ctx.IsGlobal != 1 || ctx.globalParsed {
		return
	}
	ctx.globalParsed = true

	if ctx.OutType == "rs" {
		if strings.HasPrefix(ctx.OutRSKey, regionPrefix) {
			ctx.OutRSKey = ctx.OutRSKey[len(regionPrefix):]
		} else if strings.HasPrefix(ctx.OutRSKey, regionMagicPrefix) {
			ctx.OutRSKey = ctx.OutRSKey[len(regionMagicPrefix):]
		} else {
			err = ErrGlbNoRegion
			ctx.globalParsed = false
		}
	}

	return
}

//----------------------------------------------------------------------
// Fop PersistType

const (
	SyncPersistMode  = 1
	AsyncPersistMode = 2
)

//----------------------------------------------------------------------
// Fop Output

const (
	OutTypeDC     = "dc"  // 结果存入 DC，输出结果在 DC 的 Host 和 Key
	OutTypeRS     = "rs"  // 结果存入 RS，输出结果在 RS 的 Bucket 和 Key
	OutTypeTmp    = "tmp" // 结果存入临时存储，输出结果是一个 URL，特点是阅后即焚
	OutTypeStream = ""    // 流输出，非以上 OutType 都是流输出，这里用空字符串表示
)

type Out struct {
	Type              string
	RSBucket          string
	RSKey             string
	RSDeleteAfterDays uint32
	DCKey             []byte
}

type DCOut struct {
	Host string `json:"host"`
	Key  []byte `json:"key"`
}

type RSOut struct {
	Bucket string `json:"bucket"`
	Key    string `json:"key"`
	Hash   string `json:"hash"`
}

type TmpOut struct {
	URL      string `json:"url"`
	MimeType string `json:"mimeType"`
}

const OutHeader = "X-Qiniu-Fop-Out" // 在 HTTP Header 中指定 Fop 处理结果的输出方式。

func GetOutType(resp *http.Response) string {
	outType := resp.Header.Get(OutHeader)
	switch outType {
	case OutTypeDC, OutTypeRS, OutTypeTmp:
		return outType
	default:
		return OutTypeStream
	}
}

var ErrNotAddrResp = errors.New("fop: not addr response")

func DecodeAddrOut(out interface{}, resp *http.Response) error {
	if outType := GetOutType(resp); outType == OutTypeStream {
		return ErrNotAddrResp
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

func ExtractResponse(xl *xlog.Logger, resp *http.Response) (r io.ReadCloser, length int64, mime string, needRetry, noCdnCache bool, err error) {
	outType := GetOutType(resp)
	if resp.Header.Get("X-Cdn-Cache-Control") == "no-cache" {
		noCdnCache = true
	}
	switch outType {
	case OutTypeDC:
		var out DCOut
		err = DecodeAddrOut(&out, resp)
		resp.Body.Close()
		if err != nil {
			err = errors.Info(err, "ExtractResponse: DecodeAddrOut DCOut")
			return
		}
		xl.Infof("ExtractResponse: dcAddrOut, dcHost:%s, dcKey:%s", out.Host, base64.URLEncoding.EncodeToString(out.Key))
		var metas map[string]string
		r, length, metas, err = dcutil.GetWithHost(xl, out.Host, out.Key)
		if err != nil {
			needRetry = true
			err = errors.Info(err, "ExtractResponse: dcutil.GetWithHost")
			return
		}
		mime = metas["mime"]

	case OutTypeTmp:
		var out TmpOut
		err = DecodeAddrOut(&out, resp)
		resp.Body.Close()
		if err != nil {
			err = errors.Info(err, "ExtractResponse: DecodeAddrOut TmpOut")
			return
		}
		xl.Infof("ExtractResponse: tmpAddrOut, tmpURL:%s, mimeType:%s", out.URL, out.MimeType)
		mime = out.MimeType
		resp2, err2 := rpc.DefaultClient.Get(xl, out.URL)
		if err2 != nil {
			needRetry = true
			err = errors.Info(err2, "ExtractResponse: rpcClient.Get tmpOut.URL")
			return
		}
		if resp2.StatusCode != http.StatusOK {
			err = rpc.ResponseError(resp2)
			resp2.Body.Close()
			return
		}
		r, length = resp2.Body, resp2.ContentLength

	default:
		xl.Info("ExtractResponse: streamOut")
		r, length, mime = resp.Body, resp.ContentLength, resp.Header.Get("Content-Type")
	}
	return
}

// ----------------------------------------------------------------------------
