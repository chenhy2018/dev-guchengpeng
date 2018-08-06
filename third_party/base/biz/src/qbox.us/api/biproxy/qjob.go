package biproxy

import (
	"net/url"
)

import (
	"github.com/qiniu/rpc.v1"
)

type HandleQjob struct {
	Host   string
	Client *rpc.Client
}

func NewHandleQjob(host string, client *rpc.Client) *HandleQjob {
	return &HandleQjob{host, client}
}

func (r HandleQjob) TaskList(logger rpc.Logger, req ReqTaskQuery) (status []TaskStatus, err error) {
	value := url.Values{}
	if req.Day != nil {
		value.Add("day", *req.Day)
	}
	if req.Status != nil {
		value.Add("status", *req.Status)
	}
	err = r.Client.Call(logger, &status, r.Host+"/qjob/task/list?"+value.Encode())
	return
}

func (r HandleQjob) OpList(logger rpc.Logger, req ReqOpQuery) (ops []Ops, err error) {
	value := url.Values{}
	value.Add("task", req.Task)
	value.Add("day", req.Day)
	err = r.Client.Call(logger, &ops, r.Host+"/qjob/op/list?"+value.Encode())
	return
}

// Task && Operation handle function
func (r HandleQjob) OpDone(logger rpc.Logger, req ReqOperation) (err error) {
	value := url.Values{}
	value.Add("day", req.Day)
	value.Add("key", req.Key)
	err = r.Client.CallWithForm(logger, nil, r.Host+"/qjob/op/done", map[string][]string(value))
	return
}

func (r HandleQjob) TaskReset(logger rpc.Logger, req ReqTask) (err error) {
	value := url.Values{}
	value.Add("task", req.Task)
	value.Add("day", req.Day)
	err = r.Client.CallWithForm(logger, nil, r.Host+"/qjob/task/reset", map[string][]string(value))
	return
}

func (r HandleQjob) TaskCancel(logger rpc.Logger, req ReqTask) (err error) {
	value := url.Values{}
	value.Add("task", req.Task)
	value.Add("day", req.Day)
	err = r.Client.CallWithForm(logger, nil, r.Host+"/qjob/task/cancel", map[string][]string(value))
	return
}

func (r HandleQjob) TaskStop(logger rpc.Logger, req ReqTask) (err error) {
	value := url.Values{}
	value.Add("task", req.Task)
	value.Add("day", req.Day)
	err = r.Client.CallWithForm(logger, nil, r.Host+"/qjob/task/stop", map[string][]string(value))
	return
}

func (r HandleQjob) TaskRestart(logger rpc.Logger, req ReqTask) (err error) {
	value := url.Values{}
	value.Add("task", req.Task)
	value.Add("day", req.Day)
	err = r.Client.CallWithForm(logger, nil, r.Host+"/qjob/task/restart", map[string][]string(value))
	return
}
