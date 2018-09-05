package trade

import (
	"net/url"
	"strconv"

	"qbox.us/api/pay/trade.v1/models"

	"github.com/qiniu/rpc.v1"
)

func (r HandleBase) ProductOrderList(logger rpc.Logger, params ReqProductOrderList) (productOrders []models.ProductOrder, err error) {
	value := url.Values{}
	value.Add("page", strconv.FormatInt(int64(params.Page), 10))
	value.Add("page_size", strconv.FormatInt(int64(params.PageSize), 10))
	if params.ID != nil {
		value.Add("id", strconv.FormatInt(*params.ID, 10))
	}
	if params.BuyerId != nil {
		value.Add("buyer_id", strconv.FormatInt(int64(*params.BuyerId), 10))
	}
	if params.SellerId != nil {
		value.Add("seller_id", strconv.FormatInt(*params.SellerId, 10))
	}
	if params.ProductId != nil {
		value.Add("product_id", strconv.FormatInt(*params.ProductId, 10))
	}
	if params.OrderHash != nil {
		value.Add("order_hash", *params.OrderHash)
	}
	if params.UpdateFrom != nil {
		value.Add("update_from", params.UpdateFrom.ToString())
	}
	if params.UpdateTo != nil {
		value.Add("update_to", params.UpdateTo.ToString())
	}

	err = r.Client.CallWithForm(logger, &productOrders, r.Host+"/product/order/list", map[string][]string(value))
	return
}

func (r HandleBase) ProductOrderAccomplish(logger rpc.Logger, params ReqProductOrderAccomplish) (productOrder models.ProductOrder, err error) {
	value := url.Values{}
	value.Add("id", strconv.FormatInt(int64(params.Id), 10))
	value.Add("property", params.Property)
	value.Add("force", strconv.FormatBool(params.Force))
	if params.StartTime != nil {
		value.Add("start_time", params.StartTime.Value())
	}
	err = r.Client.CallWithForm(logger, &productOrder, r.Host+"/product/order/accomplish", map[string][]string(value))
	return
}

func (r HandleBase) ProductOrderUpgrade(logger rpc.Logger, params ReqProductOrderUpgrade) (order string, err error) {
	var (
		value = url.Values{}
		resp  RespOrderNew
	)

	value.Add("buyer_id", strconv.FormatInt(int64(params.BuyerId), 10))
	value.Add("current_id", strconv.FormatInt(params.CurrentId, 10))
	value.Add("product_id", strconv.FormatInt(params.ProductId, 10))
	value.Add("force", strconv.FormatBool(params.Force))
	if params.StartTime != nil {
		value.Add("start_time", params.StartTime.Value())
	}
	if params.Quantity != nil {
		value.Add("quantity", strconv.FormatUint(uint64(*params.Quantity), 10))
	}
	value.Add("memo", params.Memo)

	err = r.Client.CallWithForm(logger, &resp, r.Host+"/product/order/upgrade", map[string][]string(value))
	if err != nil {
		return
	}

	order = resp.OrderHash
	return
}

func (r HandleBase) ProductOrderRefund(logger rpc.Logger, params *ReqProductOrderRefund) error {
	value := url.Values{}
	value.Add("product_order_id", strconv.FormatInt(params.ProductOrderID, 10))
	value.Add("property", params.Property)

	return r.Client.CallWithForm(logger, nil, r.Host+"/product/order/refund", value)
}
