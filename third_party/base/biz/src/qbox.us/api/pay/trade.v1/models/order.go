package models

import "time"

type Order struct {
	Id            int64           `json:"id"`
	OrderHash     string          `json:"order"`
	SellerId      int64           `json:"seller_id"`
	BuyerId       uint32          `json:"buyer_id"`
	Fee           float64         `json:"fee"`
	ActuallyFee   float64         `json:"actually_fee"`
	CFee          float64         `json:"c_fee"`
	Memo          string          `json:"memo"`
	PayTime       time.Time       `json:"pay_time"`
	UpdateTime    time.Time       `json:"update_time"`
	CreateTime    time.Time       `json:"create_time"`
	ExpiredTime   time.Time       `json:"expired_time"`
	Status        OrderStatus     `json:"status"`
	Products      *[]Product      `json:"products,omitempty"`
	ProductOrders *[]ProductOrder `json:"product_orders,omitempty"`
}
