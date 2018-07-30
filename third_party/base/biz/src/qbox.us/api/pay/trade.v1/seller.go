package trade

import (
	"net/url"
	"strconv"

	"qbox.us/api/pay/trade.v1/models"

	"github.com/qiniu/rpc.v1"
)

// 新建商家
func (r HandleBase) SellerNew(logger rpc.Logger, params ReqSellerNew) (seller models.Seller, err error) {
	value := url.Values{}
	value.Add("email", params.Email)
	value.Add("title", params.Title)
	value.Add("name", params.Name)
	value.Add("callback", params.Callback)
	err = r.Client.CallWithForm(logger, &seller, r.Host+"/seller/new", map[string][]string(value))
	return
}

// 根据商家id或者email获取的商家信息
func (r HandleBase) SellerGet(logger rpc.Logger, params ReqSellerGet) (seller models.Seller, err error) {
	value := url.Values{}
	value.Add("id", strconv.FormatInt(int64(params.Id), 10))
	value.Add("email", params.Email)
	err = r.Client.Call(logger, &seller, r.Host+"/seller/get?"+value.Encode())
	return
}

// 商家列表
func (r HandleBase) SellerList(logger rpc.Logger, params ReqSellerList) (sellers []models.Seller, err error) {
	value := url.Values{}
	value.Add("page", strconv.FormatInt(int64(params.Page), 10))
	value.Add("page_size", strconv.FormatInt(int64(params.PageSize), 10))
	if params.Status != nil && params.Status.Valid() {
		value.Add("status", strconv.FormatInt(int64(*params.Status), 10))
	}
	if params.Email != nil {
		value.Add("email", *params.Email)
	}
	if params.ID != nil {
		value.Add("id", strconv.FormatInt(*params.ID, 10))
	}
	if params.UpdateFrom != nil {
		value.Add("update_from", params.UpdateFrom.ToString())
	}
	if params.UpdateTo != nil {
		value.Add("update_to", params.UpdateTo.ToString())
	}

	err = r.Client.Call(logger, &sellers, r.Host+"/seller/list?"+value.Encode())
	return
}

func (r HandleBase) SellerUpdate(logger rpc.Logger, params ReqSellerUpdate) (seller models.Seller, err error) {
	value := url.Values{}
	value.Add("id", strconv.FormatInt(int64(params.Id), 10))
	if params.Title != nil {
		value.Add("title", *params.Title)
	}
	if params.Name != nil {
		value.Add("name", *params.Name)
	}
	if params.Email != nil {
		value.Add("email", *params.Email)
	}
	if params.Callback != nil {
		value.Add("callback", *params.Callback)
	}
	if params.Status != nil {
		value.Add("status", strconv.FormatInt(int64(*params.Status), 10))
	}
	err = r.Client.CallWithForm(logger, &seller, r.Host+"/seller/update", map[string][]string(value))
	return
}
