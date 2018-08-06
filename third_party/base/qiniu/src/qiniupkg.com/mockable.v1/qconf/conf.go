package qconf

import (
	. "code.google.com/p/go.net/context"
	"errors"
	"github.com/qiniu/rpc.v1"
	"qbox.us/api/uc.v2"
	"qbox.us/mockacc"
	"qiniu.com/auth/proto.v1"
	"strconv"
)

type Client struct {
}

// ------------------------------

func (cli *Client) GetBucketInfo(xl rpc.Logger, uid uint32, bucket string) (info uc.BucketInfo, err error) {

	key := strconv.Itoa(int(uid)) + ":" + bucket
	info, ok := bucketInfos[key]
	if !ok {
		err = errors.New("bucket not found")
		return
	}
	return
}

var bucketInfos = map[string]uc.BucketInfo{
	"260637563:livebc": uc.BucketInfo{
		Region: "z1",
	},
}

// --------------------------------

func (cli *Client) GetAccessInfo(ctx Context, accessKey string) (ret proto.AccessInfo, err error) {

	for _, sa := range mockacc.SaInstance {
		if sa.AccessKey == accessKey {
			return proto.AccessInfo{
				Secret: []byte(sa.SecretKey),
				Uid:    sa.Uid,
				Appid:  uint64(sa.Appid),
			}, nil
		}
	}

	err = errors.New("access is not found")
	return
}

func (cli *Client) GetUtype(ctx Context, uid uint32) (utype uint32, err error) {

	ui, _, err := mockacc.SaInstance.InfoByUid(uid)
	if err != nil {
		return
	}
	return ui.Utype, nil
}
