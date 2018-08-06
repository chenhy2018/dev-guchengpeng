package tblmgr

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"github.com/qiniu/http/httputil.v1"

	"github.com/qiniu/rpc.v1"
)

var ErrInvalidBucket = httputil.NewError(400, "invalid bucket name")

type BucketEntry struct {
	Tbl               string      `bson:"tbl" json:"tbl"`
	Itbl              uint32      `bson:"itbl" json:"itbl"`
	PhyTbl            string      `bson:"phy" json:"phy"`
	Uid               uint32      `bson:"uid" json:"uid"`
	Zone              string      `bson:"zone" json:"zone"`
	Region            string      `bson:"region" json:"region"`
	Global            bool        `bson:"global" json:"global"`
	Line              bool        `bson:"line" json:"line"`
	Ctime             int64       `bson:"ctime" json:"ctime"`
	Oitbl             uint32      `bson:"oitbl" json:"oitbl"`
	Ouid              uint32      `bson:"ouid" json:"ouid"`
	Otbl              string      `bson:"otbl" json:"otbl"`
	Perm              uint32      `bson:"perm" json:"perm"`
	ShareUsers        []ShareUser `bson:"share_users" json:"share_users"`
	VersioningEnabled bool        `bson:"versioning_enabled" json:"versioning_enabled"`
}

type ShareUser struct {
	Uid  uint32 `json:"uid"`
	Tbl  string `json:"tbl"`
	Perm uint32 `json:"perm"`
}

type BucketQuota struct {
	Size  int64 `json:"size"`
	Count int64 `json:"count"`
}

const (
	RD = 1
	RW = 2
)

// Must with authorization
type Client struct {
	rpc.Client
}

func New(c rpc.Client) Client {

	return Client{Client: c}
}

func (c Client) Bucket(l rpc.Logger, host, tbl string) (entry BucketEntry, err error) {

	if strings.HasPrefix(tbl, "/") {
		// bucket其他字符按说也不应该出现/，已经有用户在创建bucket的时候包含/，为了兼容，这里只检查前缀
		err = ErrInvalidBucket
		return
	}
	err = c.Call(l, &entry, host+"/bucket/"+tbl)
	return
}

func (c Client) Buckets(l rpc.Logger, host, region string) (entrys []BucketEntry, err error) {

	url := host + "/v2/buckets"
	if region == "" {
		err = c.Call(l, &entrys, url)
	} else {
		params := map[string][]string{"region": {region}}
		err = c.CallWithForm(l, &entrys, url, params)
	}
	return
}

func (c Client) GlbBuckets(l rpc.Logger, host, region, global string) (entrys []BucketEntry, err error) {

	url := host + "/v2/buckets"
	params := make(map[string][]string)
	if region != "" {
		params["region"] = []string{region}
	}
	if global != "" {
		params["global"] = []string{global}
	}

	if len(params) == 0 {
		err = c.Call(l, &entrys, url)
	} else {
		err = c.CallWithForm(l, &entrys, url, params)
	}
	return
}

func (c Client) LineBuckets(l rpc.Logger, host, region, line string) (entrys []BucketEntry, err error) {

	url := host + "/v2/buckets"
	params := make(map[string][]string)
	if region != "" {
		params["region"] = []string{region}
	}
	if line != "" {
		params["line"] = []string{line}
	}

	if len(params) == 0 {
		err = c.Call(l, &entrys, url)
	} else {
		err = c.CallWithForm(l, &entrys, url, params)
	}
	return
}

func (c Client) Mkbucket(l rpc.Logger, host, tbl, region string) error {

	if strings.HasPrefix(tbl, "/") {
		// bucket以/开头会导致301，这里直接直接报错返回400
		return ErrInvalidBucket
	}
	url := host + "/mkbucket/" + tbl
	if region != "" {
		url += "/region/" + region
	}
	return c.Call(l, nil, url)
}

func (c Client) MkVersioningBucket(l rpc.Logger, host, tbl, region string) error {

	if strings.HasPrefix(tbl, "/") {
		// bucket以/开头会导致301，这里直接直接报错返回400
		return ErrInvalidBucket
	}
	url := host + "/mkbucket/" + tbl
	if region != "" {
		url += "/region/" + region
	}
	url += "/versioning/true"
	return c.Call(l, nil, url)
}

func (c Client) MkbucketV2(l rpc.Logger, host, tbl, region string) error {

	url := host + "/mkbucketv2/" + base64.URLEncoding.EncodeToString([]byte(tbl))
	if region != "" {
		url += "/region/" + region
	}
	return c.Call(l, nil, url)
}

// 兼容保留
func (c Client) GlbMkbucket(l rpc.Logger, host, tbl, region string, global bool) error {
	return c.GlbMkbucketV2(l, host, tbl, region, global)
}

func (c Client) GlbMkbucketV2(l rpc.Logger, host, tbl, region string, global bool) error {

	url := host + "/mkbucketv2/" + base64.URLEncoding.EncodeToString([]byte(tbl))
	if global {
		url += "/global/true"
	}
	if region != "" {
		url += "/region/" + region
	}
	return c.Call(l, nil, url)
}

func (c Client) LineMkbucketV2(l rpc.Logger, host, tbl, region string, line bool) error {
	url := host + "/mkbucketv2/" + base64.URLEncoding.EncodeToString([]byte(tbl))
	if line {
		url += "/line/true"
	}
	if region != "" {
		url += "/region/" + region
	}
	return c.Call(l, nil, url)
}

func (c Client) MkVersioningBucketV2(l rpc.Logger, host, tbl, region string) error {
	url := host + "/mkbucketv2/" + base64.URLEncoding.EncodeToString([]byte(tbl))
	if region != "" {
		url += "/region/" + region
	}
	url += "/versioning/true"
	return c.Call(l, nil, url)
}

func (c Client) SetGlobal(l rpc.Logger, host string, uid uint32, tbl string) error {
	url := host + "/admin/setglobal"
	params := map[string][]string{"uid": {strconv.FormatUint(uint64(uid), 10)}, "tbl": {tbl}}
	return c.CallWithForm(l, nil, url, params)
}

func (c Client) EnableVersioning(l rpc.Logger, host string, uid uint32, tbl string) error {
	url := host + "/admin/enableversioning"
	params := map[string][]string{"uid": {strconv.FormatUint(uint64(uid), 10)}, "tbl": {tbl}}
	return c.CallWithForm(l, nil, url, params)
}

// Shall without authorization
type ClientNullAuth struct {
	rpc.Client
}

func NewNullAuth(c rpc.Client) ClientNullAuth {

	return ClientNullAuth{Client: c}
}

func (c ClientNullAuth) GetByItbl(l rpc.Logger, host string, itbl uint32) (entry BucketEntry, err error) {
	err = c.Call(l, &entry, fmt.Sprintf("%s/itblbucket/%d", host, itbl))
	return
}
