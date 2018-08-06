package uc

import (
	"encoding/json"
	"net/http"
	"strconv"

	"fmt"

	"github.com/qiniu/http/httputil.v1"
	"github.com/qiniu/rpc.v1"
	"github.com/qiniu/rpc.v1/lb.v2.1"
	"qbox.us/admin_api/uc"
	"qbox.us/api"
)

type Service struct {
	Conn *lb.Client
}

func New(host string, t http.RoundTripper) *Service {
	cfg := &lb.Config{
		Hosts:    []string{host},
		TryTimes: 1,
	}
	client := lb.New(cfg, t)
	return &Service{client}
}

func NewWithMultiHosts(hosts []string, t http.RoundTripper) *Service {
	cfg := &lb.Config{
		Hosts:    hosts,
		TryTimes: uint32(len(hosts)),
	}
	client := lb.New(cfg, t)
	return &Service{client}
}

// POST /copyBucket
// Content-Type: application/x-www-form-urlencoded
//
// srcOwner=<srcOwner>&srcTbl=<srcTbl>&dstOwner=<dstOwner>&dstTbl=<dstTbl>
func (r *Service) CopyBucketinfo(l rpc.Logger, srcOwner uint32, srcTbl string, dstOwner uint32, dstTbl string) (err error) {

	err = r.Conn.CallWithForm(l, nil, "/copyBucket", map[string][]string{
		"srcOwner": {strconv.FormatUint(uint64(srcOwner), 10)},
		"srcTbl":   {srcTbl},
		"dstOwner": {strconv.FormatUint(uint64(dstOwner), 10)},
		"dstTbl":   {dstTbl},
	})
	return
}

/*
	POST /get?name=<GroupName>&name=<keyName>
*/
func (r *Service) Get(l rpc.Logger, grp, name string) (val string, err error) {

	var ret struct {
		Val string `json:"val"`
	}
	err = r.Conn.CallWithForm(l, &ret, "/get", map[string][]string{
		"group": {grp},
		"name":  {name},
	})
	val = ret.Val
	return
}

/*
	POST /set?group=<GroupName>&name=<KeyName>&val=<Value>
*/
func (r *Service) Set(l rpc.Logger, grp, name string, val1 interface{}) (err error) {

	val, ok := val1.(string)
	if !ok {
		b, err2 := json.Marshal(val1)
		if err2 != nil {
			return httputil.NewError(api.FunctionFail, err2.Error())
		}
		val = string(b)
	}
	err = r.Conn.CallWithForm(l, nil, "/set", map[string][]string{
		"group": {grp},
		"name":  {name},
		"val":   {val},
	})
	return
}

/*
	POST /delete?group=<GroupName>&name=<KeyName>
*/
func (r *Service) Delete(l rpc.Logger, grp, name string) (err error) {

	err = r.Conn.CallWithForm(l, nil, "/delete", map[string][]string{
		"group": {grp},
		"name":  {name},
	})
	return
}

/*
	POST /group?name=<GroupName>
*/
func (r *Service) Group(l rpc.Logger, name string) (items []uc.GroupItem, err error) {

	err = r.Conn.CallWithForm(l, &items, "/group", map[string][]string{
		"name": {name},
	})
	return
}

/*
	POST /setUtype?uid=<Uid>&utype=<Utype>
*/
func (r *Service) SetUtype(l rpc.Logger, uid, utype string) (err error) {

	err = r.Conn.CallWithForm(l, nil, "/setUtype", map[string][]string{
		"uid":   {uid},
		"utype": {utype},
	})
	return
}

//
// POST /bindQueue/<BucketName>/uid/<Uid>
//
func (r *Service) BindQueue(l rpc.Logger, uid int64, bucketName, notifyQueue, notifyMessage, notifyMessageType string) (err error) {
	err = r.Conn.CallWithJson(l, nil,
		fmt.Sprintf("/bindQueue/%s/uid/%d", bucketName, uid),
		KmqMsg{
			NotifyQueue:       notifyQueue,
			NotifyMessage:     notifyMessage,
			NotifyMessageType: notifyMessageType,
		})
	return
}

//
// POST /unbindQueue/<BucketName>/uid/<Uid>
//
func (r *Service) UnbindQueue(l rpc.Logger, uid int64, bucketName string) (err error) {
	err = r.Conn.CallWithJson(l, nil,
		fmt.Sprintf("/unbindQueue/%s/uid/%d", bucketName, uid),
		KmqMsg{
			NotifyQueue:       "",
			NotifyMessage:     "",
			NotifyMessageType: "",
		})
	return
}

type KmqMsg struct {
	NotifyQueue       string `json:"notify_queue"`
	NotifyMessage     string `json:"notify_message"`
	NotifyMessageType string `json:"notify_message_type"`
}

//
// GET /getQueue/<BucketName>/uid/<Uid>
//
func (r *Service) GetQueue(l rpc.Logger, uid int64, bucketName string) (notifyQueue, notifyMessage, notifyMessageType string, err error) {
	var kmqMsg KmqMsg
	err = r.Conn.GetCall(l, &kmqMsg, fmt.Sprintf("/getQueue/%s/uid/%d", bucketName, uid))
	return kmqMsg.NotifyQueue, kmqMsg.NotifyMessage, kmqMsg.NotifyMessageType, err
}
