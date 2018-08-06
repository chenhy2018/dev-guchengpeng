package v3

import (
	"net/url"
	"strconv"
)

import (
	"github.com/qiniu/rpc.v1"
	. "qbox.us/api/pay/pay"
)

type HandleBaseEvm struct {
	Host   string
	Client *rpc.Client
}

func NewHandleBaseEvm(host string, client *rpc.Client) *HandleBaseEvm {
	return &HandleBaseEvm{host, client}
}

type ModelBaseEvm struct {
	KindInfo
	Items map[Item]ModelItemBaseEvmPrice `json:"items"`
}

type ModelItemBaseEvmPrice struct {
	Type      EvmPriceType `json:"type"`
	BasePrice string       `json:"base_price"` // base price id, empty if non
	BaseNum   int64        `json:"base_num"`   // base number for base price, for example, the base_num of 6Mbps is 5(Mbps)
	Price     Money        `json:"price"`
}

// 获取基础价格
func (r HandleBaseEvm) Add(logger rpc.Logger, req ModelBaseEvm) (err error) {
	err = r.Client.CallWithJson(logger, nil, r.Host+"/v3/base/evm/add", req)
	return
}

func (r HandleBaseEvm) Get(logger rpc.Logger, req ReqID) (resp ModelBaseEvm, err error) {
	value := url.Values{}
	value.Add("id", req.ID)
	err = r.Client.Call(logger, &resp, r.Host+"/v3/base/evm/get?"+value.Encode())
	return
}

func (r HandleBaseEvm) List(logger rpc.Logger, req ReqList) (resp []ModelBaseEvm, err error) {
	value := url.Values{}
	value.Add("offset", strconv.FormatInt(int64(req.Offset), 10))
	value.Add("limit", strconv.FormatInt(int64(req.Limit), 10))
	value.Add("type", req.Type)
	err = r.Client.Call(logger, &resp, r.Host+"/v3/base/evm/list?"+value.Encode())
	return
}
