package v3

import (
	"net/url"
)

import (
	"github.com/qiniu/rpc.v1"
)

type ModelProductDef struct {
	Id      string `json:"id"`
	Display string `json:"display"`
}

type ModelGroupDef struct {
	Id      string `json:"id"`
	Display string `json:"display"`
}

type ModelItemDef struct {
	Id          string                            `json:"id"`
	Group       string                            `json:"group"`
	Product     string                            `json:"product"`
	DataDef     map[ItemDataType]ModelItemDataDef `json:"data_def"`
	IsBasic     bool                              `json:"is_basic"` // true will generate bill even if no data
	IsMultiZone bool                              `json:"is_multi_zone"`
	Price       map[string]string                 `json:"price"` // default price
	Extra       map[string]string                 `json:"extra"`
	IsDisabled  bool                              `json:"is_disabled"`
}

type ModelItemDataDef struct {
	Display          string     `json:"display"`
	DeductType       DeductType `json:"deduct_type"`
	HasRatioForMonth bool       `json:"has_ratio_month"`

	HasRatioForBillPeriod bool `json:"has_ratio_billperiod"`

	// Factor for converting raw stat value to billing value
	// Example Transfer:
	// Unit: 1024 * 1024 * 1024, convert raw byte to GB
	// CustomUnitExpr: "unit * unit * unit"
	Unit           int64  `json:"unit"`
	CustomUnitExpr string `json:"custom_unit_expr"`

	// Factor for each display unit conversion
	// Example Transfer:
	// DisplayUnits: ["GB", "TB", "PB"]
	// DisplayUnitStep: 1024
	DisplayUnits    []string `json:"display_units"`
	DisplayUnitStep int64    `json:"display_unit_step"`

	StatSrc    StatSrc    `json:"stat_src"`
	StatMethod StatMethod `json:"stat_method"`

	Extra map[string]string `json:"extra"`
}

type HandleItem struct {
	Host   string
	Client *rpc.Client
}

func NewHandleItem(host string, client *rpc.Client) *HandleItem {
	return &HandleItem{host, client}
}

func (r HandleItem) Get(logger rpc.Logger, req ReqID) (def ModelItemDef, err error) {
	value := url.Values{}
	value.Add("id", req.ID)
	err = r.Client.Call(logger, &def, r.Host+"/v3/item/get?"+value.Encode())
	return
}

func (r HandleItem) Set(logger rpc.Logger, req ModelItemDef) (err error) {
	err = r.Client.CallWithJson(logger, nil, r.Host+"/v3/item/set", req)
	return
}

type ReqItemList struct {
	Products string `json:"products"` // separated by `,`
}

func (r HandleItem) List(logger rpc.Logger, req ReqItemList) (defs []ModelItemDef, err error) {
	value := url.Values{}
	value.Add("products", req.Products)
	err = r.Client.Call(logger, &defs, r.Host+"/v3/item/list?"+value.Encode())
	return
}
