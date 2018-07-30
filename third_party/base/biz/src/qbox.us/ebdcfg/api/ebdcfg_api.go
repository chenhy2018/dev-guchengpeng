package api

import (
	"net/http"
	"strconv"

	"github.com/qiniu/rpc.v1"
	"github.com/qiniu/rpc.v1/lb.v2.1"
)

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
	return Client{conn: conn}, nil
}

// -------------------------------------------------------------------------

type DiskInfo struct {
	Guid     string    `json:"guid" bson:"guid"`
	DiskId   uint32    `json:"diskId" bson:"diskId"`
	Path     string    `json:"path" bson:"path"` // 校验值,老的磁盘不一定有
	Sectors  uint32    `json:"sectors" bson:"sectors"`
	Hosts    [2]string `json:"hosts" bson:"hosts"`
	ReadOnly uint32    `json:"readonly" bson:"readonly"`
	Deleted  bool      `json:"deleted" bson:"deleted"`
}

func (self Client) DiskStore(l rpc.Logger, disk *DiskInfo) (err error) {
	err = self.conn.CallWithJson(l, nil, "/disk/store", disk)
	return
}

func (self Client) DiskModify(l rpc.Logger, disk *DiskInfo) (err error) {
	err = self.conn.CallWithJson(l, nil, "/disk/modify", disk)
	return
}

func (self Client) DiskStat(l rpc.Logger, guid string, diskId uint32) (disk *DiskInfo, err error) {
	disk = new(DiskInfo)
	err = self.conn.CallWithForm(l, disk, "/disk/stat", map[string][]string{
		"guid":   {guid},
		"diskId": {strconv.FormatUint(uint64(diskId), 10)},
	})
	return
}

func (self Client) DiskDelete(l rpc.Logger, guid string, diskId uint32) (err error) {
	err = self.conn.CallWithForm(l, nil, "/disk/delete", map[string][]string{
		"guid":   {guid},
		"diskId": {strconv.FormatUint(uint64(diskId), 10)},
	})
	return
}

func (self Client) DiskAccess(l rpc.Logger, guid string, diskIds []uint32, readOnly uint32) (err error) {
	diskIdStrs := make([]string, len(diskIds))
	for i, diskId := range diskIds {
		diskIdStrs[i] = strconv.FormatUint(uint64(diskId), 10)
	}
	err = self.conn.CallWithForm(l, nil, "/disk/access", map[string][]string{
		"guid":     {guid},
		"diskId":   diskIdStrs,
		"readonly": {strconv.FormatUint(uint64(readOnly), 10)},
	})
	return
}

type DiskPathArgs struct {
	Guid    string    `json:"guid"`
	Hosts   [2]string `json:"hosts"`
	Path    string    `json:"path"`
	History bool      `json:"histroy"`
}

func (self Client) DiskPath(l rpc.Logger, args *DiskPathArgs) (disks []DiskInfo, err error) {
	err = self.conn.CallWithJson(l, &disks, "/disk/path", args)
	return
}

func (self Client) DiskList(l rpc.Logger, guid string) (disks []*DiskInfo, err error) {
	err = self.conn.CallWithForm(l, &disks, "/disk/list", map[string][]string{
		"guid": {guid},
	})
	return
}

func (self Client) DiskListWritable(l rpc.Logger, guid string) (disks []*DiskInfo, err error) {
	err = self.conn.CallWithForm(l, &disks, "/disk/list", map[string][]string{
		"guid":     {guid},
		"readonly": {"0"},
	})
	return
}

func DiskListUrl(host, guid string) string {
	return host + "/disk/list?guid=" + guid
}

func DiskListWritableUrl(host, guid string) string {
	return host + "/disk/list?guid=" + guid + "&readonly=0"
}

// -------------------------------------------------------------------------

type EcbInfo struct {
	Guid  string    `json:"guid" bson:"guid"`
	Hosts [2]string `json:"hosts" bson:"hosts"`
}

func (self Client) EcbStore(l rpc.Logger, ecb *EcbInfo) (err error) {
	err = self.conn.CallWithJson(l, nil, "/ecb/store", ecb)
	return
}

func (self Client) EcbExist(l rpc.Logger, ecb *EcbInfo) (exist bool, err error) {
	err = self.conn.CallWithJson(l, &exist, "/ecb/exist", ecb)
	return
}

func (self Client) EcbList(l rpc.Logger, guid string) (ecbs []*EcbInfo, err error) {
	err = self.conn.CallWithForm(l, &ecbs, "/ecb/list", map[string][]string{
		"guid": {guid},
	})
	return
}

func EcbListUrl(host, guid string) string {
	return host + "/ecb/list?guid=" + guid
}

// -------------------------------------------------------------------------
