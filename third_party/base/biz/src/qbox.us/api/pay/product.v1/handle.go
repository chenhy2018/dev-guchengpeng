package product

import (
	"net/url"
	"strconv"
)

import (
	"time"

	"github.com/qiniu/rpc.v1"
	. "qbox.us/api/pay/pay"
	. "qbox.us/zone"
)

type HandleUserProduct struct {
	Host   string
	Client *rpc.Client
}

func NewHandleUserProduct(host string, client *rpc.Client) *HandleUserProduct {
	return &HandleUserProduct{host, client}
}

type RespUserProduct struct {
	ID        string    `json:"id"`
	Uid       uint32    `json:"uid"`
	Parent    uint32    `json:"parent"`
	Zone      string    `json:"zone"`
	Product   string    `json:"product"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ReqSetUserProduct struct {
	Uid     uint32  `json:"uid"`
	Parent  uint32  `json:"parent"`
	Product Product `json:"product"`
	Zone    Zone    `json:"zone"`
	Update  *bool   `json:"update"`
}

type ReqUserProducts struct {
	LastID       string `json:"last_id"`
	PageSize     int    `json:"page_size"`
	Uid          uint32 `json:"uid"`
	Parent       uint32 `json:"parent"`
	Zone         Zone   `json:"zone"`
	UpdatedAtStr string `json:"updated_at"`
	UpdatedAt    time.Time
}

type ReqProductUsers struct {
	LastID       string  `json:"last_id"`
	PageSize     int     `json:"page_size"`
	Product      Product `json:"product"`
	Zone         Zone    `json:"zone"`
	UpdatedAtStr string  `json:"updated_at"`
	UpdatedAt    time.Time
}

type ReqMeasureUids struct {
	Product  Product `json:"product"`
	Zone     Zone    `json:"zone"`
	Page     int     `json:"page"`
	PageSize int     `json:"page_size"`
}

type ReqMeasureNoZoneUids struct {
	Product  Product `json:"product"`
	Page     int     `json:"page"`
	PageSize int     `json:"page_size"`
}

type ReqProductChildren struct {
	Product Product `json:"product"`
	Zone    Zone    `json:"zone"`
	Parent  uint32  `json:"parent"`
}

type ReqProductChildrenNoZone struct {
	Product Product `json:"product"`
	Parent  uint32  `json:"parent"`
}

type ReqDataSync struct {
	Product Product `json:"product"`
	Zone    Zone    `json:"zone"`
}

type RespDataSyncStatus struct {
	Product   Product   `json:"product"`
	Zone      Zone      `json:"zone"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Done      bool      `json:"done"`
	Count     int64     `json:"count"`
}

type RespProductList struct {
	Zone Zone `json:"zone"`
	RespLabel
	Items map[Product]RespProduct `json:"items"`
}

type RespProduct struct {
	RespLabel
	Items map[Group]RespGroup `json:"items"`
}

type RespGroup struct {
	RespLabel
	Items map[Item]RespLabel `json:"items"`
}

type RespLabel struct {
	Label string `json:"label"`
}

type RespUidList struct {
	Uids   []uint32 `json:"uids"`
	LastID string   `json:"last_id"`
}

// WspSet POST:/set add new userProduct
// Content-type:application/x-www-form-urlencoded
func (r HandleUserProduct) Set(logger rpc.Logger, params ReqSetUserProduct) (err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(params.Uid), 10))
	value.Add("parent", strconv.FormatUint(uint64(params.Parent), 10))
	value.Add("product", params.Product.ToString())
	value.Add("zone", params.Zone.String())
	if params.Update != nil {
		value.Add("update", strconv.FormatBool(*params.Update))
	}
	err = r.Client.CallWithForm(logger, nil, r.Host+"/v1/user_product/set", map[string][]string(value))
	return
}

// WsUserProducts  GET:/user/products list user's UserProducts by uid, zone and updated time
func (r HandleUserProduct) UserProducts(logger rpc.Logger, params ReqUserProducts) (data []RespUserProduct, err error) {
	value := url.Values{}
	value.Add("last_id", params.LastID)
	value.Add("page_size", strconv.FormatInt(int64(params.PageSize), 10))
	value.Add("uid", strconv.FormatUint(uint64(params.Uid), 10))
	value.Add("parent", strconv.FormatUint(uint64(params.Parent), 10))
	value.Add("zone", params.Zone.String())
	value.Add("updated_at", params.UpdatedAtStr)

	err = r.Client.Call(logger, &data, r.Host+"/v1/user_product/user/products?"+value.Encode())
	return
}

// WsProductUsers GET:/product/users list UserProducts by product, zone and updated time
func (r HandleUserProduct) ProductUsers(logger rpc.Logger, params ReqProductUsers) (data []RespUserProduct, err error) {
	value := url.Values{}
	value.Add("last_id", params.LastID)
	value.Add("page_size", strconv.FormatInt(int64(params.PageSize), 10))
	value.Add("product", params.Product.ToString())
	value.Add("zone", params.Zone.String())
	value.Add("updated_at", params.UpdatedAtStr)

	err = r.Client.Call(logger, &data, r.Host+"/v1/user_product/product/users?"+value.Encode())
	return
}

func (r HandleUserProduct) ProductMeasureUids(logger rpc.Logger, params ReqMeasureUids) (uids []uint32, err error) {
	value := url.Values{}
	value.Add("product", params.Product.ToString())
	value.Add("zone", params.Zone.String())
	value.Add("page", strconv.FormatInt(int64(params.Page), 10))
	value.Add("page_size", strconv.FormatInt(int64(params.PageSize), 10))
	err = r.Client.Call(logger, &uids, r.Host+"/v1/user_product/product/measure/uids?"+value.Encode())
	return
}

func (r HandleUserProduct) ProductChildren(logger rpc.Logger, params ReqProductChildren) (children []uint32, err error) {
	value := url.Values{}
	value.Add("product", params.Product.ToString())
	value.Add("zone", params.Zone.String())
	value.Add("parent", strconv.FormatUint(uint64(params.Parent), 10))
	err = r.Client.Call(logger, &children, r.Host+"/v1/user_product/product/children?"+value.Encode())
	return
}

// WsProductUids GET:/product/uids similar with WsProductUsers, only return uids for performance
func (r HandleUserProduct) ProductUids(logger rpc.Logger, params ReqProductUsers) (data RespUidList, err error) {
	value := url.Values{}
	value.Add("last_id", params.LastID)
	value.Add("page_size", strconv.FormatInt(int64(params.PageSize), 10))
	value.Add("product", params.Product.ToString())
	value.Add("zone", params.Zone.String())
	value.Add("updated_at", params.UpdatedAtStr)

	err = r.Client.Call(logger, &data, r.Host+"/v1/user_product/product/uids?"+value.Encode())
	return
}

type ReqProductUsersNoZone struct {
	LastID       string  `json:"last_id"`
	PageSize     int     `json:"page_size"`
	Product      Product `json:"product"`
	UpdatedAtStr string  `json:"updated_at"`
	UpdatedAt    time.Time
}

func (r HandleUserProduct) ProductUidsNoZone(logger rpc.Logger, params ReqProductUsersNoZone) (data RespUidList, err error) {
	value := url.Values{}
	value.Add("last_id", params.LastID)
	value.Add("page_size", strconv.FormatInt(int64(params.PageSize), 10))
	value.Add("product", params.Product.ToString())
	value.Add("updated_at", params.UpdatedAtStr)

	err = r.Client.Call(logger, &data, r.Host+"/v1/user_product/product/uids/no/zone?"+value.Encode())
	return
}

func (r HandleUserProduct) ProductMeasureUidsNoZone(logger rpc.Logger, params ReqMeasureNoZoneUids) (uids []uint32, err error) {
	value := url.Values{}
	value.Add("product", params.Product.ToString())
	value.Add("page", strconv.FormatInt(int64(params.Page), 10))
	value.Add("page_size", strconv.FormatInt(int64(params.PageSize), 10))
	err = r.Client.Call(logger, &uids, r.Host+"/v1/user_product/product/measure/uids/no/zone?"+value.Encode())
	return
}

func (r HandleUserProduct) ProductChildrenNoZone(logger rpc.Logger, params ReqProductChildrenNoZone) (children []uint32, err error) {
	value := url.Values{}
	value.Add("product", params.Product.ToString())
	value.Add("parent", strconv.FormatUint(uint64(params.Parent), 10))
	err = r.Client.Call(logger, &children, r.Host+"/v1/user_product/product/children/no/zone?"+value.Encode())
	return
}

func (r HandleUserProduct) ProductList(logger rpc.Logger, params ReqProductUsers) (data RespProductList, err error) {
	value := url.Values{}
	value.Add("last_id", params.LastID)
	value.Add("page_size", strconv.FormatInt(int64(params.PageSize), 10))
	value.Add("product", params.Product.ToString())
	value.Add("zone", params.Zone.String())
	value.Add("updated_at", params.UpdatedAtStr)

	err = r.Client.Call(logger, &data, r.Host+"/v1/user_product/product/list?"+value.Encode())
	return
}
