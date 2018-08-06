package rsf

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/qiniu/rpc.v1"
	"github.com/qiniu/xlog.v1"
	"golang.org/x/net/context"
	"qbox.us/api"
	"qbox.us/net/httputil"
)

const (
	InvalidMarker  = 640
	TimeoutMarker  = 641
	DumpLimitUpper = 1000
	ListLimitUpper = 1000
)

var (
	EInvaildMarker = api.RegisterError(InvalidMarker, "invalid marker")
)

type Service struct {
	conn rpc.Client
	host string
}

func New(t http.RoundTripper, host string) *Service {
	client := &http.Client{Transport: t}
	return &Service{rpc.Client{client}, host}
}

type DeletedMarkerPolicy int

const (
	DeletedMarkerExcluded = DeletedMarkerPolicy(iota)
	DeletedMarkerOnly
	DeletedMarkerIncluded
)

type ListParam struct {
	Bucket              string              `json:"bucket"`
	Prefix              string              `json:"prefix"`
	Delimiter           string              `json:"delimiter"`
	Limit               int                 `json:"limit,default"`
	Marker              string              `json:"marker"`
	DeletedMarkerPolicy DeletedMarkerPolicy `json:"deleted_marker_policy"`
}

type DumpItem struct {
	Fsize   int64  `json:"fsize" bson:"fsize"`
	Time    int64  `json:"putTime" bson:"putTime"`
	Name    string `json:"key" bson:"key"`
	Hash    string `json:"hash" bson:"hash"`
	Mime    string `json:"mimeType,omitempty" bson:"mimeType,omitempty"`
	EndUser string `json:"endUser,omitempty" bson:"endUser,omitempty"`
	Type    int    `json:"type,omitempty" bson:"type,omitempty"`
}

type DumpRet struct {
	Marker         string     `json:"marker,omitempty"`
	Items          []DumpItem `json:"items"`
	CommonPrefixes []string   `json:"commonPrefixes,omitempty"`
}

func (rsf Service) ListPrefix(xl *xlog.Logger, bucketName, prefix, markerIn string, limit int) (ret DumpRet, err error) {

	u := rsf.host + "/list?bucket=" + bucketName
	if markerIn != "" {
		u += "&marker=" + markerIn
	}
	if prefix != "" {
		u += "&prefix=" + url.QueryEscape(prefix)
	}
	if limit > 0 {
		u += "&limit=" + strconv.Itoa(limit)
	}
	err = rsf.conn.Call(xl, &ret, u)
	return
}

func (rsf Service) ListPrefix1(xl *xlog.Logger, param ListParam) (ret DumpRet, err error) {

	u := rsf.host + "/list?bucket=" + param.Bucket
	if param.Marker != "" {
		u += "&marker=" + param.Marker
	}
	if param.Prefix != "" {
		u += "&prefix=" + url.QueryEscape(param.Prefix)
	}
	if param.Limit > 0 {
		u += "&limit=" + strconv.Itoa(param.Limit)
	}
	if param.Delimiter != "" {
		u += "&delimiter=" + url.QueryEscape(param.Delimiter)
	}
	u += "&deleted_marker_policy=" + strconv.Itoa(int(param.DeletedMarkerPolicy))
	err = rsf.conn.Call(xl, &ret, u)
	return
}

type V2ListItem struct {
	Item   *DumpItem `json:"item" bson:"item"`     // 返回的数据 Item 可能为 nil, 这种情况客户端应该把 Marker 记录下来，下次请求传过去
	Marker string    `json:"marker" bson:"marker"` // 每个数据都有 Marker
	Dir    string    `json:"dir" bson:"dir"`       // 只在请求参数里面指定 Delimiter 的时候才有意义
}

func (rsf Service) V2ListPrefixOnce(xl *xlog.Logger, ctx context.Context, param ListParam) (retCh chan *V2ListItem, shouldRetry bool, err error) {
	retCh = make(chan *V2ListItem, 50)
	u := rsf.host + "/v2/list?bucket=" + param.Bucket
	if param.Marker != "" {
		u += "&marker=" + param.Marker
	}
	if param.Prefix != "" {
		u += "&prefix=" + url.QueryEscape(param.Prefix)
	}
	if param.Limit >= 0 {
		u += "&limit=" + strconv.Itoa(param.Limit)
	}
	if param.Delimiter != "" {
		u += "&delimiter=" + url.QueryEscape(param.Delimiter)
	}
	u += "&deleted_marker_policy=" + strconv.Itoa(int(param.DeletedMarkerPolicy))
	resp, err := rsf.conn.Get(xl, u)
	if err != nil {
		xl.Warn(err)
		return
	}
	if resp.StatusCode != 200 {
		err = httputil.ResponseError(resp)
		resp.Body.Close()
		return
	}
	shouldRetry = true
	go func() {
		defer resp.Body.Close()
		defer close(retCh)
		dec := json.NewDecoder(resp.Body)
		for {
			var ret V2ListItem
			err := dec.Decode(&ret)
			if err != nil {
				if err != io.EOF {
					xl.Warn(err)
				}
				return
			}
			select {
			case retCh <- &ret:
			case <-ctx.Done():
				return
			}
		}
	}()
	return
}
func (rsf Service) V2ListPrefixRetry(xl *xlog.Logger, ctx context.Context, param ListParam) (retCh chan *V2ListItem, err error) {
	retCh = make(chan *V2ListItem, 50)
	var firstError = make(chan error, 1)
	go func() {
		defer close(retCh)
		var count int
		lastMarker := param.Marker
		if lastMarker == "" {
			lastMarker = "x"
		}
		for lastMarker != "" && ctx.Err() == nil && (count < param.Limit || param.Limit == 0) {
			if lastMarker != "x" {
				param.Marker = lastMarker
			}
			rets, shouldRetry, err := rsf.V2ListPrefixOnce(xl, ctx, param)
			select {
			case firstError <- err:
			default:
			}
			if err != nil {
				xl.Warn(err, shouldRetry)
				if !shouldRetry {
					return
				}
				time.Sleep(time.Second)
				continue
			}
			for ret := range rets {
				lastMarker = ret.Marker
				select {
				case retCh <- ret:
					if ret.Item != nil || ret.Dir != "" {
						count++
					}
				case <-ctx.Done():
					return
				}
			}
			if lastMarker == "x" {
				lastMarker = ""
			}
		}
	}()
	err = <-firstError
	return
}

type DumpItemAdmin struct {
	Fsize   int64  `json:"fsize" bson:"fsize"`
	Time    int64  `json:"putTime" bson:"putTime"`
	Name    string `json:"key" bson:"key"`
	Hash    string `json:"hash" bson:"hash"`
	Mime    string `json:"mimeType,omitempty" bson:"mimeType,omitempty"`
	EndUser string `json:"endUser,omitempty" bson:"endUser,omitempty"`
	Fh      []byte `json:"fh" bson:"fh"`
}

type DumpRetAdmin struct {
	Marker         string          `json:"marker,omitempty"`
	Items          []DumpItemAdmin `json:"items"`
	CommonPrefixes []string        `json:"commonPrefixes,omitempty"`
}

func (rsf Service) AdminListByItblPrefix(xl *xlog.Logger, region string, itbl uint32, prefix, markerIn string, limit int) (ret DumpRetAdmin, err error) {

	u := rsf.host + "/admin/listbyitbl?region=" + region

	u += "&itbl=" + strconv.FormatUint(uint64(itbl), 10)

	if markerIn != "" {
		u += "&marker=" + markerIn
	}
	if prefix != "" {
		u += "&prefix=" + url.QueryEscape(prefix)
	}
	if limit > 0 {
		u += "&limit=" + strconv.Itoa(limit)
	}
	err = rsf.conn.Call(xl, &ret, u)
	return
}
