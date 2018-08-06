package stat

import (
	"strconv"
	"time"
)

//----------------------------------------------------------------------------//

type HundredNanoSecond int64

func NewHundredNanoSecond(t time.Time) HundredNanoSecond {
	return HundredNanoSecond(t.UnixNano() / 100)
}

func (a HundredNanoSecond) Time() time.Time {
	return time.Unix(int64(a)/1e7, int64(a)%1e7*100)
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
	return Second(t.Unix())
}

func (r Second) Time() time.Time {
	return time.Unix(int64(r), 0)
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
	return Day(t.Format("20060102"))
}

func (r Day) Time() (time.Time, error) {
	return time.Parse("20060102 -0700", string(r)+" +0800")
}

func (r Day) ToString() string {
	return string(r)
}

type Month string // like 200601

func NewMonth(t time.Time) Month {
	return Month(t.Format("200601"))
}

func (r Month) Time() (time.Time, error) {
	return time.Parse("200601 -0700", string(r)+" +0800")
}

func (r Month) ToString() string {
	return string(r)
}

//----------------------------------------------------------------------------//

type ShowType string

var (
	P_5MIN  ShowType = "5min"
	P_DAY   ShowType = "day"
	P_MONTH ShowType = "month"
)

func (self ShowType) ToString() string {
	return string(self)
}

//----------------------------------------------------------------------------//

type ReqTimeQuery struct {
	Uid    uint32   `json:"uid"`
	Bucket *string  `json:"bucket"`
	From   Day      `json:"from"`
	To     Day      `json:"to"`
	P      ShowType `json:"p"`
}

type ReqDomainTimeQuery struct {
	Uid     uint32   `json:"uid"`
	Domain  string   `json:"domain"`
	ApiType *string  `json:"type"` //get|put
	From    Day      `json:"from"`
	To      Day      `json:"to"`
	P       ShowType `json:"p"`
}

type ReqBandwidthAdjustment struct {
	Uid    uint32  `json:"uid"`
	Bucket *string `json:"bucket"`
	From   Day     `json:"from"`
	To     Day     `json:"to"`
	Limit  int     `json:"limit"`
}

type ReqApiCallTimeQuery struct {
	Uid     uint32   `json:"uid"`
	Bucket  *string  `json:"bucket"`
	From    Day      `json:"from"`
	To      Day      `json:"to"`
	P       ShowType `json:"p"`
	ApiType string   `json:"type"`
}

type ReqMonthInfo struct {
	Uid    uint32  `json:"uid"`
	Bucket *string `json:"bucket"`
	Month  Month   `json:"month"`
}

type ReqDayInfo struct {
	Uid    uint32  `json:"uid"`
	Bucket *string `json:"bucket"`
	Day    Day     `json:"day"`
}

type RespSimpleTimeQuery struct {
	Times []Second `json:"time"`
	Datas []int64  `json:"data"`
}

type RespRtTimeQuery struct {
	Times           []Second `json:"time"`
	Datas           []int64  `json:"data"`
	RealtimeIndexes []int    `json:"realtimeIndexes"`
}

type RespBucketsQuery struct {
	Buckets []string `json:"bucket"`
	Datas   []int64  `json:"data"`
}
