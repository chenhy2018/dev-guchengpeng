package wallet

import (
	"strconv"
	"strings"
)

const (
	B  int64 = 1
	KB int64 = 1024 * B
	MB int64 = 1024 * KB
	GB int64 = 1024 * MB
	TB int64 = 1024 * GB
	PB int64 = 1024 * TB
)

var units = []int64{PB, TB, GB}
var strUnit = []string{"PB", "TB", "GB"}

var numUnitMap = map[int]string{1e4: "万次", 1e3: "千次"}

func formatDataInt(bytes int64) string {
	for i, unit := range units {
		if bytes >= unit {
			if bytes%unit != 0 {
				continue
			}
			return strconv.FormatInt(bytes/unit, 10) + " " + strUnit[i]
		}
	}
	return "0 GB"
}

func formatDataGBFloat(bytes int64, pre int) string {
	return addComma(strconv.FormatFloat(float64(bytes)/float64(GB), 'f', pre, 64)) + " GB"
}

func formatMoneyFloat(money int64, pre int) string {
	return "￥ " + strconv.FormatFloat(float64(money)/1e4, 'f', pre, 64)
}

func formatDataPriceFloat(money int64, pre int) string {
	return "￥ " + strconv.FormatFloat(float64(money)/1e4, 'f', pre, 64) + "/GB"
}

func formatNumPriceFloat(money int64, pre, base int) string {
	return "￥ " + strconv.FormatFloat(float64(money)/1e4, 'f', pre, 64) + "/" + numUnitMap[base]
}

func formatNumFloat(num int64, base, pre int) string {
	return addComma(strconv.FormatFloat(float64(num)/float64(base), 'f', pre, 64)) +
		" " + numUnitMap[base]
}

func formatDataRange(fromBytes, toBytes int64) string {
	from := formatDataInt(fromBytes)
	to := ""
	if toBytes != fromBytes {
		to = formatDataInt(toBytes)
	}
	return from + " - " + to
}

func addComma(numStr string) string {
	additional := ""
	pointIndex := strings.Index(numStr, ".")
	if pointIndex != -1 {
		additional = numStr[pointIndex:]
		numStr = numStr[0:pointIndex]
	}
	strLen := len(numStr)
	strs := make([]string, 0)
	begin := strLen % 3
	if begin > 0 {
		strs = []string{numStr[0:begin]}
	}
	for i := begin; i < strLen; i = i + 3 {
		strs = append(strs, numStr[i:i+3])
	}
	if additional == "" {
		return strings.Join(strs, ",")
	}
	return strings.Join(strs, ",") + additional
}
