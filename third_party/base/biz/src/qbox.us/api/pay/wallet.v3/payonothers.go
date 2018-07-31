package wallet

import (
	"net/url"
	"strconv"
)

import (
	"github.com/qiniu/rpc.v1"
)

type HandlePayOnOthers struct {
	Host   string
	Client *rpc.Client
}

func NewHandlePayOnOthers(host string, client *rpc.Client) *HandlePayOnOthers {
	return &HandlePayOnOthers{host, client}
}

type ReqPayTransactions struct {
	SerialNums []string `json:"serial_nums"`
	Uid        uint32   `json:"uid"`
}

func (r HandlePayOnOthers) Paytransactions(logger rpc.Logger, req ReqPayTransactions) (err error) {
	err = r.Client.CallWithJson(logger, nil, r.Host+"/v3/payonothers/paytransactions", req)
	return
}

func (r HandlePayOnOthers) GetUnpaytransactions(logger rpc.Logger, req ReqUid) (serialNums []string, err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	err = r.Client.Call(logger, &serialNums, r.Host+"/v3/payonothers/get/unpaytransactions?"+value.Encode())
	return
}
