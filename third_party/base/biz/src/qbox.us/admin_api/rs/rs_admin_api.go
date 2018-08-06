package rs

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	rs "qbox.us/api/v2/rs"
	"qbox.us/net/httputil"
	"qbox.us/rpc"

	. "github.com/qiniu/api/conf"
	qrpc "github.com/qiniu/rpc.v1"
)

// ----------------------------------------------------------

type Service struct {
	Conn httputil.Client
}

func New(t http.RoundTripper) Service {
	client := &http.Client{Transport: t}
	return Service{httputil.Client{client}}
}

// ----------------------------------------------------------

func (rs Service) Publish(uid uint32, domain, bucketName string) (code int, err error) {
	return rs.Conn.Call(nil, RS_HOST+"/admin/publish/"+rpc.EncodeURI(domain)+"/from/"+bucketName+"/uid/"+strconv.FormatUint(uint64(uid), 10))
}

func (rs Service) Unpublish(uid uint32, domain string) (code int, err error) {
	return rs.Conn.Call(nil, RS_HOST+"/admin/unpublish/"+rpc.EncodeURI(domain)+"/uid/"+strconv.FormatUint(uint64(uid), 10))
}

func (rs *Service) QueryDomain(domain []string) (info map[string]uint32, code int, err error) {
	code, err = rs.Conn.CallWithForm(&info, PUB_HOST+"/admin/query", map[string][]string{
		"domain": domain,
	})
	return
}

func (rs *Service) Buckets(uid uint32) (buckets []string, code int, err error) {
	code, err = rs.Conn.Call(&buckets, RS_HOST+"/admin/buckets/uid/"+strconv.FormatUint(uint64(uid), 10))
	return
}

func (rs *Service) Transfer(bucketName string, uidDest, utypeDest, uidSrc uint32) (code int, err error) {
	callUrl := fmt.Sprintf("%s/transfer/%s/to/%d/utypeto/%d/uid/%d", RS_HOST, bucketName, uidDest, utypeDest, uidSrc)
	code, err = rs.Conn.Call(nil, callUrl)
	return
}

func (rs *Service) AdminDel(id string, putTime int64) (code int, err error) {
	return rs.Conn.Call(nil, RS_HOST+"/admindel/"+rpc.EncodeURI(id)+"/putTime/"+fmt.Sprint(putTime))
}

func (rs *Service) AdminDelVersion(id, version string, putTime int64) (code int, err error) {
	return rs.Conn.Call(nil, RS_HOST+"/admindel/"+rpc.EncodeURI(id)+"/putTime/"+fmt.Sprint(putTime)+"/version/"+version)
}

func (rs *Service) AdminDelAllVersions(id string) (code int, err error) {
	return rs.Conn.Call(nil, RS_HOST+"/admindel/"+rpc.EncodeURI(id)+"/allVersions/1")
}

func (rs *Service) AdminDelAllVersionsWithPutTime(id string, putTime int64) (code int, err error) {
	return rs.Conn.Call(nil, RS_HOST+"/admindel/"+rpc.EncodeURI(id)+"/putTime/"+fmt.Sprint(putTime)+"/allVersions/1")
}

func (rs *Service) AdminChType(id string, putTime int64, Type rs.FileType) (code int, err error) {
	return rs.Conn.Call(nil, RS_HOST+
		"/adminchtype/"+rpc.EncodeURI(id)+
		"/putTime/"+fmt.Sprint(putTime)+
		"/type/"+fmt.Sprint(Type))
}

func (rs *Service) AdminChVersionType(id, version string, putTime int64, Type rs.FileType) (code int, err error) {
	return rs.Conn.Call(nil, RS_HOST+
		"/adminchtype/"+rpc.EncodeURI(id)+
		"/putTime/"+fmt.Sprint(putTime)+
		"/type/"+fmt.Sprint(Type)+
		"/version/"+version)
}

func (rs *Service) AdminDeleteAfterDays(uid uint32, bucket, key string, deleteAferDays int) (code int, err error) {
	return rs.Conn.Call(nil, RS_HOST+
		"/adminDeleteAfterDays/"+rpc.EncodeURI(bucket+":"+key)+
		"/"+fmt.Sprint(deleteAferDays)+
		"/uid/"+fmt.Sprint(uid))
}

func (rs *Service) AdminDeleteVersionAfterDays(uid uint32, bucket, key, version string, deleteAferDays int) (code int, err error) {
	return rs.Conn.Call(nil, RS_HOST+
		"/adminDeleteAfterDays/"+rpc.EncodeURI(bucket+":"+key)+
		"/"+fmt.Sprint(deleteAferDays)+
		"/uid/"+fmt.Sprint(uid)+
		"/version/"+version)
}

type UpdateFhArgs struct {
	Uid    uint32 `json:"uid"`
	Bucket string `json:"bucket"`
	Key    string `json:"key"`
	OldFh  []byte `json:"oldFh"`
	NewFh  []byte `json:"newFh"`
}

func (rs *Service) AdminUpdateFh(uid uint32, bucket, key string, oldfh, newfh []byte) (code int, err error) {
	args := UpdateFhArgs{
		Uid:    uid,
		Bucket: bucket,
		Key:    key,
		OldFh:  oldfh,
		NewFh:  newfh,
	}
	return rpc.Client{rs.Conn.Client}.CallWithJson(nil, RS_HOST+"/admin/updatefh/", args)
}

type SetMd5Args struct {
	Uid    uint32 `json:"uid"`
	Bucket string `json:"bucket"`
	Key    string `json:"key"`
	Fh     []byte `json:"fh"`
	Md5    []byte `json:"md5"`
}

func (rs *Service) AdminSetMd5(uid uint32, bucket, key string, fh, md5 []byte) (code int, err error) {
	args := SetMd5Args{
		Uid:    uid,
		Bucket: bucket,
		Key:    key,
		Fh:     fh,
		Md5:    md5,
	}
	return rpc.Client{rs.Conn.Client}.CallWithJson(nil, RS_HOST+"/admin/setmd5", args)
}

type GetByDomainRet struct {
	PhyTbl string `json:"phy"`
	Tbl    string `json:"tbl"`
	Uid    uint32 `json:"uid"`
	Itbl   uint32 `json:"itbl"`
}

// deprecated, use "qbox.us/api/one/domain".GetByDomain
func (rs *Service) GetByDomain(domain string) (info GetByDomainRet, err error) {
	v := url.Values{}
	v.Set("domain", domain)
	u := PUB_HOST + "/getbydomain?" + v.Encode()
	_, err = rs.Conn.Call(&info, u)
	return
}

// ----------------------------------------------------------

type EntryInfoRet struct {
	EncodedFh string            `json:"fh"`
	Hash      string            `json:"hash"`
	MimeType  string            `json:"mimeType"`
	EndUser   string            `json:"endUser"`
	Fsize     int64             `json:"fsize"`
	PutTime   int64             `json:"putTime"`
	Idc       uint16            `json:"idc"`
	IP        string            `json:"ip"`
	XMeta     map[string]string `json:"x-qn-meta"`
	Type      rs.FileType       `json:"type"`
}

func (rs Service) EntryInfo(uid uint32, bucket, key string) (info EntryInfoRet, code int, err error) {
	params := map[string][]string{
		"uid":    {strconv.FormatUint(uint64(uid), 10)},
		"bucket": {bucket},
		"key":    {key},
	}
	code, err = rs.Conn.CallWithForm(&info, RS_HOST+"/entryinfo", params)
	return
}

// ----------------------------------------------------------

type AgetRet struct {
	EncodedFh []byte            `json:"fh"`
	Hash      string            `json:"hash"`
	MimeType  string            `json:"mimeType"`
	EndUser   string            `json:"endUser"`
	Fsize     int64             `json:"fsize"`
	PutTime   int64             `json:"putTime"`
	Idc       uint16            `json:"idc"`
	IP        string            `json:"ip"`
	Type      rs.FileType       `json:"type"`
	XMeta     map[string]string `json:"x-qn-meta"`
	Version   string            `json:"version"`
	Vdel      bool              `json:"vdel"`
}

func (rs Service) Aget(uid uint32, bucket, key string) (info AgetRet, code int, err error) {
	code, err = rs.Conn.Call(&info, RS_HOST+
		"/aget/"+rpc.EncodeURI(bucket+":"+key)+
		"/user/"+fmt.Sprint(uid))
	return
}

func (rs Service) AgetVersion(uid uint32, bucket, key, version string) (info AgetRet, code int, err error) {
	code, err = rs.Conn.Call(&info, RS_HOST+
		"/aget/"+rpc.EncodeURI(bucket+":"+key)+
		"/user/"+fmt.Sprint(uid)+
		"/version/"+version)
	return
}

// ----------------------------------------------------------

type Client struct {
	qrpc.Client
}

func (p Client) EntryInfo(l qrpc.Logger, host string, uid uint32, tbl, key string) (info EntryInfoRet, err error) {

	params := map[string][]string{
		"uid":    {strconv.FormatUint(uint64(uid), 10)},
		"bucket": {tbl},
		"key":    {key},
	}
	err = p.CallWithForm(l, &info, host+"/entryinfo", params)
	return
}

func (p Client) UpdateFh(l qrpc.Logger, host string, uid uint32, tbl, key string, oldfh, newfh []byte) error {
	params := map[string][]string{
		"uid":    {strconv.FormatUint(uint64(uid), 10)},
		"tbl":    {tbl},
		"key":    {key},
		"oldefh": {base64.URLEncoding.EncodeToString(oldfh)},
		"newefh": {base64.URLEncoding.EncodeToString(newfh)},
	}
	return p.CallWithForm(l, nil, host+"/updatefh", params)
}
