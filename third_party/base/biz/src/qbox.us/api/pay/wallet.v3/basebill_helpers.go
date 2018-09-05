package wallet

import (
	P "qbox.us/api/pay/price/v3"
)

func (r ModelBaseBill) ResourceGroup() interface{} {
	return getResourceGroup(&r.Detail.Price.Item.Base.Price.ResourceGroupList, r.ResGrpIdx)
}

// ItemPrice returns the actual resource group price used for this bill
func (r ModelBaseBill) ItemPrice() P.ModelItemBasePrice {
	return getItemPrice(&r.Detail.Price.Item.Base.Price, r.ResGrpIdx)
}
