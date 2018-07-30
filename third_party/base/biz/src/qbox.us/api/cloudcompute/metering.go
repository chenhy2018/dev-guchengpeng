package cc

import (
	"github.com/qiniu/rpc.v2"
)

type MetricData struct {
	Period        int     `json:"period"`
	Duration      float32 `json:"duration"`
	Max           float32 `json:"max"`
	Min           float32 `json:"min"`
	Avg           float32 `json:"avg"`
	Sum           float32 `json:"sum"`
	Count         int     `json:"count"`
	Unit          string  `json:"unit"`
	PeriodStart   string  `json:"period_start"`
	PeriodEnd     string  `json:"period_end"`
	DurationStart string  `json:"duration_start"`
	DurationEnd   string  `json:"duration_end"`
}

func (r *Service) GetMetricdataStatistics(l rpc.Logger, metricName, resourceId, kind string) (ret []MetricData, err error) {
	params := map[string][]string{
		"metric_name": {metricName},
		"resource_id": {resourceId},
		"kind":        {kind},
	}
	err = r.Conn.CallWithForm(l, &ret, "GET", r.Host+r.ApiPrefix+"/metricdata/statistics", params)
	return
}
