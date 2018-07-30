// 相比v1，bucketInfo不会返回BindDomains
package uc

import (
	"net/http"
	"strconv"

	"github.com/qiniu/rpc.v1"
	"qbox.us/admin_api/uc"
)

type Service struct {
	uc.Service
	Conn rpc.Client
}

func New(host string, t http.RoundTripper) *Service {
	s := *uc.New(host, t)
	return &Service{Service: s, Conn: rpc.Client{s.Conn.Client}}
}

// POST /copyBucket
// Content-Type: application/x-www-form-urlencoded
//
// srcOwner=<srcOwner>&srcTbl=<srcTbl>&dstOwner=<dstOwner>&dstTbl=<dstTbl>
func (r *Service) CopyBucketinfo(l rpc.Logger, srcOwner uint32, srcTbl string, dstOwner uint32, dstTbl string) (err error) {

	err = r.Conn.CallWithForm(l, nil, r.Host+"/copyBucket", map[string][]string{
		"srcOwner": {strconv.FormatUint(uint64(srcOwner), 10)},
		"srcTbl":   {srcTbl},
		"dstOwner": {strconv.FormatUint(uint64(dstOwner), 10)},
		"dstTbl":   {dstTbl},
	})
	return
}

/*
	POST /v2/get?name=<GroupName>&name=<keyName>
*/
func (r *Service) Get(l rpc.Logger, grp, name string) (val string, err error) {

	var ret struct {
		Val string `json:"val"`
	}
	err = r.Conn.CallWithForm(l, &ret, r.Host+"/v2/get", map[string][]string{
		"group": {grp},
		"name":  {name},
	})
	val = ret.Val
	return
}

/*
	POST /delete?group=<GroupName>&name=<KeyName>
*/
func (r *Service) Delete(l rpc.Logger, grp, name string) (err error) {

	err = r.Conn.CallWithForm(l, nil, r.Host+"/delete", map[string][]string{
		"group": {grp},
		"name":  {name},
	})
	return
}
