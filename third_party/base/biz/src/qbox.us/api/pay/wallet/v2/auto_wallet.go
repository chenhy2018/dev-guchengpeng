package v2

import (
	"net/url"
	"strconv"

	"github.com/qiniu/rpc.v1"
)

type Service struct {
	Host   string
	Client *rpc.Client
}

func NewService(host string, client *rpc.Client) *Service {
	return &Service{host, client}
}

//激活优惠券
func (s *Service) Active_coupon(l rpc.Logger, modelIn ActiveCouponIn) (model Coupon, err error) {
	value := url.Values{}
	value.Add("excode", modelIn.Excode)
	value.Add("uid", strconv.FormatUint(uint64(modelIn.Uid), 10))
	value.Add("id", modelIn.Id)
	value.Add("desc", modelIn.Desc)
	err = s.Client.CallWithForm(l, &model, s.Host+"/active_coupon", map[string][]string(value))
	return
}

//欠费信息
func (s *Service) ArrearageInfo(l rpc.Logger, modelIn User) (model ArrearageInfo, err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(modelIn.Uid), 10))
	err = s.Client.Call(l, &model, s.Host+"/arrearage/info?"+value.Encode())
	return
}

//删除欠费用户
func (s *Service) ArrearageUserDelete(l rpc.Logger, modelIn User) (err error) {
	err = s.Client.CallWithJson(l, nil, s.Host+"/arrearage/user/delete", modelIn)
	return
}

//列取欠费用户
func (s *Service) ArrearageUserList(l rpc.Logger, modelIn ArrearageUserLister) (model []ArrearageUserInfo, err error) {
	value := url.Values{}
	if modelIn.CustomerGroup != nil {
		value.Add("customergroup", strconv.FormatInt(int64(*modelIn.CustomerGroup), 10))
	}
	if modelIn.From != nil {
		value.Add("from", (*modelIn.From).ToString())
	}
	if modelIn.To != nil {
		value.Add("to", (*modelIn.To).ToString())
	}
	if modelIn.Max != nil {
		value.Add("max", (*modelIn.Max).ToString())
	}
	if modelIn.Min != nil {
		value.Add("min", (*modelIn.Min).ToString())
	}
	if modelIn.Offset != nil {
		value.Add("offset", strconv.FormatInt(int64(*modelIn.Offset), 10))
	}
	if modelIn.Limit != nil {
		value.Add("limit", strconv.FormatInt(int64(*modelIn.Limit), 10))
	}
	err = s.Client.Call(l, &model, s.Host+"/arrearage/user/list?"+value.Encode())
	return
}

//新增/更新欠费用户
func (s *Service) ArrearageUserUpsert(l rpc.Logger, modelIn ArrearageUserInfo) (model bool, err error) {
	err = s.Client.CallWithJson(l, &model, s.Host+"/arrearage/user/upsert", modelIn)
	return
}

//月清单对应的帐单列表
func (s *Service) BaseBillsForMonth(l rpc.Logger, modelIn BaseBillsForMonth) (model MonthStatementBills, err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(modelIn.Uid), 10))
	value.Add("month", modelIn.Month)
	err = s.Client.Call(l, &model, s.Host+"/base/bills/for/month?"+value.Encode())
	return
}

//basebill 销帐
func (s *Service) BasebillDiscard(l rpc.Logger, modelIn BaseBillDiscard) (err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(modelIn.Uid), 10))
	value.Add("day", modelIn.Day)
	err = s.Client.CallWithForm(l, nil, s.Host+"/basebill/discard", map[string][]string(value))
	return
}

//实时价格计费接口：获取账单(详单)
func (s *Service) BasebillGet(l rpc.Logger, modelIn BaseBillGetter) (model BaseBillGet, err error) {
	value := url.Values{}
	value.Add("id", modelIn.Id)
	err = s.Client.Call(l, &model, s.Host+"/basebill/get?"+value.Encode())
	return
}

//实时价格计费接口：获取某业务指定时间之前（包含）的最后一份账单(详单)
func (s *Service) BasebillLastInMonthGet(l rpc.Logger, modelIn BaseBillLastGetter) (model BaseBillGet, err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(modelIn.Uid), 10))
	value.Add("field", modelIn.Field.ToString())
	if modelIn.Date != nil {
		value.Add("date", *modelIn.Date)
	}
	err = s.Client.Call(l, &model, s.Host+"/basebill/last/in/month/get?"+value.Encode())
	return
}

func (s *Service) BillDiscountGet(l rpc.Logger, modelIn BillDiscountGetIn) (model DiscountTransaction, err error) {
	value := url.Values{}
	value.Add("serial_num", modelIn.SerialNum)
	err = s.Client.Call(l, &model, s.Host+"/bill/discount/get?"+value.Encode())
	return
}

func (s *Service) BillDiscountList(l rpc.Logger, modelIn BillDiscountListIn) (model []DiscountTransaction, err error) {
	value := url.Values{}
	if modelIn.RewardId != nil {
		value.Add("reward_id", *modelIn.RewardId)
	}
	if modelIn.PartnerName != nil {
		value.Add("partner_name", *modelIn.PartnerName)
	}
	if modelIn.From != nil {
		value.Add("from", *modelIn.From)
	}
	if modelIn.To != nil {
		value.Add("to", *modelIn.To)
	}
	if modelIn.Offset != nil {
		value.Add("offset", strconv.FormatInt(int64(*modelIn.Offset), 10))
	}
	if modelIn.Limit != nil {
		value.Add("limit", strconv.FormatInt(int64(*modelIn.Limit), 10))
	}
	err = s.Client.Call(l, &model, s.Host+"/bill/discount/list?"+value.Encode())
	return
}

func (s *Service) BillDiscountSum(l rpc.Logger, modelIn BillDiscountSumIn) (model DiscountTransactionSummary, err error) {
	value := url.Values{}
	value.Add("reward_id", modelIn.RewardId)
	value.Add("partner_name", modelIn.PartnerName)
	value.Add("from", modelIn.From)
	value.Add("to", modelIn.To)
	err = s.Client.Call(l, &model, s.Host+"/bill/discount/sum?"+value.Encode())
	return
}

//获取流水
func (s *Service) BillGet(l rpc.Logger, modelIn BillGetIn) (model TransactionOut, err error) {
	value := url.Values{}
	if modelIn.Uid != nil {
		value.Add("uid", strconv.FormatUint(uint64(*modelIn.Uid), 10))
	}
	value.Add("excode", modelIn.Excode)
	value.Add("prefix", modelIn.Prefix)
	value.Add("type", modelIn.Type)
	err = s.Client.Call(l, &model, s.Host+"/bill/get?"+value.Encode())
	return
}

func (s *Service) BillGetbysn(l rpc.Logger, modelIn BillGetbysn) (model TransactionOut, err error) {
	value := url.Values{}
	value.Add("serial_num", modelIn.SerialNum)
	err = s.Client.Call(l, &model, s.Host+"/bill/getbysn?"+value.Encode())
	return
}

//获取流水列表
func (s *Service) BillList(l rpc.Logger, modelIn Get_billsIn) (model []TransactionOut, err error) {
	value := url.Values{}
	if modelIn.Uid != nil {
		value.Add("uid", strconv.FormatUint(uint64(*modelIn.Uid), 10))
	}
	value.Add("starttime", modelIn.StartTime.ToString())
	value.Add("endtime", modelIn.EndTime.ToString())
	value.Add("prefix", modelIn.Prefix)
	value.Add("type", modelIn.Type)
	value.Add("expenses", modelIn.Expenses)
	if modelIn.IsProcessed != nil {
		value.Add("isprocessed", *modelIn.IsProcessed)
	}
	if modelIn.IsHide != nil {
		value.Add("ishide", strconv.FormatBool(*modelIn.IsHide))
	}
	if modelIn.Offset != nil {
		value.Add("offset", strconv.FormatInt(int64(*modelIn.Offset), 10))
	}
	if modelIn.Limit != nil {
		value.Add("limit", strconv.FormatInt(int64(*modelIn.Limit), 10))
	}
	err = s.Client.Call(l, &model, s.Host+"/bill/list?"+value.Encode())
	return
}

func (s *Service) CashList(l rpc.Logger, modelIn CashListIn) (model []CashListOut, err error) {
	value := url.Values{}
	if modelIn.Offset != nil {
		value.Add("offset", strconv.FormatInt(int64(*modelIn.Offset), 10))
	}
	if modelIn.Limit != nil {
		value.Add("limit", strconv.FormatInt(int64(*modelIn.Limit), 10))
	}
	err = s.Client.Call(l, &model, s.Host+"/cash/list?"+value.Encode())
	return
}

//激活优惠券
func (s *Service) CouponActive(l rpc.Logger, modelIn CouponActiveIn) (model Coupon, err error) {
	value := url.Values{}
	value.Add("excode", modelIn.Excode)
	value.Add("uid", strconv.FormatUint(uint64(modelIn.Uid), 10))
	value.Add("id", modelIn.Id)
	value.Add("desc", modelIn.Desc)
	err = s.Client.CallWithForm(l, &model, s.Host+"/coupon/active", map[string][]string(value))
	return
}

func (s *Service) CouponAdminCount(l rpc.Logger, modelIn CouponAdminListIn) (model int, err error) {
	value := url.Values{}
	if modelIn.Offset != nil {
		value.Add("offset", strconv.FormatInt(int64(*modelIn.Offset), 10))
	}
	if modelIn.Limit != nil {
		value.Add("limit", strconv.FormatInt(int64(*modelIn.Limit), 10))
	}
	value.Add("uid", strconv.FormatInt(int64(modelIn.Uid), 10))
	value.Add("type", modelIn.Type)
	value.Add("status", strconv.FormatInt(int64(modelIn.Status), 10))
	if modelIn.Title != nil {
		value.Add("title", *modelIn.Title)
	}
	err = s.Client.Call(l, &model, s.Host+"/coupon/admin/count?"+value.Encode())
	return
}

func (s *Service) CouponAdminList(l rpc.Logger, modelIn CouponAdminListIn) (model []Coupon, err error) {
	value := url.Values{}
	if modelIn.Offset != nil {
		value.Add("offset", strconv.FormatInt(int64(*modelIn.Offset), 10))
	}
	if modelIn.Limit != nil {
		value.Add("limit", strconv.FormatInt(int64(*modelIn.Limit), 10))
	}
	value.Add("uid", strconv.FormatInt(int64(modelIn.Uid), 10))
	value.Add("type", modelIn.Type)
	value.Add("status", strconv.FormatInt(int64(modelIn.Status), 10))
	if modelIn.Title != nil {
		value.Add("title", *modelIn.Title)
	}
	err = s.Client.Call(l, &model, s.Host+"/coupon/admin/list?"+value.Encode())
	return
}

func (s *Service) CouponGet(l rpc.Logger, modelIn CouponGetIn) (model Coupon, err error) {
	value := url.Values{}
	value.Add("id", modelIn.Id)
	err = s.Client.Call(l, &model, s.Host+"/coupon/get?"+value.Encode())
	return
}

func (s *Service) CouponHistory(l rpc.Logger, modelIn CouponHistory) (model []Coupon, err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(modelIn.Uid), 10))
	value.Add("offset", strconv.FormatInt(int64(modelIn.Offset), 10))
	value.Add("limit", strconv.FormatInt(int64(modelIn.Limit), 10))
	err = s.Client.Call(l, &model, s.Host+"/coupon/history?"+value.Encode())
	return
}

func (s *Service) CouponList(l rpc.Logger, modelIn User) (model []Coupon, err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(modelIn.Uid), 10))
	err = s.Client.Call(l, &model, s.Host+"/coupon/list?"+value.Encode())
	return
}

func (s *Service) CouponNew(l rpc.Logger, modelIn CouponNewIn) (model string, err error) {
	value := url.Values{}
	value.Add("quota", modelIn.Quota.ToString())
	value.Add("day", strconv.FormatInt(int64(modelIn.Day), 10))
	value.Add("deadtime", strconv.FormatInt(int64(modelIn.DeadTime), 10))
	value.Add("type", modelIn.Type)
	value.Add("desc", modelIn.Desc)
	value.Add("title", modelIn.Title)
	err = s.Client.CallWithForm(l, &model, s.Host+"/coupon/new", map[string][]string(value))
	return
}

func (s *Service) CustomerRewardAdd(l rpc.Logger, modelIn CustomerReward) (model string, err error) {
	err = s.Client.CallWithJson(l, &model, s.Host+"/customer/reward/add", modelIn)
	return
}

func (s *Service) CustomerRewardAvailable(l rpc.Logger, modelIn CustomerRewardAvailableIn) (err error) {
	value := url.Values{}
	value.Add("id", modelIn.Id)
	value.Add("available", strconv.FormatBool(modelIn.Available))
	err = s.Client.CallWithForm(l, nil, s.Host+"/customer/reward/available", map[string][]string(value))
	return
}

func (s *Service) CustomerRewardCalculate(l rpc.Logger, modelIn CustomerRewardCalculateIn) (model int64, err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(modelIn.Uid), 10))
	value.Add("id", modelIn.Id)
	value.Add("cost", modelIn.Cost.ToString())
	err = s.Client.CallWithForm(l, &model, s.Host+"/customer/reward/calculate", map[string][]string(value))
	return
}

func (s *Service) CustomerRewardGet(l rpc.Logger, modelIn CustomerRewardGetIn) (model CustomerReward, err error) {
	value := url.Values{}
	value.Add("id", modelIn.Id)
	err = s.Client.Call(l, &model, s.Host+"/customer/reward/get?"+value.Encode())
	return
}

func (s *Service) CustomerRewardList(l rpc.Logger, modelIn CustomerRewardListIn) (model []CustomerReward, err error) {
	value := url.Values{}
	if modelIn.Name != nil {
		value.Add("name", *modelIn.Name)
	}
	if modelIn.Offset != nil {
		value.Add("offset", strconv.FormatInt(int64(*modelIn.Offset), 10))
	}
	if modelIn.Limit != nil {
		value.Add("limit", strconv.FormatInt(int64(*modelIn.Limit), 10))
	}
	err = s.Client.Call(l, &model, s.Host+"/customer/reward/list?"+value.Encode())
	return
}

func (s *Service) CustomerRewardUpdatedesc(l rpc.Logger, modelIn CustomerRewardUpdatedescIn) (err error) {
	value := url.Values{}
	value.Add("id", modelIn.Id)
	value.Add("desc", modelIn.Desc)
	err = s.Client.CallWithForm(l, nil, s.Host+"/customer/reward/updatedesc", map[string][]string(value))
	return
}

//新建日账单
func (s *Service) Day_billAdd(l rpc.Logger, modelIn DailyBillSetter) (err error) {
	err = s.Client.CallWithJson(l, nil, s.Host+"/day_bill/add", modelIn)
	return
}

//删除日账单
func (s *Service) Day_billDel(l rpc.Logger, modelIn DailyBillGetter) (err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(modelIn.Uid), 10))
	value.Add("day", modelIn.Day)
	err = s.Client.CallWithForm(l, nil, s.Host+"/day_bill/del", map[string][]string(value))
	return
}

//日帐单撤销
func (s *Service) Day_billDiscard(l rpc.Logger, modelIn DailyBillDiscard) (err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(modelIn.Uid), 10))
	value.Add("date", modelIn.Date)
	err = s.Client.CallWithForm(l, nil, s.Host+"/day_bill/discard", map[string][]string(value))
	return
}

//获取日账单
func (s *Service) Day_billGet(l rpc.Logger, modelIn DailyBillGetter) (model DailyBill, err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(modelIn.Uid), 10))
	value.Add("day", modelIn.Day)
	err = s.Client.Call(l, &model, s.Host+"/day_bill/get?"+value.Encode())
	return
}

//获取日帐单列表
func (s *Service) Day_billList(l rpc.Logger, modelIn DailyBillGetterRange) (model []DailyBill, err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(modelIn.Uid), 10))
	value.Add("from", modelIn.From)
	value.Add("to", modelIn.To)
	err = s.Client.Call(l, &model, s.Host+"/day_bill/list?"+value.Encode())
	return
}

//扣费
func (s *Service) Deduct(l rpc.Logger, modelIn DeductIn) (err error) {
	err = s.Client.CallWithJson(l, nil, s.Host+"/deduct", modelIn)
	return
}

//实时价格计费接口：扣费(根据BaseBill)
func (s *Service) DeductByBasebill(l rpc.Logger, modelIn BaseBillSet) (model string, err error) {
	err = s.Client.CallWithJson(l, &model, s.Host+"/deduct/by/basebill", modelIn)
	return
}

//扣现金
func (s *Service) Deduct_cash(l rpc.Logger, modelIn DeductCashIn) (err error) {
	value := url.Values{}
	value.Add("excode", modelIn.Excode)
	value.Add("type", modelIn.Type)
	value.Add("uid", strconv.FormatUint(uint64(modelIn.Uid), 10))
	value.Add("money", modelIn.Money.ToString())
	if modelIn.At != nil {
		value.Add("at", (*modelIn.At).ToString())
	}
	value.Add("desc", modelIn.Desc)
	value.Add("details", modelIn.Details)
	err = s.Client.CallWithForm(l, nil, s.Host+"/deduct_cash", map[string][]string(value))
	return
}

//删除冻结用户信息
func (s *Service) FreezeInfoDelete(l rpc.Logger, modelIn User) (err error) {
	err = s.Client.CallWithJson(l, nil, s.Host+"/freeze/info/delete", modelIn)
	return
}

//获取冻结用户信息
func (s *Service) FreezeInfoGet(l rpc.Logger, modelIn User) (model FreezeUserInfo, err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(modelIn.Uid), 10))
	err = s.Client.Call(l, &model, s.Host+"/freeze/info/get?"+value.Encode())
	return
}

//列取冻结用户信息
func (s *Service) FreezeInfoList(l rpc.Logger, modelIn FreezeUserLister) (model []FreezeUserInfo, err error) {
	value := url.Values{}
	value.Add("step", modelIn.Step.ToString())
	value.Add("offset", strconv.FormatInt(int64(modelIn.Offset), 10))
	value.Add("limit", strconv.FormatInt(int64(modelIn.Limit), 10))
	err = s.Client.Call(l, &model, s.Host+"/freeze/info/list?"+value.Encode())
	return
}

//设置冻结用户
func (s *Service) FreezeInfoSet(l rpc.Logger, modelIn FreezeUserInfo) (err error) {
	err = s.Client.CallWithJson(l, nil, s.Host+"/freeze/info/set", modelIn)
	return
}

//获取流水列表
func (s *Service) Get_bills(l rpc.Logger, modelIn Get_billsIn) (model []TransactionOut, err error) {
	value := url.Values{}
	if modelIn.Uid != nil {
		value.Add("uid", strconv.FormatUint(uint64(*modelIn.Uid), 10))
	}
	value.Add("starttime", modelIn.StartTime.ToString())
	value.Add("endtime", modelIn.EndTime.ToString())
	value.Add("prefix", modelIn.Prefix)
	value.Add("type", modelIn.Type)
	value.Add("expenses", modelIn.Expenses)
	if modelIn.IsProcessed != nil {
		value.Add("isprocessed", *modelIn.IsProcessed)
	}
	if modelIn.IsHide != nil {
		value.Add("ishide", strconv.FormatBool(*modelIn.IsHide))
	}
	if modelIn.Offset != nil {
		value.Add("offset", strconv.FormatInt(int64(*modelIn.Offset), 10))
	}
	if modelIn.Limit != nil {
		value.Add("limit", strconv.FormatInt(int64(*modelIn.Limit), 10))
	}
	err = s.Client.Call(l, &model, s.Host+"/get_bills?"+value.Encode())
	return
}

//获取优惠券
func (s *Service) Get_coupons(l rpc.Logger, modelIn User) (model []Coupon, err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(modelIn.Uid), 10))
	err = s.Client.Call(l, &model, s.Host+"/get_coupons?"+value.Encode())
	return
}

//隐藏流水
func (s *Service) HideBill(l rpc.Logger, modelIn HideBillIn) (err error) {
	value := url.Values{}
	value.Add("id", modelIn.Id)
	value.Add("ishide", strconv.FormatBool(modelIn.IsHide))
	err = s.Client.CallWithForm(l, nil, s.Host+"/hide/bill", map[string][]string(value))
	return
}

//获取用户信息
func (s *Service) Info(l rpc.Logger, modelIn User) (model InfoOut, err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(modelIn.Uid), 10))
	err = s.Client.Call(l, &model, s.Host+"/info?"+value.Encode())
	return
}

//获取批量用户信息
func (s *Service) InfoBatch(l rpc.Logger, modelIn Users) (model InfoOuts, err error) {
	value := url.Values{}
	value.Add("uids", modelIn.Uids)
	err = s.Client.Call(l, &model, s.Host+"/info/batch?"+value.Encode())
	return
}

//新建月账单
func (s *Service) Month_billAdd(l rpc.Logger, modelIn MonthBillSetter) (err error) {
	err = s.Client.CallWithJson(l, nil, s.Host+"/month_bill/add", modelIn)
	return
}

//月帐单撤销
func (s *Service) Month_billDiscard(l rpc.Logger, modelIn MonthBillDiscard) (model MonthBill, err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(modelIn.Uid), 10))
	value.Add("month", modelIn.Month)
	err = s.Client.CallWithForm(l, &model, s.Host+"/month_bill/discard", map[string][]string(value))
	return
}

//admin获取格式化的账单列表
func (s *Service) Month_billFormatedList(l rpc.Logger, modelIn MonthBillListerForAdmin) (model []MonthBillFormated, err error) {
	value := url.Values{}
	if modelIn.Uid != nil {
		value.Add("uid", strconv.FormatInt(int64(*modelIn.Uid), 10))
	}
	if modelIn.Month != nil {
		value.Add("month", *modelIn.Month)
	}
	if modelIn.Status != nil {
		value.Add("status", (*modelIn.Status).ToString())
	}
	if modelIn.Offset != nil {
		value.Add("offset", strconv.FormatInt(int64(*modelIn.Offset), 10))
	}
	if modelIn.Limit != nil {
		value.Add("limit", strconv.FormatInt(int64(*modelIn.Limit), 10))
	}
	err = s.Client.Call(l, &model, s.Host+"/month_bill/formated/list?"+value.Encode())
	return
}

//获取格式化的账单列表
func (s *Service) Month_billFormatedRangelist(l rpc.Logger, modelIn MonthBillLister) (model []MonthBillFormated, err error) {
	value := url.Values{}
	if modelIn.Uid != nil {
		value.Add("uid", strconv.FormatInt(int64(*modelIn.Uid), 10))
	}
	if modelIn.From != nil {
		value.Add("from", *modelIn.From)
	}
	if modelIn.To != nil {
		value.Add("to", *modelIn.To)
	}
	if modelIn.Offset != nil {
		value.Add("offset", strconv.FormatInt(int64(*modelIn.Offset), 10))
	}
	if modelIn.Limit != nil {
		value.Add("limit", strconv.FormatInt(int64(*modelIn.Limit), 10))
	}
	err = s.Client.Call(l, &model, s.Host+"/month_bill/formated/rangelist?"+value.Encode())
	return
}

//获取月账单
func (s *Service) Month_billGet(l rpc.Logger, modelIn MonthBillGetter) (model MonthBill, err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(modelIn.Uid), 10))
	value.Add("month", modelIn.Month)
	value.Add("id", modelIn.Id)
	err = s.Client.Call(l, &model, s.Host+"/month_bill/get?"+value.Encode())
	return
}

//获取格式化月账单
func (s *Service) Month_billGetformated(l rpc.Logger, modelIn MonthBillGetter) (model MonthBillFormated, err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(modelIn.Uid), 10))
	value.Add("month", modelIn.Month)
	value.Add("id", modelIn.Id)
	err = s.Client.Call(l, &model, s.Host+"/month_bill/getformated?"+value.Encode())
	return
}

//一次获取多个用户的格式化账单列表
func (s *Service) Month_billGetformatedList(l rpc.Logger, modelIn MonthBillGetterList) (model []MonthBillFormated, err error) {
	value := url.Values{}
	value.Add("uids", modelIn.Uids)
	value.Add("month", modelIn.Month)
	err = s.Client.Call(l, &model, s.Host+"/month_bill/getformated/list?"+value.Encode())
	return
}

//admin获取月账单列表
func (s *Service) Month_billList(l rpc.Logger, modelIn MonthBillListerForAdmin) (model []MonthBill, err error) {
	value := url.Values{}
	if modelIn.Uid != nil {
		value.Add("uid", strconv.FormatInt(int64(*modelIn.Uid), 10))
	}
	if modelIn.Month != nil {
		value.Add("month", *modelIn.Month)
	}
	if modelIn.Status != nil {
		value.Add("status", (*modelIn.Status).ToString())
	}
	if modelIn.Offset != nil {
		value.Add("offset", strconv.FormatInt(int64(*modelIn.Offset), 10))
	}
	if modelIn.Limit != nil {
		value.Add("limit", strconv.FormatInt(int64(*modelIn.Limit), 10))
	}
	err = s.Client.Call(l, &model, s.Host+"/month_bill/list?"+value.Encode())
	return
}

//get month bill month list
func (s *Service) Month_billMonths(l rpc.Logger, modelIn Month_billMonthsIn) (model []int64, err error) {
	value := url.Values{}
	if modelIn.Uid != nil {
		value.Add("uid", strconv.FormatUint(uint64(*modelIn.Uid), 10))
	}
	value.Add("lastmonth", modelIn.LastMonth)
	value.Add("limit", strconv.FormatInt(int64(modelIn.Limit), 10))
	err = s.Client.Call(l, &model, s.Host+"/month_bill/months?"+value.Encode())
	return
}

//获取月账单列表
func (s *Service) Month_billRangelist(l rpc.Logger, modelIn MonthBillLister) (model []MonthBill, err error) {
	value := url.Values{}
	if modelIn.Uid != nil {
		value.Add("uid", strconv.FormatInt(int64(*modelIn.Uid), 10))
	}
	if modelIn.From != nil {
		value.Add("from", *modelIn.From)
	}
	if modelIn.To != nil {
		value.Add("to", *modelIn.To)
	}
	if modelIn.Offset != nil {
		value.Add("offset", strconv.FormatInt(int64(*modelIn.Offset), 10))
	}
	if modelIn.Limit != nil {
		value.Add("limit", strconv.FormatInt(int64(*modelIn.Limit), 10))
	}
	err = s.Client.Call(l, &model, s.Host+"/month_bill/rangelist?"+value.Encode())
	return
}

//更新月账单状态
func (s *Service) Month_billStatusUpdate(l rpc.Logger, modelIn MonthBillStatusSetter) (err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(modelIn.Uid), 10))
	value.Add("month", modelIn.Month)
	value.Add("id", modelIn.Id)
	value.Add("status", modelIn.Status.ToString())
	err = s.Client.CallWithForm(l, nil, s.Host+"/month_bill/status/update", map[string][]string(value))
	return
}

//更新月账单
func (s *Service) Month_billUpdate(l rpc.Logger, modelIn MonthBillSetter) (err error) {
	err = s.Client.CallWithJson(l, nil, s.Host+"/month_bill/update", modelIn)
	return
}

//get month snapshot
func (s *Service) Month_snapshotGet(l rpc.Logger, modelIn Month_snapshotIn) (model Snapshot, err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(modelIn.Uid), 10))
	value.Add("month", modelIn.Month)
	err = s.Client.Call(l, &model, s.Host+"/month_snapshot/get?"+value.Encode())
	return
}

//set month snapshot
func (s *Service) Month_snapshotSet(l rpc.Logger, modelIn Snapshot) (err error) {
	err = s.Client.CallWithJson(l, nil, s.Host+"/month_snapshot/set", modelIn)
	return
}

//月账单与月对账单的月份列表
func (s *Service) MonthbillMonthstatementMonths(l rpc.Logger, modelIn MonthBillMonthStatementMonthsIn) (model MonthBillMonthStatementMonthsOut, err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(modelIn.Uid), 10))
	value.Add("lastmonth", modelIn.LastMonth)
	value.Add("limit", strconv.FormatInt(int64(modelIn.Limit), 10))
	err = s.Client.Call(l, &model, s.Host+"/monthbill/monthstatement/months?"+value.Encode())
	return
}

//实时价格计费接口：获取月对账单
func (s *Service) MonthstatementGet(l rpc.Logger, modelIn MonthStatementGetter) (model MonthStatement, err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(modelIn.Uid), 10))
	value.Add("month", modelIn.Month)
	value.Add("id", modelIn.Id)
	err = s.Client.Call(l, &model, s.Host+"/monthstatement/get?"+value.Encode())
	return
}

//实时价格计费接口：获取批量用户月对账单
func (s *Service) MonthstatementList(l rpc.Logger, modelIn MonthStatementGetterList) (model []MonthStatement, err error) {
	value := url.Values{}
	value.Add("uids", modelIn.Uids)
	value.Add("month", modelIn.Month)
	err = s.Client.Call(l, &model, s.Host+"/monthstatement/list?"+value.Encode())
	return
}

//实时价格计费接口：月对账单月份列表
func (s *Service) MonthstatementMonths(l rpc.Logger, modelIn MonthStatementMonthsIn) (model []int64, err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(modelIn.Uid), 10))
	value.Add("lastmonth", modelIn.LastMonth)
	value.Add("limit", strconv.FormatInt(int64(modelIn.Limit), 10))
	err = s.Client.Call(l, &model, s.Host+"/monthstatement/months?"+value.Encode())
	return
}

//实时价格计费接口：月对账单列表
func (s *Service) MonthstatementRangeList(l rpc.Logger, modelIn MonthStatementLister) (model []MonthStatement, err error) {
	value := url.Values{}
	if modelIn.Uid != nil {
		value.Add("uid", strconv.FormatUint(uint64(*modelIn.Uid), 10))
	}
	if modelIn.From != nil {
		value.Add("from", *modelIn.From)
	}
	if modelIn.To != nil {
		value.Add("to", *modelIn.To)
	}
	if modelIn.Status != nil {
		value.Add("status", (*modelIn.Status).ToString())
	}
	if modelIn.Offset != nil {
		value.Add("offset", strconv.FormatInt(int64(*modelIn.Offset), 10))
	}
	if modelIn.Limit != nil {
		value.Add("limit", strconv.FormatInt(int64(*modelIn.Limit), 10))
	}
	err = s.Client.Call(l, &model, s.Host+"/monthstatement/range/list?"+value.Encode())
	return
}

//实时价格计费接口：设置月对账单
func (s *Service) MonthstatementSet(l rpc.Logger, modelIn MonthStatement) (err error) {
	err = s.Client.CallWithJson(l, nil, s.Host+"/monthstatement/set", modelIn)
	return
}

//新建用户
func (s *Service) Newuser(l rpc.Logger, modelIn NewUserModel) (err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(modelIn.Uid), 10))
	value.Add("excode", modelIn.Excode)
	value.Add("desc", modelIn.Desc)
	err = s.Client.CallWithForm(l, nil, s.Host+"/newuser", map[string][]string(value))
	return
}

func (s *Service) PartnerAdd(l rpc.Logger, modelIn PartnerAddIn) (err error) {
	err = s.Client.CallWithJson(l, nil, s.Host+"/partner/add", modelIn)
	return
}

func (s *Service) PartnerGet(l rpc.Logger, modelIn PartnerGetIn) (model PartnerListOut, err error) {
	value := url.Values{}
	value.Add("name", modelIn.Name)
	err = s.Client.Call(l, &model, s.Host+"/partner/get?"+value.Encode())
	return
}

func (s *Service) PartnerList(l rpc.Logger, modelIn PartnerListIn) (model []PartnerListOut, err error) {
	value := url.Values{}
	if modelIn.Type != nil {
		value.Add("type", *modelIn.Type)
	}
	if modelIn.Offset != nil {
		value.Add("offset", strconv.FormatInt(int64(*modelIn.Offset), 10))
	}
	if modelIn.Limit != nil {
		value.Add("limit", strconv.FormatInt(int64(*modelIn.Limit), 10))
	}
	err = s.Client.Call(l, &model, s.Host+"/partner/list?"+value.Encode())
	return
}

func (s *Service) PartnerRewardAdd(l rpc.Logger, modelIn PartnerRewardAddIn) (err error) {
	err = s.Client.CallWithJson(l, nil, s.Host+"/partner/reward/add", modelIn)
	return
}

func (s *Service) PartnerRewardAvailable(l rpc.Logger, modelIn PartnerRewardAvailableIn) (err error) {
	value := url.Values{}
	value.Add("id", modelIn.Id)
	value.Add("available", strconv.FormatBool(modelIn.Available))
	err = s.Client.CallWithForm(l, nil, s.Host+"/partner/reward/available", map[string][]string(value))
	return
}

func (s *Service) PartnerRewardGet(l rpc.Logger, modelIn PartnerRewardGet) (model PartnerReward, err error) {
	value := url.Values{}
	value.Add("id", modelIn.Id)
	value.Add("name", modelIn.Name)
	err = s.Client.Call(l, &model, s.Host+"/partner/reward/get?"+value.Encode())
	return
}

func (s *Service) PartnerUpdate(l rpc.Logger, modelIn PartnerAddIn) (err error) {
	err = s.Client.CallWithJson(l, nil, s.Host+"/partner/update", modelIn)
	return
}

//实时扣费记录增加
func (s *Service) RealtimeDeduct(l rpc.Logger, modelIn RealtimeDeductIn) (err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(modelIn.Uid), 10))
	value.Add("month", modelIn.Month)
	value.Add("money", modelIn.Money.ToString())
	err = s.Client.CallWithForm(l, nil, s.Host+"/realtime/deduct", map[string][]string(value))
	return
}

//实时扣费记录删除
func (s *Service) RealtimeDeductDelete(l rpc.Logger, modelIn RealtimeDeductDelete) (err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(modelIn.Uid), 10))
	value.Add("month", modelIn.Month)
	value.Add("delete_before", strconv.FormatBool(modelIn.DeleteBefore))
	err = s.Client.CallWithForm(l, nil, s.Host+"/realtime/deduct/delete", map[string][]string(value))
	return
}

//获取用户实时扣费信息
func (s *Service) RealtimeInfo(l rpc.Logger, modelIn User) (model RealtimeInfoOut, err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(modelIn.Uid), 10))
	err = s.Client.Call(l, &model, s.Host+"/realtime/info?"+value.Encode())
	return
}

//充值
func (s *Service) Recharge(l rpc.Logger, modelIn RechargeIn) (model string, err error) {
	value := url.Values{}
	value.Add("excode", modelIn.Excode)
	value.Add("type", modelIn.Type)
	value.Add("uid", strconv.FormatUint(uint64(modelIn.Uid), 10))
	value.Add("money", modelIn.Money.ToString())
	if modelIn.At != nil {
		value.Add("at", (*modelIn.At).ToString())
	}
	value.Add("desc", modelIn.Desc)
	err = s.Client.CallWithForm(l, &model, s.Host+"/recharge", map[string][]string(value))
	return
}

//充值优惠券
func (s *Service) RechargeAdd_reward(l rpc.Logger, modelIn AddRewardIn) (err error) {
	value := url.Values{}
	value.Add("excode", modelIn.Excode)
	value.Add("desc", modelIn.Desc)
	value.Add("type", modelIn.Type)
	value.Add("uid", strconv.FormatUint(uint64(modelIn.Uid), 10))
	value.Add("money", modelIn.Money.ToString())
	if modelIn.At != nil {
		value.Add("at", (*modelIn.At).ToString())
	}
	value.Add("serial_num", modelIn.SerialNum)
	err = s.Client.CallWithForm(l, nil, s.Host+"/recharge/add_reward", map[string][]string(value))
	return
}

//自定义的充值赠送。开放式赠送金额，全额充入FreeNB
func (s *Service) RechargeFreeReward(l rpc.Logger, modelIn RechargeMini) (err error) {
	value := url.Values{}
	value.Add("excode", modelIn.Excode)
	value.Add("uid", strconv.FormatUint(uint64(modelIn.Uid), 10))
	value.Add("money", strconv.FormatInt(int64(modelIn.Money), 10))
	if modelIn.At != nil {
		value.Add("at", (*modelIn.At).ToString())
	}
	value.Add("desc", modelIn.Desc)
	value.Add("details", modelIn.Details)
	err = s.Client.CallWithForm(l, nil, s.Host+"/recharge/free/reward", map[string][]string(value))
	return
}

//获取充值列表
func (s *Service) RechargeList(l rpc.Logger, modelIn RechargeListIn) (model []TransactionOut, err error) {
	value := url.Values{}
	if modelIn.Uid != nil {
		value.Add("uid", strconv.FormatUint(uint64(*modelIn.Uid), 10))
	}
	if modelIn.StartTime != nil {
		value.Add("starttime", (*modelIn.StartTime).ToString())
	}
	if modelIn.EndTime != nil {
		value.Add("endtime", (*modelIn.EndTime).ToString())
	}
	value.Add("type", modelIn.Type)
	if modelIn.IsHide != nil {
		value.Add("ishide", strconv.FormatBool(*modelIn.IsHide))
	}
	if modelIn.Offset != nil {
		value.Add("offset", strconv.FormatInt(int64(*modelIn.Offset), 10))
	}
	if modelIn.Limit != nil {
		value.Add("limit", strconv.FormatInt(int64(*modelIn.Limit), 10))
	}
	err = s.Client.Call(l, &model, s.Host+"/recharge/list?"+value.Encode())
	return
}

//根据序列号获取优惠券
func (s *Service) RechargeRewardBy_serial_num(l rpc.Logger, modelIn RechargeRewardBy_serial_numIn) (model RewardInfo, err error) {
	value := url.Values{}
	value.Add("serial_num", modelIn.SerialNum)
	err = s.Client.Call(l, &model, s.Host+"/recharge/reward/by_serial_num?"+value.Encode())
	return
}

//获取充值优惠券
func (s *Service) RechargeRewards(l rpc.Logger, modelIn RewardModel) (model []RewardInfo, err error) {
	value := url.Values{}
	if modelIn.Money != nil {
		value.Add("money", strconv.FormatInt(int64(*modelIn.Money), 10))
	}
	if modelIn.Offset != nil {
		value.Add("offset", strconv.FormatInt(int64(*modelIn.Offset), 10))
	}
	if modelIn.Limit != nil {
		value.Add("limit", strconv.FormatInt(int64(*modelIn.Limit), 10))
	}
	err = s.Client.Call(l, &model, s.Host+"/recharge/rewards?"+value.Encode())
	return
}

//渠道分成充值
func (s *Service) Recharge_reward(l rpc.Logger, modelIn Recharge_rewardIn) (err error) {
	value := url.Values{}
	value.Add("excode", modelIn.Excode)
	value.Add("type", modelIn.Type)
	value.Add("desc", modelIn.Desc)
	value.Add("uid", strconv.FormatUint(uint64(modelIn.Uid), 10))
	value.Add("money", modelIn.Money.ToString())
	value.Add("cost", modelIn.Cost.ToString())
	value.Add("customer_reward_id", modelIn.CustomerRewardId)
	if modelIn.At != nil {
		value.Add("at", (*modelIn.At).ToString())
	}
	err = s.Client.CallWithForm(l, nil, s.Host+"/recharge_reward", map[string][]string(value))
	return
}

//渠道分成充值，并返回充值流水ID
func (s *Service) RechargeReward(l rpc.Logger, modelIn Recharge_rewardIn) (id string, err error) {
	value := url.Values{}
	value.Add("excode", modelIn.Excode)
	value.Add("type", modelIn.Type)
	value.Add("desc", modelIn.Desc)
	value.Add("uid", strconv.FormatUint(uint64(modelIn.Uid), 10))
	value.Add("money", modelIn.Money.ToString())
	value.Add("cost", modelIn.Cost.ToString())
	value.Add("customer_reward_id", modelIn.CustomerRewardId)
	if modelIn.At != nil {
		value.Add("at", (*modelIn.At).ToString())
	}
	err = s.Client.CallWithForm(l, &id, s.Host+"/recharge/reward", map[string][]string(value))
	return
}

func (s *Service) Refund(l rpc.Logger, modelIn Refund) (err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(modelIn.Uid), 10))
	value.Add("money", modelIn.Money.ToString())
	value.Add("free_money", modelIn.FreeMoney.ToString())
	value.Add("excode", modelIn.Excode)
	if modelIn.At != nil {
		value.Add("at", (*modelIn.At).ToString())
	}
	value.Add("desc", modelIn.Desc)
	value.Add("detail", modelIn.Detail)
	err = s.Client.CallWithForm(l, nil, s.Host+"/refund", map[string][]string(value))
	return
}

//充值卡设置充值额度
func (s *Service) ThirdpartyLimit(l rpc.Logger) (model int64, err error) {
	err = s.Client.Call(l, &model, s.Host+"/thirdparty/limit")
	return
}

//充值卡设置充值额度
func (s *Service) ThirdpartyLimitUpdate(l rpc.Logger, modelIn ThirdpartyLimit) (err error) {
	value := url.Values{}
	value.Add("limit", strconv.FormatInt(int64(modelIn.Limit), 10))
	err = s.Client.CallWithForm(l, nil, s.Host+"/thirdparty/limit/update", map[string][]string(value))
	return
}

//更新流水的At字段（业务操作时间)
func (s *Service) TransactionUpdateAt(l rpc.Logger, modelIn TransactionAtReq) (err error) {
	value := url.Values{}
	value.Add("serial_num", modelIn.Serial_num)
	value.Add("at", modelIn.At.ToString())
	err = s.Client.CallWithForm(l, nil, s.Host+"/transaction/update/at", map[string][]string(value))
	return
}

func (s *Service) UserGet(l rpc.Logger, modelIn User) (model UserInfo, err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(modelIn.Uid), 10))
	err = s.Client.Call(l, &model, s.Host+"/user/get?"+value.Encode())
	return
}

func (s *Service) UserList(l rpc.Logger, modelIn UserListIn) (model []UserInfo, err error) {
	value := url.Values{}
	if modelIn.UserType != nil {
		value.Add("user_type", strconv.FormatInt(int64(*modelIn.UserType), 10))
	}
	if modelIn.UserTag != nil {
		value.Add("user_tag", strconv.FormatInt(int64(*modelIn.UserTag), 10))
	}
	if modelIn.DeductType != nil {
		value.Add("deduct_type", strconv.FormatInt(int64(*modelIn.DeductType), 10))
	}
	if modelIn.Offset != nil {
		value.Add("offset", strconv.FormatInt(int64(*modelIn.Offset), 10))
	}
	if modelIn.Limit != nil {
		value.Add("limit", strconv.FormatInt(int64(*modelIn.Limit), 10))
	}
	err = s.Client.Call(l, &model, s.Host+"/user/list?"+value.Encode())
	return
}

func (s *Service) UserUpdate(l rpc.Logger, modelIn UserUpdateIn) (err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(modelIn.Uid), 10))
	if modelIn.Company != nil {
		value.Add("company", *modelIn.Company)
	}
	if modelIn.Type != nil {
		value.Add("type", (*modelIn.Type).ToString())
	}
	if modelIn.Tag != nil {
		value.Add("tag", (*modelIn.Tag).ToString())
	}
	if modelIn.DeductType != nil {
		value.Add("deduct_type", (*modelIn.DeductType).ToString())
	}
	if modelIn.Remark != nil {
		value.Add("remark", *modelIn.Remark)
	}
	if modelIn.NickName != nil {
		value.Add("nickname", *modelIn.NickName)
	}
	err = s.Client.CallWithForm(l, nil, s.Host+"/user/update", map[string][]string(value))
	return
}

//销售销帐接口
func (s *Service) Writeoff(l rpc.Logger, modelIn WriteoffIn) (model string, err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(modelIn.Uid), 10))
	value.Add("money", strconv.FormatInt(int64(modelIn.Money), 10))
	value.Add("prefix", modelIn.Prefix)
	value.Add("desc", modelIn.Desc)
	value.Add("excode", modelIn.Excode)
	err = s.Client.CallWithForm(l, &model, s.Host+"/writeoff", map[string][]string(value))
	return
}
