package models

import (
	"time"
)

type Product struct {
	Id          int64         `json:"id"`
	Name        string        `json:"name"`
	SellerId    int64         `json:"seller_id"`
	Model       string        `json:"model"`
	SPU         string        `json:"spu"`
	Unit        ProductUnit   `json:"unit"`
	Price       float64       `json:"price"`
	ExpiresIn   uint64        `json:"expires_in"`
	Property    string        `json:"property"`
	Description string        `json:"description"`
	UpdateTime  time.Time     `json:"update_time"`
	CreateTime  time.Time     `json:"create_time"`
	StartTime   time.Time     `json:"start_time"`
	EndTime     time.Time     `json:"end_time"`
	Status      ProductStatus `json:"status"`
	Version     int           `json:"version"`
}
