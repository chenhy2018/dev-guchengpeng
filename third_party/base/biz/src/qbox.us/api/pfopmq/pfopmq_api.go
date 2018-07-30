package pfopmq

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/qiniu/log.v1"
	"github.com/qiniu/rpc.v1"
)

const (
	Zero2OneTag = "zero2one"
)

type UserInfo struct {
	Id    string `bson:"_id"   json:"id"`
	Limit int    `bson:"limit" json:"limit"`
}

type SlaveInfo struct {
	Region string `bson:"region" 	json:"region"`
	Host   string `bson:"host" 		json:"host"`
}

type Pipeline struct {
	Id        string `bson:"_id"   		json:"id"`
	Owner     string `bson:"owner" 		json:"owner"`
	Name      string `bson:"name"  		json:"name"`
	Timestamp int64  `bson:"timestamp" 	json:"timestamp"`
	Action    string `bson:"action" 	json:"action"`
	ActionId  string `bson:"actionId"  	json:"actionId"`
}

type Message struct {
	Id         string    `bson:"_id"         json:"id"`
	PipelineId string    `bson:"pipelineId"  json:"pipelineId"`
	Doby       string    `bson:"doby"        json:"doby"`
	Tags       []string  `bson:"tags"        json:"tags"`
	PutTime    time.Time `bson:"putTime"     json:"putTime"`
	Deadline   time.Time `bson:"deadline"    json:"deadline"`
	State      int32     `bson:"state"       json:"state"`
	Attempts   int32     `bson:"attempts"    json:"attempts"`
	Body       []byte    `bson:"body"        json:"body"`
}

type StatRet struct {
	Todo  int `json:"todo"`
	Doing int `json:"doing"`
}

type PipelineStatTags struct {
	PipelineId string `json:"pipelineId"`
	Todo       int    `json:"todo"`
	Doing      int    `json:"doing"`
	Tags       []Tags `json:"tags"`
}

type PipelineStat struct {
	Todo      int                 `json:"todo"`
	Doing     int                 `json:"doing"`
	Pipelines []*PipelineStatTags `json:"pipelines"`
}

type Tags struct {
	Tag   string `json:"tag"`
	Todo  int    `json:"todo"`
	Doing int    `json:"doing"`
}

type MsgsListRet struct {
	Marker string     `json:"marker,omitempty"`
	Items  []*Message `json:"items"`
}

type IdRet struct {
	Id string `json:"id"`
}

type Client struct {
	Hosts []string
	Conn  rpc.Client
}

func New(hosts []string, t http.RoundTripper) *Client {
	conn := &http.Client{Transport: t, Timeout: time.Second * 5}
	return &Client{hosts, rpc.Client{conn}}
}

func haCall(hosts []string, call func(host string) error) (err error) {
	if len(hosts) == 0 {
		return errors.New("no server available")
	}
	for i, host := range hosts {
		err = call(host)
		if err == nil || !shouldRetry(err) {
			return
		}
		log.Warnf("pfopmq: call [%d][%s] failed: %v\n", i, host, err)
	}
	return
}

// 网络错误需要重试
func shouldRetry(err error) bool {
	if _, ok := err.(*rpc.ErrorInfo); ok {
		return false
	}
	return true
}

func (c *Client) CreatePipeline(l rpc.Logger, zone, uid, name string) (pipelineId string, err error) {
	var ret IdRet
	err = haCall(c.Hosts, func(host string) error {
		params := map[string][]string{"zone": {zone}, "uid": {uid}, "name": {name}}
		return c.Conn.CallWithForm(l, &ret, host+"/pipeline/new", params)
	})
	return ret.Id, err
}

func (c *Client) RemovePipeline(l rpc.Logger, pipelineId string) (err error) {
	err = haCall(c.Hosts, func(host string) error {
		params := map[string][]string{"id": {pipelineId}}
		return c.Conn.CallWithForm(l, nil, host+"/pipeline/rm", params)
	})
	return
}

func (c *Client) StatPipeline(l rpc.Logger, pipelineId string) (ret StatRet, err error) {
	err = haCall(c.Hosts, func(host string) error {
		params := map[string][]string{"id": {pipelineId}}
		return c.Conn.CallWithForm(l, &ret, host+"/pipeline/stat", params)
	})
	return
}

func (c *Client) GlobalStatPipeline(l rpc.Logger, pipelineId string) (ret StatRet, err error) {
	err = haCall(c.Hosts, func(host string) error {
		params := map[string][]string{"id": {pipelineId}}
		return c.Conn.CallWithForm(l, &ret, host+"/glb/pipeline/stat", params)
	})
	return
}

// due to some terrible func name. this func return stat with tags
func (c *Client) StatPipelineWithTag(l rpc.Logger, todo string) (ret PipelineStat, err error) {
	err = haCall(c.Hosts, func(host string) error {
		params := map[string][]string{"todo": {todo}}
		return c.Conn.CallWithForm(l, &ret, host+"/stat/pipeline", params)
	})

	return
}

func (c *Client) EmptyPipeline(l rpc.Logger, pipelineId string) (err error) {
	err = haCall(c.Hosts, func(host string) error {
		params := map[string][]string{"id": {pipelineId}}
		return c.Conn.CallWithForm(l, nil, host+"/pipeline/empty", params)
	})
	return
}

func (c *Client) GetPipeline(l rpc.Logger, owner, name string) (info Pipeline, err error) {
	err = haCall(c.Hosts, func(host string) error {
		params := map[string][]string{"uid": {owner}, "name": {name}}
		return c.Conn.CallWithForm(l, &info, host+"/pipeline/get", params)
	})
	return
}

func (c *Client) ListPipelines(l rpc.Logger, zone, uid string) (ret []*Pipeline, err error) {
	err = haCall(c.Hosts, func(host string) error {
		params := map[string][]string{"zone": {zone}, "uid": {uid}}
		return c.Conn.CallWithForm(l, &ret, host+"/pipeline/ls", params)
	})
	return
}

func (c *Client) PutMessage(l rpc.Logger, pipelineId string, msg []byte) (msgId string, err error) {
	var ret IdRet
	err = haCall(c.Hosts, func(host string) error {
		return c.putMessage(l, &ret, host, pipelineId, nil, bytes.NewReader(msg), len(msg))
	})
	return ret.Id, err
}

func (c *Client) PutMessageStr(l rpc.Logger, pipelineId string, msg string) (msgId string, err error) {
	var ret IdRet
	err = haCall(c.Hosts, func(host string) error {
		return c.putMessage(l, &ret, host, pipelineId, nil, strings.NewReader(msg), len(msg))
	})
	return ret.Id, err
}

func (c *Client) PutMessageEx(l rpc.Logger, pipelineId string, tags []string, msg []byte) (msgId string, err error) {
	var ret IdRet
	err = haCall(c.Hosts, func(host string) error {
		return c.putMessage(l, &ret, host, pipelineId, tags, bytes.NewReader(msg), len(msg))
	})
	return ret.Id, err
}

func (c *Client) putMessage(l rpc.Logger, ret interface{}, host string, pipelineId string, tags []string, msgr io.Reader, bytes int) (err error) {
	v := url.Values{}
	v.Set("pipelineId", pipelineId)
	for _, tag := range tags {
		v.Add("tags", tag)
	}
	u := host + "/message/put?" + v.Encode()
	return c.Conn.CallWith(l, ret, u, "application/octet-stream", msgr, bytes)
}

func (c *Client) GetMessage(l rpc.Logger) (msg []byte, msgId string, pipelineId string, err error) {
	err = haCall(c.Hosts, func(host string) error {
		msg, msgId, pipelineId, err = c.getMessage(l, host, "")
		return err
	})
	return
}

func (c *Client) GetMessageEx(l rpc.Logger, tag string) (msg []byte, msgId string, pipelineId string, err error) {
	err = haCall(c.Hosts, func(host string) error {
		msg, msgId, pipelineId, err = c.getMessage(l, host, tag)
		return err
	})
	return
}

func (c *Client) getMessage(l rpc.Logger, host string, tag string) (msg []byte, msgId string, pipelineId string, err error) {
	u := host + "/message/get"
	if tag != "" {
		v := url.Values{}
		v.Add("tags", tag)
		u += "?" + v.Encode()
	}
	resp, err := c.Conn.Get(l, u)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		err = rpc.ResponseError(resp)
		return
	}

	msgId = resp.Header.Get("X-Id")
	pipelineId = resp.Header.Get("X-PipelineId")
	msg, err = ioutil.ReadAll(resp.Body)
	return
}

func (c *Client) RemoveMessage(l rpc.Logger, pipeline, tag, msgId string) (err error) {
	err = haCall(c.Hosts, func(host string) error {
		params := map[string][]string{"id": {msgId}, "tag": {tag}, "pipeline": {pipeline}}
		return c.Conn.CallWithForm(l, nil, host+"/message/rm", params)
	})
	return
}

func (c *Client) PingMessage(l rpc.Logger, msgId string) (err error) {

	err = haCall(c.Hosts, func(host string) error {
		params := map[string][]string{"id": {msgId}}
		return c.Conn.CallWithForm(l, nil, host+"/message/ping", params)
	})
	return
}

func (c *Client) StatMessage(l rpc.Logger, msgId string) (ret Message, err error) {
	err = haCall(c.Hosts, func(host string) error {
		params := map[string][]string{"id": {msgId}}
		return c.Conn.CallWithForm(l, &ret, host+"/message/stat", params)
	})
	return
}

func (c *Client) ListMessages(l rpc.Logger, pipelineId string, marker string, limit int) (ret *MsgsListRet, err error) {
	err = haCall(c.Hosts, func(host string) error {
		params := map[string][]string{
			"pipelineId": {pipelineId},
			"marker":     {marker},
			"limit":      {strconv.Itoa(limit)},
		}
		return c.Conn.CallWithForm(l, &ret, host+"/message/ls", params)
	})
	return
}

func (c *Client) Stat(l rpc.Logger) (ret StatRet, err error) {
	err = haCall(c.Hosts, func(host string) error {
		return c.Conn.Call(l, &ret, host+"/stat")
	})
	return
}
