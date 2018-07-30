package api

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/qiniu/rpc.v1"
	"github.com/qiniu/rpc.v1/lb.v2.1"

	"gopkg.in/mgo.v2/bson"
)

type DiskType int

const (
	DEFAULT DiskType = iota
	SSD
)

type DiskGroupInfo struct {
	Guid     string      `json:"guid" bson:"guid"`
	Dgid     uint32      `json:"dgid" bson:"dgid"`
	Hosts    [][2]string `json:"hosts" bson:"hosts"`
	IsAsync  []bool      `json:"is_async" bson:"is_async"`
	IsBackup []bool      `json:"is_backup" bson:"is_backup"`
	Idc      []string    `json:"idc" bson:"idc"`
	Path     []string    `json:"path" bson:"path"`
	ReadOnly uint32      `json:"readonly" bson:"readonly"`
	DiskType DiskType    `json:"disk_type" bson:"disk_type"`
	Weight   uint32      `json:"weight" bson:"weight"`
	Repair   []bool      `json:"repair" bson:"repair"`
}

type DiskInfo struct {
	Guid     string    `json:"guid"`
	Dgid     uint32    `json:"dgid"`
	Host     [2]string `json:"host"`
	Idc      string    `json:"idc"`
	Path     string    `json:"path"`
	DiskType DiskType  `json:"disk_type"`
	GroupIdx int       `json:"group_idx"`
	GroupCnt int       `json:"group_cnt"`
}

type Client struct {
	conn *lb.Client
}

func shouldRetry(code int, err error) bool {
	if code == http.StatusServiceUnavailable {
		return true
	}
	return lb.ShouldRetry(code, err)
}

func New(hosts []string, tr http.RoundTripper) (c Client, err error) {

	cfg := &lb.Config{
		Hosts:              hosts,
		ShouldRetry:        shouldRetry,
		FailRetryIntervalS: -1,
		TryTimes:           uint32(len(hosts)),
	}

	conn := lb.New(cfg, tr)
	if err != nil {
		return
	}
	return Client{conn: conn}, nil
}

func (self Client) AllWritableDgs(l rpc.Logger, guid string) (dgInfos []*DiskGroupInfo, err error) {

	err = self.conn.CallWithForm(l, &dgInfos, "/dgs", map[string][]string{
		"guid":     {guid},
		"readonly": {"0"},
	})
	return
}

func (self *Client) IdcDgs(l rpc.Logger, guid, idc string) (dgs []*DiskGroupInfo, err error) {

	err = self.conn.CallWithForm(l, &dgs, "/dgs", map[string][]string{
		"guid": {guid},
		"idc":  {idc},
	})
	return
}

func (self *Client) ListBroken(l rpc.Logger, guid string, repair bool) (dgs []*DiskGroupInfo, err error) {
	err = self.conn.CallWithForm(l, &dgs, "/dgs", map[string][]string{
		"guid":   {guid},
		"repair": {strconv.FormatBool(repair)},
	})
	return
}

func (self Client) AllDgs(l rpc.Logger, guid string) (dgInfos []*DiskGroupInfo, err error) {

	err = self.conn.CallWithForm(l, &dgInfos, "/dgs", map[string][]string{
		"guid": {guid},
	})
	return
}

func (self Client) DGInfo(l rpc.Logger, guid string, dgid uint32) (dgInfo *DiskGroupInfo, err error) {

	m := make(url.Values)
	m.Add("id", "dg:"+guid+":"+strconv.FormatUint(uint64(dgid), 36))
	body := m.Encode()
	req, err := lb.NewRequest("POST", "/getb", strings.NewReader(body))
	if err != nil {
		return
	}
	req.Header.Set("Content-Length", strconv.Itoa(len(body)))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := self.conn.Do(l, req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		err = rpc.ResponseError(resp)
		return
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	dgInfo = new(DiskGroupInfo)
	err = bson.Unmarshal(b, dgInfo)
	if err != nil {
		return
	}
	return
}

func (self Client) Dgs(l rpc.Logger, guid string, hosts [2]string) (dgInfos []*DiskGroupInfo, err error) {

	err = self.conn.CallWithForm(l, &dgInfos, "/dgs", map[string][]string{
		"guid":   {guid},
		"host":   {hosts[0]},
		"echost": {hosts[1]},
	})
	return
}

func (self Client) DgStore(l rpc.Logger, dgInfo *DiskGroupInfo) (err error) {

	err = self.conn.CallWithJson(l, nil, "/dg/store", dgInfo)
	return
}

func (self Client) DgAccess(l rpc.Logger, guid string, dgids []uint32, readOnly uint32) (err error) {

	dgidStrs := make([]string, len(dgids))
	for i, dgid := range dgids {
		dgidStrs[i] = strconv.FormatUint(uint64(dgid), 10)
	}

	err = self.conn.CallWithForm(l, nil, "/dg/access",
		map[string][]string{
			"guid":     {guid},
			"dgid":     dgidStrs,
			"readonly": {strconv.FormatUint(uint64(readOnly), 10)},
		})
	return
}

func (self Client) DgWeight(l rpc.Logger, guid string, dgids []uint32, weight uint32) (err error) {

	dgidStrs := make([]string, len(dgids))
	for i, dgid := range dgids {
		dgidStrs[i] = strconv.FormatUint(uint64(dgid), 10)
	}

	err = self.conn.CallWithForm(l, nil, "/dg/weight",
		map[string][]string{
			"guid":   {guid},
			"dgid":   dgidStrs,
			"weight": {strconv.FormatUint(uint64(weight), 10)},
		})
	return
}

func (self Client) DgBackup(l rpc.Logger, guid string, dgid uint32, backup bool, grpIdx int) (err error) {

	err = self.conn.CallWithForm(l, nil, "/dg/backup",
		map[string][]string{
			"guid":   {guid},
			"dgid":   {strconv.FormatUint(uint64(dgid), 10)},
			"backup": {strconv.FormatBool(backup)},
			"grpIdx": {strconv.Itoa(grpIdx)},
		})
	return
}

func (self Client) DgRepair(l rpc.Logger, guid string, dgid uint32, repair bool, grpIdx int) (err error) {

	err = self.conn.CallWithForm(l, nil, "/dg/repair",
		map[string][]string{
			"guid":   {guid},
			"dgid":   {strconv.FormatUint(uint64(dgid), 10)},
			"repair": {strconv.FormatBool(repair)},
			"grpIdx": {strconv.Itoa(grpIdx)},
		})
	return
}

func (self Client) DiskRegister(l rpc.Logger, diskInfo *DiskInfo) (err error) {

	err = self.conn.CallWithJson(l, nil, "/disk/register", diskInfo)
	return
}
