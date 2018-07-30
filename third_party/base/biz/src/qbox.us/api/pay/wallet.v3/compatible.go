package wallet

import (
	"net/url"
	"strconv"
)

import (
	"github.com/qiniu/rpc.v1"
)

type HandleCompatible struct {
	Host   string
	Client *rpc.Client
}

func NewHandleCompatible(host string, client *rpc.Client) *HandleCompatible {
	return &HandleCompatible{host, client}
}

func (r HandleCompatible) Comfirmrecharge(logger rpc.Logger, req ReqUid) (err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	err = r.Client.CallWithForm(logger, nil, r.Host+"/v3/compatible/comfirmrecharge", map[string][]string(value))
	return
}
