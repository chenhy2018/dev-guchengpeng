package bucketinfo

import (
	"strconv"
	"strings"

	"github.com/qiniu/errors"
	"github.com/qiniu/http/httputil.v1"
	"github.com/qiniu/rpc.v1"
	"qbox.us/api/qconf/tblg"
	"qbox.us/api/tblmgr"
	"qbox.us/api/uc.v2"
	qconf "qbox.us/qconf/qconfapi"
)

// ------------------------------------------------------------------------

var ErrInvalidBucket = httputil.NewError(400, "invalid bucket name")

type Client struct {
	Conn *qconf.Client
}

func (r Client) GetBucketInfo(l rpc.Logger, uid uint32, bucket string) (info uc.BucketInfo, err error) {

	if strings.HasPrefix(bucket, "/") {
		// bucket其他字符按说也不应该出现/，已经有用户在创建bucket的时候包含/，为了兼容，这里只检查前缀
		err = ErrInvalidBucket
		return
	}
	tblgCli := tblg.Client{r.Conn}
	e, err := tblgCli.GetBucketEntry(l, uid, bucket)
	if err != nil {
		return
	}
	if e.Ouid != 0 {
		uid = e.Ouid
		if e.Otbl != "" {
			bucket = e.Otbl
		}
	}
	id := MakeId(uid, bucket)
	err = r.Conn.Get(l, &info, id, 0)
	updateBucketInfo(&e, &info)
	return
}

// ------------------------------------------------------------------------

const groupPrefix = "bucketInfo:"
const groupPrefixLen = len(groupPrefix)

var ErrInvalidGroup = errors.New("invalid group")

func MakeId(uid uint32, bucket string) string {
	key := strconv.FormatInt(int64(uid), 36) + ":" + bucket
	return groupPrefix + key
}

func ParseId(id string) (uid uint32, bucket string, err error) {
	if !strings.HasPrefix(id, groupPrefix) {
		return 0, "", ErrInvalidGroup
	}
	s := strings.SplitN(id[groupPrefixLen:], ":", 2)
	uid64, err := strconv.ParseUint(s[0], 36, 32)
	if err != nil {
		return
	}
	return uint32(uid64), s[1], err
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
