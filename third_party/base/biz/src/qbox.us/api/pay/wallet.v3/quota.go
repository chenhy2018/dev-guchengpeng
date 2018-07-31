package wallet

import (
	"net/url"
	"strconv"
)

import (
	"github.com/qiniu/rpc.v1"
	. "qbox.us/api/pay/pay"
	P "qbox.us/api/pay/price/v3"
)

type HandleQuota struct {
	Host   string
	Client *rpc.Client
}

func NewHandleQuota(host string, client *rpc.Client) *HandleQuota {
	return &HandleQuota{host, client}
}

type ModelUserQuota struct {
	Uid         uint32                 `json:"uid"`
	Month       Month                  `json:"month"`
	ModTime     HundredNanoSecond      `json:"modtime"`
	Packages    []ModelPackageQuota    `json:"packages"`
	Rebates     []ModelRebateQuota     `json:"rebates"`
	EvmPackages []ModelEvmPackageQuota `json:"evm_packages"`
	EvmRebates  []ModelRebateQuota     `json:"evm_rebates"`
}

type ModelPackageQuota struct {
	OP    string                  `json:"op"`
	Items map[Item]ModelItemQuota `json:"items"`
}

type ModelEvmPackageQuota struct {
	OP    string                    `json:"op"`
	Items map[string]map[Item]int64 `json:"items"`
}

type ModelItemQuota struct {
	DataType P.ItemDataType `json:"data"`
	Quota    int64          `json:"quota"`
}

type ModelRebateQuota struct {
	OP     string `json:"op"`
	Quota  Money  `json:"quota"`
	Reduce Money  `json:"reduce"`
}

func (r HandleQuota) Get(logger rpc.Logger, req ReqUidAndMonth) (resp ModelUserQuota, err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	value.Add("month", req.Month.ToString())
	value.Add("limit", strconv.FormatInt(int64(req.Limit), 10))
	err = r.Client.Call(logger, &resp, r.Host+"/v3/quota/get?"+value.Encode())
	return
}

func (r HandleQuota) Set(logger rpc.Logger, req ModelUserQuota) (err error) {
	err = r.Client.CallWithJson(logger, nil, r.Host+"/v3/quota/set", req)
	return
}
