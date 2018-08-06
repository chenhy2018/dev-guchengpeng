package wallet

import (
	P "qbox.us/api/pay/price/v3"
)

func (u ModelMonthStatementUnit) ResourceGroupType() int {
	return u.Price.Item.Base.Price.ResourceGroupList.Type
}

func (u ModelMonthStatementUnit) ResourceGroup() interface{} {
	return getResourceGroup(&u.Price.Item.Base.Price.ResourceGroupList, u.ResGrpIdx)
}

func (u ModelMonthStatementUnit) ItemPrice() P.ModelItemBasePrice {
	return getItemPrice(&u.Price.Item.Base.Price, u.ResGrpIdx)
}
