package ios

import (
	"encoding/base64"
	"io"
	"math/rand"
	"net/http"
	"qbox.us/api"
	"qbox.us/cc/time"
	"qbox.us/objid"
	"qbox.us/rpc"
	"strconv"
)

// ----------------------------------------------------------

const (
	InvalidChan = 701 // IO: 无效Channel，可能情况：Channel已经关闭、Channel是其他用户的、伪造的Channel号
	InvalidFd   = 702 // IO: fd out of range or not open
	IoError     = 703 // IO: i/o error
	Expired     = 704 // IO: url 已经过期
	NoFile      = 705 // IO: 无当前文件 (chan.Stat)
)

var (
	EInvalidChan = api.RegisterError(InvalidChan, "invalid chan")
	EInvalidFd   = api.RegisterError(InvalidFd, "fd out of range or not open") // EBADF
	EIoError     = api.RegisterError(IoError, "i/o error")                     // EIO
	EExpired     = api.RegisterError(Expired, "expired")
	ENoFile      = api.RegisterError(NoFile, "no file")
)

// ----------------------------------------------------------

const (
	ImageForView     = 0 // Web 图片大图预览。规格为：image(w800h600q85)，格式可能是 jpeg, png, gif，视传入的原始格式而定，尽量保留原始格式。
	ImageForGridview = 1 // Web 图片小图预览。规格为：image(w128h128q80)，格式可能是 jpeg, png, gif，视传入的原始格式而定，尽量保留原始格式。
)

const (
	NormalMode  = 0 // 正常上传
	PipedMode   = 1 // 采用“边上传边下载”模式上传
	PrepareMode = 3 // 包含PipedMode，并在开始上传前调用callback(with null etag)
)

// ----------------------------------------------------------

type Service struct {
	Host       string
	Conn       rpc.Client
	WindowSize int64
}

type ChannelImpl struct {
	Io         *Service
	Id         string
	fidBase    int
	RetryTimes int
}

type Channel struct {
	*ChannelImpl
}

const DefaultWindowSize = 1024 * 1024 // UDP: max 1472

func New(host string, t http.RoundTripper) *Service {
	client := &http.Client{Transport: t}
	return &Service{host, rpc.Client{client}, DefaultWindowSize}
}

type mkchanRet struct {
	Id        string `json:"id"`
	ChunkSize int    `json:"chunkSize"`
}

func (p *Service) mkchan() (id string, code int, err error) {
	var ret mkchanRet
	code, err = p.Conn.Call(&ret, p.Host+"/mkchan")
	if err != nil {
		return
	}
	if ret.Id == "" || ret.ChunkSize <= 0 {
		code, err = api.UnexpectedResponse, api.EUnexpectedResponse
		return
	}
	id = ret.Id
	return
}

func (p *Service) Mkchan() (channel Channel, code int, err error) {
	id, code, err := p.mkchan()
	if err != nil {
		return
	}
	channel = Channel{&ChannelImpl{p, id, int(time.Seconds()), 3}}
	return
}

type PutAuthRet struct {
	URL       string `json:"url"`
	ExpiresIn int64  `json:"expiresIn"`
}

func (p *Service) PutAuth() (ret PutAuthRet, code int, err error) {
	code, err = p.Conn.Call(&ret, p.Host+"/put-auth/")
	if err == nil {
		ret.ExpiresIn += time.Seconds()
	}
	return
}

func (p *Service) PutAuthEx(expires int, callback string) (ret PutAuthRet, code int, err error) {
	url := p.Host + "/put-auth/"
	url += strconv.Itoa(expires)
	url += "/callback/"
	url += base64.URLEncoding.EncodeToString([]byte(callback))
	code, err = p.Conn.Call(&ret, url)
	if err == nil {
		ret.ExpiresIn += time.Seconds()
	}
	return
}

type CreateFileRet struct {
	Fid      string      `json:"fid"`
	Data     interface{} `json:"data"`
	URL      string      `json:"url"`
	Required int64       `json:"required"`
}

func (c Channel) Good() bool {
	return c.ChannelImpl != nil
}

func (c Channel) CreateFile(
	data interface{}, callback string, fsize int64, first io.Reader) (fid string, code int, err error) {

	ret, code, err := c.CreateFileEx(data, callback, fsize, first, 0)
	fid = ret.Fid
	return
}

func (c Channel) CreateFileEx(
	data interface{}, callback string, fsize int64, first io.Reader, mode int) (ret CreateFileRet, code int, err error) {

	ret.Data = data
	url := c.Io.Host + "/chan/" + c.Id + "/create/" + rpc.EncodeURI(callback) + "/fsize/" + strconv.FormatInt(fsize, 10)
	if mode != 0 {
		url += "/mode/" + strconv.Itoa(mode)
	}
	code, err = c.Io.Conn.CallWithBinary1(&ret, url, first)
	if err != nil {
		return
	}
	if ret.Fid == "" {
		code, err = api.UnexpectedResponse, api.EUnexpectedResponse
		return
	}
	return
}

func (c Channel) CreateFileWithKeys(
	data interface{}, callback string, fsize int64, first []byte, mode int) (ret CreateFileRet, code int, err error) {

	ret.Data = data
	url := c.Io.Host + "/chan/" + c.Id + "/createWithKeys/" + rpc.EncodeURI(callback) + "/fsize/" + strconv.FormatInt(fsize, 10)
	if mode != 0 {
		url += "/mode/" + strconv.Itoa(mode)
	}
	code, err = c.Io.Conn.CallWithParam(&ret, url, "application/octet-stream", first)
	if err != nil {
		return
	}
	if ret.Fid == "" {
		code, err = api.UnexpectedResponse, api.EUnexpectedResponse
		return
	}
	return
}

type writeAtRet struct {
	Data     interface{} `json:"data"`
	Required int64       `json:"required"`
}

func (c Channel) WriteAt(
	data interface{}, fid string, offset int64, next io.Reader) (code int, err error) {

	ret := writeAtRet{Data: data}
	url := c.Io.Host + "/chan/" + c.Id + "/put/" + fid + "/" + strconv.FormatInt(offset, 10)
	return c.Io.Conn.CallWithBinary1(&ret, url, next)
}

func (c Channel) WriteAtWithKeys(
	data interface{}, fid string, offset int64, next []byte) (required int64, code int, err error) {

	ret := writeAtRet{Data: data}
	url := c.Io.Host + "/chan/" + c.Id + "/putWithKeys/" + fid + "/" + strconv.FormatInt(offset, 10)
	code, err = c.Io.Conn.CallWithParam(&ret, url, "application/octet-stream", next)
	required = ret.Required
	return
}

type ChannelInfo struct {
	Fid      string `json:"fid"`
	Callback string `json:"callback"`
	Pos      int64  `json:"pos"`
	Fsize    int64  `json:"fsize"`
	Fails    int    `json:"fails"`
}

func (c Channel) Stat() (ci ChannelInfo, code int, err error) {
	code, err = c.Io.Conn.Call(&ci, c.Io.Host+"/chan/"+c.Id+"/stat")
	return
}

// ----------------------------------------------------------

func CreateSid() string {

	v1 := rand.Int63()
	v2 := time.Nanoseconds()
	oidLow := uint32(v2 / 1000)
	return objid.Encode(oidLow, uint64(v1^v2))
}

func Get(url string, sid string) (resp *http.Response, err error) {

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}

	req.AddCookie(&http.Cookie{Name: "sid", Value: sid})

	return http.DefaultClient.Do(req)
}

func GetRangeEx(url string, sid string, rg string) (resp *http.Response, err error) {

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}

	req.AddCookie(&http.Cookie{Name: "sid", Value: sid})
	req.Header.Set("Range", rg)

	return http.DefaultClient.Do(req)
}

func GetRange(url string, sid string, from, to int64) (resp *http.Response, err error) {

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}

	rg := strconv.FormatInt(from, 10) + "-"
	if to > 0 {
		rg += strconv.FormatInt(to-1, 10)
	}

	req.AddCookie(&http.Cookie{Name: "sid", Value: sid})
	req.Header.Set("Range", "bytes="+rg)

	return http.DefaultClient.Do(req)
}

// ----------------------------------------------------------
