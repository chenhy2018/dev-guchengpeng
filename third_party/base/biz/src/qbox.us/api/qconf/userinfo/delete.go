package userinfo

import (
	"github.com/qiniu/rpc.v1"
	"qbox.us/qconf/qconfapi"
)

// ------------------------------------------------------------------------

const DeleteIPWhitelistGroupPrefix = GroupPrefix + "deleteIPWhitelist:"

func init() {
	prefixes = append(prefixes, DeleteIPWhitelistGroupPrefix)
}

// ------------------------------------------------------------------------

type IPWhitelist struct {
	Whitelist []string `bson:"ip_whitelist"`
}

func (r Client) GetDeleteIPWhitelist(l rpc.Logger, uid uint32) (wl IPWhitelist, err error) {
	err = r.Conn.Get(l, &wl, MakeId(DeleteIPWhitelistGroupPrefix, uid), qconfapi.Cache_Normal)
	return
}
