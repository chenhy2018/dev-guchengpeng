package ufop

import (
	"github.com/qiniu/rpc.v1"
	"net/http"
	"strconv"
)

const (
	ACCMODE_PUBLIC = iota
	ACCMODE_PROTECTED
	ACCMODE_PRIVATE
)

const (
	ACCMETHOD_UAPP = iota
	ACCMETHOD_URL
)

type Client struct {
	Host string
	Conn rpc.Client
}

func New(host string, t http.RoundTripper) Client {
	client := &http.Client{Transport: t}
	return Client{
		Host: host,
		Conn: rpc.Client{client},
	}
}

type RegisterArgs struct {
	Ufop    string `json:"ufop"`
	AclMode int    `json:"acl_mode"`
	Desc    string `json:"desc"`
}

func (p Client) Register(l rpc.Logger, args *RegisterArgs) (err error) {

	params := map[string][]string{
		"ufop":     {args.Ufop},
		"acl_mode": {strconv.Itoa(args.AclMode)},
		"desc":     {args.Desc},
	}
	return p.Conn.CallWithForm(l, nil, p.Host+"/register", params)
}

func (p Client) Unregister(l rpc.Logger, ufop string) (err error) {

	params := map[string][]string{
		"ufop": {ufop},
	}
	return p.Conn.CallWithForm(l, nil, p.Host+"/unregister", params)
}

func (p Client) Bind(l rpc.Logger, ufop, entry string) (err error) {

	params := map[string][]string{
		"ufop":           {ufop},
		"resource_entry": {entry},
	}
	return p.Conn.CallWithForm(l, nil, p.Host+"/bind", params)
}

func (p Client) Apply(l rpc.Logger, ufop string) (err error) {

	params := map[string][]string{
		"ufop": {ufop},
	}
	return p.Conn.CallWithForm(l, nil, p.Host+"/apply", params)
}

func (p Client) Unapply(l rpc.Logger, ufop string) (err error) {

	params := map[string][]string{
		"ufop": {ufop},
	}
	return p.Conn.CallWithForm(l, nil, p.Host+"/unapply", params)
}

type UfopInfo struct {
	Ufop          string   `json:"ufop"`
	Owner         uint32   `json:"owner"`
	AclMode       byte     `json:"acl_mode"`
	AclList       []uint32 `json:"acl_list"`
	ResourceEntry string   `json:"resource_entry"`
	CreateTime    int64    `json:"create_time"`
	Method        byte     `json:"method"`
	Url           string   `json:"url"`
	Desc          string   `json:"desc"`
}

func (p Client) GetUfopInfo(l rpc.Logger, ufop string) (ui UfopInfo, err error) {

	params := map[string][]string{
		"ufop": {ufop},
	}
	err = p.Conn.CallWithForm(l, &ui, p.Host+"/ufop/info", params)
	return
}

func (p Client) ListUfop(l rpc.Logger) (ufops []string, err error) {

	err = p.Conn.Call(l, &ufops, p.Host+"/list/ufop")
	return
}

func (p Client) GetUfopExist(l rpc.Logger, ufop string) (ret bool, err error) {

	params := map[string][]string{
		"ufop": {ufop},
	}

	err = p.Conn.CallWithForm(l, &ret, p.Host+"/ufop/exist", params)
	return
}

func (p Client) ListSelfUfops(l rpc.Logger) (ufops []UfopInfo, err error) {

	err = p.Conn.Call(l, &ufops, p.Host+"/ufops/self")
	return
}

// 该 API 暂不对外提供
func (p Client) Authorize(l rpc.Logger, ufop string, to uint32) (err error) {

	params := map[string][]string{
		"ufop": {ufop},
		"to":   {strconv.Itoa(int(to))},
	}
	return p.Conn.CallWithForm(l, nil, p.Host+"/authorize", params)
}

// 该 API 暂不对外提供
func (p Client) Unauthorize(l rpc.Logger, ufop string, from uint32) (err error) {

	params := map[string][]string{
		"ufop": {ufop},
		"from": {strconv.Itoa(int(from))},
	}
	return p.Conn.CallWithForm(l, nil, p.Host+"/unauthorize", params)
}

func (p Client) ChangeAclmode(l rpc.Logger, ufop string, aclMode int) (err error) {

	params := map[string][]string{
		"ufop":     {ufop},
		"acl_mode": {strconv.Itoa(aclMode)},
	}
	return p.Conn.CallWithForm(l, nil, p.Host+"/aclmode", params)
}

func (p Client) ChangeDesc(l rpc.Logger, ufop string, desc string) (err error) {

	params := map[string][]string{
		"ufop": {ufop},
		"desc": {desc},
	}
	return p.Conn.CallWithForm(l, nil, p.Host+"/ufop/desc", params)
}

func (p Client) ChangeUrl(l rpc.Logger, ufop string, url string) (err error) {

	params := map[string][]string{
		"ufop": {ufop},
		"url":  {url},
	}
	return p.Conn.CallWithForm(l, nil, p.Host+"/ufop/url", params)
}

func (p Client) ChangeMethod(l rpc.Logger, ufop string, method int) (err error) {

	params := map[string][]string{
		"ufop":   {ufop},
		"method": {strconv.Itoa(method)},
	}
	return p.Conn.CallWithForm(l, nil, p.Host+"/ufop/method", params)
}

//----------------------------------------------------------------------------
type QueryUfopsArgs struct {
	Uid uint32 `json:"uid"`
}

func (p Client) QueryUfops(l rpc.Logger, args *QueryUfopsArgs) (ret []UfopInfo, err error) {
	params := map[string][]string{
		"uid": {strconv.Itoa(int(args.Uid))},
	}
	err = p.Conn.CallWithForm(l, &ret, p.Host+"/ufops/query", params)
	return
}
