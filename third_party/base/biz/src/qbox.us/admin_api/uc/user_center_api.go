package uc

import (
	"encoding/json"
	"net/http"
	"qbox.us/api"
	"qbox.us/rpc"
)

type Service struct {
	Host string
	Conn rpc.Client
}

func New(host string, t http.RoundTripper) *Service {
	client := &http.Client{Transport: t}
	return &Service{host, rpc.Client{client}}
}

/*
	POST /get?name=<GroupName>&name=<keyName>
*/
func (r *Service) Get(grp, name string) (val string, code int, err error) {

	var ret struct {
		Val string `json:"val"`
	}
	code, err = r.Conn.CallWithForm(&ret, r.Host+"/get", map[string][]string{
		"group":{grp},
		"name": {name},
	})
	val = ret.Val
	return
}

/*
	POST /set?group=<GroupName>&name=<KeyName>&val=<Value>
*/
func (r *Service) Set(grp, name string, val1 interface{}) (code int, err error) {

	val, ok := val1.(string)
	if !ok {
		b, err2 := json.Marshal(val1)
		if err2 != nil {
			return api.FunctionFail, err2
		}
		val = string(b)
	}
	code, err = r.Conn.CallWithForm(nil, r.Host+"/set", map[string][]string{
		"group": {grp},
		"name":  {name},
		"val":   {val},
	})
	return
}

/*
	POST /delete?group=<GroupName>&name=<KeyName>
*/
func (r *Service) Delete(grp, name string) (code int, err error) {

	code, err = r.Conn.CallWithForm(nil, r.Host+"/delete", map[string][]string{
		"group": {grp},
		"name":  {name},
	})
	return
}

type GroupItem struct {
	Key string `json:"key" bson:"key"`
	Val string `json:"val" bson:"val"`
}

/*
	POST /group?name=<GroupName>
*/
func (r *Service) Group(name string) (items []GroupItem, code int, err error) {

	code, err = r.Conn.CallWithForm(&items, r.Host+"/group", map[string][]string{
		"name": {name},
	})
	return
}

/*
	POST /setUtype?uid=<Uid>&utype=<Utype>
*/
func (r *Service) SetUtype(uid, utype string) (code int, err error) {

	code, err = r.Conn.CallWithForm(nil, r.Host+"/setUtype", map[string][]string{
		"uid":   {uid},
		"utype": {utype},
	})
	return
}
