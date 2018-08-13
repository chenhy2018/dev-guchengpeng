package rs

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"strconv"
	"strings"

	"qbox.us/api"
	"qbox.us/api/tblmgr"
	"qbox.us/api/up"
	"qbox.us/cc/time"
	"qbox.us/errors"
	"qbox.us/rpc"

	"github.com/qiniu/http/httputil.v1"
	"github.com/qiniu/log.v1"
	qiniurpc "github.com/qiniu/rpc.v1"
)

// ----------------------------------------------------------

const (
	FILE_NOT_FOUND_KEY = "errno-404"
	RS_PUT             = "/rs-put/"
	UNDEFINED_KEY      = "?"
)
const (
	FileModified   = 608 // RS: 文件被修改（see fs.GetIfNotModified）
	NoSuchEntry    = 612 // RS: 指定的 Entry 不存在或已经 Deleted
	EntryExists    = 614 // RS: 要创建的 Entry 已经存在
	TooManyBuckets = 630 // RS: 创建的 Bucket 个数过多
	NoSuchBucket   = 631 // RS: 指定的 Bucket 不存在
)

var (
	EFileModified   = api.RegisterError(FileModified, "file modified")
	ENoSuchEntry    = api.RegisterError(NoSuchEntry, "no such file or directory")
	EEntryExists    = api.RegisterError(EntryExists, "file exists")
	ETooManyBuckets = api.RegisterError(TooManyBuckets, "too many buckets")
	ENoSuchBucket   = httputil.NewError(NoSuchBucket, "no such bucket")
)

// ----------------------------------------------------------
// rsf只提供list接口，用于列出bucket中所有的文件
// rs提供增删改查操作
type Service struct {
	client  rpc.Client
	rsHost  string
	rsfHost string
	upHost  string
}

func NewService(t http.RoundTripper, rsHost, rsfHost, upHost string) *Service {
	client := &http.Client{Transport: t}
	return &Service{
		client:  rpc.Client{client},
		rsHost:  rsHost,
		rsfHost: rsfHost,
		upHost:  upHost,
	}
}

// ----------------------------------------------------------

type PutRet struct {
	Hash string `json:"hash"`
	Key  string `json:"key"`
}

//
// Put 用于上传一个文件到 RS Bucket 中。
//
// entryURI 指定了该文件在 RS Bucket 中的路径。其格式为 <BucketName>:<Path>
// mimeType 指定了文件的 mime 类型。如果 mimeType 为空，则自动检测该文件的 mime 类型。
// f, fsize 指定了文件内容及文件大小。
// customMeta 指定了该文件的用户自定义元信息。如果 customMeta 为空表示无用户自定义元信息。
// crc32 指定了该文件内容的 crc32 校验值。如果 crc32 值非空，则服务端对上传的文件内容进行 crc 校验。
//
func (rs *Service) Put(
	entryURI, mimeType string, f io.Reader, fsize int64,
	customMeta, callbackParams, crc32 string) (ret PutRet, code int, err error) {

	return Put(rs.client, rs.upHost, entryURI, mimeType, f, fsize, customMeta, callbackParams, crc32, "")
}

func (rs *Service) PutEx(
	entryURI, mimeType string, f io.Reader, fsize int64,
	customMeta, callbackParams, crc32, endUser string) (ret PutRet, code int, err error) {

	return Put(rs.client, rs.upHost, entryURI, mimeType, f, fsize, customMeta, callbackParams, crc32, endUser)
}

func Put(
	conn rpc.Client, ioHost string,
	entryURI, mimeType string, f io.Reader, fsize int64,
	customMeta, callbackParams, crc32, endUser string) (ret PutRet, code int, err error) {

	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	url := ioHost + RS_PUT + rpc.EncodeURI(entryURI) + "/mimeType/" + rpc.EncodeURI(mimeType)
	if customMeta != "" {
		url += "/meta/" + rpc.EncodeURI(customMeta)
	}
	if endUser != "" {
		url += "/endUser/" + endUser
	}
	if crc32 != "" {
		url += "/crc32/" + crc32
	}
	log.Debug("rs.Put", url)
	code, err = conn.CallWithBinary64(&ret, url, f, fsize)
	return
}

func (rs *Service) Put2(ret interface{}, b64 bool, key string, localFile string,
	mimeType string, crc32 string, params map[string]string) error {

	f, err := os.Open(localFile)
	if err != nil {
		return err
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return err
	}
	fsize := fi.Size()

	var url string
	if b64 {
		url = rs.upHost + "/putb64/" + strconv.FormatInt(fsize, 10)
	} else {
		url = rs.upHost + "/put/" + strconv.FormatInt(fsize, 10)
	}
	if mimeType != "" {
		url += "/mimeType/" + base64.URLEncoding.EncodeToString([]byte(mimeType))
	}
	if crc32 != "" {
		url += "/crc32/" + crc32
	}
	for k, v := range params {
		if strings.HasPrefix(k, "x:") && v != "" {
			url += "/" + k + "/" + base64.URLEncoding.EncodeToString([]byte(v))
		}
	}
	if key != UNDEFINED_KEY {
		url += "/key/" + base64.URLEncoding.EncodeToString([]byte(key))
	}
	log.Info("rs.Put2:", url)

	if !b64 {
		_, err = rs.client.CallWithBinary64(ret, url, f, fsize)
		return err
	}

	// base64 encoding file
	ch := make(chan error)
	pr, pw := io.Pipe()
	go func(ch chan error) {
		_, e := rs.client.CallWithBinary(ret, url, pr, base64.StdEncoding.EncodedLen(int(fsize)))
		pr.CloseWithError(e)
		ch <- e
	}(ch)

	w := base64.NewEncoder(base64.StdEncoding, pw)
	_, err = io.Copy(w, f)
	if err != nil {
		pw.CloseWithError(err)
		return <-ch
	}
	pw.CloseWithError(w.Close())
	return <-ch
}

func (rs *Service) Put2WithStream(ret interface{}, b64 bool, key string, b64f io.Reader, fsize int64,
	mimeType string, crc32 string, params map[string]string) error {

	var url string
	if b64 {
		url = rs.upHost + "/putb64/" + strconv.FormatInt(fsize, 10)
	} else {
		url = rs.upHost + "/put/" + strconv.FormatInt(fsize, 10)
	}
	if mimeType != "" {
		url += "/mimeType/" + base64.URLEncoding.EncodeToString([]byte(mimeType))
	}
	if crc32 != "" {
		url += "/crc32/" + crc32
	}
	for k, v := range params {
		if strings.HasPrefix(k, "x:") && v != "" {
			url += "/" + k + "/" + base64.URLEncoding.EncodeToString([]byte(v))
		}
	}
	if key != UNDEFINED_KEY {
		url += "/key/" + base64.URLEncoding.EncodeToString([]byte(key))
	}
	log.Info("rs.Put2:", url)

	_, err := rs.client.CallWithBinary64(ret, url, b64f, -1)
	return err
}

//
// ResumablePut 用于上传一个文件到 RS Bucket 中。与 Put 不一样的是，ResumablePut 支持断点续上传。
// 断点续上传的机理是：先将文件内容按 4M 切分为多个 Chunk。各个 Chunk 可并行上传，互不干扰。
// 每个 Chunk 上传的时候，可自由的切分为多个 Block 上传。Block 大小可自由控制，最佳值应按用户网络状况智能确定。
// Block 太小会导致上传过慢，太大则可能导致 Block 上传失败，做无用功。
//
// notify 指定了上传一个 Chunk/Block 成功的回调。它通常用于持久化上传进度用。
//
func (rs *Service) ResumablePut(
	uploader up.Service, checksums []string, progs []up.BlockputProgress,
	blockNotify func(blockIdx int, checksum string), chunkNotify func(blockIdx int, prog *up.BlockputProgress),
	entryURI, mimeType string, f io.ReaderAt, fsize int64, customMeta, callbackParams string) (ret PutRet, code int, err error) {

	return rs.ResumablePutEx(
		uploader, checksums, progs, blockNotify, chunkNotify,
		entryURI, mimeType, f, fsize, customMeta, callbackParams, "")
}

func (rs *Service) ResumablePutEx(
	uploader up.Service, checksums []string, progs []up.BlockputProgress,
	blockNotify func(blockIdx int, checksum string),
	chunkNotify func(blockIdx int, prog *up.BlockputProgress),
	entryURI, mimeType string, f io.ReaderAt, fsize int64,
	customMeta, callbackParams, customer string) (ret PutRet, code int, err error) {

	code, err = uploader.Put(rs.client, f, fsize, checksums, progs, blockNotify, chunkNotify)
	if err != nil {
		err = errors.Info(err, "rs.ResumablePut", entryURI, fsize).Detail(err)
		return
	}

	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	params := "/mimeType/" + rpc.EncodeURI(mimeType)
	if customMeta != "" {
		params += "/meta/" + rpc.EncodeURI(customMeta)
	}
	if customer != "" {
		params += "/customer/" + customer
	}

	ctxs := make([]string, len(progs))
	for i, prog := range progs {
		ctxs[i] = prog.Ctx
	}

	action := RS_PUT[:len(RS_PUT)-4] + "mkfile/"
	code, err = up.Mkfile(rs.client, &ret, action, entryURI, fsize, params, callbackParams, ctxs)
	if err != nil {
		err = errors.Info(err, "rs.ResumablePut", entryURI, fsize).Detail(err)
	}
	return
}

func (rs Service) ResumablePut2(
	ret interface{}, uploader up.Service, checksums []string, progs []up.BlockputProgress,
	blockNotify func(blockIdx int, checksum string),
	chunkNotify func(blockIdx int, prog *up.BlockputProgress),
	key, mimeType string, f io.ReaderAt, fsize int64, params map[string]string) (code int, err error) {

	code, err = uploader.Put(rs.client, f, fsize, checksums, progs, blockNotify, chunkNotify)
	if err != nil {
		err = errors.Info(err, "rs.ResumablePut2", key, fsize).Detail(err)
		return
	}

	ctxs := make([]string, len(progs))
	for i, prog := range progs {
		ctxs[i] = prog.Ctx
	}

	return up.Mkfile2(rs.client, &ret, key, mimeType, fsize, params, ctxs)
}

// ----------------------------------------------------------

type GetRet struct {
	EncodedFh string `json:"fh"` // base64 编码. Note: 对 pfd 需要 urlsafe base64 编码
	URL       string `json:"url"`
	Hash      string `json:"hash"`
	MimeType  string `json:"mimeType"`
	Fsize     int64  `json:"fsize"`
	Expiry    int64  `json:"expires"`
}

type Entry struct {
	Hash     string            `json:"hash"`
	Fsize    int64             `json:"fsize"`
	PutTime  int64             `json:"putTime"`
	MimeType string            `json:"mimeType"`
	EndUser  string            `json:"endUser"`
	Type     FileType          `json:"type,omitempty"`
	XMeta    map[string]string `json:"x-qn-meta,omitempty"`
}

type ListItem struct {
	Name     string `json:"name"`
	Hash     string `json:"hash"`
	Fsize    int64  `json:"fsize"`
	PutTime  int64  `json:"putTime"`
	MimeType string `json:"mimeType"`
	EndUser  string `json:"endUser"`
}

func (rs *Service) Get(entryURI string, attName string) (data GetRet, code int, err error) {
	return rs.GetWithExpires(entryURI, attName, -1)
}

func (rs *Service) GetWithExpires(entryURI string, attName string, expires int) (data GetRet, code int, err error) {
	url := rs.rsHost + "/get/" + rpc.EncodeURI(entryURI)
	if attName != "" {
		url = url + "/attName/" + rpc.EncodeURI(attName)
	}
	if expires > 0 {
		url = url + "/expires/" + strconv.Itoa(expires)
	}

	code, err = rs.client.Call(&data, url)
	if code == 200 {
		data.Expiry += time.Seconds()
	}
	return
}

func (rs *Service) GetIfNotModified(entryURI string, attName string, base string) (data GetRet, code int, err error) {
	url := rs.rsHost + "/get/" + rpc.EncodeURI(entryURI) + "/base/" + base
	if attName != "" {
		url = url + "/attName/" + rpc.EncodeURI(attName)
	}

	code, err = rs.client.Call(&data, url)
	if code == 200 {
		data.Expiry += time.Seconds()
	}
	return
}

func (rs *Service) Stat(entryURI string) (entry Entry, code int, err error) {
	code, err = rs.client.Call(&entry, rs.rsHost+"/stat/"+rpc.EncodeURI(entryURI))
	return
}

func (rs *Service) Delete(entryURI string) (code int, err error) {
	return rs.client.Call(nil, rs.rsHost+"/delete/"+rpc.EncodeURI(entryURI))
}

func (rs *Service) Move(entryURISrc, entryURIDest string) (code int, err error) {
	return rs.client.Call(nil, rs.rsHost+"/move/"+rpc.EncodeURI(entryURISrc)+"/"+rpc.EncodeURI(entryURIDest))
}

func (rs *Service) Copy(entryURISrc, entryURIDest string) (code int, err error) {
	return rs.client.Call(nil, rs.rsHost+"/copy/"+rpc.EncodeURI(entryURISrc)+"/"+rpc.EncodeURI(entryURIDest))
}

func (rs *Service) Chgm(entryURI, mimeType string) (code int, err error) {
	return rs.client.Call(nil, rs.rsHost+"/chgm/"+rpc.EncodeURI(entryURI)+"/mime/"+rpc.EncodeURI(mimeType))
}

type FileType uint32

const (
	TypeNormal = iota
	TypeLine
)

func (rs *Service) ChType(entryURI string, Type FileType) (code int, err error) {
	return rs.client.Call(nil, rs.rsHost+"/chtype/"+rpc.EncodeURI(entryURI)+"/type/"+fmt.Sprint(Type))
}

func (rs *Service) ChMeta(entryURI, metaKey, metaValue string) (code int, err error) {
	return rs.client.Call(nil, rs.rsHost+"/chgm/"+rpc.EncodeURI(entryURI)+"/x-qn-meta-"+metaKey+"/"+rpc.EncodeURI(metaValue))
}

func (rs *Service) ChMetas(entryURI string, kv map[string]string) (code int, err error) {
	u := rs.rsHost + "/chgm/" + rpc.EncodeURI(entryURI)
	for k, v := range kv {
		u += "/x-qn-meta-" + k + "/" + rpc.EncodeURI(v)
	}
	return rs.client.Call(nil, u)
}

func (rs Service) Mkbucket(bucketName string) (code int, err error) {
	return rs.client.Call(nil, rs.rsHost+"/mkbucket/"+bucketName)
}

func (rs *Service) Mkbucket2(bucketName string, public int) (code int, err error) {
	return rs.client.Call(nil, rs.rsHost+"/mkbucket2/"+bucketName+"/public/"+strconv.Itoa(public))
}

type BucketRet struct {
	Tbl    string `json:"tbl"`
	Itbl   uint32 `json:"itbl"`
	PhyTbl string `json:"phy"`
	Uid    uint32 `json:"uid"`
	Zone   string `json:"zone"`
}

func (rs *Service) Bucket(bucketName string) (bucketEntry BucketRet, code int, err error) {
	code, err = rs.client.Call(&bucketEntry, rs.rsHost+"/bucket/"+bucketName)
	return
}

func (rs *Service) MkbucketWithZone(bucketName, zone string) (code int, err error) {
	url := rs.rsHost + "/mkbucket/" + bucketName
	if zone != "" {
		url += "/zone/" + zone
	}

	return rs.client.Call(nil, url)
}

func (rs Service) Buckets() (buckets []string, code int, err error) {
	return rs.BucketsWithSharedV2(tblmgr.RW)
}

//兼容废弃
func (rs Service) BucketsWithShared(shared bool) (buckets []string, code int, err error) {
	var url = rs.rsHost + "/buckets"
	if shared {
		url += "?shared=true"
	}
	code, err = rs.client.Call(&buckets, url)
	return
}

func (rs Service) BucketsWithSharedV2(share_flag int32) (buckets []string, code int, err error) {
	var url = rs.rsHost + "/buckets"
	shared := "rw"
	if share_flag == tblmgr.RD {
		shared = "rd"
	} else if share_flag != tblmgr.RW {
		err = errors.New("invalid share_flag")
		return
	}
	url += fmt.Sprintf("?shared=%s", shared)
	code, err = rs.client.Call(&buckets, url)
	return
}

// ----------------------------------------------------------

func (rs *Service) List(entryURI string, skip, limit int) (listRet []ListItem, code int, err error) {
	url := rs.rsfHost + "/list/" + rpc.EncodeURI(entryURI)
	if skip > 0 {
		url = url + "/skip/" + strconv.Itoa(skip)
	}
	if limit > 0 {
		url = url + "/limit/" + strconv.Itoa(limit)
	}
	code, err = rs.client.Call(&listRet, url)
	return
}

func (rs *Service) Mkdir(entryURI string) (code int, err error) {
	return rs.client.Call(nil, rs.rsHost+"/mkdir/"+rpc.EncodeURI(entryURI))
}

// ----------------------------------------------------------

func (rs *Service) GetRealTable(bucketAlias string) (bucket string, err error) {
	resp, err := rs.client.Get(rs.rsHost + "/realbucket?bucket=" + bucketAlias)
	if err != nil {
		log.Warn("GetRealTable response error : ", err)
		err = api.EInternalError
		return
	}
	if resp.StatusCode != 200 {
		if resp.StatusCode == NoSuchBucket {
			err = ENoSuchBucket
			return
		}
		err = api.EInternalError
		log.Warn("GetRealTable response code error ", resp.StatusCode)
		return
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Warn("GetRealTable response body error : ", err)
		err = api.EInternalError
		return
	}

	bucket = string(data)
	return
}

func (rs *Service) Drop(bucketName string) (code int, err error) {
	return rs.client.Call(nil, rs.rsHost+"/drop/"+bucketName)
}

// ----------------------------------------------------------

type BatchRet struct {
	Data  interface{} `json:"data"`
	Code  int         `json:"code"`
	Error string      `json:"error"`
}

type Batcher struct {
	op  []string
	ret []BatchRet
}

func (b *Batcher) operate(entryURI string, method string) {
	b.op = append(b.op, method+rpc.EncodeURI(entryURI))
	b.ret = append(b.ret, BatchRet{})
}

func (b *Batcher) operate2(entryURISrc, entryURIDest string, method string) {
	b.op = append(b.op, method+rpc.EncodeURI(entryURISrc)+"/"+rpc.EncodeURI(entryURIDest))
	b.ret = append(b.ret, BatchRet{})
}

func (b *Batcher) Get(entryURI string) {
	b.operate(entryURI, "/get/")
}

func (b *Batcher) Delete(entryURI string) {
	b.operate(entryURI, "/delete/")
}

func (b *Batcher) Move(entryURISrc, entryURIDest string) {
	b.operate2(entryURISrc, entryURIDest, "/move/")
}

func (b *Batcher) Copy(entryURISrc, entryURIDest string) {
	b.operate2(entryURISrc, entryURIDest, "/copy/")
}

func (b *Batcher) Reset() {
	b.op = nil
	b.ret = nil
}

func (b *Batcher) Len() int {
	return len(b.op)
}

func (b *Batcher) Do(rs *Service) (ret []BatchRet, code int, err error) {
	code, err = rs.client.CallWithForm(&b.ret, rs.rsHost+"/batch", map[string][]string{"op": b.op})
	ret = b.ret
	return
}

// ----------------------------------------------------------

func Upload(upHost, entryURI, localFile, mimeType, customMeta, callbackParam, upToken string) (ret PutRet, code int, err error) {

	code, err = UploadEx2(&ret, upHost, upToken, localFile, entryURI, mimeType, customMeta, callbackParam, -1, -1)
	return
}

func UploadEx(upHost, upToken, localFile, entryURI, mimeType, customMeta, callbackParam string,
	crc int64, rotate int) (ret PutRet, code int, err error) {

	code, err = UploadEx2(&ret, upHost, upToken, localFile, entryURI, mimeType, customMeta, callbackParam, crc, rotate)
	return
}

func Upload2(upHost, upToken, localFile, entryURI, mimeType, customMeta, callbackParam string,
	crc int64, rotate int) (ret map[string]interface{}, code int, err error) {

	code, err = UploadEx2(&ret, upHost, upToken, localFile, entryURI, mimeType, customMeta, callbackParam, crc, rotate)
	return
}

func UploadEx2(ret interface{}, upHost, upToken, localFile, entryURI, mimeType, customMeta, callbackParam string,
	crc int64, rotate int) (code int, err error) {

	action := RS_PUT + rpc.EncodeURI(entryURI)
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	action += "/mimeType/" + rpc.EncodeURI(mimeType)
	if customMeta != "" {
		action += "/meta/" + rpc.EncodeURI(customMeta)
	}
	if crc >= 0 {
		action += "/crc32/" + strconv.FormatInt(crc, 10)
	}
	if rotate >= 0 {
		action += "/rotate/" + strconv.FormatInt(int64(rotate), 10)
	}
	log.Debug("action:", action)

	url := upHost + "/upload"

	multiParams := map[string][]string{
		"action": {action},
		"file":   {"@" + localFile},
		"auth":   {upToken},
	}
	if callbackParam != "" {
		multiParams["params"] = []string{callbackParam}
	}

	resp, err := rpc.DefaultClient.PostMultipart(url, multiParams)
	if err != nil {
		return api.NetworkError, err
	}
	defer resp.Body.Close()

	code = resp.StatusCode
	if code/100 == 2 {
		if ret != nil && resp.ContentLength != 0 {
			err = json.NewDecoder(resp.Body).Decode(ret)
			return
		}
	} else {
		b, _ := ioutil.ReadAll(resp.Body)
		msg := "unknown error"
		if len(b) != 0 {
			msg = string(b)
		}
		msg = "upload error: " + msg
		err = errors.New(msg)
	}
	return
}

func RsUpload(ret interface{}, upHost, upToken, entryURI, mimeType, localFile string, crc32 int64, vars map[string]string) (err error) {

	f, err := os.Open(localFile)
	if err != nil {
		return
	}
	defer f.Close()

	r, w := io.Pipe()
	defer r.Close()
	writer := multipart.NewWriter(w)

	go func() {
		err := writeMultipart(writer, f, upToken, entryURI, mimeType, localFile, crc32, vars)
		writer.Close()
		w.CloseWithError(err)
	}()

	contentType := writer.FormDataContentType()
	client := qiniurpc.Client{&http.Client{Transport: nil}}
	return client.CallWith64(nil, ret, upHost, contentType, r, 0)
}

func RsUploadWithBinary(ret interface{}, upHost, upToken, entryURI, mimeType string, f io.Reader, crc32 int64, vars map[string]string) (err error) {
	r, w := io.Pipe()
	defer r.Close()
	writer := multipart.NewWriter(w)

	go func() {
		err := writeMultipart(writer, f, upToken, entryURI, mimeType, "file", crc32, vars)
		writer.Close()
		w.CloseWithError(err)
	}()
	contentType := writer.FormDataContentType()
	client := qiniurpc.Client{&http.Client{Transport: nil}}
	return client.CallWith64(nil, ret, upHost, contentType, r, 0)
}

func writeMultipart(writer *multipart.Writer, data io.Reader,
	upToken, entryURI, mimeType, localFile string, crc32 int64, vars map[string]string) (err error) {

	// user customize variables
	if vars != nil {
		for k, v := range vars {
			err = writer.WriteField(k, v)
			if err != nil {
				return
			}
		}
	}

	// token
	err = writer.WriteField("token", upToken)
	if err != nil {
		return
	}

	// key
	if pos := strings.Index(entryURI, ":"); pos >= 0 {
		err = writer.WriteField("key", entryURI[pos+1:])
		if err != nil {
			return
		}
	}

	// crc32
	if crc32 >= 0 {
		err = writer.WriteField("crc32", strconv.FormatInt(crc32, 10))
		if err != nil {
			return
		}
	}

	// file
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition",
		fmt.Sprintf(`form-data; name="%s"; filename="%s"`,
			escapeQuotes("file"), escapeQuotes(localFile)))
	if mimeType != "" {
		h.Set("Content-Type", mimeType)
	}

	writerBuf, err := writer.CreatePart(h)
	if err != nil {
		return
	}
	_, err = io.Copy(writerBuf, data)
	return
}

var quoteEscaper = strings.NewReplacer("\\", "\\\\", `"`, "\\\"")

func escapeQuotes(s string) string {
	return quoteEscaper.Replace(s)
}