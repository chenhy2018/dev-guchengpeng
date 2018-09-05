package tblg

import (
	"strconv"
	"strings"

	"github.com/qiniu/rpc.v1"
	"qbox.us/api/tblmgr"
	qconf "qbox.us/qconf/qconfapi"
)

const groupBucketQuotaPrefix = "bucketquota:"
const groupBucketQuotaPrefixLen = len(groupBucketQuotaPrefix)

func (r Client) GetBucketQuota(l rpc.Logger, uid uint32, tbl string) (bucketQuota tblmgr.BucketQuota, err error) {
	e, err := r.GetBucketEntry(l, uid, tbl)
	if err != nil {
		return
	}
	if e.Ouid != 0 {
		uid = e.Ouid
		if e.Otbl != "" {
			tbl = e.Otbl
		}
	}
	err = getPhybuckStub(r.Conn, l, &bucketQuota, MakeBucketQuotaId(uid, tbl), qconf.Cache_NoSuchEntry)
	if err != nil {
		// 当根据 uid tbl 获取 itbl 的情况下返回 612，业务层将错误重置为 631 (no such bucket)。
		if e, ok := err.(*rpc.ErrorInfo); ok && e.Code == 612 {
			err = ErrNoSuchBucket
		}
	}
	return
}

func MakeBucketQuotaId(uid uint32, tbl string) string {
	key := strconv.FormatInt(int64(uid), 36) + ":" + tbl
	return groupBucketQuotaPrefix + key
}

func ParseBucketQuotaId(id string) (uid uint32, tbl string, err error) {
	if !strings.HasPrefix(id, groupBucketQuotaPrefix) {
		return 0, "", ErrInvalidGroup
	}
	key := id[groupBucketQuotaPrefixLen:]
	pos := strings.Index(key, ":")
	if pos < 0 {
		return 0, "", ErrInvalidGroup
	}
	v, err := strconv.ParseUint(key[:pos], 36, 32)
	if err != nil {
		return 0, "", ErrInvalidGroup
	}
	return uint32(v), key[pos+1:], nil
}

const groupBucketStatPrefix = "bucketstat:"
const groupBucketStatPrefixLen = len(groupBucketStatPrefix)

func (r Client) GetBucketStat(l rpc.Logger, uid uint32, tbl string) (bucketStat tblmgr.BucketQuota, err error) {
	e, err := r.GetBucketEntry(l, uid, tbl)
	if err != nil {
		return
	}
	if e.Ouid != 0 {
		uid = e.Ouid
		if e.Otbl != "" {
			tbl = e.Otbl
		}
	}
	err = getPhybuckStub(r.Conn, l, &bucketStat, MakeBucketStatId(uid, tbl), qconf.Cache_NoSuchEntry)
	if err != nil {
		return
	}
	return
}

func MakeBucketStatId(uid uint32, tbl string) string {
	key := strconv.FormatInt(int64(uid), 36) + ":" + tbl
	return groupBucketStatPrefix + key
}

func ParseBucketStatId(id string) (uid uint32, tbl string, err error) {
	if !strings.HasPrefix(id, groupBucketStatPrefix) {
		return 0, "", ErrInvalidGroup
	}
	key := id[groupBucketStatPrefixLen:]
	pos := strings.Index(key, ":")
	if pos < 0 {
		return 0, "", ErrInvalidGroup
	}
	v, err := strconv.ParseUint(key[:pos], 36, 32)
	if err != nil {
		return 0, "", ErrInvalidGroup
	}
	return uint32(v), key[pos+1:], nil
}
