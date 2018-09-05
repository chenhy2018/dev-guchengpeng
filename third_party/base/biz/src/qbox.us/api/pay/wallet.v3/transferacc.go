package wallet

import (
	"net/url"
	"strconv"
)

import (
	"github.com/qiniu/rpc.v1"
	. "qbox.us/api/pay/pay"
)

type HandleTransferAccount struct {
	Host   string
	Client *rpc.Client
}

func NewHandleTransferAccount(host string, client *rpc.Client) *HandleTransferAccount {
	return &HandleTransferAccount{host, client}
}

type ReqTransferAcc struct {
	FromUid uint32 `json:"from_uid"`
	ToUid   uint32 `json:"to_uid"`
	Cash    Money  `json:"cash"`
	FreeNb  Money  `json:"freenb"`
	Desc    string `json:"desc"`
}

// 转账 FromUid -> ToUid
// 把FromUid的余额转入ToUid账户，包含cash，freenb， coupon, 可指定cash和freenb
func (r HandleTransferAccount) Transferaccount(logger rpc.Logger, req ReqTransferAcc) (err error) {
	value := url.Values{}
	value.Add("from_uid", strconv.FormatUint(uint64(req.FromUid), 10))
	value.Add("to_uid", strconv.FormatUint(uint64(req.ToUid), 10))
	value.Add("cash", req.Cash.ToString())
	value.Add("freenb", req.FreeNb.ToString())
	value.Add("desc", req.Desc)
	err = r.Client.CallWithForm(logger, nil, r.Host+"/v3/transferaccount/transferaccount", map[string][]string(value))
	return
}
