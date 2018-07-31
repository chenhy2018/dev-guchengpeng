package api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/qiniu/errors"
	"github.com/qiniu/io/crc32util"
	"github.com/qiniu/rpc.v1"
	"github.com/qiniu/rpc.v1/lb.v2.1"
	"qbox.us/pfd/api/types"
)

var (
	DefaultTimeoutMs          = 1000
	DefaultProxyRespTimeoutMs = 2000
	IsPfdForceGet             = os.Getenv("PFD_FORCE_GET") != "" // 仅用于文件恢复的时候强制从 pfdstg 读取已经删除的文件，其他场景不需要设置此环境变量
)

var PostTimeoutTransport http.RoundTripper = &http.Transport{
	Proxy: http.ProxyFromEnvironment,
	Dial: (&net.Dialer{
		Timeout:   time.Duration(DefaultTimeoutMs) * time.Millisecond,
		KeepAlive: 30 * time.Second,
	}).Dial,
}

var GetTimeoutTransport http.RoundTripper = &http.Transport{
	Proxy: http.ProxyFromEnvironment,
	Dial: (&net.Dialer{
		Timeout:   time.Duration(DefaultTimeoutMs) * time.Millisecond,
		KeepAlive: 30 * time.Second,
	}).Dial,
	ResponseHeaderTimeout: time.Duration(DefaultTimeoutMs) * time.Millisecond,
}

var ProxyGetTimeoutTransport = GetTimeoutTransport
var ProxyPostTimeoutTransport = PostTimeoutTransport

type TimeoutOption struct {
	DialMs                int `json:"dial_ms"`
	GetRespMs             int `json:"get_resp_ms"`
	ProxyGetRespMs        int `json:"proxy_get_resp_ms"`
	DeleteClientTimeoutMs int `json:"delete_client_timeout_ms"`
}

func shouldReproxy(code int, err error) bool {
	if err != nil {
		return strings.Contains(err.Error(), "connecting to proxy") && strings.Contains(err.Error(), "dial tcp")
	}
	return code == http.StatusServiceUnavailable
}

func Init(option TimeoutOption, proxies []string) {
	if option.DialMs == 0 {
		option.DialMs = DefaultTimeoutMs
	}
	if option.GetRespMs == 0 {
		option.GetRespMs = DefaultTimeoutMs
	}
	if option.ProxyGetRespMs == 0 {
		option.ProxyGetRespMs = DefaultProxyRespTimeoutMs
	}
	if option.DeleteClientTimeoutMs == 0 {
		option.DeleteClientTimeoutMs = DefaultTimeoutMs
	}
	PostTimeoutTransport = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		Dial: (&net.Dialer{
			Timeout:   time.Duration(option.DialMs) * time.Millisecond,
			KeepAlive: 30 * time.Second,
		}).Dial,
	}
	GetTimeoutTransport = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		Dial: (&net.Dialer{
			Timeout:   time.Duration(option.DialMs) * time.Millisecond,
			KeepAlive: 30 * time.Second,
		}).Dial,
		ResponseHeaderTimeout: time.Duration(option.GetRespMs) * time.Millisecond,
	}
	if len(proxies) > 0 {
		ProxyGetTimeoutTransport = lb.NewTransport(&lb.TransportConfig{
			DialTimeoutMS: option.DialMs,
			RespTimeoutMS: option.ProxyGetRespMs,
			Proxys:        proxies,
			ShouldReproxy: shouldReproxy,
		})
		ProxyPostTimeoutTransport = lb.NewTransport(&lb.TransportConfig{
			DialTimeoutMS: option.DialMs,
			Proxys:        proxies,
			ShouldReproxy: shouldReproxy,
		})
	}
	if IsPfdForceGet {
		log.Println("PFD_FORCE_GET enabled")
	}
}

// -----------------------------------------------------------

const (
	StatusAllocedEntry = 613
	StatusSyncDelayed  = 620
)

type FwdReq int

const (
	FwdReqAlloc FwdReq = iota + 1
	FwdReqPutat
	FwdReqPut
)

// -----------------------------------------------------------

type Client struct {
	Host            string
	PostClient      rpc.Client
	GetClient       rpc.Client
	ProxyGetClient  rpc.Client
	ProxyPostClient rpc.Client
}

func NewClient(host string) Client {
	return Client{
		Host:       host,
		PostClient: rpc.DefaultClient,
		GetClient:  rpc.DefaultClient,
	}
}

func NewClientWithTimeout(host string) Client {
	return Client{
		Host:            host,
		PostClient:      rpc.Client{&http.Client{Transport: PostTimeoutTransport}},
		GetClient:       rpc.Client{&http.Client{Transport: GetTimeoutTransport}},
		ProxyGetClient:  rpc.Client{&http.Client{Transport: ProxyGetTimeoutTransport}},
		ProxyPostClient: rpc.Client{&http.Client{Transport: ProxyPostTimeoutTransport}},
	}
}

func NewClientWithTimeoutEx(host string, clientTimeout time.Duration) Client {
	if clientTimeout == 0 {
		clientTimeout = time.Duration(DefaultTimeoutMs) * time.Millisecond
	}
	return Client{
		Host:            host,
		PostClient:      rpc.Client{&http.Client{Transport: PostTimeoutTransport, Timeout: clientTimeout}},
		GetClient:       rpc.Client{&http.Client{Transport: GetTimeoutTransport, Timeout: clientTimeout}},
		ProxyGetClient:  rpc.Client{&http.Client{Transport: ProxyGetTimeoutTransport, Timeout: clientTimeout}},
		ProxyPostClient: rpc.Client{&http.Client{Transport: ProxyPostTimeoutTransport, Timeout: clientTimeout}},
	}
}

func (c Client) Put(l rpc.Logger, data io.Reader, fsize int64, dgid uint32) (fh []byte, md5 []byte, err error) {

	if fsize == 0 {
		// see https://github.com/golang/go/issues/20257
		data = nil
	}
	u := fmt.Sprintf("%v/put/%v", c.Host, dgid)
	var ret struct {
		Efh string `json:"fh"`
		Md5 string `json:"md5"`
	}
	err = c.PostClient.CallAfterCrcEncoded(l, &ret, u, "application/octet-stream", data, fsize)
	if err != nil {
		return
	}
	fh, err = base64.URLEncoding.DecodeString(ret.Efh)
	if err != nil {
		return
	}
	if ret.Md5 != "" {
		md5, err = base64.URLEncoding.DecodeString(ret.Md5)
	}
	return
}

func (c Client) PutWithoutDgid(l rpc.Logger, data io.Reader, fsize int64) (fh []byte, err error) {

	if fsize == 0 {
		// see https://github.com/golang/go/issues/20257
		data = nil
	}
	u := fmt.Sprintf("%v/put", c.Host)
	var ret struct {
		Efh string `json:"fh"`
	}
	err = c.PostClient.CallWith64(l, &ret, u, "application/octet-stream", data, fsize)
	if err != nil {
		return
	}
	fh, err = base64.URLEncoding.DecodeString(ret.Efh)
	return
}

func (c Client) Fwd(l rpc.Logger, data io.Reader, bodySize int64, fh []byte, req FwdReq, nbroken int) (err error) {
	if bodySize == 0 {
		// see https://github.com/golang/go/issues/20257
		data = nil
	}
	u := fmt.Sprintf("%v/fwd/%v/req/%v/nbroken/%v", c.Host, base64.URLEncoding.EncodeToString(fh), req, nbroken)
	err = c.PostClient.CallAfterCrcEncoded(l, nil, u, "application/octet-stream", data, bodySize)
	return
}

func (c Client) Alloc(l rpc.Logger, fsize int64, hash [20]byte, dgid uint32) (fh []byte, err error) {

	u := fmt.Sprintf("%v/alloc/%v/hash/%v/dgid/%v", c.Host, fsize, base64.URLEncoding.EncodeToString(hash[:]), dgid)
	var ret struct {
		Efh string `json:"fh"`
	}
	err = c.PostClient.Call(l, &ret, u)
	if err != nil {
		return
	}
	fh, err = base64.URLEncoding.DecodeString(ret.Efh)
	return
}

func (c Client) AllocWithoutDgid(l rpc.Logger, fsize int64, hash [20]byte) (fh []byte, err error) {

	u := fmt.Sprintf("%v/alloc/%v/hash/%v", c.Host, fsize, base64.URLEncoding.EncodeToString(hash[:]))
	var ret struct {
		Efh string `json:"fh"`
	}
	err = c.PostClient.Call(l, &ret, u)
	if err != nil {
		return
	}
	fh, err = base64.URLEncoding.DecodeString(ret.Efh)
	return
}

func (c Client) PutAt(l rpc.Logger, data io.Reader, fh []byte) (err error) {

	fhi, err := types.DecodeFh(fh)
	if err != nil {
		err = errors.Info(err, "DecodeFh").Detail(err)
		return
	}
	fsize := fhi.Fsize
	if fsize == 0 {
		// see https://github.com/golang/go/issues/20257
		data = nil
	}

	u := c.Host + "/putat/" + base64.URLEncoding.EncodeToString(fh)
	err = c.PostClient.CallAfterCrcEncoded(l, nil, u, "application/octec-stream", data, fsize)
	return
}

func (c Client) Hash(l rpc.Logger, fh []byte) (hash [20]byte, err error) {
	ret := struct {
		Hash string `json:"hash"`
	}{}
	u := c.Host + "/hash/" + base64.URLEncoding.EncodeToString(fh)
	err = c.GetClient.GetCall(l, &ret, u)
	if err != nil {
		return
	}
	hash0, err := base64.URLEncoding.DecodeString(ret.Hash)
	copy(hash[:], hash0)
	return
}

func (c Client) Md5(l rpc.Logger, fh []byte) (hash []byte, err error) {
	return c.md5(l, c.GetClient, fh)
}

func (c Client) ProxyMd5(l rpc.Logger, fh []byte) (hash []byte, err error) {
	return c.md5(l, c.ProxyGetClient, fh)
}

func (c Client) md5(l rpc.Logger, cli rpc.Client, fh []byte) (hash []byte, err error) {
	ret := struct {
		Md5 string `json:"md5"`
	}{}
	u := c.Host + "/md5/" + base64.URLEncoding.EncodeToString(fh)
	err = cli.GetCall(l, &ret, u)
	if err != nil {
		return
	}
	hash, err = base64.URLEncoding.DecodeString(ret.Md5)
	return
}

type nullReadCloser struct{}

func (self nullReadCloser) Read(p []byte) (int, error) {
	return 0, io.EOF
}

func (self nullReadCloser) Close() error {
	return nil
}

func (c Client) ProxyGet(l rpc.Logger, fh []byte, from, to int64) (rc io.ReadCloser, n int64, err error) {
	return c.get(l, c.ProxyGetClient, fh, from, to, false)
}

func (c Client) Get(l rpc.Logger, fh []byte, from, to int64) (rc io.ReadCloser, n int64, err error) {
	return c.get(l, c.GetClient, fh, from, to, false)
}

func (c Client) ForceGet(l rpc.Logger, fh []byte, from, to int64) (rc io.ReadCloser, n int64, err error) {
	return c.get(l, c.GetClient, fh, from, to, true)
}

func (c Client) get(l rpc.Logger, cli rpc.Client, fh []byte, from, to int64, force bool) (rc io.ReadCloser, n int64, err error) {
	fhi, err := types.DecodeFh(fh)
	if err != nil {
		err = errors.Info(err, "DecodeFh").Detail(err)
		return
	}
	fsize := fhi.Fsize

	if from >= to || fsize == 0 {
		return nullReadCloser{}, 0, nil
	}

	u := c.Host + "/get/" + base64.URLEncoding.EncodeToString(fh)
	if force || IsPfdForceGet {
		u += "/force/true"
	}
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return
	}
	if !(from == 0 && to >= fsize) {
		rangeStr := fmt.Sprintf("bytes=%v-%v", from, to-1)
		req.Header.Set("Range", rangeStr)
	}
	resp, err := cli.Do(l, req)
	if err != nil {
		return
	}
	if resp.StatusCode/100 != 2 {
		defer resp.Body.Close()
		err = rpc.ResponseError(resp)
		return
	}
	rc, n = resp.Body, resp.ContentLength
	return
}

type StatusRet struct {
	DiskInfo []DiskInfo `json:"disk_info"`
}

type DiskInfo struct {
	Dgid   uint32 `json:"dgid"`
	Path   string `json:"path"`
	Avail  int64  `json:"avail"`
	Total  int64  `json:"total"`
	Repair bool   `json:"repair"`

	Putting int64 `json:"putting"`
}

func (c Client) Status(l rpc.Logger) (ret StatusRet, err error) {
	u := c.Host + "/status"
	err = c.GetClient.Call(l, &ret, u)
	return
}

func (c Client) Delete(l rpc.Logger, fh []byte) (err error) {
	return c.delete(l, c.PostClient, fh)
}

func (c Client) ProxyDelete(l rpc.Logger, fh []byte) (err error) {
	return c.delete(l, c.ProxyPostClient, fh)
}

func (c Client) delete(l rpc.Logger, cli rpc.Client, fh []byte) (err error) {
	u := c.Host + "/delete/" + base64.URLEncoding.EncodeToString(fh)
	return cli.Call(l, nil, u)
}

func (c Client) Undelete(l rpc.Logger, fh []byte) (err error) {
	u := c.Host + "/undelete/" + base64.URLEncoding.EncodeToString(fh)
	return c.PostClient.Call(l, nil, u)
}

func (c Client) Rotate(l rpc.Logger) (err error) {
	u := c.Host + "/rotate"
	err = c.PostClient.Call(l, nil, u)
	return
}

func (c Client) Gfile(l rpc.Logger, egid string) (rc io.ReadCloser, n int64, name string, err error) {
	return c.gfile(l, c.GetClient, egid)
}

func (c Client) ProxyGfile(l rpc.Logger, egid string) (rc io.ReadCloser, n int64, name string, err error) {
	return c.gfile(l, c.ProxyGetClient, egid)
}

func (c Client) gfile(l rpc.Logger, cli rpc.Client, egid string) (rc io.ReadCloser, n int64, name string, err error) {
	u := c.Host + "/gfile/" + egid
	resp, err := cli.Get(l, u)
	if err != nil {
		return
	}
	if resp.StatusCode/100 != 2 {
		defer resp.Body.Close()
		err = rpc.ResponseError(resp)
		return
	}
	n = crc32util.DecodeSize(resp.ContentLength)
	rc = struct {
		io.Reader
		io.Closer
	}{
		crc32util.Decoder(resp.Body, n, nil),
		io.Closer(resp.Body),
	}
	name = resp.Header.Get("X-Gid-Filename")
	return
}

type GidInfo struct {
	Egid    string
	Dgid    uint32
	ModTime time.Time
	Size    int64
	Compact bool
}

func (c Client) Gids(l rpc.Logger) (gidInfos []GidInfo, err error) {
	return c.gids(l, c.GetClient)
}
func (c Client) ProxyGids(l rpc.Logger) (gidInfos []GidInfo, err error) {
	return c.gids(l, c.ProxyGetClient)
}

func (c Client) gids(l rpc.Logger, cli rpc.Client) (gidInfos []GidInfo, err error) {
	u := c.Host + "/gids"
	err = cli.Call(l, &gidInfos, u)
	return
}

func (c Client) Transfer(l rpc.Logger, egid string, dgid uint32) (err error) {
	return c.transfer(l, c.PostClient, egid, dgid, false)
}

func (c Client) ProxyTransfer(l rpc.Logger, egid string, dgid uint32) (err error) {
	return c.transfer(l, c.ProxyPostClient, egid, dgid, true)
}

func (c Client) Transfer2(l rpc.Logger, egid string, dgid uint32) (err error) {
	return c.transfer2(l, c.PostClient, egid, dgid)
}

func (c Client) ProxyTransfer2(l rpc.Logger, egid string, dgid uint32) (err error) {
	return c.transfer2(l, c.ProxyPostClient, egid, dgid)
}

func (c Client) SlaveTransfer(l rpc.Logger, egid string, dgid uint32) (err error) {
	return c.slaveTransfer(l, c.PostClient, egid, dgid, false)
}

func (c Client) SlaveTransferEx(l rpc.Logger, egid string, dgid uint32, egidFromEdgeNode bool) (err error) {
	return c.slaveTransfer(l, c.PostClient, egid, dgid, egidFromEdgeNode)
}

func (c Client) ProxySlaveTransfer(l rpc.Logger, egid string, dgid uint32) (err error) {
	return c.slaveTransfer(l, c.ProxyPostClient, egid, dgid, true)
}

func (c Client) SlaveTransfer2(l rpc.Logger, egid string, dgid uint32) (err error) {
	return c.slaveTransfer2(l, c.PostClient, egid, dgid)
}

func (c Client) ProxySlaveTransfer2(l rpc.Logger, egid string, dgid uint32) (err error) {
	return c.slaveTransfer2(l, c.ProxyPostClient, egid, dgid)
}

func (c Client) transfer(l rpc.Logger, cli rpc.Client, egid string, dgid uint32, egidFromEdgeNode bool) (err error) {
	u := fmt.Sprintf("%v/transfer/%v/edge/%v/to/%v/fsync/%v", c.Host, egid, egidFromEdgeNode, dgid, true)
	err = cli.Call(l, nil, u)
	return
}

func (c Client) transfer2(l rpc.Logger, cli rpc.Client, egid string, dgid uint32) (err error) {
	u := fmt.Sprintf("%v/transfer2/%v/to/%v/fsync/%v", c.Host, egid, dgid, true)
	err = cli.Call(l, nil, u)
	return
}

func (c Client) slaveTransfer(l rpc.Logger, cli rpc.Client, egid string, dgid uint32, egidFromEdgeNode bool) (err error) {

	u := fmt.Sprintf("%v/transfer/%v/edge/%v/to/%v/fsync/%v?slaveOk=1", c.Host, egid, egidFromEdgeNode, dgid, true)
	err = cli.Call(l, nil, u)
	return
}

func (c Client) slaveTransfer2(l rpc.Logger, cli rpc.Client, egid string, dgid uint32) (err error) {
	u := fmt.Sprintf("%v/transfer2/%v/to/%v/fsync/%v?slaveOk=1", c.Host, egid, dgid, true)
	err = cli.Call(l, nil, u)
	return
}

func (c Client) Remove(l rpc.Logger, egid string) (err error) {
	u := c.Host + "/remove/" + egid
	err = c.PostClient.Call(l, nil, u)
	return
}

func (c Client) RepairGid(l rpc.Logger, egid string, pos int64, step, interval, ratelimitMB int, async bool, dryrun bool) (err error) {
	u := fmt.Sprintf("%s/repairgid/%v/pos/%v/step/%v/interval/%v/ratelimit/%v/async/%v/dryrun/%v",
		c.Host, egid, pos, step, interval, ratelimitMB, async, dryrun)
	resp, err := c.PostClient.Get(l, u)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		err = rpc.ResponseError(resp)
		return
	}
	return
}

type WalkFileInfo struct {
	Fid      uint64 `json:"fid"`
	FhOffset int64  `json:"fhoffset"`
	Size     int64  `json:"size"`
	Flag     uint8  `json:"flag"`
	Hash     string `json:"hash,omitempty"`
	Error    string `json:"error,omitempty"`
	Finish   bool   `json:"finish,omitempty"`
}

func (c Client) WalkFunc(l rpc.Logger, egid string, pos int64, step, interval, ratelimitMB int, hash bool, f func(xl rpc.Logger, fi *WalkFileInfo) error) (err error) {
	u := fmt.Sprintf("%s/walk/%v/pos/%v/step/%v/interval/%v/ratelimit/%v/hash/%v",
		c.Host, egid, pos, step, interval, ratelimitMB, hash)
	resp, err := c.PostClient.Get(l, u)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		err = rpc.ResponseError(resp)
		return
	}
	decoder := json.NewDecoder(resp.Body)
	for {
		one := WalkFileInfo{}
		err = decoder.Decode(&one)
		if err != nil {
			return
		}
		if one.Finish {
			if one.Error != "" {
				err = errors.New(one.Error)
				return
			}
			break
		}
		err = f(l, &one)
		if err != nil {
			return
		}
	}
	return nil
}

type CompactStats struct {
	FilesCount     uint64 `json:"filesCount"`
	DeletedCount   uint64 `json:"deletedCount"`
	AllocatedCount uint64 `json:"allocatedCount"`
	FilesSize      int64  `json:"filesSize"`
	DeletedSize    int64  `json:"deletedSize"`
	AllocatedSize  int64  `json:"allocatedSize"`
}

func (c Client) CompactStat(l rpc.Logger, egid string) (stats *CompactStats, err error) {
	u := c.Host + "/compact/stat/gid/" + egid
	var statsVal CompactStats
	err = c.GetClient.Call(l, &statsVal, u)
	return &statsVal, err
}

func (c Client) CompactWalkStat(l rpc.Logger, egid string, step int, interval string) (stats *CompactStats, err error) {
	u := c.Host + "/compact/stat/gid/" + egid + "/step/" + strconv.Itoa(step) + "/interval/" + interval
	var statsVal CompactStats
	err = c.GetClient.Call(l, &statsVal, u)
	return &statsVal, err
}

func (c Client) CompactExec(l rpc.Logger, egid string, step int, interval string) error {
	u := c.Host + "/compact/exec/gid/" + egid + "/step/" + strconv.Itoa(step) + "/interval/" + interval
	return c.PostClient.Call(l, nil, u)
}

func (c Client) CompactExecAndCheck(l rpc.Logger, egid string, step int, interval string) error {
	u := c.Host + "/compact/exec/gid/" + egid + "/step/" + strconv.Itoa(step) + "/interval/" + interval + "?check=1"
	return c.PostClient.Call(l, nil, u)
}

func (c Client) CompactSlaveExec(l rpc.Logger, egid string, step int, interval string) error {
	u := c.Host + "/compact/exec/gid/" + egid + "/step/" + strconv.Itoa(step) + "/interval/" + interval + "?slaveOk=1"
	return c.PostClient.Call(l, nil, u)
}

func (c Client) CompactSlaveExecAndCheck(l rpc.Logger, egid string, step int, interval string) error {
	u := c.Host + "/compact/exec/gid/" + egid + "/step/" + strconv.Itoa(step) + "/interval/" + interval + "?slaveOk=1&check=1"
	return c.PostClient.Call(l, nil, u)
}

func (c Client) Compact2Exec(l rpc.Logger, egid string, step int, interval string) error {
	u := c.Host + "/compact2/exec/gid/" + egid + "/step/" + strconv.Itoa(step) + "/interval/" + interval
	return c.PostClient.Call(l, nil, u)
}

func (c Client) Compact2ExecAndCheck(l rpc.Logger, egid string, step int, interval string) error {
	u := c.Host + "/compact2/exec/gid/" + egid + "/step/" + strconv.Itoa(step) + "/interval/" + interval + "?check=1"
	return c.PostClient.Call(l, nil, u)
}

func (c Client) Compact2SlaveExec(l rpc.Logger, egid string, step int, interval string) error {
	u := c.Host + "/compact2/exec/gid/" + egid + "/step/" + strconv.Itoa(step) + "/interval/" + interval + "?slaveOk=1"
	return c.PostClient.Call(l, nil, u)
}

func (c Client) Compact2SlaveExecAndCheck(l rpc.Logger, egid string, step int, interval string) error {
	u := c.Host + "/compact2/exec/gid/" + egid + "/step/" + strconv.Itoa(step) + "/interval/" + interval + "?slaveOk=1&check=1"
	return c.PostClient.Call(l, nil, u)
}

func (c Client) CompactCommit(l rpc.Logger, egid string) error {
	u := c.Host + "/compact/commit/gid/" + egid
	return c.PostClient.Call(l, nil, u)
}

type RepairRet struct {
	Progress int    `json:"progress"`
	Msg      string `json:"msg"`
}

func (c Client) Repair(l rpc.Logger, dgid uint32) (ret RepairRet, err error) {
	u := c.Host + "/repair/" + strconv.Itoa(int(dgid))
	err = c.PostClient.Call(l, &ret, u)
	return
}

type ProgressRet struct {
	TotalFsize    int64 `json:"total_fsize"`
	RepairedFsize int64 `json:"repaired_fsize"`
	TotalGids     int   `json:"total_gids"`
	RepairedGids  int   `json:"repair_gids"`
}

func (c Client) RepairProgress(l rpc.Logger, dgid uint32) (ret ProgressRet, err error) {
	u := c.Host + "/repair/progress/dgid/" + strconv.Itoa(int(dgid))
	err = c.PostClient.Call(l, &ret, u)
	return
}

func (c Client) ReportRestart(l rpc.Logger) (err error) {
	u := c.Host + "/report/restart"
	err = c.PostClient.Call(l, nil, u)
	return
}

func (c Client) RepairShakehands(l rpc.Logger, dgid uint32) (err error) {
	u := c.Host + "/repair/shakehands/dgid/" + strconv.Itoa(int(dgid))
	err = c.PostClient.Call(l, nil, u)
	return
}

func (c Client) Debug(l rpc.Logger, dgid uint32, code int, msg string) (err error) {
	u := c.Host + "/debug/" + strconv.Itoa(int(dgid)) + "/code/" + strconv.Itoa(code) + "/msg/" + base64.URLEncoding.EncodeToString([]byte(msg))
	err = c.PostClient.Call(l, nil, u)
	return
}

func (c Client) SetBroken(l rpc.Logger, dgid uint32) (err error) {
	u := c.Host + "/setbroken/" + strconv.Itoa(int(dgid))
	err = c.PostClient.Call(l, nil, u)
	return
}
