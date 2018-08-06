package mockacc

import (
	. "code.google.com/p/go.net/context"
	"errors"
	authp "qiniu.com/auth/proto.v1"
)

type AuthpAcc struct {
	auth map[string]*authInfo
}

var AuthpAccInstance AuthpAcc

func (aa AuthpAcc) GetAccessInfo(ctx Context, accessKey string) (ret authp.AccessInfo, err error) {

	auth := aa.auth
	if auth == nil {
		auth = defaultImpl.auth
	}

	info, ok := auth[accessKey]
	if !ok {
		err = errors.New("not found")
		return
	}
	ret.Secret = info.secret
	ret.Appid = uint64(info.appId)
	ret.Uid = info.uid
	return
}

func (aa AuthpAcc) GetUtype(ctx Context, uid uint32) (utype uint32, err error) {

	auth := aa.auth
	if auth == nil {
		auth = defaultImpl.auth
	}

	for _, v := range auth {
		if v.uid == uid {
			utype = v.utype
			return
		}
	}
	err = errors.New("not found")
	return
}
