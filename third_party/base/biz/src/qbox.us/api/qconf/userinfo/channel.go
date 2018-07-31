package userinfo

import (
	"github.com/qiniu/rpc.v1"
	"qbox.us/qconf/qconfapi"
)

// ------------------------------------------------------------------------

const ChannelGroupPrefix = GroupPrefix + "channel:"

func init() {
	prefixes = append(prefixes, ChannelGroupPrefix)
}

// ------------------------------------------------------------------------

type ChannelInfo struct {
	Channels []string `bson:"channel"`
}

func (r Client) GetChannel(l rpc.Logger, uid uint32) (ret ChannelInfo, err error) {

	err = r.Conn.Get(l, &ret, MakeId(ChannelGroupPrefix, uid), qconfapi.Cache_Normal)
	return
}

// ------------------------------------------------------------------------
