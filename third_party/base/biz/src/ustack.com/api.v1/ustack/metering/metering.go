package metering

import (
	"strconv"

	"github.com/qiniu/rpc.v2"
	"ustack.com/api.v1/ustack"
)

// --------------------------------------------------

type Client struct {
	ProjectId string
	Conn      ustack.Conn
}

func New(services ustack.Services, project string) Client {
	conn, ok := services.Find("metering")
	if !ok {
		panic("metering api not found")
	}
	return Client{
		ProjectId: project,
		Conn:      conn,
	}
}

func fakeError(err error) bool {
	if rpc.HttpCodeOf(err)/100 == 2 {
		return true
	}
	return false
}

// --------------------------------------------------
// 获得 metric 的列表

type Metric struct {
	ResourceName string `json:"resource_name"`
	MetricName   string `json:"metric_name"`
	ResourceId   string `json:"resource_id"`
	MetricUnit   string `json:"metric_unit"`
	VerboseName  string `json:"verbose_name"`
	ResourceType string `json:"resource_type"`
}

type SrvMetricArgs struct {
	CpuUtil        bool `json:"cpu_util"`         // cpu 使用量 %
	MemUtil        bool `json:"mem_util"`         // 内存使用率 %
	DiskReadRate   bool `json:"disk_read_rate"`   // 磁盘读速率 B/s
	DiskWriteRate  bool `json:"disk_write_rate"`  // 磁盘写速率 B/s
	NetworkInRate  bool `json:"network_in_rate"`  // 网络进流量 B/s
	NetworkOutRate bool `json:"network_out_rate"` // 网络出流量 B/s
}

type VolMetricArgs struct {
	VolReadBytesRate  bool `json:"volume_read_bytes_rate"`     // 云硬盘读速率 B/s
	VolWriteBytesRate bool `json:"volume_write_bytes_rate"`    // 云硬盘写速率 B/s
	VolReadReqRate    bool `json:"volume_read_requests_rate"`  // 云硬盘读请求速率 requests/s
	VolWriteReqRate   bool `json:"volume_write_requests_rate"` // 云硬盘写请求速率 requests/s
}

type FipMetricArgs struct {
	FipBytesInRate   bool `json:"fip.bytes.in.rate"`     // 网络进流量 B/s
	FipBbytesOutRate bool `json:"fip.bytes.out.rate"`    // 网络出流量 B/s
	FipPktsInRate    bool `json:"fip.packages.in.rate"`  // 网络进流量 pkts/s
	FipPktsOutRate   bool `json:"fip.packages.out.rate"` // 网络出流量 pkts/s
}

type MetricArgs struct {
	Srv SrvMetricArgs `json:"instance"`
	Vol VolMetricArgs `json:"volume"`
	Fip FipMetricArgs `json:"fip"`
}

func (p Client) FetchMetrics(l rpc.Logger, args MetricArgs) (ret []Metric, err error) {
	err = p.Conn.CallWithJson(l, &ret, "POST", "/v2/metrics", args)
	if err != nil && fakeError(err) {
		err = nil
	}
	return
}

// --------------------------------------------------
// 获得某个资源某个指标的监控数据

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

type MetricDataArgs struct {
	Period     int
	Duration   string
	ResourceId string
	MetricName string
}

func (p Client) FetchMetricsData(l rpc.Logger, args MetricDataArgs) (ret []MetricData, err error) {
	params := map[string][]string{
		"q.field":  {"resource_id"},
		"q.op":     {"eq"},
		"q.value":  {args.ResourceId},
		"duration": {args.Duration},
		"period":   {strconv.Itoa(args.Period)},
	}
	err = p.Conn.CallWithForm(l, &ret, "GET", "/v2/meters/"+args.MetricName+"/statistics", params)
	if err != nil && fakeError(err) {
		err = nil
	}
	return
}
