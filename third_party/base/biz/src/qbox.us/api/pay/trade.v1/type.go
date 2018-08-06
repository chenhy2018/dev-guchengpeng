package trade

import (
	"net/url"
	"strconv"
	"strings"
	"time"

	"qbox.us/api/pay/trade.v1/models"
)

type CustomTime struct {
	time.Time
}

func (ct *CustomTime) ToString() string {
	return ct.Time.Format(time.RFC3339)
}

func (ct *CustomTime) UnmarshalJSON(b []byte) (err error) {
	if b[0] == '"' && b[len(b)-1] == '"' {
		b = b[1 : len(b)-1]
	}
	ct.Time, err = time.Parse(time.RFC3339, string(b))
	return
}

func (ct *CustomTime) MarshalJSON() ([]byte, error) {
	return []byte(ct.ToString()), nil
}

func (ct *CustomTime) ParseValue(str string) (err error) {
	ct.Time, err = time.Parse(time.RFC3339, str)
	return err
}

func (ct *CustomTime) Value() string {
	return ct.ToString()
}

type ReqSellerNew struct {
	Email    string `json:"email"`
	Title    string `json:"title"`
	Name     string `json:"name"`
	Callback string `json:"callback"`
}

type ReqSellerGet struct {
	Id    int64  `json:"id"`
	Email string `json:"email"`
}

type ReqSellerList struct {
	Page       int                  `json:"page"`
	PageSize   int                  `json:"page_size"`
	Status     *models.SellerStatus `json:"status"`
	UpdateFrom *CustomTime          `json:"update_from"`
	UpdateTo   *CustomTime          `json:"update_to"`
	ID         *int64               `json:"id"`
	Email      *string              `json:"email"`
}

type ReqSellerUpdate struct {
	Id       int64                `json:"id"`
	Title    *string              `json:"title"`
	Name     *string              `json:"name"`
	Email    *string              `json:"email"`
	Callback *string              `json:"callback"`
	Status   *models.SellerStatus `json:"status"`
}

type ReqProductNew struct {
	Email       string             `json:"email"`
	Name        string             `json:"name"`
	Model       string             `json:"model"`
	SPU         string             `json:"spu"`
	Unit        models.ProductUnit `json:"unit"`
	Price       *float64           `json:"price"`
	Property    string             `json:"property"`
	Description string             `json:"description"`
	ExpiresIn   uint64             `json:"expires_in"`
	StartTime   *CustomTime        `json:"start_time"`
	EndTime     *CustomTime        `json:"end_time"`
}

type ReqProductGet struct {
	Id           int64  `json:"id"`
	ProductModel string `json:"model"`
}

type ReqSellerProduct struct {
	SellerId int64  `json:"seller_id"`
	Email    string `json:"email"`
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
}

type ReqProductList struct {
	ID         *int64                `json:"id"`
	SellerId   *int64                `json:"seller_id"`
	Model      *string               `json:"model"`
	SPU        *string               `json:"spu"`
	Unit       *models.ProductUnit   `json:"unit"`
	Status     *models.ProductStatus `json:"status"`
	UpdateFrom *CustomTime           `json:"update_from"`
	UpdateTo   *CustomTime           `json:"update_to"`
	Page       int                   `json:"page"`
	PageSize   int                   `json:"page_size"`
}

type ReqProductIdsGet struct {
	Ids []int64 `json:"ids"`
}

type ReqProductUpdate struct {
	Id          int64                 `json:"id"` //required
	Name        *string               `json:"name"`
	Price       *float64              `json:"price"`
	Model       *string               `json:"model"`
	SPU         *string               `json:"spu"`
	Property    *string               `json:"property"`
	Unit        *models.ProductUnit   `json:"unit"`
	Description *string               `json:"description"`
	ExpiresIn   *uint64               `json:"expires_in"`
	Status      *models.ProductStatus `json:"status"`
	StartTime   *CustomTime           `json:"start_time"`
	EndTime     *CustomTime           `json:"end_time"`
}

type ReqProductRelease struct {
	Id int64 `json:"id"`
}

type ReqProductHistory struct {
	Id       int64 `json:"id"`
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
}

type ReqProductOrderNew struct {
	ProductId    int64    `json:"product_id"`
	Duration     uint     `json:"duration"`
	TimeDuration uint64   `json:"time_duration"`
	Quantity     uint     `json:"quantity"`
	Property     *string  `json:"property"`
	Fee          *float64 `json:"fee"`
}

type ReqOrderNew struct {
	Orders  []ReqProductOrderNew `json:"orders"`
	BuyerId uint32               `json:"uid"`
	Memo    string               `json:"memo"`
}

type ReqOrderPay struct {
	OrderHash string `json:"order_hash"`
}

type ReqOrderGet struct {
	Id         int64  `json:"id"`
	OrderHash  string `json:"order_hash"`
	WithDetail bool   `json:"with_detail"`
}

type ReqOrderUpdate struct {
	Id          int64              `json:"id"`
	OrderHash   string             `json:"order_hash"`
	ActuallyFee *float64           `json:"actually_fee"`
	Status      models.OrderStatus `json:"status"`
}

type ReqOrderList struct {
	WithDetail    bool                 `json:"with_detail"`
	ID            *int64               `json:"id"`
	OrderHash     *string              `json:"order_hash"`
	SellerId      *int64               `json:"seller_id"`
	BuyerId       *uint32              `json:"uid"`
	Status        []models.OrderStatus `json:"status"`
	CreateFrom    *CustomTime          `json:"create_from"`
	CreateTo      *CustomTime          `json:"create_to"`
	UpdateFrom    *CustomTime          `json:"update_from"`
	UpdateTo      *CustomTime          `json:"update_to"`
	PayFrom       *CustomTime          `json:"pay_from"`
	PayTo         *CustomTime          `json:"pay_to"`
	ExpiredFilter bool                 `json:"expired_filter"`
	Page          int                  `json:"page"`
	PageSize      int                  `json:"page_size"`
}

func (params *ReqOrderList) ToValues() url.Values {
	value := url.Values{}
	value.Add("with_detail", strconv.FormatBool(params.WithDetail))
	value.Add("page", strconv.FormatInt(int64(params.Page), 10))
	value.Add("page_size", strconv.FormatInt(int64(params.PageSize), 10))
	if params.ID != nil {
		value.Add("id", strconv.FormatInt(*params.ID, 10))
	}
	if params.OrderHash != nil {
		value.Add("order_hash", *params.OrderHash)
	}
	if params.SellerId != nil {
		value.Add("seller_id", strconv.FormatInt(*params.SellerId, 10))
	}
	if params.BuyerId != nil {
		value.Add("buyer_id", strconv.FormatUint(uint64(*params.BuyerId), 10))
	}
	if l := len(params.Status); l > 0 {
		ss := make([]string, l)
		for i, s := range params.Status {
			ss[i] = strconv.Itoa(int(s))
		}
		value.Add("status", strings.Join(ss, ","))
	}
	if params.CreateFrom != nil {
		value.Add("create_from", params.CreateFrom.ToString())
	}
	if params.CreateTo != nil {
		value.Add("create_to", params.CreateTo.ToString())
	}
	if params.PayFrom != nil {
		value.Add("pay_from", params.PayFrom.ToString())
	}
	if params.PayTo != nil {
		value.Add("pay_to", params.PayTo.ToString())
	}
	if params.UpdateFrom != nil {
		value.Add("update_from", params.UpdateFrom.ToString())
	}
	if params.UpdateTo != nil {
		value.Add("update_to", params.UpdateTo.ToString())
	}

	if params.ExpiredFilter {
		value.Add("expired_filter", "true")
	}

	return value
}

type ReqUserOrderList struct {
	Uid        uint32 `json:"uid"`
	WithDetail bool   `json:"with_detail"`
	Page       int    `json:"page"`
	PageSize   int    `json:"page_size"`
}

type ReqSellerOrderList struct {
	Email      string `json:"email"`
	SellerId   int64  `json:"seller_id"`
	WithDetail bool   `json:"with_detail"`
	Page       int    `json:"page"`
	PageSize   int    `json:"page_size"`
}

type RespOrderNew struct {
	OrderHash string `json:"order_hash"`
}

type ReqProductOrderAccomplish struct {
	Id        int64       `json:"id"`
	Property  string      `json:"property"`
	StartTime *CustomTime `json:"start_time"`
	Force     bool        `json:"force"`
}

type ReqProductOrderUpgrade struct {
	BuyerId   uint32      `json:"buyer_id"`
	CurrentId int64       `json:"current_id"`
	ProductId int64       `json:"product_id"`
	Quantity  *uint       `json:"quantity"`
	StartTime *CustomTime `json:"start_time"`
	Force     bool        `json:"bool"`
	Memo      string      `json:"memo"`
}

type ReqProductOrderList struct {
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	ID         *int64      `json:"id"`
	ProductId  *int64      `json:"product_id"`
	SellerId   *int64      `json:"seller_id"`
	BuyerId    *uint32     `json:"buyer_id"`
	OrderHash  *string     `json:"order_hash"`
	UpdateFrom *CustomTime `json:"update_from"`
	UpdateTo   *CustomTime `json:"update_to"`
}

type ReqProductOrderRefund struct {
	ProductOrderID int64  `json:"product_order_id"`
	Property       string `json:"property"`
}
