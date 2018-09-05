package v3

import (
	"net/url"
	"strconv"
)

import (
	"github.com/qiniu/rpc.v1"
	. "qbox.us/api/pay/pay"
)

type HandleUserEvm struct {
	Host   string
	Client *rpc.Client
}

func NewHandleUserEvm(host string, client *rpc.Client) *HandleUserEvm {
	return &HandleUserEvm{host, client}
}

type ModelPriceEvmForPortal struct {
	Uid       uint32                                           `json:"uid"`
	Frees     map[Group]map[Item]int64                         `json:"frees"`
	Bases     map[Group]map[Item]ModelUserItemBaseEvmForPortal `json:"bases"`
	Packages  []ModelUserPackageEvm                            `json:"packages"`
	Discounts []ModelUserDiscount                              `json:"discounts"`
	Rebates   []ModelUserRebate                                `json:"rebates"`
}

type ModelUserItemBaseEvmForPortal struct {
	OP string `json:"op"`
	LifeCycle
	ModelItemBaseEvmPriceForPortal
}

type ModelItemBaseEvmPriceForPortal struct {
	Type      EvmPriceType                  `json:"type"`
	BasePrice string                        `json:"base_price"` // base price id, empty if non
	BaseNum   int64                         `json:"base_num"`   // base number for base price, for example, the base_num of 6Mbps is 5(Mbps)
	Ranges    []ModelEvmPriceRangeForPortal `json:"ranges"`
}

type ModelEvmPriceRangeForPortal struct {
	From  int64 `json:"from"`
	To    int64 `json:"to"`
	Price Money `json:"price"`
}

type ModelUserPackageEvm struct {
	OP        string `json:"op"`
	LifeCycle `json:"lifecycle"`
	ModelPackageEvm
}

func (r HandleUserEvm) PortalGet(logger rpc.Logger, req ReqUidAndWhen) (resp ModelPriceEvmForPortal, err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	if req.When != nil {
		value.Add("when", (*req.When).ToString())
	}
	value.Add("product", req.Product.ToString())
	value.Add("zone", req.Zone.String())
	err = r.Client.Call(logger, &resp, r.Host+"/v3/user/evm/portal/get?"+value.Encode())
	return
}

type ModelEvmPrice struct {
	Uid       uint32                        `json:"uid"`
	Bases     map[Item]ModelUserItemBaseEvm `json:"bases"`
	Packages  []ModelUserPackageEvm         `json:"packages"`
	Discounts []ModelUserDiscount           `json:"discounts"`
	Rebates   []ModelUserRebate             `json:"rebates"`
}

type ModelUserItemBaseEvm struct {
	OP string `json:"op"`
	LifeCycle
	ModelItemBaseEvmPrice
}

func (r HandleUserEvm) Get(logger rpc.Logger, req ReqUidAndWhen) (resp ModelEvmPrice, err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	if req.When != nil {
		value.Add("when", (*req.When).ToString())
	}
	value.Add("product", req.Product.ToString())
	value.Add("zone", req.Zone.String())
	err = r.Client.Call(logger, &resp, r.Host+"/v3/user/evm/get?"+value.Encode())
	return
}

type RespBindingEvmPrice struct {
	Uid       uint32                `json:"uid"`
	Bases     []ModelUserBaseEvm    `json:"bases"`
	Packages  []ModelUserPackageEvm `json:"packages"`
	Discounts []ModelUserDiscount   `json:"discoutns"`
	Rebates   []ModelUserRebate     `json:"rebates"`
}

type ModelUserBaseEvm struct {
	OP string `json:"op"`
	LifeCycle
	ModelBaseEvm
}

func (r HandleUserEvm) GetAll(logger rpc.Logger, req ReqUid) (resp RespBindingEvmPrice, err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	value.Add("zone", req.Zone.String())
	err = r.Client.Call(logger, &resp, r.Host+"/v3/user/evm/get/all?"+value.Encode())
	return
}

type RespEvmPriceWithLifeCycle struct {
	Uid    uint32                       `json:"uid"`
	Prices map[Item][]RespEvmPriceEntry `json:"prices"`
}

type RespEvmPriceEntry struct {
	From  Day                   `json:"from"`
	To    Day                   `json:"to"`
	All   RespPriceAllEntry     `json:"all"`
	Group RespPriceGroupEntry   `json:"group"`
	Item  RespEvmPriceItemEntry `json:"item"`
}

type RespEvmPriceItemEntry struct {
	Base      RespBindingItemBaseEvm      `json:"base"`
	Packages  []RespBindingItemPackageEvm `json:"packages"`
	Discounts []RespBindingDiscount       `json:"discounts"`
	Rebates   []RespBindingRebate         `json:"rebates"`
}

type RespBindingItemBaseEvm struct {
	BindingInfo
	KindInfo
	Price ModelItemBaseEvmPrice `json:"price"`
}

type RespBindingItemPackageEvm struct {
	BindingInfo
	KindInfo
	RecId string `json:"rec_id"`
	Quota int64  `json:"quota"`
}

func (r HandleUserEvm) GetTime(logger rpc.Logger, req ReqUidAndTime) (resp RespEvmPriceWithLifeCycle, err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	value.Add("zone", req.Zone.String())
	value.Add("product", req.Product.ToString())
	value.Add("from_sec", req.FromSec.ToString())
	value.Add("to_sec", req.ToSec.ToString())
	value.Add("from", req.From.ToString())
	value.Add("to", req.To.ToString())
	err = r.Client.Call(logger, &resp, r.Host+"/v3/user/evm/get/time?"+value.Encode())
	return
}

func (r HandleUserEvm) UserList(logger rpc.Logger, req ReqUserList) (resp []ModelUserListItem, err error) {
	value := url.Values{}
	value.Add("id", req.ID)
	value.Add("offset", strconv.FormatInt(int64(req.Offset), 10))
	value.Add("limit", strconv.FormatInt(int64(req.Limit), 10))
	err = r.Client.Call(logger, &resp, r.Host+"/v3/user/evm/user/list?"+value.Encode())
	return
}
