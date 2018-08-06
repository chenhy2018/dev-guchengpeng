package pay

import (
	"strconv"
	"time"
)

type Money int64 //精确到0.01分

func (m Money) ToString() string {
	return strconv.FormatInt(int64(m), 10)
}

func (m Money) AsYuan() float64 {
	return float64(m) / 10000
}

//----------------------------------------------------------------------------//

var (
	TZ_CST   *time.Location = time.FixedZone("CST", 8*3600)
	TIME_MIN time.Time      = time.Date(2000, 1, 1, 0, 0, 0, 0, TZ_CST)
	TIME_MAX time.Time      = time.Date(2200, 1, 1, 0, 0, 0, 0, TZ_CST)
)

//----------------------------------------------------------------------------//

type HundredNanoSecond int64

func NewHundredNanoSecond(t time.Time) HundredNanoSecond {
	return HundredNanoSecond(t.In(TZ_CST).UnixNano() / 100)
}

func (a HundredNanoSecond) Time() time.Time {
	return time.Unix(int64(a)/1e7, int64(a)%1e7*100).In(TZ_CST)
}

func (a HundredNanoSecond) ToString() string {
	return strconv.FormatInt(int64(a), 10)
}

func (r *HundredNanoSecond) Value() int64 {
	if r == nil {
		return 0
	}
	return int64(*r)
}

type Second int64

func NewSecond(t time.Time) Second {
	return Second(t.In(TZ_CST).Unix())
}

func (r Second) Time() time.Time {
	return time.Unix(int64(r), 0).In(TZ_CST)
}

func (r Second) ToString() string {
	return strconv.FormatInt(int64(r), 10)
}

func (r *Second) Value() int64 {
	if r == nil {
		return 0
	}
	return int64(*r)
}

type Day string // like 20060102

func NewDay(t time.Time) Day {
	return Day(t.In(TZ_CST).Format("20060102"))
}

func (r Day) Time() (time.Time, error) {
	t, err := time.Parse("20060102 -0700", string(r)+" +0800")
	if err != nil {
		return t, err
	}
	t = t.In(TZ_CST)
	return t, err
}

func (r Day) ToString() string {
	return string(r)
}

type Month string // like 200601

func NewMonth(t time.Time) Month {
	return Month(t.In(TZ_CST).Format("200601"))
}

func (r Month) Time() (time.Time, error) {
	t, err := time.Parse("200601 -0700", string(r)+" +0800")
	if err != nil {
		return t, err
	}
	t = t.In(TZ_CST)
	return t, err
}

func (r Month) ToString() string {
	return string(r)
}
