package trade

import (
	"net/url"
	"strconv"
	"strings"

	"qbox.us/api/pay/trade.v1/models"

	"github.com/qiniu/rpc.v1"
)

// 新建商品
func (r HandleBase) ProductNew(logger rpc.Logger, params ReqProductNew) (product models.Product, err error) {
	value := url.Values{}
	value.Add("email", params.Email)
	value.Add("name", params.Name)
	value.Add("model", params.Model)
	value.Add("spu", params.SPU)
	value.Add("unit", strconv.FormatInt(int64(params.Unit), 10))
	if params.Price != nil {
		value.Add("price", strconv.FormatFloat(*params.Price, 'f', -1, 64))
	}
	value.Add("property", params.Property)
	value.Add("description", params.Description)
	value.Add("expires_in", strconv.FormatUint(params.ExpiresIn, 10))
	if params.StartTime != nil {
		value.Add("start_time", params.StartTime.Value())
	}
	if params.EndTime != nil {
		value.Add("end_time", params.EndTime.Value())
	}

	err = r.Client.CallWithForm(logger, &product, r.Host+"/product/new", map[string][]string(value))
	return
}

// 根据id获取商品信息
func (r HandleBase) ProductGet(logger rpc.Logger, params ReqProductGet) (product models.Product, err error) {
	value := url.Values{}
	value.Add("id", strconv.FormatInt(int64(params.Id), 10))
	value.Add("model", params.ProductModel)
	err = r.Client.Call(logger, &product, r.Host+"/product/get?"+value.Encode())
	return
}

// 根据id获取商品信息，多id接口
func (r HandleBase) ProductIds(logger rpc.Logger, params ReqProductIdsGet) (products []models.Product, err error) {
	value := url.Values{}
	var ids []string = make([]string, len(params.Ids))
	for i, id := range params.Ids {
		ids[i] = strconv.FormatInt(id, 10)
	}
	value.Add("ids", strings.Join(ids, ","))
	err = r.Client.Call(logger, &products, r.Host+"/product/ids?"+value.Encode())
	return
}

// 获取指定商家id对应的商家所有商品信息
func (r HandleBase) SellerProduct(logger rpc.Logger, params ReqSellerProduct) (products []models.Product, err error) {
	value := url.Values{}
	value.Add("seller_id", strconv.FormatInt(int64(params.SellerId), 10))
	value.Add("email", params.Email)
	value.Add("page", strconv.FormatInt(int64(params.Page), 10))
	value.Add("page_size", strconv.FormatInt(int64(params.PageSize), 10))
	err = r.Client.Call(logger, &products, r.Host+"/seller/product?"+value.Encode())
	return
}

// 商品列表
func (r HandleBase) ProductList(logger rpc.Logger, params ReqProductList) (products []models.Product, err error) {
	value := url.Values{}
	value.Add("page", strconv.FormatInt(int64(params.Page), 10))
	value.Add("page_size", strconv.FormatInt(int64(params.PageSize), 10))
	if params.ID != nil {
		value.Add("id", strconv.FormatInt(*params.ID, 10))
	}
	if params.Model != nil {
		value.Add("model", *params.Model)
	}
	if params.SPU != nil {
		value.Add("spu", *params.SPU)
	}
	if params.Unit != nil && params.Unit.Valid() {
		value.Add("unit", strconv.Itoa(int(*params.Unit)))
	}
	if params.SellerId != nil {
		value.Add("seller_id", strconv.FormatInt(*params.SellerId, 10))
	}
	if params.Status != nil {
		value.Add("status", strconv.Itoa(int(*params.Status)))
	}
	if params.UpdateFrom != nil {
		value.Add("update_from", params.UpdateFrom.ToString())
	}
	if params.UpdateTo != nil {
		value.Add("update_to", params.UpdateTo.ToString())
	}

	err = r.Client.Call(logger, &products, r.Host+"/product/list?"+value.Encode())
	return
}

// 更新商品信息
func (r HandleBase) ProductUpdate(logger rpc.Logger, params ReqProductUpdate) (product models.Product, err error) {
	value := url.Values{}
	value.Add("id", strconv.FormatInt(int64(params.Id), 10))
	if params.Name != nil {
		value.Add("name", *params.Name)
	}
	if params.Price != nil {
		value.Add("price", strconv.FormatFloat(*params.Price, 'f', -1, 64))
	}
	if params.Model != nil {
		value.Add("model", *params.Model)
	}
	if params.SPU != nil {
		value.Add("spu", *params.SPU)
	}
	if params.Property != nil {
		value.Add("property", *params.Property)
	}
	if params.Unit != nil {
		value.Add("unit", strconv.FormatUint(uint64(*params.Unit), 10))
	}
	if params.Description != nil {
		value.Add("description", *params.Description)
	}
	if params.ExpiresIn != nil {
		value.Add("expires_in", strconv.FormatUint(*params.ExpiresIn, 10))
	}
	if params.Status != nil {
		value.Add("status", strconv.FormatInt(int64(*params.Status), 10))
	}
	if params.StartTime != nil {
		value.Add("start_time", params.StartTime.Value())
	}
	if params.EndTime != nil {
		value.Add("end_time", params.EndTime.Value())
	}

	err = r.Client.CallWithForm(logger, &product, r.Host+"/product/update", map[string][]string(value))
	return
}

// 商品上线
func (r HandleBase) ProductRelease(logger rpc.Logger, params ReqProductRelease) (product models.Product, err error) {
	value := url.Values{}
	value.Add("id", strconv.FormatInt(int64(params.Id), 10))
	err = r.Client.CallWithForm(logger, &product, r.Host+"/product/release", map[string][]string(value))
	return
}

// 商品历史版本
func (r HandleBase) ProductHistory(logger rpc.Logger, params ReqProductHistory) (product []models.Product, err error) {
	value := url.Values{}
	value.Add("id", strconv.FormatInt(params.Id, 10))
	value.Add("page", strconv.FormatInt(int64(params.Page), 10))
	value.Add("page_size", strconv.FormatInt(int64(params.PageSize), 10))
	err = r.Client.CallWithForm(logger, &product, r.Host+"/product/history", map[string][]string(value))
	return
}
