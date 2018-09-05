package v2

import (
	"github.com/qiniu/rpc.v1"
	"net/url"
	"strconv"
)

type Service struct {
	Host   string
	Client *rpc.Client
}

func NewService(host string, client *rpc.Client) *Service {
	return &Service{host, client}
}

//get the base price by id
func (s *Service) BasepriceGet(l rpc.Logger, modelIn BasePriceId) (model BasePrice, err error) {
	value := url.Values{}
	value.Add("id", modelIn.Id)
	err = s.Client.Call(l, &model, s.Host+"/baseprice/get?"+value.Encode())
	return
}

//list base prices by type, paging by offset, limit
func (s *Service) BasepriceList(l rpc.Logger, modelIn BasePriceListReq) (model []BasePrice, err error) {
	value := url.Values{}
	value.Add("type", modelIn.Type.ToString())
	value.Add("offset", strconv.FormatInt(int64(modelIn.Offset), 10))
	value.Add("limit", strconv.FormatInt(int64(modelIn.Limit), 10))
	err = s.Client.Call(l, &model, s.Host+"/baseprice/list?"+value.Encode())
	return
}

//set a new base price, id could be auto generated if not set
func (s *Service) BasepriceSet(l rpc.Logger, modelIn BasePrice) (model string, err error) {
	err = s.Client.CallWithJson(l, &model, s.Host+"/baseprice/set", modelIn)
	return
}

//count the baseprice users by id, type, t
func (s *Service) BasepriceUsersCount(l rpc.Logger, modelIn PriceUsersCount) (model int, err error) {
	value := url.Values{}
	value.Add("id", modelIn.Id)
	value.Add("type", modelIn.Type)
	value.Add("time", modelIn.Time)
	err = s.Client.Call(l, &model, s.Host+"/baseprice/users/count?"+value.Encode())
	return
}

//list the baseprice users by id, type, t, paging by offset, limit
func (s *Service) BasepriceUsersList(l rpc.Logger, modelIn PriceUsersList) (model []UserListItem, err error) {
	value := url.Values{}
	value.Add("id", modelIn.Id)
	value.Add("type", modelIn.Type)
	value.Add("time", modelIn.Time)
	value.Add("offset", strconv.FormatInt(int64(modelIn.Offset), 10))
	value.Add("limit", strconv.FormatInt(int64(modelIn.Limit), 10))
	err = s.Client.Call(l, &model, s.Host+"/baseprice/users/list?"+value.Encode())
	return
}

//add discount, id could be auto generated if not set
func (s *Service) DiscountAdd(l rpc.Logger, modelIn Discount) (model string, err error) {
	err = s.Client.CallWithJson(l, &model, s.Host+"/discount/add", modelIn)
	return
}

//update discount's description
func (s *Service) DiscountDescUpdate(l rpc.Logger, modelIn RewardDesc) (err error) {
	value := url.Values{}
	value.Add("id", modelIn.Id)
	value.Add("desc", modelIn.Desc)
	err = s.Client.CallWithForm(l, nil, s.Host+"/discount/desc/update", map[string][]string(value))
	return
}

//get discount by id
func (s *Service) DiscountGet(l rpc.Logger, modelIn DiscountId) (model Discount, err error) {
	value := url.Values{}
	value.Add("id", modelIn.Id)
	err = s.Client.Call(l, &model, s.Host+"/discount/get?"+value.Encode())
	return
}

//list discounts by type, paging by offset, limit
func (s *Service) DiscountList(l rpc.Logger, modelIn DiscountListReq) (model []Discount, err error) {
	value := url.Values{}
	value.Add("type", modelIn.Type)
	value.Add("offset", strconv.FormatInt(int64(modelIn.Offset), 10))
	value.Add("limit", strconv.FormatInt(int64(modelIn.Limit), 10))
	err = s.Client.Call(l, &model, s.Host+"/discount/list?"+value.Encode())
	return
}

//update discount's name
func (s *Service) DiscountNameUpdate(l rpc.Logger, modelIn DiscountName) (err error) {
	value := url.Values{}
	value.Add("id", modelIn.Id)
	value.Add("name", modelIn.Name)
	err = s.Client.CallWithForm(l, nil, s.Host+"/discount/name/update", map[string][]string(value))
	return
}

//count the discount users by id, type, t
func (s *Service) DiscountUsersCount(l rpc.Logger, modelIn PriceUsersCount) (model int, err error) {
	value := url.Values{}
	value.Add("id", modelIn.Id)
	value.Add("type", modelIn.Type)
	value.Add("time", modelIn.Time)
	err = s.Client.Call(l, &model, s.Host+"/discount/users/count?"+value.Encode())
	return
}

//list the discount users by id, type, t, paging by offset, limit
func (s *Service) DiscountUsersList(l rpc.Logger, modelIn PriceUsersList) (model []UserListItem, err error) {
	value := url.Values{}
	value.Add("id", modelIn.Id)
	value.Add("type", modelIn.Type)
	value.Add("time", modelIn.Time)
	value.Add("offset", strconv.FormatInt(int64(modelIn.Offset), 10))
	value.Add("limit", strconv.FormatInt(int64(modelIn.Limit), 10))
	err = s.Client.Call(l, &model, s.Host+"/discount/users/list?"+value.Encode())
	return
}

//add reward, id could be auto generated if not set
func (s *Service) RewardAdd(l rpc.Logger, modelIn Reward) (model string, err error) {
	err = s.Client.CallWithJson(l, &model, s.Host+"/reward/add", modelIn)
	return
}

//update reward's description
func (s *Service) RewardDescUpdate(l rpc.Logger, modelIn RewardDesc) (err error) {
	value := url.Values{}
	value.Add("id", modelIn.Id)
	value.Add("desc", modelIn.Desc)
	err = s.Client.CallWithForm(l, nil, s.Host+"/reward/desc/update", map[string][]string(value))
	return
}

//get reward by id
func (s *Service) RewardGet(l rpc.Logger, modelIn RewardId) (model Reward, err error) {
	value := url.Values{}
	value.Add("id", modelIn.Id)
	err = s.Client.Call(l, &model, s.Host+"/reward/get?"+value.Encode())
	return
}

//list reward by type, paging by offset, limit
func (s *Service) RewardList(l rpc.Logger, modelIn RewardListReq) (model []Reward, err error) {
	value := url.Values{}
	value.Add("type", modelIn.Type)
	value.Add("offset", strconv.FormatInt(int64(modelIn.Offset), 10))
	value.Add("limit", strconv.FormatInt(int64(modelIn.Limit), 10))
	err = s.Client.Call(l, &model, s.Host+"/reward/list?"+value.Encode())
	return
}

//update reward's name
func (s *Service) RewardNameUpdate(l rpc.Logger, modelIn RewardName) (err error) {
	value := url.Values{}
	value.Add("id", modelIn.Id)
	value.Add("name", modelIn.Name)
	err = s.Client.CallWithForm(l, nil, s.Host+"/reward/name/update", map[string][]string(value))
	return
}

//count the reward users by id, type, t
func (s *Service) RewardUsersCount(l rpc.Logger, modelIn PriceUsersCount) (model int, err error) {
	value := url.Values{}
	value.Add("id", modelIn.Id)
	value.Add("type", modelIn.Type)
	value.Add("time", modelIn.Time)
	err = s.Client.Call(l, &model, s.Host+"/reward/users/count?"+value.Encode())
	return
}

//list the reward users by id, type, t, paging by offset, limit
func (s *Service) RewardUsersList(l rpc.Logger, modelIn PriceUsersList) (model []UserListItem, err error) {
	value := url.Values{}
	value.Add("id", modelIn.Id)
	value.Add("type", modelIn.Type)
	value.Add("time", modelIn.Time)
	value.Add("offset", strconv.FormatInt(int64(modelIn.Offset), 10))
	value.Add("limit", strconv.FormatInt(int64(modelIn.Limit), 10))
	err = s.Client.Call(l, &model, s.Host+"/reward/users/list?"+value.Encode())
	return
}

func (s *Service) UpdateUsercustomgroup(l rpc.Logger, modelIn UserCustomgroupUpdater) (err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(modelIn.Uid), 10))
	value.Add("customer_group", modelIn.CustomerGroup.ToString())
	value.Add("time", strconv.FormatInt(int64(modelIn.Time), 10))
	err = s.Client.CallWithForm(l, nil, s.Host+"/update/usercustomgroup", map[string][]string(value))
	return
}

//set user base price, id could be auto generated if not set
func (s *Service) UserBasepriceSet(l rpc.Logger, modelIn UserBasePriceSetter) (model string, err error) {
	err = s.Client.CallWithJson(l, &model, s.Host+"/user/baseprice/set", modelIn)
	return
}

//update the baseprice's timerange(effect_time, dead_time) of user
func (s *Service) UserBasepriceTimerangeUpdate(l rpc.Logger, modelIn UserBasePriceTimeRangeUpdater) (err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(modelIn.Uid), 10))
	value.Add("customer_group", modelIn.CustomerGroup.ToString())
	value.Add("id", modelIn.Id)
	value.Add("effect_time", strconv.FormatInt(int64(modelIn.EffectTime), 10))
	value.Add("dead_time", strconv.FormatInt(int64(modelIn.DeadTime), 10))
	value.Add("op_id", modelIn.OpId)
	err = s.Client.CallWithForm(l, nil, s.Host+"/user/baseprice/timerange/update", map[string][]string(value))
	return
}

//bind an available discount to user
func (s *Service) UserDiscountAdd(l rpc.Logger, modelIn UserDiscountSetter) (err error) {
	err = s.Client.CallWithJson(l, nil, s.Host+"/user/discount/add", modelIn)
	return
}

//unbind user from a discount
func (s *Service) UserDiscountDelete(l rpc.Logger, modelIn UserPriceId) (err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(modelIn.Uid), 10))
	value.Add("id", modelIn.Id)
	value.Add("op_id", modelIn.OpId)
	err = s.Client.CallWithForm(l, nil, s.Host+"/user/discount/delete", map[string][]string(value))
	return
}

//update discount's timerange(effect_time, dead_time) of user
func (s *Service) UserDiscountTimerangeUpdate(l rpc.Logger, modelIn UserPriceTimeRange) (err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(modelIn.Uid), 10))
	value.Add("id", modelIn.Id)
	value.Add("effect_time", strconv.FormatInt(int64(modelIn.EffectTime), 10))
	value.Add("dead_time", strconv.FormatInt(int64(modelIn.DeadTime), 10))
	value.Add("op_id", modelIn.OpId)
	err = s.Client.CallWithForm(l, nil, s.Host+"/user/discount/timerange/update", map[string][]string(value))
	return
}

//get user all formated price items
func (s *Service) UserPriceFormatedGet(l rpc.Logger, modelIn UserPriceGetter) (model UserAllPriceFormated, err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(modelIn.Uid), 10))
	value.Add("customer_group", modelIn.CustomerGroup.ToString())
	err = s.Client.Call(l, &model, s.Host+"/user/price/formated/get?"+value.Encode())
	return
}

//get user available formated price items in given time
func (s *Service) UserPriceFormatedTimeGet(l rpc.Logger, modelIn UserPriceWithTimeGetter) (model UserPriceFormated, err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(modelIn.Uid), 10))
	value.Add("customer_group", modelIn.CustomerGroup.ToString())
	value.Add("time", strconv.FormatInt(int64(modelIn.Time), 10))
	err = s.Client.Call(l, &model, s.Host+"/user/price/formated/time/get?"+value.Encode())
	return
}

//get user all price items
func (s *Service) UserPriceGet(l rpc.Logger, modelIn UserPriceGetter) (model UserAllPrice, err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(modelIn.Uid), 10))
	value.Add("customer_group", modelIn.CustomerGroup.ToString())
	err = s.Client.Call(l, &model, s.Host+"/user/price/get?"+value.Encode())
	return
}

//get user price in given time for portal
func (s *Service) UserPricePortalGet(l rpc.Logger, modelIn UserPriceWithTimeGetter) (model UserPricePortal, err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(modelIn.Uid), 10))
	value.Add("customer_group", modelIn.CustomerGroup.ToString())
	value.Add("time", strconv.FormatInt(int64(modelIn.Time), 10))
	err = s.Client.Call(l, &model, s.Host+"/user/price/portal/get?"+value.Encode())
	return
}

//get user all available field price items in given range time
func (s *Service) UserPriceRangetimeGet(l rpc.Logger, modelIn UserPriceWithRangeTimeGetter) (model UserPriceWithTimeRangeByFiled, err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(modelIn.Uid), 10))
	value.Add("customer_group", modelIn.CustomerGroup.ToString())
	value.Add("start", strconv.FormatInt(int64(modelIn.Start), 10))
	value.Add("end", strconv.FormatInt(int64(modelIn.End), 10))
	err = s.Client.Call(l, &model, s.Host+"/user/price/rangetime/get?"+value.Encode())
	return
}

//get user available price items in given time
func (s *Service) UserPriceTimeGet(l rpc.Logger, modelIn UserPriceWithTimeGetter) (model UserPrice, err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(modelIn.Uid), 10))
	value.Add("customer_group", modelIn.CustomerGroup.ToString())
	value.Add("time", strconv.FormatInt(int64(modelIn.Time), 10))
	err = s.Client.Call(l, &model, s.Host+"/user/price/time/get?"+value.Encode())
	return
}

//bind an available reward to user
func (s *Service) UserRewardAdd(l rpc.Logger, modelIn UserRewardSetter) (err error) {
	err = s.Client.CallWithJson(l, nil, s.Host+"/user/reward/add", modelIn)
	return
}

//unbind user from a reward
func (s *Service) UserRewardDelete(l rpc.Logger, modelIn UserPriceId) (err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(modelIn.Uid), 10))
	value.Add("id", modelIn.Id)
	value.Add("op_id", modelIn.OpId)
	err = s.Client.CallWithForm(l, nil, s.Host+"/user/reward/delete", map[string][]string(value))
	return
}

//update reward's timerange(effect_time, dead_time) of user
func (s *Service) UserRewardTimerangeUpdate(l rpc.Logger, modelIn UserPriceTimeRange) (err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(modelIn.Uid), 10))
	value.Add("id", modelIn.Id)
	value.Add("effect_time", strconv.FormatInt(int64(modelIn.EffectTime), 10))
	value.Add("dead_time", strconv.FormatInt(int64(modelIn.DeadTime), 10))
	value.Add("op_id", modelIn.OpId)
	err = s.Client.CallWithForm(l, nil, s.Host+"/user/reward/timerange/update", map[string][]string(value))
	return
}

//bind an available valuetype to user
func (s *Service) UserValuetypeAdd(l rpc.Logger, modelIn UserValueTypeSetter) (err error) {
	err = s.Client.CallWithJson(l, nil, s.Host+"/user/valuetype/add", modelIn)
	return
}

//unbind user from a valuetype
func (s *Service) UserValuetypeDelete(l rpc.Logger, modelIn UserPriceId) (err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(modelIn.Uid), 10))
	value.Add("id", modelIn.Id)
	value.Add("op_id", modelIn.OpId)
	err = s.Client.CallWithForm(l, nil, s.Host+"/user/valuetype/delete", map[string][]string(value))
	return
}

//update valuetype's timerange(effect_time, dead_time) of user
func (s *Service) UserValuetypeTimerangeUpdate(l rpc.Logger, modelIn UserPriceTimeRange) (err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(modelIn.Uid), 10))
	value.Add("id", modelIn.Id)
	value.Add("effect_time", strconv.FormatInt(int64(modelIn.EffectTime), 10))
	value.Add("dead_time", strconv.FormatInt(int64(modelIn.DeadTime), 10))
	value.Add("op_id", modelIn.OpId)
	err = s.Client.CallWithForm(l, nil, s.Host+"/user/valuetype/timerange/update", map[string][]string(value))
	return
}

//add valuetype, id could be auto generated if not set
func (s *Service) ValuetypeAdd(l rpc.Logger, modelIn ValueType) (model string, err error) {
	err = s.Client.CallWithJson(l, &model, s.Host+"/valuetype/add", modelIn)
	return
}

//update valuetype's description
func (s *Service) ValuetypeDescUpdate(l rpc.Logger, modelIn ValueTypeDesc) (err error) {
	value := url.Values{}
	value.Add("id", modelIn.Id)
	value.Add("desc", modelIn.Desc)
	err = s.Client.CallWithForm(l, nil, s.Host+"/valuetype/desc/update", map[string][]string(value))
	return
}

//get valuetype by id
func (s *Service) ValuetypeGet(l rpc.Logger, modelIn ValueTypeId) (model ValueType, err error) {
	value := url.Values{}
	value.Add("id", modelIn.Id)
	err = s.Client.Call(l, &model, s.Host+"/valuetype/get?"+value.Encode())
	return
}

//list the valuetypes by field, type, paging by offset, limit
func (s *Service) ValuetypeList(l rpc.Logger, modelIn ValueTypeListReq) (model []ValueType, err error) {
	value := url.Values{}
	value.Add("field", modelIn.Field.ToString())
	value.Add("type", modelIn.Type.ToString())
	value.Add("offset", strconv.FormatInt(int64(modelIn.Offset), 10))
	value.Add("limit", strconv.FormatInt(int64(modelIn.Limit), 10))
	err = s.Client.Call(l, &model, s.Host+"/valuetype/list?"+value.Encode())
	return
}

//count the valuetype users by id, type, t
func (s *Service) ValuetypeUsersCount(l rpc.Logger, modelIn PriceUsersCount) (model int, err error) {
	value := url.Values{}
	value.Add("id", modelIn.Id)
	value.Add("type", modelIn.Type)
	value.Add("time", modelIn.Time)
	err = s.Client.Call(l, &model, s.Host+"/valuetype/users/count?"+value.Encode())
	return
}

//list the valuetype users by id, type, t, paging by offset, limit
func (s *Service) ValuetypeUsersList(l rpc.Logger, modelIn PriceUsersList) (model []UserListItem, err error) {
	value := url.Values{}
	value.Add("id", modelIn.Id)
	value.Add("type", modelIn.Type)
	value.Add("time", modelIn.Time)
	value.Add("offset", strconv.FormatInt(int64(modelIn.Offset), 10))
	value.Add("limit", strconv.FormatInt(int64(modelIn.Limit), 10))
	err = s.Client.Call(l, &model, s.Host+"/valuetype/users/list?"+value.Encode())
	return
}
