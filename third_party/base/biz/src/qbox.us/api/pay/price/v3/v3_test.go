package v3

import (
	"github.com/qiniu/rpc.v1"
	. "qbox.us/api/pay/pay"
)

type MockHandleUser struct {
}

func (r MockHandleUser) PortalGet(logger rpc.Logger, req ReqUidAndWhen) (resp ModelPriceForPortal, err error) {

	resp = ModelPriceForPortal{
		Uid: 123456789,
		Frees: map[Group]map[Item]int64{
			GROUP_COMMON: map[Item]int64{
				SPACE:    10,
				TRANSFER: 10,
				APIGET:   100 * 10000,
				APIPUT:   10 * 10000,
			},
		},
		Bases: map[Group]map[Item]ModelUserItemBaseForPortal{
			GROUP_COMMON: map[Item]ModelUserItemBaseForPortal{
				SPACE: ModelUserItemBaseForPortal{
					"xxxxxxxxxxxxxxxxxxxxx",
					LifeCycle{"20140701", "201507001"},
					ModelItemBasePriceForPortal{
						ITEM_BASE_COMMON, "", "",
						CumulativeTypeMonth,
						BillPeriodMonth,
						ModelRangePriceForPortal{
							UNITPRICE,
							[]ModelPriceRangeForPortal{
								ModelPriceRangeForPortal{0, 10, 0},
								ModelPriceRangeForPortal{10, 50 * 1024, 1650},
								ModelPriceRangeForPortal{50 * 1024, 500 * 1024, 1620},
								ModelPriceRangeForPortal{500 * 1024, 5 * 1024 * 1024, 1590},
								ModelPriceRangeForPortal{5 * 1024 * 1024, 5 * 1024 * 1024, 1560},
							},
						},
						0,
						ModelResourceGroupListForPortal{},
					},
				},
			},
		},
	}

	return
}
