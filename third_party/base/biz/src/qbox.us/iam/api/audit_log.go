package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/qiniu/rpc.v1"
	"qbox.us/iam/entity"
)

const (
	ListAuditLogsPath = "/iam/u/%d/audits?%s"
	PostAuditLogsPath = "/iam/audits"
)

type ListAuditLogsInput struct {
	IUID      uint32
	Service   string
	Action    string
	Resource  string
	StartTime time.Time
	EndTime   time.Time
	Marker    string
	Limit     int
}

func (i *ListAuditLogsInput) GetQueryString() string {
	q := &url.Values{}
	q.Set("iuid", strconv.FormatUint(uint64(i.IUID), 10))
	q.Set("service", i.Service)
	q.Set("action", i.Action)
	q.Set("resource", i.Resource)
	q.Set("start_time", formatTime(i.StartTime))
	q.Set("end_time", formatTime(i.EndTime))
	q.Set("marker", i.Marker)
	q.Set("limit", strconv.Itoa(i.Limit))
	return q.Encode()
}

type ListAuditLogsOutput struct {
	List   []*entity.AuditLog `json:"list"`
	Marker string             `json:"marker"`
}

func (c *Client) ListAuditLogs(l rpc.Logger, rootUID uint32, query *ListAuditLogsInput) (out *ListAuditLogsOutput, err error) {
	var output struct {
		CommonResponse `json:",inline"`
		Data           *ListAuditLogsOutput `json:"data"`
	}
	reqPath := fmt.Sprintf(ListAuditLogsPath, rootUID, query.GetQueryString())
	if err = c.client.GetCall(l, &output, reqPath); err != nil {
		return
	}
	out = output.Data
	return
}

type PostAuditLogInput struct {
	RootUID       uint32                 `json:"root_uid"`       // 根用户ID
	IUID          uint32                 `json:"iuid"`           // Iam用户UID
	Service       string                 `json:"service"`        // 服务（产品）
	Action        string                 `json:"action"`         // 动作
	Resources     []string               `json:"resources"`      // 资源
	SourceIP      string                 `json:"source_ip"`      // 发送API请求的源IP地址
	UserAgent     string                 `json:"user_agent"`     // 用户代理
	RequestParams map[string]interface{} `json:"request_params"` // 请求参数
	RequestID     string                 `json:"request_id"`     // 请求ID
	EventTime     time.Time              `json:"event_time"`     // 发出请求的时间
	EventVersion  string                 `json:"event_version"`  // 日志事件格式的版本
	EventSource   string                 `json:"event_source"`   // 处理请求的server
	ErrorCode     int                    `json:"error_code"`     // 错误码
	ErrorMessage  string                 `json:"error_message"`  // 错误信息
}

func (c *Client) PostAuditLogs(l rpc.Logger, logs []PostAuditLogInput) (int, error) {
	buf := bytes.NewBuffer(nil)
	for _, auditLog := range logs {
		err := json.NewEncoder(buf).Encode(auditLog)
		if err != nil {
			return 0, err
		}
		err = buf.WriteByte('\n')
		if err != nil {
			return 0, err
		}
	}
	body := bytes.NewReader(buf.Bytes())
	resp, err := c.client.PostWith(l, PostAuditLogsPath, "text/plain", body, buf.Len())
	if err != nil {
		return 0, err
	}
	ret := struct {
		Count int `json:"count"`
	}{}
	err = json.NewDecoder(resp.Body).Decode(&ret)
	return ret.Count, err
}
