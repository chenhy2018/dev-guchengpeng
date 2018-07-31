package domaing

import (
	"strconv"
	"strings"

	"github.com/qiniu/rpc.v1"
	"qbox.us/api/one/domain"
	qconf "qbox.us/qconf/qconfapi"
)

func (r Client) List(l rpc.Logger, owner uint32, tblName string) (domains []domain.Entry, err error) {

	var ret struct {
		Domains []domain.Entry `bson:"domains"`
	}
	err = r.Conn.Get(l, &ret, MakeListId(owner, tblName), qconf.Cache_Normal)
	return ret.Domains, err
}

const groupListPrefix = "domainList:"
const groupListPrefixLen = len(groupListPrefix)

func MakeListId(owner uint32, tblName string) string {
	key := strconv.FormatInt(int64(owner), 36) + ":" + tblName
	return groupListPrefix + key
}

func ParseListId(id string) (owner uint32, tblName string, err error) {
	if !strings.HasPrefix(id, groupListPrefix) {
		return 0, "", ErrInvalidGroup
	}
	s := strings.SplitN(id[groupListPrefixLen:], ":", 2)
	uid64, err := strconv.ParseUint(s[0], 36, 32)
	if err != nil {
		return
	}
	return uint32(uid64), s[1], err
}
