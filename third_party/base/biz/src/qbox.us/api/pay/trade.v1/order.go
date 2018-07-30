package trade

import (
	"encoding/json"
	"net/url"
	"strconv"

	"github.com/qiniu/rpc.v1"

	"qbox.us/api/pay/trade.v1/models"
)

func (r HandleBase) OrderNew(logger rpc.Logger, params ReqOrderNew) (order string, err error) {
	var (
		value = url.Values{}
		resp  RespOrderNew
	)

	data, err := json.Marshal(params)
	if err != nil {
		return
	}
	value.Add("data", string(data))

	err = r.Client.CallWithForm(logger, &resp, r.Host+"/order/new", map[string][]string(value))
	if err == nil {
		order = resp.OrderHash
	}

	return
}

func (r HandleBase) OrderPay(logger rpc.Logger, params ReqOrderPay) (err error) {
	value := url.Values{}
	value.Add("order_hash", params.OrderHash)
	err = r.Client.CallWithForm(logger, nil, r.Host+"/order/pay", map[string][]string(value))
	return
}

func (r HandleBase) OrderGet(logger rpc.Logger, params ReqOrderGet) (order models.Order, err error) {
	value := url.Values{}
	value.Add("id", strconv.FormatInt(int64(params.Id), 10))
	value.Add("order_hash", params.OrderHash)
	value.Add("with_detail", strconv.FormatBool(params.WithDetail))
	err = r.Client.Call(logger, &order, r.Host+"/order/get?"+value.Encode())
	return
}

func (r HandleBase) OrderUpdate(logger rpc.Logger, params ReqOrderUpdate) (order models.Order, err error) {
	value := url.Values{}
	value.Add("id", strconv.FormatInt(int64(params.Id), 10))
	value.Add("order_hash", params.OrderHash)
	if params.ActuallyFee != nil {
		value.Add("actually_fee", strconv.FormatFloat(*params.ActuallyFee, 'f', -1, 64))
	}
	value.Add("status", strconv.FormatUint(uint64(params.Status), 10))
	err = r.Client.CallWithForm(logger, &order, r.Host+"/order/update", map[string][]string(value))
	return
}

// 订单列表
func (r HandleBase) OrderList(logger rpc.Logger, params ReqOrderList) (orders []models.Order, err error) {
	value := params.ToValues()
	err = r.Client.Call(logger, &orders, r.Host+"/order/list?"+value.Encode())
	return
}

// 用户订单列表
// deprecated: please use HandleBase.OrderList func
func (r HandleBase) UserOrderList(logger rpc.Logger, params ReqOrderList) (orders []models.Order, err error) {
	value := params.ToValues()
	err = r.Client.Call(logger, &orders, r.Host+"/order/list?"+value.Encode())
	return
}

//商家订单列表
// deprecated: please use HandleBase.OrderList func
func (r HandleBase) SellerOrderList(logger rpc.Logger, params ReqOrderList) (orders []models.Order, err error) {
	value := params.ToValues()
	err = r.Client.Call(logger, &orders, r.Host+"/order/list?"+value.Encode())
	return
}
