package api

import (
	"crypto/md5"
	"encoding/base64"
	"io"
	"strconv"
	"strings"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	qio "github.com/qiniu/io"
	"github.com/qiniu/log.v1"
	rpc "github.com/qiniu/rpc.v1"
	xlog "github.com/qiniu/xlog.v1"
	"gopkg.in/mgo.v2/bson"
	"qbox.us/etag"
	"qbox.us/fh/ossbd"
	"qbox.us/net/httputil"
)

const (
	bigFileSize int64 = 5 * 1000 * 1000 * 1000         // 5G
	maxFileSize int64 = 50 * 1000 * 1000 * 1000 * 1000 // 50T
)

type Config struct {
	Endpoint          string   `json:"endpoint"`         // oss-cn-hangzhou.aliyuncs.com or https://oss-cn-hangzhou.aliyuncs.com
	Proxies           []string `json:"proxies"`          // http://ip:port
	BackupEndpoint    string   `json:"backup"`           // https://oss-cn-hangzhou.aliyuncs.com
	RetryPutBackup    bool     `json:"retry_put_backup"` // put 失败是否重试公网，当前重试pfd
	AK                string   `json:"ak"`
	SK                string   `json:"sk"`
	Bucket            string   `json:"bucket"`
	ConnectTimeoutS   int64    `json:"connect_timeout_s"`    // HTTP链接超时时间，单位是秒，默认10秒。0表示永不超时。
	ReadWriteTimeoutS int64    `json:"read_write_timeout_s"` // HTTP发送接受数据超时时间，单位是秒，默认20秒。0表示永不超时。
	RetryTimes        int      `json:"retry_times"`
	PutSplitSize      int64    `json:"put_split_size"` // 大文件分片上传每片大小
}

func New(cfg *Config) (c *Client, err error) {
	c = &Client{conf: cfg}
	if cfg.PutSplitSize <= 100*1024 {
		cfg.PutSplitSize = bigFileSize
	}
	if len(cfg.Proxies) == 0 {
		cfg.Proxies = append(cfg.Proxies, "")
	}
	if len(cfg.BackupEndpoint) != 0 {
		c.backup, err = oss.New(cfg.BackupEndpoint, cfg.AK, cfg.SK,
			oss.Timeout(cfg.ConnectTimeoutS, cfg.ReadWriteTimeoutS))
		if err != nil {
			log.Warn(err)
			return nil, err
		}
	}
	for _, proxy := range cfg.Proxies {
		var cli *oss.Client
		if proxy == "" {
			cli, err = oss.New(cfg.Endpoint, cfg.AK, cfg.SK,
				oss.Timeout(cfg.ConnectTimeoutS, cfg.ReadWriteTimeoutS))
		} else {
			cli, err = oss.New(cfg.Endpoint, cfg.AK, cfg.SK, oss.Proxy(proxy),
				oss.Timeout(cfg.ConnectTimeoutS, cfg.ReadWriteTimeoutS))
		}
		if err != nil {
			log.Warn(err)
			return nil, err
		}
		c.clis = append(c.clis, cli)
	}
	return c, nil
}

type Client struct {
	conf   *Config
	clis   []*oss.Client
	backup *oss.Client
}

var (
	ErrPutFailed        = httputil.NewError(500, "put failed")
	ErrGetFailed        = httputil.NewError(500, "get failed")
	ErrDelFailed        = httputil.NewError(500, "delete failed")
	ErrDocumentNotFound = httputil.NewError(404, "document not found")
	ErrFileTooLarge     = httputil.NewError(500, "file too large")
)

func (c *Client) Put(l rpc.Logger, f io.Reader, fsize int64) (fh []byte, md5 []byte, err error) {
	if fsize > bigFileSize {
		return c.putBigFile(l, f, fsize)
	}
	var (
		xl = xlog.NewWith(l)
		h  = newclients(c.clis)
	)
	for i := 0; i <= c.conf.RetryTimes; i++ {
		fh, md5, err = c.put(xl, h.Get(), f, fsize)
		if err == nil {
			return
		}
		if rt, ok := f.(io.ReaderAt); ok {
			f = &qio.Reader{rt, 0}
		} else {
			xl.Error("need retry, but reader is not ReaderAt")
			return
		}
	}
	if !c.conf.RetryPutBackup || c.backup == nil {
		err = ErrPutFailed
		return
	}
	xl.Info("put retry backup")
	return c.put(xl, c.backup, f, fsize)
}

func (c *Client) putBigFile(l rpc.Logger, f io.Reader, fsize int64) (fh []byte, md5 []byte, err error) {
	if fsize > maxFileSize {
		return nil, nil, ErrFileTooLarge
	}
	var (
		xl    = xlog.NewWith(l)
		h     = newclients(c.clis)
		objid = Reverse(bson.NewObjectId())
		key   = objid.Hex()
	)
	// init
	var initRet oss.InitiateMultipartUploadResult
	for i := 0; i <= c.conf.RetryTimes; i++ {
		initRet, err = h.GetBucket(c.conf.Bucket).InitiateMultipartUpload(key)
		if err == nil {
			break
		}
		xl.Warn("bucket.InitiateMultipartUpload failed", err)
		err = ErrPutFailed
	}
	if err != nil {
		return
	}
	xl.Info("bigfile UploadID", initRet.UploadID)
	// put parts
	h = newclients(c.clis)
	var etag []byte
	var parts []oss.UploadPart
	for i := 0; i <= c.conf.RetryTimes; i++ {
		etag, md5, parts, err = c.putParts(xl, h.GetBucket(c.conf.Bucket), initRet, f, fsize)
		if err == nil {
			break
		}
		xl.Warn("p.putParts failed", err)
		fh, err = nil, ErrPutFailed
		if rt, ok := f.(io.ReaderAt); ok {
			f = &qio.Reader{rt, 0}
		} else {
			xl.Error("need retry, but reader is not ReaderAt")
			xl.Warn("abort upload", c.AbortMultipartUpload(xl.Spawn(), initRet))
			return
		}
	}
	if err != nil {
		fh, err = nil, ErrPutFailed
		xl.Warn("abort upload", c.AbortMultipartUpload(xl.Spawn(), initRet))
		return
	}
	// complate
	h = newclients(c.clis)
	for i := 0; i <= c.conf.RetryTimes; i++ {
		_, err = h.GetBucket(c.conf.Bucket).CompleteMultipartUpload(initRet, parts)
		if err == nil {
			break
		}
		xl.Warn("CompleteMultipartUpload failed", err)
		err = ErrPutFailed
	}
	if err != nil {
		xl.Warn("abort upload", c.AbortMultipartUpload(xl.Spawn(), initRet))
		return
	}
	fhi := ossbd.NewInstance(fsize, etag, c.conf.Bucket, []byte(objid))
	xl.Info("put fh", fhi)
	return fhi, md5, nil
}

var testfunc = func(initRet oss.InitiateMultipartUploadResult) {}

func (c *Client) AbortMultipartUpload(xl *xlog.Logger, initRet oss.InitiateMultipartUploadResult) (err error) {
	testfunc(initRet)
	h := newclients(c.clis)
	for i := 0; i <= c.conf.RetryTimes; i++ {
		err = h.GetBucket(c.conf.Bucket).AbortMultipartUpload(initRet)
		if err == nil {
			break
		}
		xl.Warn(err)
	}
	return
}

func (c *Client) putParts(xl *xlog.Logger, bucket *oss.Bucket, initRet oss.InitiateMultipartUploadResult, f io.Reader, fsize int64) (fileEtag, fileMd5 []byte, parts []oss.UploadPart, err error) {
	hash := etag.New(1 << 22)
	md5 := md5.New()
	pr, pw := io.Pipe()
	ws := io.MultiWriter(hash, pw, md5)
	go func(xl *xlog.Logger) {
		_, err := bcopy(ws, f)
		if err != nil {
			xl.Warn("copy data failed", err)
			pw.CloseWithError(err)
		}
	}(xl.Spawn())
	partCount := int(fsize / c.conf.PutSplitSize)
	for i := 0; i <= partCount; i++ {
		size := c.conf.PutSplitSize
		if i == partCount {
			size = fsize % c.conf.PutSplitSize
		}
		if size == 0 {
			break
		}
		part, err := bucket.UploadPart(initRet, io.LimitReader(pr, size), size, i+1)
		if err != nil {
			xl.Warn("bucket.UploadPart failed", err)
			return nil, nil, nil, err
		}
		parts = append(parts, part)
	}
	fileEtag, fileMd5 = hash.Sum(), md5.Sum(nil)
	return
}

func (c *Client) put(xl *xlog.Logger, cli *oss.Client, f io.Reader, fsize int64) (fh []byte, fileMd5 []byte, err error) {
	objid := Reverse(bson.NewObjectId())
	key := objid.Hex()
	bucket, oerr := cli.Bucket(c.conf.Bucket)
	if oerr != nil {
		xl.Warn("get bucket failed", oerr)
		fh, err = nil, ErrPutFailed
		return
	}
	hash := etag.New(1 << 22)
	pr, pw := io.Pipe()
	ws := io.MultiWriter(hash, pw)
	go func(xl *xlog.Logger) {
		_, err := bcopy(ws, f)
		if err != nil {
			xl.Warn("copy data failed", err)
			pw.CloseWithError(err)
		}
	}(xl.Spawn())
	request := &oss.PutObjectRequest{
		ObjectKey: key,
		Reader:    io.LimitReader(pr, fsize),
	}
	resp, err := bucket.DoPutObject(request, nil)
	if err != nil {
		xl.Warn("put failed", err)
		err = ErrPutFailed
		return
	}
	defer resp.Body.Close()
	fileMd5, _ = base64.StdEncoding.DecodeString(strings.Trim(resp.Headers.Get("Content-Md5"), "\""))
	if len(fileMd5) != md5.Size {
		fileMd5 = nil
	}
	etag := hash.Sum()
	fhi := ossbd.NewInstance(fsize, etag, c.conf.Bucket, []byte(objid))
	xl.Info("put fh", fhi)
	return fhi, fileMd5, nil
}

func (c *Client) Get(l rpc.Logger,
	fh []byte, from, to int64) (rc io.ReadCloser, fsize int64, err error) {
	rc, _, fsize, err = c.GetWithMd5(l, fh, from, to)
	return
}

// oss只有normal object才会有文件的md5，分片上传的文件没有md5，返回 nil
func (c *Client) GetWithMd5(l rpc.Logger,
	fh []byte, from, to int64) (rc io.ReadCloser, md5 []byte, fsize int64, err error) {
	var (
		xl = xlog.NewWith(l)
		h  = newclients(c.clis)
	)
	fhi := ossbd.Instance(fh)
	xl.Info("get fh:", fhi)
	for i := 0; i <= c.conf.RetryTimes; i++ {
		rc, md5, fsize, err = c.get(xl, h.Get(), fhi, from, to)
		if err == nil || err == ErrDocumentNotFound {
			return
		}
	}
	if c.backup != nil {
		xl.Info("get retry backup")
		return c.get(xl, c.backup, fhi, from, to)
	}
	return
}

func (c *Client) get(xl *xlog.Logger, cli *oss.Client,
	fhi ossbd.Instance, from, to int64) (rc io.ReadCloser, md5 []byte, fsize int64, err error) {

	bucket, oerr := cli.Bucket(fhi.Bucket())
	if oerr != nil {
		xl.Warn("get bucket failed", oerr)
		err = ErrGetFailed
		return
	}
	if from >= to || fhi.Fsize() == 0 {
		return nullReadCloser{}, nil, 0, nil
	}
	result, oerr := bucket.DoGetObject(&oss.GetObjectRequest{fhi.Key()}, []oss.Option{oss.Range(from, to-1)})
	if oerr != nil {
		xl.Warn("get failed", fhi.OssEncode(), oerr)
		if strings.Contains(oerr.Error(), "ErrorCode=NoSuchKey") {
			err = ErrDocumentNotFound
			return
		}
		err = ErrGetFailed
		return
	}
	md5, _ = base64.StdEncoding.DecodeString(strings.Trim(result.Response.Headers.Get("Content-Md5"), "\""))
	if len(md5) == 0 {
		md5 = nil
	}
	fsize, err = strconv.ParseInt(result.Response.Headers.Get(oss.HTTPHeaderContentLength), 10, 64)
	if err != nil {
		xl.Warn("Parse Content-Length failed", err, result.Response.Headers)
		err = ErrGetFailed
		result.Response.Body.Close()
		return
	}
	rc = result.Response.Body
	return
}

func (c *Client) Delete(l rpc.Logger, fh []byte) (err error) {
	var (
		xl = xlog.NewWith(l)
		h  = newclients(c.clis)
	)
	fhi := ossbd.Instance(fh)
	xl.Info("del fh:", fhi)
	for i := 0; i <= c.conf.RetryTimes; i++ {
		err = c.delete(xl, h.Get(), fhi)
		if err == nil {
			return
		}
	}
	if c.backup != nil {
		xl.Info("del retry backup")
		return c.delete(xl, c.backup, fhi)
	}
	return
}

func (c *Client) delete(xl *xlog.Logger, cli *oss.Client, fhi ossbd.Instance) (err error) {
	bucket, oerr := cli.Bucket(fhi.Bucket())
	if oerr != nil {
		xl.Warn("get bucket failed", oerr)
		err = ErrGetFailed
		return
	}
	err = bucket.DeleteObject(fhi.Key())
	if err != nil {
		xl.Warn("delete failed", fhi.OssEncode(), err)
		err = ErrDelFailed
		return
	}
	return
}

type nullReadCloser struct{}

func (self nullReadCloser) Read(p []byte) (int, error) {
	return 0, io.EOF
}

func (self nullReadCloser) Close() error {
	return nil
}
