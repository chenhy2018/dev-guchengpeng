// 和 rs 包一样，但是有 logger
package rs

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	. "github.com/qiniu/api/conf"
	"github.com/qiniu/rpc.v1"
	"qbox.us/net/uri"
)

// ----------------------------------------------------------

type Service struct {
	Conn rpc.Client
}

func New(t http.RoundTripper) Service {
	client := &http.Client{Transport: t}
	return Service{rpc.Client{client}}
}

// ----------------------------------------------------------

func (rs Service) Publish(l rpc.Logger, uid uint32, domain, bucketName string) error {
	return rs.Conn.Call(l, nil, RS_HOST+"/admin/publish/"+uri.Encode(domain)+"/from/"+bucketName+"/uid/"+strconv.FormatUint(uint64(uid), 10))
}

func (rs Service) Unpublish(l rpc.Logger, uid uint32, domain string) error {
	return rs.Conn.Call(l, nil, RS_HOST+"/admin/unpublish/"+uri.Encode(domain)+"/uid/"+strconv.FormatUint(uint64(uid), 10))
}

func (rs *Service) QueryDomain(l rpc.Logger, domain []string) (info map[string]uint32, err error) {
	err = rs.Conn.CallWithForm(l, &info, PUB_HOST+"/admin/query", map[string][]string{
		"domain": domain,
	})
	return
}

func (rs *Service) Buckets(l rpc.Logger, uid uint32) (buckets []string, err error) {
	err = rs.Conn.Call(l, &buckets, RS_HOST+"/admin/buckets/uid/"+strconv.FormatUint(uint64(uid), 10))
	return
}

func (rs *Service) Transfer(l rpc.Logger, bucketName string, uidDest, utypeDest, uidSrc uint32) (err error) {
	callUrl := fmt.Sprintf("%s/transfer/%s/to/%d/utypeto/%d/uid/%d", RS_HOST, bucketName, uidDest, utypeDest, uidSrc)
	err = rs.Conn.Call(l, nil, callUrl)
	return
}

func (rs *Service) AdminDel(l rpc.Logger, id string, putTime int64) (err error) {
	return rs.Conn.Call(l, nil, RS_HOST+"/admindel/"+uri.Encode(id)+"/putTime/"+fmt.Sprint(putTime))
}

func (rs *Service) AdminDel2(l rpc.Logger, itbl uint32, key string) (err error) {
	itblStr := strconv.FormatUint(uint64(itbl), 36)
	return rs.Conn.Call(l, nil, RS_HOST+"/admindel/"+uri.Encode(itblStr+":"+key))
}

func (rs *Service) AdminDeleteAfterDays(l rpc.Logger, uid uint32, bucket, key string, deleteAferDays int) (err error) {
	return rs.Conn.Call(l, nil, RS_HOST+
		"/adminDeleteAfterDays/"+uri.Encode(bucket+":"+key)+
		"/"+fmt.Sprint(deleteAferDays)+
		"/uid/"+fmt.Sprint(uid))
}

type UpdateFhArgs struct {
	Uid    uint32 `json:"uid"`
	Bucket string `json:"bucket"`
	Key    string `json:"key"`
	OldFh  []byte `json:"oldFh"`
	NewFh  []byte `json:"newFh"`
}

func (rs *Service) AdminUpdateFh(l rpc.Logger, uid uint32, bucket, key string, oldfh, newfh []byte) (err error) {
	args := UpdateFhArgs{
		Uid:    uid,
		Bucket: bucket,
		Key:    key,
		OldFh:  oldfh,
		NewFh:  newfh,
	}
	return rs.Conn.CallWithJson(l, nil, RS_HOST+"/admin/updatefh/", args)
}

type SetMd5Args struct {
	Uid    uint32 `json:"uid"`
	Bucket string `json:"bucket"`
	Key    string `json:"key"`
	Fh     []byte `json:"fh"`
	Md5    []byte `json:"md5"`
}

func (rs *Service) AdminSetMd5(l rpc.Logger, uid uint32, bucket, key string, fh, md5 []byte) (err error) {
	args := SetMd5Args{
		Uid:    uid,
		Bucket: bucket,
		Key:    key,
		Fh:     fh,
		Md5:    md5,
	}
	return rs.Conn.CallWithJson(l, nil, RS_HOST+"/admin/setmd5", args)
}

type GetByDomainRet struct {
	PhyTbl string `json:"phy"`
	Tbl    string `json:"tbl"`
	Uid    uint32 `json:"uid"`
	Itbl   uint32 `json:"itbl"`
}

// deprecated, use "qbox.us/api/one/domain".GetByDomain
func (rs *Service) GetByDomain(l rpc.Logger, domain string) (info GetByDomainRet, err error) {
	v := url.Values{}
	v.Set("domain", domain)
	u := PUB_HOST + "/getbydomain?" + v.Encode()
	err = rs.Conn.Call(l, &info, u)
	return
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

func (rs Service) EntryInfo(l rpc.Logger, uid uint32, bucket, key string) (info EntryInfoRet, err error) {
	params := map[string][]string{
		"uid":    {strconv.FormatUint(uint64(uid), 10)},
		"bucket": {bucket},
		"key":    {key},
	}
	err = rs.Conn.CallWithForm(l, &info, RS_HOST+"/entryinfo", params)
	return
}

// ----------------------------------------------------------

type Client struct {
	rpc.Client
}

func (p Client) EntryInfo(l rpc.Logger, host string, uid uint32, tbl, key string) (info EntryInfoRet, err error) {

	params := map[string][]string{
		"uid":    {strconv.FormatUint(uint64(uid), 10)},
		"bucket": {tbl},
		"key":    {key},
	}
	err = p.CallWithForm(l, &info, host+"/entryinfo", params)
	return
}

func (p Client) UpdateFh(l rpc.Logger, host string, uid uint32, tbl, key string, oldfh, newfh []byte) error {
	params := map[string][]string{
		"uid":    {strconv.FormatUint(uint64(uid), 10)},
		"tbl":    {tbl},
		"key":    {key},
		"oldefh": {base64.URLEncoding.EncodeToString(oldfh)},
		"newefh": {base64.URLEncoding.EncodeToString(newfh)},
	}
	return p.CallWithForm(l, nil, host+"/updatefh", params)
}
