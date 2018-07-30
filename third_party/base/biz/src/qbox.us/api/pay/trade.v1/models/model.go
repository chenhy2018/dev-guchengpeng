package models

import "time"

type SellerStatus int

const (
	SellerStatusEnabled SellerStatus = iota + 1
	SellerStatusDisabled
)

func (s SellerStatus) Valid() bool {
	if s == SellerStatusEnabled || s == SellerStatusDisabled {
		return true
	}

	return false
}

func (s SellerStatus) IsEnable() bool {
	return s == SellerStatusEnabled
}

type OrderStatus int

const (
	OrderStatusUnPay OrderStatus = iota + 1
	OrderStatusPayed
	OrderStatusCancelled
	_
)

func (o OrderStatus) Valid() bool {
	switch o {
	case OrderStatusUnPay, OrderStatusPayed, OrderStatusCancelled:
		return true
	default:
		return false
	}
}

func (o OrderStatus) IsUnPay() bool {
	return o == OrderStatusUnPay
}

func (o OrderStatus) IsPayed() bool {
	return o == OrderStatusPayed
}

func (o OrderStatus) IsCancelled() bool {
	return o == OrderStatusCancelled
}

func (o OrderStatus) Humanize() string {
	switch o {
	case OrderStatusUnPay:
		return "未支付"
	case OrderStatusPayed:
		return "已支付"
	case OrderStatusCancelled:
		return "作废"
	default:
		return "未知订单状态"
	}
}

type OrderType int

const (
	OrderTypeBuy OrderType = iota + 1
	OrderTypeReNew
	OrderTypeUpgrade
	OrderTypeCompensation
	OrderTypeRefund
)

func (s OrderType) Valid() bool {
	switch s {
	case OrderTypeBuy, OrderTypeReNew, OrderTypeUpgrade, OrderTypeCompensation, OrderTypeRefund:
		return true
	default:
		return false
	}
}

func (p OrderType) Humanize() string {
	switch p {
	case OrderTypeBuy:
		return "新购"
	case OrderTypeReNew:
		return "续费"
	case OrderTypeUpgrade:
		return "升级"
	case OrderTypeCompensation:
		return "补偿"
	case OrderTypeRefund:
		return "退款"
	default:
		return "未知订单类型"
	}
}

type ProductUnit int

const (
	Yearly ProductUnit = iota + 1
	Monthly
	Weekly
	Daily
	UnLimited ProductUnit = 99
)

func (pu ProductUnit) Valid() bool {
	switch pu {
	case Yearly, Monthly, Weekly, Daily, UnLimited:
		return true
	default:
		return false
	}
}

func (pu ProductUnit) String() string {
	switch pu {
	case Yearly:
		return "year"
	case Monthly:
		return "month"
	case Weekly:
		return "week"
	case Daily:
		return "day"
	case UnLimited:
		return "unlimited"
	default:
		return "known ProductUnit"
	}
}

func (pu ProductUnit) Humanize() string {
	switch pu {
	case Yearly:
		return "按年"
	case Monthly:
		return "按月"
	case Weekly:
		return "按周"
	case Daily:
		return "按天"
	case UnLimited:
		return "一次性购买"
	default:
		return "未知计费单位"
	}
}

func (pu ProductUnit) AddDuration(baseTime time.Time, duration int) time.Time {
	switch pu {
	case Yearly:
		return baseTime.AddDate(duration, 0, 0)
	case Monthly:
		return baseTime.AddDate(0, duration, 0)
	case Weekly:
		return baseTime.AddDate(0, 0, duration*7)
	case Daily:
		return baseTime.AddDate(0, 0, duration)
	case UnLimited:
		//mysql only support 2038-01-01 00:00:00 for timestamp datetype
		return time.Date(2038, time.January, 1, 0, 0, 0, 0, time.Local)
	default:
		return baseTime
	}
}

type ProductStatus int

const (
	ProductStatusNew ProductStatus = iota + 1
	ProductStatusOnline
	ProductStatusDeprecated
	ProductStatusDeleted
)

func (p ProductStatus) Humanize() string {
	switch p {
	case ProductStatusNew:
		return "新建"
	case ProductStatusOnline:
		return "在线"
	case ProductStatusDeprecated:
		return "已失效"
	case ProductStatusDeleted:
		return "已删除"
	default:
		return "未知产品状态"
	}
}

func (p ProductStatus) Valid() bool {
	switch p {
	case ProductStatusNew, ProductStatusOnline, ProductStatusDeprecated, ProductStatusDeleted:
		return true
	default:
		return false
	}
}

func (p ProductStatus) IsNew() bool {
	return p == ProductStatusNew
}

func (p ProductStatus) IsOnline() bool {
	return p == ProductStatusOnline
}

func (p ProductStatus) IsDeprecated() bool {
	return p == ProductStatusDeprecated
}

func (p ProductStatus) IsDeleted() bool {
	return p == ProductStatusDeleted
}

type ProductOrderStatus int

const (
	ProductOrderStatusNew ProductOrderStatus = iota + 1
	ProductOrderStatusComplete
)
