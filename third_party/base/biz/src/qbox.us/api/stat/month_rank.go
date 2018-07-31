// 单项数据的月排名接口
// 需要admin权限

package stat

import "strconv"

type MonthRankEntry struct {
	Uid   []uint32 `json:"uid"`
	Data  []int64  `json:"data"`
	Start uint     `json:"start"`
	End   uint     `json:"end"`
}

// AdminMonthRankTraffic 获取流量月排名信息
// month 为要查询月份的日期字符串(200601)
// skip 跳过条目数
// limit 条目限制数，默认为50, 最大为1000
// 当 skip = 5, limit = 5时，则返回的排名为[6,7,8,9,10]
func (r Service) AdminMonthRankTraffic(host, month string, skip, limit uint) (
	data MonthRankEntry, code int, err error) {
	code, err = r.Conn.CallWithForm(&data, host+"/rank/traffic",
		map[string][]string{
			"month": []string{month},
			"skip":  []string{strconv.FormatUint(uint64(skip), 10)},
			"limit": []string{strconv.FormatUint(uint64(limit), 10)},
		})
	return
}

// AdminMonthRankBandwidth 获取带宽月排名信息
// month 为要查询月份的日期字符串(200601)
// skip 跳过条目数
// limit 条目限制数，默认为50, 最大为1000
// 当 skip = 5, limit = 5时，则返回的排名为[6,7,8,9,10]
func (r Service) AdminMonthRankBandwidth(host, month string, skip, limit uint) (
	data MonthRankEntry, code int, err error) {
	code, err = r.Conn.CallWithForm(&data, host+"/rank/bandwidth",
		map[string][]string{
			"month": []string{month},
			"skip":  []string{strconv.FormatUint(uint64(skip), 10)},
			"limit": []string{strconv.FormatUint(uint64(limit), 10)},
		})
	return
}

// AdminMonthRankSpace 获取空间月排名信息
// month为要查询月份的日期字符串(200601)
// skip 跳过条目数
// limit 条目限制数，默认为50, 最大为1000
// 当 skip = 5, limit = 5时，则返回的排名为[6,7,8,9,10]
func (r Service) AdminMonthRankSpace(host, month string, skip, limit uint) (
	data MonthRankEntry, code int, err error) {
	code, err = r.Conn.CallWithForm(&data, host+"/rank/space",
		map[string][]string{
			"month": []string{month},
			"skip":  []string{strconv.FormatUint(uint64(skip), 10)},
			"limit": []string{strconv.FormatUint(uint64(limit), 10)},
		})
	return
}

// AdminMonthRankApicall 获取api请求数月排名信息
// month 为要查询月份的日期字符串(200601)
// typ 为apicall类型(get/put)
// skip 跳过条目数
// limit 条目限制数，默认为50, 最大为1000
// 当 skip = 5, limit = 5时，则返回的排名为[6,7,8,9,10]
func (r Service) AdminMonthRankApicall(host, month string, typ string, skip, limit uint) (
	data MonthRankEntry, code int, err error) {
	code, err = r.Conn.CallWithForm(&data, host+"/rank/apicall",
		map[string][]string{
			"month": []string{month},
			"type":  []string{typ},
			"skip":  []string{strconv.FormatUint(uint64(skip), 10)},
			"limit": []string{strconv.FormatUint(uint64(limit), 10)},
		})
	return
}
