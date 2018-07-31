package types

import (
	"regexp"
	"strings"
	"time"
)

type Second int64
type NanoSecond int64
type HundredNanoSecond int64

func (a HundredNanoSecond) Time() time.Time {
	return time.Unix(int64(a)/1e7, int64(a)%1e7*100)
}

func MakeHunderedNanoSecondFromTime(t time.Time) HundredNanoSecond {
	return HundredNanoSecond(t.UnixNano() / 100)
}

type Duration time.Duration

func (d Duration) Localize() string {
	str := time.Duration(d).String()
	regex := regexp.MustCompile(`(\d+)(ms|us|µs|ns|h|m|s)`)
	allUnit := regex.FindAllStringSubmatch(str, -1)

	localizeStr := ""
	for index, item := range allUnit {
		if len(item) != 3 {
			continue
		}

		num := item[1]
		unit := item[2]

		if num == "0" {
			continue
		}

		unit = strings.Replace(unit, "ms", "毫秒", -1)
		unit = strings.Replace(unit, "us", "微秒", -1)

		unit = strings.Replace(unit, "µs", "微秒", -1)
		unit = strings.Replace(unit, "ns", "纳秒", -1)

		unit = strings.Replace(unit, "s", "秒", -1)
		unit = strings.Replace(unit, "m", "分钟", -1)
		unit = strings.Replace(unit, "h", "小时", -1)

		if index != 0 {
			localizeStr += " "
		}

		localizeStr += num + unit
	}

	return localizeStr
}
