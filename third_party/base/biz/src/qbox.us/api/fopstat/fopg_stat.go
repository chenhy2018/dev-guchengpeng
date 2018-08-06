package fopstat

import (
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/qiniu/rpc.v2"
	"github.com/qiniu/rpc.v2/failover"
)

type FopgStat struct {
	Conn *failover.Client
}

func NewFopgStat(hosts []string, tr http.RoundTripper) *FopgStat {
	return &FopgStat{
		Conn: failover.New(hosts, &failover.Config{
			Http: &http.Client{
				Transport: tr,
			},
		}),
	}
}

//-------------------------------------

// G
type Granularity string

const (
	G_5MIN Granularity = "5min"
	G_DAY  Granularity = "day"
)

func (g Granularity) ToString() string {
	return string(g)
}

// value type
type Type string

const (
	HITS     Type = "hits"
	RESPTIME Type = "resptime"
)

func (t Type) ToString() string {
	return string(t)
}

//------------------------------------

type Timestamp int64
type RespQuery struct {
	Times []Timestamp `json:"times"`
	Datas []int64     `json:"datas"`
}

//-------------------------------------

type HitsOrResptime struct {
	Cmd   string      `json:"cmd"`
	Code  uint        `json:"code"`
	Type  Type        `json:"value"`
	Begin string      `json:"begin"` //like 20150801000000 (means 2015-08-01 00:00:00)
	End   string      `json:"end"`   //like 20150801000000 (means 2015-08-01 00:00:00)
	G     Granularity `json:"g"`
}

// ****************注意: 调用此Client时需要考虑到限制用户访问次数
func (f *FopgStat) GetHitsOrResptime(logger rpc.Logger, req HitsOrResptime) (ret RespQuery, err error) {

	value := url.Values{}
	value.Add("value", req.Type.ToString())
	value.Add("begin", req.Begin)
	value.Add("end", req.End)
	value.Add("g", req.G.ToString())

	if req.Cmd != "" {
		value.Add("api", strings.ToUpper(req.Cmd))
	}

	if req.Code != 0 {
		value.Add("code", strconv.FormatUint(uint64(req.Code), 10))
	}

	err = f.Conn.Call(logger, &ret, "GET", "/get/fopg?"+value.Encode())
	return
}

type AveRespTime struct {
	Cmd   string      `json:"cmd"`
	Code  uint        `json:"code"`
	Begin string      `json:"begin"` //like 20150801000000 (means 2015-08-01 00:00:00)
	End   string      `json:"end"`   //like 20150801000000 (means 2015-08-01 00:00:00)
	G     Granularity `json:"g"`
}

// ****************注意: 调用此Client时需要考虑到限制用户访问次数
func (f *FopgStat) AverageResptime(logger rpc.Logger, req AveRespTime) (ret uint64, err error) {

	respTimeValue := url.Values{}
	respTimeValue.Add("value", RESPTIME.ToString())
	respTimeValue.Add("begin", req.Begin)
	respTimeValue.Add("end", req.End)
	respTimeValue.Add("g", req.G.ToString())

	if req.Cmd != "" {
		respTimeValue.Add("api", strings.ToUpper(req.Cmd))
	}

	if req.Code != 0 {
		respTimeValue.Add("code", strconv.FormatUint(uint64(req.Code), 10))
	}

	var respTimeRet RespQuery
	var totalResptime uint64 = 0
	err = f.Conn.Call(logger, &respTimeRet, "GET", "/get/fopg?"+respTimeValue.Encode())
	if err != nil {
		return
	}
	for _, resptime := range respTimeRet.Datas {
		totalResptime += uint64(resptime)
	}
	//--------------------------

	hitsValue := url.Values{}
	hitsValue.Add("value", HITS.ToString())
	hitsValue.Add("begin", req.Begin)
	hitsValue.Add("end", req.End)
	hitsValue.Add("g", req.G.ToString())

	if req.Cmd != "" {
		hitsValue.Add("api", strings.ToUpper(req.Cmd))
	}

	if req.Code != 0 {
		hitsValue.Add("code", strconv.FormatUint(uint64(req.Code), 10))
	}

	var hitsRet RespQuery
	var totalHits uint64 = 0
	err = f.Conn.Call(logger, &hitsRet, "GET", "/get/fopg?"+hitsValue.Encode())
	if err != nil {
		return
	}
	for _, hits := range hitsRet.Datas {
		totalHits += uint64(hits)
	}

	if totalHits == 0 {
		err = errors.New("no data")
		return
	}

	return totalResptime / totalHits, nil

}
