package tblg

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/qiniu/http/httputil.v1"
	"github.com/qiniu/rpc.v1"
	"qbox.us/api/tblmgr"
	"qbox.us/api/uc"
	qconf "qbox.us/qconf/qconfapi"
)

// ------------------------------------------------------------------------

var ErrInvalidBucket = httputil.NewError(400, "invalid bucket name")

func (r Client) GetBucketInfo(l rpc.Logger, owner uint32, bucket string) (info uc.BucketInfo, err error) {

	if strings.HasPrefix(bucket, "/") {
		// bucket其他字符按说也不应该出现/，已经有用户在创建bucket的时候包含/，为了兼容，这里只检查前缀
		err = ErrInvalidBucket
		return
	}
	var ret struct {
		Val string `bson:"v"`
	}
	e, err := r.GetBucketEntry(l, owner, bucket)
	if err != nil {
		return
	}
	if e.Ouid != 0 {
		owner = e.Ouid
		if e.Otbl != "" {
			bucket = e.Otbl
		}
	}
	id := MakeBucketInfoId(owner, bucket)
	err = r.Conn.Get(l, &ret, id, qconf.Cache_NoSuchEntry)
	if err != nil {
		return
	}

	err = json.NewDecoder(strings.NewReader(ret.Val)).Decode(&info)
	updateBucketInfo(&e, &info)
	return
}

func MakeBucketInfoId(uid uint32, tbl string) string {
	key := strconv.FormatInt(int64(uid), 36) + ":" + tbl
	return "uc:pubinfo:" + key
}

func updateBucketInfo(e *tblmgr.BucketEntry, bi *uc.BucketInfo) {
	if e == nil || bi == nil {
		return
	}
	bi.Ouid = e.Ouid
	bi.Perm = e.Perm
	bi.ShareUsers = e.ShareUsers
	bi.Otbl = e.Otbl
}

// ------------------------------------------------------------------------
