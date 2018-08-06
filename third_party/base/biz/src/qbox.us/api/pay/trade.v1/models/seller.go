package models

import (
	"time"
)

type Seller struct {
	Id         int64        `json:"id"`
	Title      string       `json:"title"` //for human
	Name       string       `json:"name"`  //for product model prefix
	Email      string       `json:"email"`
	Callback   string       `json:"callback"`
	UpdateTime time.Time    `json:"update_time"`
	CreateTime time.Time    `json:"create_time"`
	Status     SellerStatus `json:"status"`
}
