package v3

import (
	"net/url"
	"strconv"
)

import (
	rpc "github.com/qiniu/rpc.v1"
	. "qbox.us/api/pay/pay"
	. "qbox.us/zone"
)

type HandleUser struct {
	Host   string
	Client *rpc.Client
}

func NewHandleUser(host string, client *rpc.Client) *HandleUser {
	return &HandleUser{host, client}
}

type ReqUidAndWhen struct {
	Uid     uint32  `json:"uid"`
	When    *Day    `json:"when"`
	Product Product `json:"product"`
	Zone    Zone    `json:"zone"`
}

type ModelPriceForPortal struct {
	Uid       uint32                                        `json:"uid"`
	Frees     map[Group]map[Item]int64                      `json:"frees"`
	Bases     map[Group]map[Item]ModelUserItemBaseForPortal `json:"bases"`
	Packages  []ModelUserPackage                            `json:"packages"`
	Discounts []ModelUserDiscount                           `json:"discounts"`
	Rebates   []ModelUserRebate                             `json:"rebates"`
}

type ModelUserItemBaseForPortal struct {
	OP string `json:"op"`
	LifeCycle
	ModelItemBasePriceForPortal
}

type ModelItemBasePriceForPortal struct {
	Type            ItemBasePriceType               `json:"type"`
	DataType        ItemDataType                    `json:"datatype"`
	CountType       ItemCountType                   `json:"counttype"`
	CumulativeCycle CumulativeType                  `json:"cumulative_cycle"`
	BillPeriodType  BillPeriodType                  `json:"bill_period_type"`
	Price           ModelRangePriceForPortal        `json:"price"`
	Unit            int64                           `json:"unit"`
	ResourceGroups  ModelResourceGroupListForPortal `json:"resource_groups"`
}

type ModelRangePriceForPortal struct {
	Type   RangePriceType             `json:"type"`
	Ranges []ModelPriceRangeForPortal `json:"ranges"`
}

type ModelPriceRangeForPortal struct {
	From  int64 `json:"from"`
	To    int64 `json:"to"`
	Price Money `json:"price"`
}

func (r HandleUser) PortalGet(logger rpc.Logger, req ReqUidAndWhen) (resp ModelPriceForPortal, err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	if req.When != nil {
		value.Add("when", (*req.When).ToString())
	}
	value.Add("product", req.Product.ToString())
	value.Add("zone", req.Zone.String())
	err = r.Client.Call(logger, &resp, r.Host+"/v3/user/portal/get?"+value.Encode())
	return
}

type ModelPrice struct {
	Uid       uint32                     `json:"uid"`
	Bases     map[Item]ModelUserItemBase `json:"bases"`
	Packages  []ModelUserPackage         `json:"packages"`
	Discounts []ModelUserDiscount        `json:"discounts"`
	Rebates   []ModelUserRebate          `json:"rebates"`
	ItemDefs  map[Item]ModelItemDef      `json:"item_defs"`
}

type ModelUserItemBase struct {
	Id string `json:"id"`
	OP string `json:"op"`
	LifeCycle
	ModelItemBasePrice
}

type ModelUserPackage struct {
	OP        string `json:"op"`
	LifeCycle `json:"lifecycle"`
	ModelPackage
}

type ModelUserDiscount struct {
	OP        string `json:"op"`
	LifeCycle `json:"lifecycle"`
	ModelDiscount
}

type ModelUserRebate struct {
	OP        string `json:"op"`
	LifeCycle `json:"lifecycle"`
	ModelRebate
}

func (r HandleUser) Get(logger rpc.Logger, req ReqUidAndWhen) (resp ModelPrice, err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	if req.When != nil {
		value.Add("when", (*req.When).ToString())
	}
	value.Add("product", req.Product.ToString())
	value.Add("zone", req.Zone.String())
	err = r.Client.Call(logger, &resp, r.Host+"/v3/user/get?"+value.Encode())
	return
}

type RespBindingPrice struct {
	Uid       uint32              `json:"uid"`
	Zone      Zone                `json:"zone"`
	Bases     []ModelUserBase     `json:"bases"`
	Packages  []ModelUserPackage  `json:"packages"`
	Discounts []ModelUserDiscount `json:"discoutns"`
	Rebates   []ModelUserRebate   `json:"rebates"`
}

type ModelUserBase struct {
	OP string `json:"op"`
	LifeCycle
	ModelBase
}

func (r HandleUser) GetAll(logger rpc.Logger, req ReqUid) (resp RespBindingPrice, err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	value.Add("zone", req.Zone.String())
	err = r.Client.Call(logger, &resp, r.Host+"/v3/user/get/all?"+value.Encode())
	return
}

type RespBindingUserPrice struct {
	Uid       uint32                   `json:"uid"`
	Bases     map[Zone][]ModelUserBase `json:"bases"`
	Packages  []ModelUserPackage       `json:"packages"`
	Discounts []ModelUserDiscount      `json:"discounts"`
	Rebates   []ModelUserRebate        `json:"rebates"`
}

func (r HandleUser) ListUserPrice(logger rpc.Logger, req ReqListUserPrice) (resp RespBindingUserPrice, err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	if req.When != nil {
		value.Add("when", (*req.When).ToString())
	}
	if req.Zone != nil {
		value.Add("zone", (*req.Zone).String())
	}
	err = r.Client.Call(logger, &resp, r.Host+"/v3/user/list/user/price?"+value.Encode())
	return
}

func (r HandleUser) ListTime(logger rpc.Logger, req ReqListTime) (resp []RespBindingPrice, err error) {
	value := url.Values{}
	value.Add("marker", strconv.FormatUint(uint64(req.Marker), 10))
	value.Add("zone", req.Zone.String())
	value.Add("limit", strconv.FormatInt(int64(req.Limit), 10))
	value.Add("kind", req.Kind)
	value.Add("from", req.From.ToString())
	value.Add("to", req.To.ToString())
	err = r.Client.Call(logger, &resp, r.Host+"/v3/user/list/time?"+value.Encode())
	return
}

type RespPriceWithLifeCycle struct {
	Uid      uint32                    `json:"uid"`
	Prices   map[Item][]RespPriceEntry `json:"prices"`
	ItemDefs map[Item]ModelItemDef     `json:"item_defs"`
}

type RespPriceEntry struct {
	From    Day                 `json:"from"`
	To      Day                 `json:"to"`
	All     RespPriceAllEntry   `json:"all"`
	Product RespPriceAllEntry   `json:"product"`
	Group   RespPriceGroupEntry `json:"group"`
	Item    RespPriceItemEntry  `json:"item"`
}

type RespPriceAllEntry struct {
	Rebates []RespBindingRebate `json:"rebates"`
}

type RespPriceGroupEntry struct {
	Discounts []RespBindingDiscount `json:"discounts"`
	Rebates   []RespBindingRebate   `json:"rebates"`
}

type RespPriceItemEntry struct {
	Base      RespBindingItemBase      `json:"base"`
	Packages  []RespBindingItemPackage `json:"packages"`
	Discounts []RespBindingDiscount    `json:"discounts"`
	Rebates   []RespBindingRebate      `json:"rebates"`
	ResPacks  []RespBindingResPack     `json:"respacks"`
}

type RespBindingItemBase struct {
	BindingInfo
	KindInfo
	Price ModelItemBasePrice `json:"price"`
}

type RespBindingResPack struct {
	BindingInfo
	KindInfo
	DataType        ItemDataType    `json:"datatype"`
	Quota           int64           `json:"quota"`
	Item            Item            `json:"item"`
	CarryOverPolicy CarryOverPolicy `json:"carry_over_policy"`
}

type RespBindingItemPackage struct {
	BindingInfo
	KindInfo
	DataType ItemDataType `json:"datatype"`
	Quota    int64        `json:"quota"`
}

type RespBindingDiscount struct {
	BindingInfo
	KindInfo
	Percent int `json:"percent"`
}

type RespBindingRebate struct {
	BindingInfo
	KindInfo
	Quota  Money `bson:"quota"`
	Rebate Money `json:"rebate"`
}

func (r HandleUser) GetTime(logger rpc.Logger, req ReqUidAndTime) (resp RespPriceWithLifeCycle, err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	value.Add("zone", req.Zone.String())
	value.Add("product", req.Product.ToString())
	value.Add("from_sec", req.FromSec.ToString())
	value.Add("to_sec", req.ToSec.ToString())
	value.Add("from", req.From.ToString())
	value.Add("to", req.To.ToString())
	err = r.Client.Call(logger, &resp, r.Host+"/v3/user/get/time?"+value.Encode())
	return
}

type ReqUserList struct {
	ID     string `json:"id"`
	Offset int    `json:"offset"`
	Limit  int    `json:"limit"`
}

type ModelUserListItem struct {
	Uid        uint32 `json:"uid"`
	EffectTime Day    `json:"effect_time"`
	DeadTime   Day    `json:"dead_time"`
}

func (r HandleUser) UserList(logger rpc.Logger, req ReqUserList) (resp []ModelUserListItem, err error) {
	value := url.Values{}
	value.Add("id", req.ID)
	value.Add("offset", strconv.FormatInt(int64(req.Offset), 10))
	value.Add("limit", strconv.FormatInt(int64(req.Limit), 10))
	err = r.Client.Call(logger, &resp, r.Host+"/v3/user/user/list?"+value.Encode())
	return
}
