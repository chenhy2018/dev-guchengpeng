package models

import (
	"time"
)

type ProductOrder struct {
	Id              int64              `json:"id"`
	ProductId       int64              `json:"product_id"`
	SellerId        int64              `json:"seller_id"`
	BuyerId         uint32             `json:"buyer_id"`
	OrderId         int64              `json:"order_id"`
	OrderHash       string             `json:"order_hash"`
	OrderType       OrderType          `json:"order_type"`
	ProductOrderId  int64              `json:"product_order_id"`
	ProductVersion  int                `json:"product_version"`
	ProductName     string             `json:"product_name"`
	ProductProperty string             `json:"product_property"`
	Property        string             `json:"property"`
	Duration        uint               `json:"duration"`
	TimeDuration    time.Duration      `json:"time_duration"`
	Quantity        uint               `json:"quantity"`
	ItemFee         float64            `json:"item_fee"`
	Fee             float64            `json:"fee"`
	CFee            float64            `json:"c_fee"`
	UpdateTime      time.Time          `json:"update_time"`
	CreateTime      time.Time          `json:"create_time"`
	StartTime       time.Time          `json:"start_time"`
	EndTime         time.Time          `json:"end_time"`
	ExpiredTime     time.Time          `json:"expired_time"`
	Status          ProductOrderStatus `json:"status"`
	Product         *Product           `json:"product,omitempty"`
}
