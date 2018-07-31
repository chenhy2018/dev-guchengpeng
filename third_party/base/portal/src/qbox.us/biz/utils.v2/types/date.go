package types

import (
	"time"
)

// store the date in the "2013-02-14" style
type Date string

func MakeDateFromTime(t time.Time) Date {
	return Date(t.Format("2006-01-02"))
}

func (d Date) Time() (time.Time, error) {
	return time.ParseInLocation("2006-01-02", string(d), time.Local)
}

// 找到当前日期下的下一天刚开始的时刻
// 如：
// 2015-01-08 输入duration=0将会得到:
// 2015-01-09 00:00:00 +0800 CST
//
// 输入 duration = 7 将会得到
// 2015-01-02 00:00:00 +0800 CST
func (d Date) DaysAgoInit(duration int) (time.Time, error) {

	dateTime, err := d.Time()
	if err != nil {
		return dateTime, err
	}

	tmpDayAgo := dateTime.AddDate(0, 0, -(duration - 1))
	year, month, day := tmpDayAgo.Date()
	loc := tmpDayAgo.Location()

	// set second && nsec to zero
	dayAgo := time.Date(year, month, day, 0, 0, 0, 0, loc)
	return dayAgo, nil
}

// -------------------------------

// 范例: 2013-06
type Month string

func MakeMonthFromTime(t time.Time) Month {
	return Month(t.Format("2006-01"))
}

func (m Month) StartDate() Date {
	return Date(m.Time().Format("2006-01-02"))
}

func (m Month) EndDate() Date {
	return Date(m.Time().AddDate(0, 1, 0).AddDate(0, 0, -1).Format("2006-01-02"))
}

func (m Month) NextMonth() Month {
	return MakeMonthFromTime(m.Time().AddDate(0, 1, 0))
}

func (m Month) PrevMonth() Month {
	return MakeMonthFromTime(m.Time().AddDate(0, -1, 0))
}

func (m Month) Time() time.Time {
	res, err := time.ParseInLocation("2006-01", string(m), time.Local)
	if err != nil {
		panic(err)
	}
	return res
}

func (m Month) HundredNanoSecond() HundredNanoSecond {
	return MakeHunderedNanoSecondFromTime(m.Time())
}
