package cc

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"qbox.us/qcc/uapp"

	"github.com/qiniu/rpc.v1"
	"github.com/qiniu/rpc.v1/lb.v2"
)

type Client struct {
	Conn *lb.Client
}

type Item struct {
	State string `json:"state"`
}

type UappInfo struct {
	Name       string `json:"name"`
	CreateTime int64  `json:"create_time"`
	Desc       string `json:"desc"`

	ImageURL string `json:"image_url"`
	ImgVer   int    `json:"img_ver"`

	ReqType uint32 `json:"req_type"`
	SrvType uint32 `json:"srv_type"`
	AclMode byte   `json:"acl_mode"`

	InstQuota uint32        `json:"inst_quota"`
	RsCap     uapp.Resource `json:"rs_cap"`
}

type OpRet struct {
	State uapp.InsState `json:"state"`
	Error string        `json:"error"`
}

var bodyType = "application/x-www-form-urlencoded"

type InstanceStatus map[string]OpRet

func shouldRetry(code int, err error) bool {
	if code == http.StatusServiceUnavailable {
		return true
	}
	return lb.ShouldRetry(code, err)
}

func New(hosts []string, tr http.RoundTripper) (c Client, err error) {

	cfg := &lb.Config{
		Http:              &http.Client{Transport: tr},
		ShouldRetry:       shouldRetry,
		FailRetryInterval: -1,
		TryTimes:          uint32(len(hosts)),
	}
	conn, err := lb.New(hosts, cfg)
	if err != nil {
		return
	}
	c = Client{Conn: conn}
	return
}

func (cc *Client) Mkapp(l rpc.Logger, uapp, desc, aclMode, reqType, srvType string) (err error) {

	params := url.Values(
		map[string][]string{
			"name":     {uapp},
			"req_type": {reqType},
			"srv_type": {srvType},
			"acl_mode": {aclMode},
			"desc":     {desc},
		}).Encode()

	return cc.Conn.CallWith(l, nil, "/mkapp", bodyType, strings.NewReader(params), len(params))
}

func (cc *Client) DropUapp(l rpc.Logger, uapp string) (err error) {

	params := url.Values(
		map[string][]string{
			"name": {uapp},
		}).Encode()

	return cc.Conn.CallWith(l, nil, "/drop", bodyType, strings.NewReader(params), len(params))
}

func (cc *Client) Bind(l rpc.Logger, uapp, imageurl, bucket, key string) (err error) {

	if imageurl == "" {
		imageurl = "qiniu:" + bucket + ":" + key
	}

	params := url.Values(
		map[string][]string{
			"name":      {uapp},
			"image_url": {imageurl},
		}).Encode()

	return cc.Conn.CallWith(l, nil, "/bind", bodyType, strings.NewReader(params), len(params))
}

func (cc *Client) InstnQuota(l rpc.Logger, uapp string, quota int) (err error) {

	params := url.Values(
		map[string][]string{
			"name":  {uapp},
			"quota": {strconv.Itoa(quota)},
		}).Encode()

	return cc.Conn.CallWith(l, nil, "/quota", bodyType, strings.NewReader(params), len(params))
}

func (cc *Client) SetCap(l rpc.Logger, uapp string, mem, cpu, disk uint64) (err error) {

	memstr := strconv.FormatUint(mem, 10)
	cpustr := strconv.FormatUint(cpu, 10)
	diskstr := strconv.FormatUint(disk, 10)

	params := url.Values(map[string][]string{
		"name": {uapp},
		"mem":  {memstr},
		"cpu":  {cpustr},
		"disk": {diskstr},
	}).Encode()

	return cc.Conn.CallWith(l, nil, "/cap", bodyType, strings.NewReader(params), len(params))
}

func (cc *Client) StartUapp(l rpc.Logger, uapp, idx string) (state InstanceStatus, err error) {

	params := url.Values(
		map[string][]string{
			"name": {uapp},
			"idx":  {idx},
		}).Encode()

	err = cc.Conn.CallWith(l, &state, "/start", bodyType, strings.NewReader(params), len(params))
	return
}

func (cc *Client) StopUapp(l rpc.Logger, uapp, idx string) (state InstanceStatus, err error) {

	params := url.Values(
		map[string][]string{
			"name": {uapp},
			"idx":  {idx},
		}).Encode()

	err = cc.Conn.CallWith(l, &state, "/stop", bodyType, strings.NewReader(params), len(params))
	return
}

func (cc *Client) StateUapp(l rpc.Logger, uapp, idx string) (state InstanceStatus, err error) {

	params := url.Values(
		map[string][]string{
			"name": {uapp},
			"idx":  {idx},
		}).Encode()

	err = cc.Conn.CallWith(l, &state, "/state", bodyType, strings.NewReader(params), len(params))
	return
}

func (cc *Client) InfoUapp(l rpc.Logger, uapp string) (uappinfo UappInfo, err error) {

	params := url.Values(
		map[string][]string{
			"name": {uapp},
		}).Encode()

	err = cc.Conn.CallWith(l, &uappinfo, "/info", bodyType, strings.NewReader(params), len(params))
	return
}

// <Host> <Disk> <Mem> <Iops> <Net> <Cpu>
func (cc *Client) NewHouse(l rpc.Logger, houseHost, disk, mem, iops, net, cpu string) (err error) {

	params := url.Values(
		map[string][]string{
			"host": {houseHost},
			"disk": {disk},
			"mem":  {mem},
			"iops": {iops},
			"net":  {net},
			"cpu":  {cpu},
		}).Encode()
	return cc.Conn.CallWith(l, nil, "/newhouse", bodyType, strings.NewReader(params), len(params))
}

func (cc *Client) DelHouse(l rpc.Logger, houseHost string) (err error) {

	params := url.Values(
		map[string][]string{
			"host": {houseHost},
		}).Encode()

	return cc.Conn.CallWith(l, nil, "/delhouse", bodyType, strings.NewReader(params), len(params))
}
