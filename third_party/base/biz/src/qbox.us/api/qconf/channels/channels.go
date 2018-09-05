// 这个包没用了(现在所有的域名都在数据库中)
// 通过 uid->channels, bucket->channels, uid->utype 三个qconfapi， 得到 uid,bucket -> channels 的qconfapi
package channels

import (
	"strings"

	"github.com/qiniu/rpc.v1"
	"qbox.us/api/qconf/bucketinfo.v2"
	"qbox.us/api/qconf/uidg"
	"qbox.us/api/qconf/userinfo"
	"qbox.us/api/uc"
	"qbox.us/qconf/qconfapi"
	auth "qiniu.com/auth/proto.v1"
)

// --------------------------------------------------------------------------------

type Client struct {
	Conn *qconfapi.Client
}

// --------------------------------------------------------------------------------

var UtypeDefaultChannels = map[string]uint32{
	"com1": auth.USER_TYPE_USERS | auth.USER_TYPE_SUDOERS | auth.USER_TYPE_VIP,
	"com2": auth.USER_TYPE_STDUSER2 | auth.USER_TYPE_VIP,
}

func utype2Channels(utype uint32) (channels []string) {
	for channel, utypeAllowed := range UtypeDefaultChannels {
		if utypeAllowed&utype != 0 {
			channels = append(channels, channel)
		}
	}
	return
}

func (p Client) getDefaultChannels(l rpc.Logger, uid uint32, bucketType uc.BucketType) (channels []string, err error) {

	switch bucketType {
	case uc.TYPE_COM:
		var utype uint32
		utype, err = uidg.Client{p.Conn}.GetUtype(l, uid)
		if err != nil {
			return
		}
		channels = utype2Channels(utype)
	case uc.TYPE_MEDIA:
		channels = []string{"media1"}
	case uc.TYPE_DL:
		channels = []string{"dl1"}
	}
	return
}

func (p Client) GetChannels(l rpc.Logger, uid uint32, bucket string) (channels []string, err error) {

	userInfo, err := userinfo.Client{p.Conn}.GetChannel(l, uid)
	if err != nil {
		return
	}
	bucketInfo, err := bucketinfo.Client{p.Conn}.GetBucketInfo(l, uid, bucket)
	if err != nil {
		return
	}
	denied := getDeniedChannels(userInfo)

	channels, err = p.getDefaultChannels(l, uid, bucketInfo.Type)
	if err != nil {
		return
	}

	channels = union(channels, intersection(userInfo.Channels, bucketInfo.Channel))
	channels = difference(channels, denied)
	return
}

// 如果userInfo里的channel开头带感叹号，认为是被禁用的
func getDeniedChannels(userInfo userinfo.ChannelInfo) (denied []string) {
	for _, c := range userInfo.Channels {
		if strings.HasPrefix(c, "!") {
			denied = append(denied, c[1:])
		}
	}
	return
}

// --------------------------------------------------------------------------------
