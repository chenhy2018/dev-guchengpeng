package gaea

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/qiniu/rpc.v1"

	"qbox.us/biz/api/gaea/enums"
)

type OpLogService struct {
	host   string
	client rpc.Client
}

func NewOpLogService(host string, t http.RoundTripper) *OpLogService {
	return &OpLogService{
		host: host,
		client: rpc.Client{
			&http.Client{Transport: t},
		},
	}
}

type opLogCreatePayload struct {
	Uid       uint32       `json:"uid"`
	IP        string       `json:"ip"`
	UserAgent string       `json:"user_agent"`
	OpType    enums.OpType `json:"op_type"`
	Op        enums.Op     `json:"op"`
	Extra     string       `json:"extra"`
	Bucket    string       `json:"bucket"`
}

// OpLogService.Create creates oplog for a user operation.
func (s *OpLogService) Create(preq *http.Request, l rpc.Logger, uid uint32, ip, userAgent string, opType enums.OpType, op enums.Op,
	extra, bucket string) (err error) {

	payload := &opLogCreatePayload{
		Uid:       uid,
		IP:        ip,
		UserAgent: userAgent,
		OpType:    opType,
		Op:        op,
		Extra:     extra,
		Bucket:    bucket,
	}

	msg, err := json.Marshal(payload)
	if err != nil {
		return
	}

	cookies := preq.Cookies()

	body := bytes.NewReader(msg)

	req, err := http.NewRequest("POST", s.host+"/api/gaea/oauth/oplog/create", body)
	if err != nil {
		return
	}

	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	var resp CommonResponse
	err = CallWithCookie(req, &resp, l, s.client)
	if err != nil {
		return
	}

	if resp.Code != CodeOK {
		err = errors.New(fmt.Sprintf("OpLogService::Create() failed with code: %d", resp.Code))
		return
	}

	return
}

type OpLog struct {
	Uid        uint32       `json:"uid"`
	Type       enums.OpType `json:"type"`
	Time       time.Time    `json:"time"`
	RemoteAddr string       `json:"remote_addr"`
	UserAgent  string       `json:"user_agent"`
	Op         enums.Op     `json:"op"`
	Extra      string       `json:"extra"`
	Bucket     string       `json:"bucket"`
}

type queryOpLogResp struct {
	CommonResponse

	Logs []*OpLog `json:"data"`
}

// OpLogService.Query querys oplog of certain user and cluster type.
// `bucket` is optional and filters oplog from certain bucket.
func (s *OpLogService) Query(l rpc.Logger, uid uint32, cluster enums.OpTypeCluster, bucket string,
	offset, limit int) (logs []*OpLog, err error) {

	payload := map[string][]string{
		"uid":     []string{strconv.FormatUint(uint64(uid), 10)},
		"cluster": []string{strconv.Itoa(int(cluster))},
		"bucket":  []string{bucket},
		"offset":  []string{strconv.Itoa(offset)},
		"limit":   []string{strconv.Itoa(limit)},
	}

	var resp queryOpLogResp
	err = s.client.GetCallWithForm(l, &resp, s.host+"/api/oauth/oplog/query", payload)
	if err != nil {
		return
	}

	if resp.Code != CodeOK {
		err = errors.New(fmt.Sprintf("OpLogService::Query() failed with code: %d", resp.Code))
		return
	}

	logs = resp.Logs

	return
}
