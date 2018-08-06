package wallet

import (
	"net/url"
	"strconv"
)

import (
	"github.com/qiniu/rpc.v1"
	. "qbox.us/api/pay/pay"
)

type MonthStatementV1 struct {
	Id       string        `json:"id" bson:"_id"`
	Uid      uint32        `json:"uid" bson:"uid"`
	Month    string        `json:"month" bson:"month"` // 2006-01
	Desc     string        `json:"desc" bson:"desc"`
	Status   int           `json:"status" bson:"status"`
	Money    Money         `json:"money" bson:"money"`
	Bills    []OldBillInfo `json:"bills" bson:"bills"`
	CreateAt Second        `json:"create_at"`
	UpdateAt Second        `json:"update_at"`
	Version  string        `json:"version" bson:"version"`
}

type OldBillInfo struct {
	Details map[string]OldBaseBill `json:"details"` // BASEBILL_FORMAT:map[DataType]BaseBillFormat
	Type    BillType               `json:"type"`    // BASEBILL, BASEBILL_FORMAT
}

type OldBaseBill struct {
	Money   Money         `json:"money"`
	Details []OldBBDetail `json:"details"`
}

type OldBBDetail struct {
	Start     string            `json:"start"`      // 2006-01-02
	End       string            `json:"end"`        // 2006-01-02
	ValueType string            `json:"value_type"` // 计费模式，空间:日均，流量/请求：累积,带宽：top95...
	Money     Money             `josn:"money"`
	Value     int64             `json:"value"`
	Units     OldBBDetailUnits  `json:"units"`
	Rewards   []OldBBRewardUnit `json:"rewards"`
	Discounts []OldBBDiscount   `json:"discounts"`
	Rebates   []OldBBRebate     `json:"rebates"`
}

type OldBBDetailUnits struct {
	Type  string            `json:"type"` // UNITPRICE, FIRST_BUYOUT, EACH_BUYOUT ...
	Units []OldBBDetailUnit `json:"units"`
}

type OldBBDetailUnit struct {
	From     int64 `json:"from"`
	To       int64 `json:"to"`
	Price    Money `json:"price"`
	Value    int64 `json:"value"`     // 实际产生费用部分
	Money    Money `json:"money"`     // 实际费用
	AllValue int64 `json:"all_value"` // 全额使用量，包含未产生最终费用的部分
	AllMoney Money `json:"all_money"` // 全额使用量对应的收入费用
}

type OldBBRewardUnit struct {
	Id      string            `json:"id"`
	OpId    string            `json:"opid"`
	Type    string            `json:"type"`
	Desc    string            `json:"desc"`
	Quota   int64             `json:"quota"`
	Value   int64             `json:"value"`
	Balance int64             `json:"balance"`
	Reduce  Money             `json:"reduce"`
	Details []OldBBRangeMoney `json:"details"` //阶梯使用详情
	Overdue bool              `json:"overdue"`
}

type OldBBRangeMoney struct {
	From   int64 `json:"from"`
	To     int64 `json:"to"`
	Value  int64 `json:"value"`
	Reduce Money `json:"money"` // 实际费用
}

type OldBBDiscount struct {
	Id      string `json:"id"`
	Type    string `json:"type"`
	Name    string `json:"name"`
	Desc    string `json:"desc"`
	Before  Money  `json:"before"`
	Change  Money  `json:"change"`
	Percent int64  `json:"percent"`
	After   Money  `json:"after"`
}

type OldBBRebate struct {
	Id     string `json:"id"`
	Type   string `json:"type"`
	Name   string `json:"name"`
	Desc   string `json:"desc"`
	Free   Money  `json:"free"`
	Before Money  `json:"before"`
	Change Money  `json:"change"`
	After  Money  `json:"after"`
}

type OldHandleMonthStatement struct {
	Host   string
	Client *rpc.Client
}

func NewOldHandleMonthStatement(host string, client *rpc.Client) *OldHandleMonthStatement {
	return &OldHandleMonthStatement{host, client}
}

type ReqOldIDOrMonth struct {
	ID    string `json:"id"`
	Uid   uint32 `json:"uid"`
	Month Month  `json:"month"`
}

func (r OldHandleMonthStatement) MonthstatementGet(logger rpc.Logger, req ReqOldIDOrMonth) (resp MonthStatementV1, err error) {
	value := url.Values{}
	value.Add("id", req.ID)
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	value.Add("month", req.Month.ToString())
	err = r.Client.Call(logger, &resp, r.Host+"/monthstatement/get?"+value.Encode())
	return
}

func (r OldHandleMonthStatement) MockMonthstatementGet(logger rpc.Logger, req ReqOldIDOrMonth) (resp MonthStatementV1, err error) {
	resp = MonthStatementV1{
		Id:       "53b35cfc0b12e03ea8089f11",
		Uid:      123456789,
		Money:    10400,
		Month:    "2014-06",
		Desc:     "2014-06 对账单",
		Status:   3,
		CreateAt: 0,
		UpdateAt: 0,
		Version:  "V1",

		Bills: []OldBillInfo{
			OldBillInfo{
				Details: map[string]OldBaseBill{
					APIGET.ToString(): OldBaseBill{
						Details: []OldBBDetail{
							OldBBDetail{
								Start: "2014-06-01",
								End:   "2014-06-30",
								Money: 10400,
								Discounts: []OldBBDiscount{
									OldBBDiscount{
										Id:      "xxxxxx",
										Type:    "SPONSOR",
										Name:    "资源互换90%",
										Desc:    "9折",
										Before:  17200,
										Change:  -1800,
										Percent: 90,
										After:   15400,
									},
								},
								Rebates: []OldBBRebate{
									OldBBRebate{
										Id:     "yyyyyyyy",
										Type:   "FREE",
										Name:   "Test",
										Desc:   "免费0.5元",
										Free:   5000,
										Before: 15400,
										Change: -5000,
										After:  10400,
									},
								},
								Rewards: []OldBBRewardUnit{
									OldBBRewardUnit{
										Balance: 0,
										Desc:    "",
										Details: []OldBBRangeMoney{
											OldBBRangeMoney{
												From:   0,
												Reduce: 10000,
												To:     0,
												Value:  1000000,
											},
										},
										Id:      "",
										OpId:    "",
										Overdue: false,
										Quota:   0,
										Reduce:  10000,
										Type:    "",
										Value:   1000000,
									},
								},
								Units: OldBBDetailUnits{
									Type: "UNITPRICE",
									Units: []OldBBDetailUnit{
										OldBBDetailUnit{
											AllMoney: 27200,
											AllValue: 2721724,
											From:     0,
											Money:    17200,
											Price:    10,
											To:       0,
											Value:    1721724,
										},
									},
								},
								Value:     2721724,
								ValueType: "",
							},
						},
						Money: 10400,
					},
				},
				Type: BILLTYPE_BASEBILL,
			},
		},
	}
	return
}

func (r HandleMonthStatement) MockGet(logger rpc.Logger, req ReqIDOrMonth) (resp ModelMonthStatement, err error) {
	resp = ModelMonthStatement{
		ID:       "xxxxxxxxxxxx",
		Uid:      123456,
		Money:    1710000,
		Month:    Month("201406"),
		Desc:     "2014-06 对账单",
		Status:   3,
		CreateAt: 0,
		UpdateAt: 0,
		Version:  "V2",
		Detail: ModelMonthStatementDetail{
			Discounts: []ModelBaseBillDiscount{
				{
					Before: 2137500,
					Change: -427500,
					After:  1710000,
				},
			},
			Groups: map[Group]ModelMonthStatementGroup{
				GROUP_COMMON: ModelMonthStatementGroup{
					Money: 2137500,
					Discounts: []ModelBaseBillDiscount{
						{
							Before: 2375000,
							Change: -237500,
							After:  2137500,
						},
					},
					Items: map[Item]ModelMonthStatementItem{
						APIGET: ModelMonthStatementItem{
							Money: 2375000,
							Units: []ModelMonthStatementUnit{
								{
									From:  Day("20140601"),
									To:    Day("20140630"),
									Money: 2375000,
									Value: 3000000,
									Base: ModelBaseBillBase{
										Units: []ModelBaseBillBaseUnit{
											{
												AllMoney: 1000000,
												AllValue: 1000000,
												From:     0,
												To:       1000000,
												Money:    500000,
												Value:    500000,
											},
											{
												AllMoney: 2000000,
												AllValue: 2000000,
												From:     1000000,
												To:       10000000,
												Money:    2000000,
												Value:    2000000,
											},
										},
									},
									Packages: []ModelBaseBillPackage{
										{
											Value:   500000,
											Balance: 0,
											Reduce:  500000,
											Units: []ModelRangeMoney{
												{
													From:   0,
													To:     1000000,
													Reduce: 500000,
													Value:  500000,
												},
											},
											Overdue: false,
										},
									},
									Discounts: []ModelBaseBillDiscount{
										{
											Before: 2500000,
											Change: -125000,
											After:  2375000,
										},
									},
								},
							},
						},
					},
				},
				GROUP_MPS: ModelMonthStatementGroup{
					Money: 0,
					Rebates: []ModelBaseBillRebate{
						{
							Before: 600000,
							Change: -600000,
							After:  0,
						},
					},
					Items: map[Item]ModelMonthStatementItem{
						MPS_SD: ModelMonthStatementItem{
							Money: 600000,
							Units: []ModelMonthStatementUnit{
								{
									From:  Day("20140601"),
									To:    Day("20140630"),
									Money: 600000,
									Value: 36000,
									Base: ModelBaseBillBase{
										Units: []ModelBaseBillBaseUnit{
											{
												AllMoney: 600000,
												AllValue: 36000,
												From:     0,
												To:       0,
												Money:    600000,
												Value:    36000,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	return
}
