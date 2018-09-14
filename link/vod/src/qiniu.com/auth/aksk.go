package auth

import (
	"errors"
	"os"

	xlog "github.com/qiniu/xlog.v1"
	"qbox.us/qconf/qconfapi"
	proto "qiniu.com/auth/proto.v1"
	"qiniu.com/system"
)

const (
	AK_PREFIX = "ak:"
)

var (
	localSk []byte
)

var QConfClient *qconfapi.Client

func Init(conf *system.Configuration) {
	QConfClient = qconfapi.New(&conf.Qconf)
	xl := xlog.NewDummy()
	if QConfClient == nil {
		xl.Error("init qconf client failed")
		os.Exit(3)
	}
}
func GetUserInfoFromQconf(xl *xlog.Logger, accessKey string) (*proto.AccessInfo, error) {
	resp := proto.AccessInfo{}
	if QConfClient == nil {
		resp.Uid = 0
		resp.Appid = 0
		resp.Secret = localSk
	} else {
		err := QConfClient.Get(nil, &resp, AK_PREFIX+accessKey, qconfapi.Cache_NoSuchEntry)
		if err != nil {
			return nil, errors.New("get account info failed")

		}
	}
	return &resp, nil
}

func SetSkFromUser(xl *xlog.Logger, sk []byte) {
	localSk = sk
}
