package rs

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"

	qrpc "github.com/qiniu/rpc.v1"
	"qbox.us/net/httputil"
	"qbox.us/rpc"
)

// ----------------------------------------------------------

type Service struct {
	client  httputil.Client
	rsHost  string
	pubHost string
}

func NewService(t http.RoundTripper, rsHost, pubHost string) *Service {
	client := &http.Client{Transport: t}
	return &Service{httputil.Client{client}, rsHost, pubHost}
}

// ----------------------------------------------------------

func (rs *Service) Publish(uid uint32, domain, bucketName string) (code int, err error) {
	return rs.client.Call(nil, rs.rsHost+"/admin/publish/"+rpc.EncodeURI(domain)+"/from/"+bucketName+"/uid/"+strconv.FormatUint(uint64(uid), 10))
}

func (rs *Service) Unpublish(uid uint32, domain string) (code int, err error) {
	return rs.client.Call(nil, rs.rsHost+"/admin/unpublish/"+rpc.EncodeURI(domain)+"/uid/"+strconv.FormatUint(uint64(uid), 10))
}

// 获取每个数据库的Entry信息
func (rs *Service) Inspect(uid uint32, entryURI string) (infos []string, code int, err error) {
	code, err = rs.client.Call(&infos, rs.rsHost+"/admin/inspect/"+rpc.EncodeURI(entryURI)+"/user/"+strconv.FormatUint(uint64(uid), 10))
	return
}

func (rs *Service) QueryDomain(domain []string) (info map[string]uint32, code int, err error) {
	code, err = rs.client.CallWithForm(&info, rs.pubHost+"/admin/query", map[string][]string{
		"domain": domain,
	})
	return
}

func (rs *Service) Buckets(uid uint32) (buckets []string, code int, err error) {
	code, err = rs.client.Call(&buckets, rs.rsHost+"/admin/buckets/uid/"+strconv.FormatUint(uint64(uid), 10))
	return
}

func (rs *Service) Transfer(bucketName string, uidDest, utypeDest, uidSrc uint32) (code int, err error) {
	callUrl := fmt.Sprintf("%s/transfer/%s/to/%d/utypeto/%d/uid/%d", rs.rsHost, bucketName, uidDest, utypeDest, uidSrc)
	code, err = rs.client.Call(nil, callUrl)
	return
}

type GetByDomainRet struct {
	PhyTbl string `json:"phy"`
	Tbl    string `json:"tbl"`
	Uid    uint32 `json:"uid"`
	Itbl   uint32 `json:"itbl"`
}

// ----------------------------------------------------------

type EntryInfoRet struct {
	EncodedFh string `json:"fh"`
	Hash      string `json:"hash"`
	MimeType  string `json:"mimeType"`
	EndUser   string `json:"endUser"`
	Fsize     int64  `json:"fsize"`
	PutTime   int64  `json:"putTime"`
	Idc       uint16 `json:"idc"`
}

func (rs *Service) EntryInfo(uid uint32, bucket, key string) (info EntryInfoRet, code int, err error) {
	params := map[string][]string{
		"uid":    {strconv.FormatUint(uint64(uid), 10)},
		"bucket": {bucket},
		"key":    {key},
	}
	code, err = rs.client.CallWithForm(&info, rs.rsHost+"/entryinfo", params)
	return
}

// ----------------------------------------------------------

type Client struct {
	qrpc.Client
}

func (p *Client) EntryInfo(l qrpc.Logger, host string, uid uint32, tbl, key string) (info EntryInfoRet, err error) {

	params := map[string][]string{
		"uid":    {strconv.FormatUint(uint64(uid), 10)},
		"bucket": {tbl},
		"key":    {key},
	}
	err = p.CallWithForm(l, &info, host+"/entryinfo", params)
	return
}

func (p *Client) UpdateFh(l qrpc.Logger, host string, uid uint32, tbl, key string, oldfh, newfh []byte) error {
	params := map[string][]string{
		"uid":    {strconv.FormatUint(uint64(uid), 10)},
		"tbl":    {tbl},
		"key":    {key},
		"oldefh": {base64.URLEncoding.EncodeToString(oldfh)},
		"newefh": {base64.URLEncoding.EncodeToString(newfh)},
	}
	return p.CallWithForm(l, nil, host+"/updatefh", params)
}
