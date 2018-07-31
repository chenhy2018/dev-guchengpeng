package userinfo

import (
	"github.com/qiniu/rpc.v1"
	"qbox.us/qconf/qconfapi"
)

// ------------------------------------------------------------------------

const DropIPWhitelistGroupPrefix = GroupPrefix + "dropIPWhitelist:"

func init() {
	prefixes = append(prefixes, DropIPWhitelistGroupPrefix)
}

// ------------------------------------------------------------------------

func (r Client) GetDropIPWhitelist(l rpc.Logger, uid uint32) (wl IPWhitelist, err error) {
	err = r.Conn.Get(l, &wl, MakeId(DropIPWhitelistGroupPrefix, uid), qconfapi.Cache_Normal)
	return
}
