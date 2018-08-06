package price

import (
	"encoding/json"
	"errors"
	"net/http"
	neturl "net/url"
	"strconv"
	"time"

	adminacc "qbox.us/admin_api/v2/account"
	"qbox.us/rpc"
	"qbox.us/servend/account"
	"qbox.us/servend/oauth"
)

//-----------------------------------------------------------------------------------------------//

const (
	TYPE_BASE     = "base"
	TYPE_DISCOUNT = "discount"
	TYPE_PACKAGE  = "package"
)

const (
	TYPE_PACKAGE_NORMAL = "normal"
	TYPE_PACKAGE_REWARD = "reward"
)

const (
	BASE_PRICE_TYPE_DEFAULT                 = ""
	BASE_PRICE_TYPE_TOP4TH                  = "top4th"
	BASE_PRICE_TYPE_TOP95_PERCENT           = "top95%"
	BASE_PRICE_TYPE_DAY_AVERAGE             = "day_average"
	BASE_PRICE_TYPE_AGREEMENT               = "agreement"
	BASE_PRICE_TYPE_AGREEMENT_TOP4TH        = "agreement_top4th"
	BASE_PRICE_TYPE_AGREEMENT_TOP95_PERCENT = "agreement_top95%"
	BASE_PRICE_TYPE_AGREEMENT_DAY_AVERAGE   = "agreement_day_average"
)

type RangePrice struct {
	Range int64 `json:"range" bson:"range"`
	Price int64 `json:"price" bson:"price"`
}

type BasePrice struct {
	Id   string `json:"id" bson:"id"`
	Type string `json:"type" bson:"type"`
	Desc string `json:"desc" bson:"desc"`

	Space        []RangePrice `json:"space" bson:"space"`
	Transfer_out []RangePrice `json:"transfer_out" bson:"transfer_out"`
	Bandwidth    []RangePrice `json:"bandwidth" bson:"bandwidth"`
	Api_Get      []RangePrice `json:"api_get" bson:"api_get"`
	Api_Put      []RangePrice `json:"api_put" bson:"api_put"`
}

type DiscountInfo struct {
	Id   string `json:"id" bson:"id"`
	Type string `json:"type" bson:"type"`
	Desc string `json:"desc" bson:"desc"`

	Money   int64 `json:"money" bson:"money"`
	Percent int   `json:"percent" bson:"percent"`
}

type PackageInfo struct {
	Id   string `json:"id" bson:"id"`
	Type string `json:"type" bson:"type"`
	Desc string `json:"desc" bson:"desc"`

	Money        int64 `json:"money" bson:"money"`
	Space        int64 `json:"space" bson:"space"`
	Transfer_out int64 `json:"transfer_out" bson:"transfer_out"`
	Bandwidth    int64 `json:"bandwidth" bson:"bandwidth"`
	Api_Get      int64 `json:"api_get" bson:"api_get"`
	Api_Put      int64 `json:"api_put" bson:"api_put"`
}

type PriceUnit struct {
	Type       string      `json:"type" bson:"type"`
	EffectTime int64       `json:"effecttime" bson:"effecttime"`
	DeadTime   int64       `json:"deadtime" bson:"deadtime"`
	Info       interface{} `json:"info" bson:"info"`
}

type Price struct {
	Uid   uint32      `json:"uid" bson:"uid"`
	Units []PriceUnit `json:"units" bson:"units"`
}

//---------------------------------------------------------------------------//

type EncodingPriceUnit struct {
	Type       string `json:"type" bson:"type"`
	EffectTime int64  `json:"effecttime" bson:"effecttime"`
	DeadTime   int64  `json:"deadtime" bson:"deadtime"`
	Info       string `json:"info" bson:"info"`
}

type EncodingPrice struct {
	Uid   uint32              `json:"uid" bson:"uid"`
	Units []EncodingPriceUnit `json:"units" bson:"units"`
}

//-----------------------------------------------------------------------------------------------//

func MarshalPriceUnit(unit PriceUnit) (unit2 EncodingPriceUnit, err error) {
	switch unit.Type {
	case TYPE_BASE:
		var bs []byte
		bs, err = json.Marshal(unit.Info.(BasePrice))
		if err != nil {
			return
		}
		unit2 = EncodingPriceUnit{unit.Type, unit.EffectTime, unit.DeadTime,
			string(bs)}
		return
	case TYPE_DISCOUNT:
		var bs []byte
		bs, err = json.Marshal(unit.Info.(DiscountInfo))
		if err != nil {
			return
		}
		unit2 = EncodingPriceUnit{unit.Type, unit.EffectTime, unit.DeadTime,
			string(bs)}
		return
	case TYPE_PACKAGE:
		var bs []byte
		bs, err = json.Marshal(unit.Info.(PackageInfo))
		if err != nil {
			return
		}
		unit2 = EncodingPriceUnit{unit.Type, unit.EffectTime, unit.DeadTime,
			string(bs)}
		return
	}
	err = errors.New("type out of range:should be base, discount and package")
	return
}

func MarshalPrice(price1 Price) (price2 EncodingPrice, err error) {
	price2.Uid = price1.Uid
	if price1.Units != nil {
		price2.Units = make([]EncodingPriceUnit, len(price1.Units))
		for i, unit := range price1.Units {
			price2.Units[i], err = MarshalPriceUnit(unit)
			if err != nil {
				return
			}
		}
	}
	return
}

func UnmarshalPriceUnit(unit EncodingPriceUnit) (unit2 PriceUnit, err error) {
	switch unit.Type {
	case TYPE_BASE:
		var price_ BasePrice
		err = json.Unmarshal([]byte(unit.Info), &price_)
		if err != nil {
			return
		}
		unit2 = PriceUnit{unit.Type, unit.EffectTime, unit.DeadTime, price_}
		return
	case TYPE_DISCOUNT:
		var price_ DiscountInfo
		err = json.Unmarshal([]byte(unit.Info), &price_)
		if err != nil {
			return
		}
		unit2 = PriceUnit{unit.Type, unit.EffectTime, unit.DeadTime, price_}
		return
	case TYPE_PACKAGE:
		var price_ PackageInfo
		err = json.Unmarshal([]byte(unit.Info), &price_)
		if err != nil {
			return
		}
		unit2 = PriceUnit{unit.Type, unit.EffectTime, unit.DeadTime, price_}
		return
	}
	err = errors.New("type out of range:should be base, discount and package")
	return
}

func UnmarshalPrice(price1 EncodingPrice) (price2 Price, err error) {
	price2.Uid = price1.Uid
	if price1.Units != nil {
		price2.Units = make([]PriceUnit, len(price1.Units))
		for i, unit := range price1.Units {
			price2.Units[i], err = UnmarshalPriceUnit(unit)
			if err != nil {
				return
			}
		}
	}
	return
}

//---------------------------------------------------------------------------//

func GetPrice(c rpc.Client, host string, uid uint32, customerGroup adminacc.CustomerGroup, time_ time.Time) (price Price, code int, err error) {
	url := host + "/get"
	url += "?uid=" + strconv.FormatUint(uint64(uid), 10)
	url += "&customer_group=" + strconv.Itoa(int(customerGroup))
	url += "&time=" + strconv.FormatInt(time_.UnixNano()/100, 10)
	var price_ EncodingPrice
	code, err = c.Call(&price_, url)
	if err == nil {
		price, err = UnmarshalPrice(price_)
	}
	return
}

func GetPriceEx(c rpc.Client, host string, uid uint32, customerGroup adminacc.CustomerGroup) (price Price, code int, err error) {
	url := host + "/get"
	url += "?uid=" + strconv.FormatUint(uint64(uid), 10)
	url += "&customer_group=" + strconv.Itoa(int(customerGroup))
	var price_ EncodingPrice
	code, err = c.Call(&price_, url)
	if err == nil {
		price, err = UnmarshalPrice(price_)
	}
	return
}

func GetPriceListEx(c rpc.Client, host string, time_ time.Time, offset, limit int) (prices []Price, code int, err error) {
	url := host + "/list"
	url += "?time=" + strconv.FormatInt(time_.UnixNano()/100, 10)
	url += "&offset=" + strconv.Itoa(offset)
	url += "&limit=" + strconv.Itoa(limit)
	price_s := make([]EncodingPrice, 0)
	code, err = c.Call(&price_s, url)
	if err == nil {
		prices = make([]Price, len(price_s))
		for i, price_ := range price_s {
			prices[i], err = UnmarshalPrice(price_)
			if err != nil {
				code = http.StatusInternalServerError
				return
			}
		}
	}
	return
}

func SetPrice(c rpc.Client, host string, price Price) (code int, err error) {
	price_, err := MarshalPrice(price)
	if err != nil {
		return
	}
	code, err = c.CallWithJson(nil, host+"/set", price_)
	return
}

func UpdatePriceUnit(c rpc.Client, host string, unit PriceUnit) (code int, err error) {
	unit_, err := MarshalPriceUnit(unit)
	if err != nil {
		return
	}
	code, err = c.CallWithJson(nil, host+"/updateunit", unit_)
	return
}

func GetPriceUnit(c rpc.Client, host, unitType string, id string) (unit PriceUnit, code int, err error) {
	url := host + "/getunit"
	url += "?type=" + unitType
	url += "&id=" + id
	var priceUnit EncodingPriceUnit
	code, err = c.Call(&priceUnit, url)
	if err == nil {
		unit, err = UnmarshalPriceUnit(priceUnit)
	}
	return
}

func GetPriceUnitList(c rpc.Client, host, unitType string, offset, limit int) (units []PriceUnit, code int, err error) {
	url := host + "/getunitlist"
	url += "?type=" + unitType
	url += "&offset=" + strconv.Itoa(offset)
	url += "&limit=" + strconv.Itoa(limit)
	var encodingPriceUnitList = make([]EncodingPriceUnit, limit)
	code, err = c.Call(&encodingPriceUnitList, url)
	if err == nil {
		for _, v := range encodingPriceUnitList {
			var unit PriceUnit
			unit, err = UnmarshalPriceUnit(v)
			if err != nil {
				return
			}
			units = append(units, unit)
		}
	}
	return
}

func GetUserIdsFromUnitId(c rpc.Client, host, unitType, id string) (userIds []int, code int, err error) {
	params := neturl.Values{}
	params.Add("type", unitType)
	params.Add("id", id)
	code, err = c.CallWithForm(&userIds, host+"/getuserids", params)
	return
}

//---------------------------------------------------------------------------//

type ServiceIn struct {
	host string
	acc  account.InterfaceEx
}

func NewServiceIn(host string, acc account.InterfaceEx) *ServiceIn {
	return &ServiceIn{host: host, acc: acc}
}

func (r *ServiceIn) getClient(user account.UserInfo) rpc.Client {
	token := r.acc.MakeAccessToken(user)
	return rpc.Client{oauth.NewClient(token, nil)}
}

func (r *ServiceIn) GetPrice(user account.UserInfo, uid uint32, customerGroup adminacc.CustomerGroup, time_ time.Time) (price Price, code int, err error) {
	return GetPrice(r.getClient(user), r.host, uid, customerGroup, time_)
}

func (r *ServiceIn) GetPriceEx(user account.UserInfo, uid uint32, customerGroup adminacc.CustomerGroup) (price Price, code int, err error) {
	return GetPriceEx(r.getClient(user), r.host, uid, customerGroup)
}

func (r *ServiceIn) GetPriceListEx(user account.UserInfo, time_ time.Time, offset, limit int) (prices []Price, code int, err error) {
	return GetPriceListEx(r.getClient(user), r.host, time_, offset, limit)
}

func (r *ServiceIn) SetPrice(user account.UserInfo, price Price) (code int, err error) {
	return SetPrice(r.getClient(user), r.host, price)
}

//---------------------------------------------------------------------------//

type Service struct {
	Conn rpc.Client
}

func New(t http.RoundTripper) Service {
	client := &http.Client{Transport: t}
	return Service{rpc.Client{client}}
}

func (r Service) GetPrice(host string, uid uint32, customerGroup adminacc.CustomerGroup, time_ time.Time) (price Price, code int, err error) {
	return GetPrice(r.Conn, host, uid, customerGroup, time_)
}

func (r Service) GetPriceEx(host string, uid uint32, customerGroup adminacc.CustomerGroup) (price Price, code int, err error) {
	return GetPriceEx(r.Conn, host, uid, customerGroup)
}

func (r Service) GetPriceListEx(host string, time_ time.Time, offset, limit int) (prices []Price, code int, err error) {
	return GetPriceListEx(r.Conn, host, time_, offset, limit)
}

func (r Service) SetPrice(host string, price Price) (code int, err error) {
	return SetPrice(r.Conn, host, price)
}

func (r Service) UpdatePriceUnit(host string, unit PriceUnit) (code int, err error) {
	return UpdatePriceUnit(r.Conn, host, unit)
}

func (r Service) GetPriceUnit(host, unitType, id string) (unit PriceUnit, code int, err error) {
	return GetPriceUnit(r.Conn, host, unitType, id)
}

func (r Service) GetPriceUnitList(host, unitType string, offset, limit int) (unit []PriceUnit, code int, err error) {
	return GetPriceUnitList(r.Conn, host, unitType, offset, limit)
}

func (r Service) GetUserIdsFromUnitId(host, unitType, id string) (userIds []int, code int, err error) {
	return GetUserIdsFromUnitId(r.Conn, host, unitType, id)
}
