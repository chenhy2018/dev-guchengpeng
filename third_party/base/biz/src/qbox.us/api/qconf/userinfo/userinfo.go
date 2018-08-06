package userinfo

import (
	"errors"
	"strconv"
	"strings"

	"qbox.us/qconf/qconfapi"
)

const GroupPrefix = "userInfo:"
const GroupPrefixLen = len(GroupPrefix)

var ErrInvalidGroup = errors.New("invalid group")

type Client struct {
	Conn *qconfapi.Client
}

func MakeId(prefix string, uid uint32) string {
	return prefix + strconv.FormatUint(uint64(uid), 10)
}

var prefixes = []string{}

func ParseId(id string) (uid uint32, err error) {
	for _, prefix := range prefixes {
		if strings.HasPrefix(id, prefix) {
			uid64, err := strconv.ParseUint(id[len(prefix):], 10, 32)
			return uint32(uid64), err
		}
	}
	return 0, ErrInvalidGroup
}
