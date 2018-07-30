package tblg

import (
	"errors"
	"strconv"
	"strings"

	"github.com/qiniu/rpc.v1"
	"qbox.us/api/tblmgr"
	qconf "qbox.us/qconf/qconfapi"
)

var ErrNoSuchBucket = &rpc.ErrorInfo{Code: 631, Err: "no such bucket"}

// ------------------------------------------------------------------------

type Client struct {
	Conn *qconf.Client
}

func (r Client) GetItblEntry(l rpc.Logger, itbl uint32) (entry tblmgr.BucketEntry, err error) {
	err = getPhybuckStub(r.Conn, l, &entry, MakeItblId(itbl), qconf.Cache_NoSuchEntry)
	if err != nil {
		// 当根据 itbl 获取 uid,tbl 的情况下返回 612，业务层将错误重置为 631 (no such bucket)。
		if e, ok := err.(*rpc.ErrorInfo); ok && e.Code == 612 {
			err = ErrNoSuchBucket
		}
	}
	return
}
func (r Client) GetPhybuck(l rpc.Logger, uid uint32, tbl string) (itbl uint32, phyTbl string, err error) {

	var ret tblmgr.BucketEntry
	ret, err = r.GetBucketEntry(l, uid, tbl)
	if err != nil {
		return
	}
	itbl, phyTbl = ret.Itbl, ret.PhyTbl
	return
}

func (r Client) GetBucketEntry(l rpc.Logger, uid uint32, tbl string) (entry tblmgr.BucketEntry, err error) {

	err = getPhybuckStub(r.Conn, l, &entry, MakeId(uid, tbl), qconf.Cache_NoSuchEntry)
	if err != nil {
		// 当根据 uid tbl 获取 itbl 的情况下返回 612，业务层将错误重置为 631 (no such bucket)。
		if e, ok := err.(*rpc.ErrorInfo); ok && e.Code == 612 {
			err = ErrNoSuchBucket
		}
	}
	return
}

var getPhybuckStub = func(conn *qconf.Client, l rpc.Logger, ret interface{}, id string, cacheFlags int) (err error) {
	return conn.Get(l, ret, id, cacheFlags)
}

// ------------------------------------------------------------------------

const groupPrefix = "tbl:"
const groupPrefixLen = len(groupPrefix)

var ErrInvalidGroup = errors.New("invalid group")

func MakeId(uid uint32, tbl string) string {
	key := strconv.FormatInt(int64(uid), 36) + ":" + tbl
	return groupPrefix + key
}

func ParseId(id string) (uid uint32, tbl string, err error) {
	if !strings.HasPrefix(id, groupPrefix) {
		return 0, "", ErrInvalidGroup
	}
	key := id[groupPrefixLen:]
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

// ------------------------------------------------------------------------

const groupItblPrefix = "itbl:"
const groupItblPrefixLen = len(groupItblPrefix)

func MakeItblId(itbl uint32) string {
	return groupItblPrefix + strconv.FormatInt(int64(itbl), 36)
}

func ParseItblId(id string) (itbl uint32, err error) {
	if !strings.HasPrefix(id, groupItblPrefix) {
		return 0, ErrInvalidGroup
	}
	key := id[groupItblPrefixLen:]
	v, err := strconv.ParseUint(key, 36, 32)
	if err != nil {
		return 0, ErrInvalidGroup
	}
	return uint32(v), nil
}
