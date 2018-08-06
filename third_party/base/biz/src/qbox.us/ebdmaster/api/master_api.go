package api

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"

	"github.com/qiniu/http/httputil.v1"
	"github.com/qiniu/rpc.v1"
	"github.com/qiniu/rpc.v1/lb.v2.1"

	"qbox.us/ebd/api/types"
	"qiniupkg.com/x/log.v7"
)

const (
	N = types.N
	M = types.M
)

const (
	StatusEcTaskOutdated = 410
)

var IsForceGet = os.Getenv("PFD_FORCE_GET") != "" // 如果设置环境变量PFD_FORCE_GET，pfd和ebd都强制尝试读取数据

type FileInfo struct {
	Soff     uint32   `json:"soff"`
	Fsize    int64    `json:"fsize"`
	Suids    []uint64 `json:"suids"`
	Psectors []uint64 `json:"psectors"`
}

// ---------------------------------------------------------

type LbClient struct {
	conn *lb.Client
}

func shouldRetry(code int, err error) bool {
	if code/100 == 5 || code == httputil.StatusGracefulQuit || code == httputil.StatusOverload || code == 612 {
		return true
	}
	return lb.ShouldRetry(code, err)
}

func shouldFailover(code int, err error) bool {
	if code/100 == 5 || code == httputil.StatusGracefulQuit || code == httputil.StatusOverload {
		return true
	}
	return lb.ShouldFailover(code, err)
}

func NewLbClient(hosts, masterBackupHosts []string, tr http.RoundTripper) (c *LbClient) {

	cfg := &lb.Config{
		Hosts:              hosts,
		ShouldRetry:        shouldRetry,
		FailRetryIntervalS: -1,
		TryTimes:           uint32(len(hosts)),
	}

	var conn *lb.Client
	if len(masterBackupHosts) == 0 {
		conn = lb.New(cfg, tr)
	} else {
		backupCfg := &lb.Config{
			Hosts:              masterBackupHosts,
			ShouldRetry:        shouldFailover,
			FailRetryIntervalS: -1,
			TryTimes:           uint32(len(masterBackupHosts)),
		}
		conn = lb.NewWithFailover(cfg, backupCfg, tr, tr, shouldFailover)
	}
	if IsForceGet {
		log.Println("PFD_FORCE_GET enabled")
	}
	return &LbClient{conn: conn}
}

func (r *LbClient) Get(l rpc.Logger, fid uint64) (fi *FileInfo, err error) {
	fi = new(FileInfo)
	err = r.conn.CallWithForm(l, fi, "/get", map[string][]string{
		"fid":   {strconv.FormatUint(fid, 10)},
		"force": {strconv.FormatBool(IsForceGet)},
	})
	return
}

func (r *LbClient) Gets(l rpc.Logger, sid uint64) (psects [N + M]uint64, err error) {
	resp, err := r.conn.PostWithForm(l, "/gets", map[string][]string{
		"sid": {strconv.FormatUint(sid, 10)},
	})
	if err != nil {
		return
	}
	return retPsects(resp)
}

func (r *LbClient) DiskStat(l rpc.Logger, diskId uint32) (ret DiskStatRet, err error) {

	params := map[string][]string{
		"diskid": []string{strconv.FormatUint(uint64(diskId), 10)},
	}
	err = r.conn.CallWithForm(l, &ret, "/disk/stat", params)
	return
}

// =========================================================

type Client struct {
	Host string
}

// -----------------------------------------------------------------------------

func (c Client) Get(l rpc.Logger, fid uint64) (fi *FileInfo, err error) {
	fi = new(FileInfo)
	err = rpc.DefaultClient.CallWithForm(l, fi, c.Host+"/get", map[string][]string{
		"fid":   {strconv.FormatUint(fid, 10)},
		"force": {strconv.FormatBool(IsForceGet)},
	})
	return
}

func (c Client) Gets(l rpc.Logger, sid uint64) (psects [N + M]uint64, err error) {
	resp, err := rpc.DefaultClient.PostWithForm(l, c.Host+"/gets", map[string][]string{
		"sid": {strconv.FormatUint(sid, 10)},
	})
	if err != nil {
		return
	}
	return retPsects(resp)
}

// -----------------------------------------------------------------------------

type BadInfo struct {
	Idx    uint32
	Reason uint32 // 1: 磁盘错误; 2. 网络错误; 3. 扇区被占用（注意这种情况的处理） 4. unknown
}

func (r Client) Reclaim(l rpc.Logger, sid uint64, psects [N + M]uint64, oldSuids [N + M]uint64, badi []BadInfo) (newPsects [N + M]uint64, suids [N + M]uint64, err error) {
	ePsects := EncodePsects(&psects)
	eSuids := EncodeSuids(&oldSuids)
	eBadi := EncodeBadis(badi)

	resp, err := rpc.DefaultClient.PostWithForm(l, r.Host+"/reclaim", map[string][]string{
		"id":    {strconv.FormatUint(sid, 10)},
		"psect": {ePsects},
		"suids": {eSuids},
		"badi":  {eBadi},
	})
	if err != nil {
		return
	}
	var ret ReclaimRet
	ret, err = retReclaim(resp)
	return ret.Psectors, ret.Suids, err
}

// ---------------------------------------------------------

func (r Client) Complete(l rpc.Logger, sid uint64, psects [N + M]uint64, crc32s [N + M]uint32) (err error) {

	ePsects := EncodePsects(&psects)
	eScrcs := EncodeCrc32s(&crc32s)
	err = rpc.DefaultClient.CallWithForm(l, nil, r.Host+"/complete", map[string][]string{
		"id":     {strconv.FormatUint(sid, 10)},
		"psect":  {ePsects},
		"crc32s": {eScrcs},
	})
	return
}

const (
	State_Success = "success"
	State_Unmatch = "unmatch"
	State_Fail    = "fail"
)

func (r Client) CompleteCheck(l rpc.Logger, sid uint64, status string) (err error) {

	err = rpc.DefaultClient.CallWithForm(l, nil, r.Host+"/completecheck", map[string][]string{
		"id":     {strconv.FormatUint(sid, 10)},
		"status": {status},
	})
	return
}

// ---------------------------------------------------------

func (r Client) Cancel(l rpc.Logger, sid uint64, psects [N + M]uint64) error {

	ePsects := EncodePsects(&psects)
	return rpc.DefaultClient.CallWithForm(l, nil, r.Host+"/cancel", map[string][]string{
		"id":    {strconv.FormatUint(sid, 10)},
		"psect": {ePsects},
	})
}

func (r Client) CancelCheck(l rpc.Logger, sid uint64) error {

	return rpc.DefaultClient.CallWithForm(l, nil, r.Host+"/cancelcheck", map[string][]string{
		"id": {strconv.FormatUint(sid, 10)},
	})
}

// ---------------------------------------------------------

type ReclaimRet struct {
	Psectors [N + M]uint64 `json:"psectors"`
	Suids    [N + M]uint64 `json:"suids"`
}

func retReclaim(resp *http.Response) (ret ReclaimRet, err error) {

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		err = rpc.ResponseError(resp)
		return
	}

	err = json.NewDecoder(resp.Body).Decode(&ret)
	return ret, err
}

// ---------------------------------------------------------

type PsectsRet struct {
	Psectors [N + M]uint64 `json:"psectors"`
}

func retPsects(resp *http.Response) (psects [N + M]uint64, err error) {

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		err = rpc.ResponseError(resp)
		return
	}

	var ret PsectsRet
	err = json.NewDecoder(resp.Body).Decode(&ret)
	return ret.Psectors, err
}

// -----------------------------------------------------------------------------

type MigrateStats struct {
	Paused       bool   `json:"paused"`
	Prepared     int    `json:"prepared"`
	Migrating    int    `json:"migrating"`
	Completed    int    `json:"completed"`
	Finishing    int    `json:"finishing"`
	UnCheckedGid int    `json:"uncheckedgid"`
	UnRemovedGid int    `json:"unremovedgid"`
	StripeCnt    string `json:"stripe"`
}

type RepairStats struct {
	Paused    bool   `json:"paused"`
	Bads      int    `json:"bads"`
	Repairing int    `json:"repairing"`
	Completed int    `json:"completed"`
	StripeCnt string `json:"stripe"`
}

type RecycleStats struct {
	Paused    bool   `json:"paused"`
	Prepared  int    `json:"prepared"`
	Recycling int    `json:"recycling"`
	Completed int    `json:"completed"`
	Finishing int    `json:"finishing"`
	StripeCnt string `json:"stripe"`
}

type DiskStats struct {
	Normal    int `json:"normal"`
	Broken    int `json:"broken"`
	Repairing int `json:"repairing"`
	Repaired  int `json:"repaired"`
	Failed    int `json:"failed"`
}

type SectorStats struct {
	Paused bool `json:"paused"`
}

type Stats struct {
	Readonly bool         `json:"readonly"`
	Migrate  MigrateStats `json:"migrate"`
	Repair   RepairStats  `json:"repair"`
	Recycle  RecycleStats `json:"recycle"`
	Disk     DiskStats    `json:"disk"`
	Realloc  SectorStats  `json:"realloc"`
}

func (c Client) Stats(l rpc.Logger) (ret Stats, err error) {

	err = rpc.DefaultClient.Call(l, &ret, c.Host+"/stats")
	return
}

// -----------------------------------------------------------------------------

type DiskGetRet struct {
	DiskIds []uint32 `json:"diskids"`
}

const (
	DiskState_Normal    = "normal"
	DiskState_Broken    = "broken"
	DiskState_Repairing = "repairing"
	DiskState_Repaired  = "repaired"
)

func (c Client) DiskGet(l rpc.Logger, state string) (ret DiskGetRet, err error) {

	params := map[string][]string{
		"state": []string{state},
	}
	err = rpc.DefaultClient.CallWithForm(l, &ret, c.Host+"/disk/get", params)
	return
}

func (c Client) DiskSet(l rpc.Logger, diskId uint32, state string) error {

	params := map[string][]string{
		"diskid": []string{strconv.FormatUint(uint64(diskId), 10)},
		"state":  []string{state},
	}
	return rpc.DefaultClient.CallWithForm(l, nil, c.Host+"/disk/set", params)
}

type DiskStatRet struct {
	Used         int    `json:"used(MB)"`
	Recycled     int    `json:"recycled(MB)"`
	Total        int    `json:"total(MB)"`
	State        string `json:"state"`
	RepairStart  string `json:"repair_start,omitempty"`
	RepairFinish string `json:"repair_finish,omitempty"`
}

func (c Client) DiskStat(l rpc.Logger, diskId uint32) (ret DiskStatRet, err error) {

	params := map[string][]string{
		"diskid": []string{strconv.FormatUint(uint64(diskId), 10)},
	}
	err = rpc.DefaultClient.CallWithForm(l, &ret, c.Host+"/disk/stat", params)
	return
}

func (c Client) DiskSpace(l rpc.Logger) (ret map[string]interface{}, err error) {

	err = rpc.DefaultClient.Call(l, &ret, c.Host+"/disk/space")
	return
}

// -----------------------------------------------------------------------------

func (c Client) MigratePause(l rpc.Logger) error {

	return rpc.DefaultClient.Call(l, nil, c.Host+"/migrate/pause")
}

func (c Client) MigrateResume(l rpc.Logger) error {

	return rpc.DefaultClient.Call(l, nil, c.Host+"/migrate/resume")
}

func (c Client) MigrateAdd(l rpc.Logger, dgids []uint32) error {

	strDgids := make([]string, len(dgids))
	for i, dgid := range dgids {
		strDgids[i] = strconv.FormatUint(uint64(dgid), 10)
	}
	params := map[string][]string{"dgids": strDgids}
	return rpc.DefaultClient.CallWithForm(l, nil, c.Host+"/migrate/add", params)
}

func (c Client) MigrateDel(l rpc.Logger, dgids []uint32) error {

	strDgids := make([]string, len(dgids))
	for i, dgid := range dgids {
		strDgids[i] = strconv.FormatUint(uint64(dgid), 10)
	}
	params := map[string][]string{"dgids": strDgids}
	return rpc.DefaultClient.CallWithForm(l, nil, c.Host+"/migrate/del", params)
}

type MigrateGetRet struct {
	Dgids []uint32 `json:"dgids"`
}

func (c Client) MigrateGet(l rpc.Logger) (ret MigrateGetRet, err error) {

	err = rpc.DefaultClient.Call(l, &ret, c.Host+"/migrate/get")
	return
}

func (c Client) RepairPause(l rpc.Logger) error {

	return rpc.DefaultClient.Call(l, nil, c.Host+"/repair/pause")
}

func (c Client) RepairResume(l rpc.Logger) error {

	return rpc.DefaultClient.Call(l, nil, c.Host+"/repair/resume")
}

// -----------------------------------------------------------------------------

func (c Client) RecyclePause(l rpc.Logger) error {

	return rpc.DefaultClient.Call(l, nil, c.Host+"/recycle/pause")
}

func (c Client) RecycleResume(l rpc.Logger) error {

	return rpc.DefaultClient.Call(l, nil, c.Host+"/recycle/resume")
}

func (c Client) Readonly(l rpc.Logger) error {

	return rpc.DefaultClient.Call(l, nil, c.Host+"/readonly")
}

func (c Client) Writable(l rpc.Logger) error {

	return rpc.DefaultClient.Call(l, nil, c.Host+"/writable")
}

//-------------------------------------------------------------------------------

func (c Client) Delete(l rpc.Logger, fid uint64) error {
	return rpc.DefaultClient.CallWithForm(l, nil, c.Host+"/delete", map[string][]string{
		"fid": {strconv.FormatUint(uint64(fid), 10)},
	})
}
