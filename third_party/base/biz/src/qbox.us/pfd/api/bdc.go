package api

import (
	"io"
	"io/ioutil"
	"time"

	"github.com/qiniu/http/httputil.v1"
	qio "github.com/qiniu/io"
	"github.com/qiniu/rpc.v1"
	"github.com/qiniu/xlog.v1"

	"net/http"

	"encoding/base64"
	"qbox.us/errors"
	"qbox.us/pfd/api/types"
	cfgapi "qbox.us/pfdcfg/api"
	stgapi "qbox.us/pfdstg/api"
	trackerapi "qbox.us/pfdtracker/api"
	"qbox.us/pfdtracker/stater"
)

const (
	dftRefreshIntervalS    = 300
	fastRefreshIntervalS   = 10
	NotEnoughSpaceHttpCode = 555
)

var (
	ErrProxyGetFails = httputil.NewError(http.StatusServiceUnavailable, "conneting to proxy")
)

type IsAllBackup interface {
	AllBackup(l rpc.Logger, dgid uint32) (bool, error)
}

type Client struct {
	cfg                 *Config
	Guid                string
	DgInfoer            DgInfoer
	GidStater           stater.EntryStater
	PutTryTimes         int
	DeleteTryTimes      int
	DgsRefreshIntervalS int
	SmallFileLimit      int64
	pfdNodeMgr          *PfdNodeMgr
}

func NewClientWithConfig(cfg *Config, infoer DgInfoer, stater stater.EntryStater) (c *Client, err error) {
	c = &Client{
		Guid:                cfg.Guid,
		DgInfoer:            infoer,
		GidStater:           stater,
		PutTryTimes:         cfg.PutTryTimes,
		DeleteTryTimes:      cfg.DeleteTryTimes,
		DgsRefreshIntervalS: cfg.DgsRefreshIntervalS,
		SmallFileLimit:      cfg.SmallFileLimit,
		pfdNodeMgr:          NewPfdNodeMgr(cfg.Idc, cfg.RemoteIdcOrder),
	}

	stgapi.Init(cfg.Timeouts, cfg.Proxies)

	if c.PutTryTimes == 0 {
		c.PutTryTimes = 1
	}
	if c.DeleteTryTimes == 0 {
		c.DeleteTryTimes = 1
	}
	if c.DgsRefreshIntervalS == 0 {
		c.DgsRefreshIntervalS = dftRefreshIntervalS
	}

	err = c.refreshDgs()
	if err != nil {
		return
	}
	go c.loopRefreshDgs()
	return c, nil
}

func NewClient(guid string, infoer DgInfoer, stater stater.EntryStater, tryTimes, deleteTryTimes, refreshInterval int,
	smallFileLimit int64, timeouts stgapi.TimeoutOption, idc string, proxies []string) (c *Client, err error) {

	c = &Client{
		Guid:                guid,
		DgInfoer:            infoer,
		GidStater:           stater,
		PutTryTimes:         tryTimes,
		DeleteTryTimes:      deleteTryTimes,
		DgsRefreshIntervalS: refreshInterval,
		SmallFileLimit:      smallFileLimit,
		pfdNodeMgr:          NewPfdNodeMgr(idc, nil),
	}
	stgapi.Init(timeouts, proxies)

	if c.PutTryTimes == 0 {
		c.PutTryTimes = 1
	}
	if c.DeleteTryTimes == 0 {
		c.DeleteTryTimes = 1
	}
	if c.DgsRefreshIntervalS == 0 {
		c.DgsRefreshIntervalS = dftRefreshIntervalS
	}

	err = c.refreshDgs()
	if err != nil {
		return
	}
	go c.loopRefreshDgs()
	return c, nil
}

type countReader struct {
	r     io.Reader
	count int
}

func (self *countReader) Read(p []byte) (n int, err error) {
	n, err = self.r.Read(p)
	self.count += n
	return
}

func (c *Client) Put(l rpc.Logger, r io.Reader, fsize int64) (fh []byte, md5 []byte, err error) {

	var retried int
	xl := xlog.NewWith(l)

	diskType := cfgapi.DEFAULT
	if fsize <= c.SmallFileLimit {
		diskType = cfgapi.SSD
	}

	cr := &countReader{r: r}
	var badHostUrls []string
	var badDgids []uint32
retrySata:
	for try := 0; try < c.PutTryTimes; try++ {
		var hostUrl string
		var dgid uint32
		hostUrl, dgid, fh, md5, err = c.selectAndPut(xl, diskType, cr, fsize, badHostUrls, badDgids)
		if err == nil {
			xl.Debugf("put file success. selected host is:%v", hostUrl)
			return
		}

		if hostUrl == "" {
			xl.Errorf("can not find host to put")
			break
		}
		if httputil.DetectCode(err) == NotEnoughSpaceHttpCode {
			badDgids = append(badDgids, dgid)
		} else {
			badHostUrls = append(badHostUrls, hostUrl)
		}
		xl.Warnf("put failed to host %v, try: %v - error: %v\n", hostUrl, try, err)

		// TODO-fix 可能存在并发问题
		if cr.count != 0 {
			if rt, ok := r.(io.ReaderAt); ok {
				cr = &countReader{r: &qio.Reader{ReaderAt: rt}}
			} else {
				xl.Error("need retry, but reader is not ReaderAt")
				return
			}
		}
		if httputil.DetectCode(err) == NotEnoughSpaceHttpCode {
			try--
		}
	}

	if diskType == cfgapi.SSD {
		diskType = cfgapi.DEFAULT
		xl.Info("all ssds failed, try satas")
		goto retrySata
	}

	err = errors.Info(err, "Put: put failed, retried", retried).Detail(err)
	return
}

func (c *Client) selectAndPut(xl *xlog.Logger, diskType cfgapi.DiskType,
	r io.Reader, fsize int64, badHostUrls []string, badDgids []uint32) (hostUrl string, dgid uint32, fh []byte, md5 []byte, err error) {

	diskNode, err := c.pfdNodeMgr.SelectUpDisk(xl, diskType, badHostUrls, badDgids)
	if err != nil {
		xl.Errorf("select disk from pfd cache failed. err=%v", err)
		return "", 0, nil, nil, err
	}

	defer c.pfdNodeMgr.ReleaseUpDisk(diskNode)

	xl.Info("host and dgid selected for put:", diskNode.HostUrl, diskNode.Dgid)

	cli := stgapi.NewClientWithTimeout(diskNode.HostUrl)
	r = ioutil.NopCloser(r) // https://pm.qbox.me/issues/11037
	fh, md5, err = cli.Put(xl, r, fsize, diskNode.Dgid)
	return diskNode.HostUrl, diskNode.Dgid, fh, md5, err
}

func (c *Client) Alloc(l rpc.Logger, fsize int64, hash [20]byte) (fh []byte, err error) {

	var retried int
	xl := xlog.NewWith(l)

	var hostUrl string
	var dgid uint32
	var badHostUrls []string
	var badDgids []uint32

	for ; retried < c.PutTryTimes; retried++ {
		hostUrl, dgid, fh, err = c.selectAndAlloc(xl, cfgapi.DEFAULT, fsize, hash, badHostUrls, badDgids)
		if err == nil {
			xl.Debugf("alloc success. selected host is:%v", hostUrl)
			return
		} else {
			// 如果是磁盘空间不够返回的错误，就忽视一次retry
			if httputil.DetectCode(err) == NotEnoughSpaceHttpCode {
				retried--
			}
		}
		if hostUrl == "" {
			xl.Errorf("alloc failed. no host aviable.")
			return
		}

		if _, ok := err.(rpc.RespError); ok {
			badDgids = append(badDgids, dgid)
		} else {
			badHostUrls = append(badHostUrls, hostUrl)
		}
		xl.Warnf("alloc retry failed to host %v - error: %v\n", hostUrl, err)
	}

	err = errors.Info(err, "alloc: alloc failed, retried", retried).Detail(err)
	return
}

func (c *Client) selectAndAlloc(xl *xlog.Logger, diskType cfgapi.DiskType, fsize int64, hash [20]byte, badHostUrls []string, badDgids []uint32) (hostUrl string, dgid uint32, fh []byte, err error) {

	diskNode, err := c.pfdNodeMgr.SelectUpDisk(xl, diskType, badHostUrls, badDgids)
	if diskNode == nil {
		xl.Errorf("select disk from pfd cache for alloc failed. err=%v", err)
		return "", 0, nil, err
	}

	defer c.pfdNodeMgr.ReleaseUpDisk(diskNode)

	xl.Info("host and dgid selected for alloc:", diskNode.HostUrl, diskNode.Dgid)

	cli := stgapi.NewClientWithTimeout(diskNode.HostUrl)
	fh, err = cli.Alloc(xl, fsize, hash, diskNode.Dgid)
	return diskNode.HostUrl, diskNode.Dgid, fh, err
}

func (c *Client) parseFhAndGetDGid(l rpc.Logger, fh []byte) (egid string, dgid uint32, isECed bool, ecing int32, err error) {

	xl := xlog.NewWith(l)
	fhi, err := types.DecodeFh(fh)
	if err != nil {
		xl.Errorf("decodeFh failed. fh:%v", fh)
		return
	}

	egid = types.EncodeGid(fhi.Gid)
	entry, err := c.GidStater.StateEntry(xl, egid)
	if err != nil {
		xl.Errorf("state egid(%v) failed: %v\n", egid, err)
		return
	}
	dgid = entry.Dgid
	ecing = entry.Ecing
	isECed = entry.EC
	return
}

func (c *Client) Get(l rpc.Logger,
	fh []byte, from, to int64) (rc io.ReadCloser, fsize int64, err error) {
	var proxyOk = true
	err = c.selectDo(l, fh, func(xl *xlog.Logger, diskNode *DiskNode) (innerErr error) {
		rc, fsize, innerErr = c.get(xl, diskNode, fh, from, to, &proxyOk)
		return innerErr
	})
	return
}

func (c *Client) Md5(l rpc.Logger, fh []byte) (md5 []byte, err error) {
	var proxyOk = true
	err = c.selectDo(l, fh, func(xl *xlog.Logger, diskNode *DiskNode) (innerErr error) {
		md5, innerErr = c.md5(xl, diskNode, fh, &proxyOk)
		return innerErr
	})
	return
}

func (c *Client) selectDo(l rpc.Logger, fh []byte, fn func(*xlog.Logger, *DiskNode) error) (err error) {

	xl := xlog.NewWith(l)

	// 下面 get 失败的几种情况:
	// 	1. gid->dgid 映射变更了 (stg 返回 612), 造成缓存错误 -> 刷新缓存
	// 	2. dgid->hosts 映射变更 (stg 返回 612), 造成缓存错误 -> 刷新缓存
	// 	3. dgid 对应的 hosts 被下线，判断依据为: 所有 hosts 访问错误均为网络错误 -> 刷新缓存并完整重试一次
	// 	4. 第三异步副本同步有延迟，返回 620, 需要重试第一二副本

	egid, dgid, isECed, _, err := c.parseFhAndGetDGid(l, fh)
	if err != nil {
		xl.Errorf("parser fh failed.")
		return
	}

	var (
		networkErrCount int  = 0
		needUpdate      bool = false
		alreadyUpdate   bool = false

		allocedErr error = nil
	)

	var resNotExistHostUrls string
	var badHostUrls []string
tryget:
	var hostUrl string
	var diskNode *DiskNode

	for {

		// selectGetNode会根据dgid去缓存中挑选出执行get操作的disk，
		// 如果缓存中没有dgid对应的dg，则会去pfdcfg拉取然后更新缓存
		diskNode, err = c.selectGetNode(xl, dgid, isECed, badHostUrls)
		if err != nil {
			return
		}
		if diskNode.HostUrl == resNotExistHostUrls {
			xl.Debugf("selected host is:%v, which return 612 lasttime, selectGetNode again", diskNode.HostUrl)
			badHostUrls = append(badHostUrls, diskNode.HostUrl)
			diskNode, err = c.selectGetNode(xl, dgid, isECed, badHostUrls)
			if err != nil {
				return
			}
		}
		defer c.pfdNodeMgr.ReleaseDownDisk(diskNode)
		hostUrl = diskNode.HostUrl
		err = fn(xl, diskNode)
		if err == nil {
			xl.Debugf("get file ok. selected host:%v", hostUrl)
			return
		}
		code := httputil.DetectCode(err)

		if code == 612 {
			xl.Errorf("get file failed, ret 612, selected host is:%v", hostUrl)
			needUpdate = true
			resNotExistHostUrls = hostUrl
			break
		}

		if code == stgapi.StatusAllocedEntry {
			xl.Errorf("get file failed, ret 613, selected host is:%v", hostUrl)
			allocedErr = err
		}

		if isNetworkError(err) {
			networkErrCount++
		}

		xl.Warnf("get failed from host %v - error: %v\n", hostUrl, err)
		badHostUrls = append(badHostUrls, hostUrl)

		//3. dgid 对应的 hosts 被下线，判断依据为: 所有 hosts 访问错误均为网络错误 -> 先break，然后刷新缓存并完整重试一次

		if c.pfdNodeMgr.UsedAllHost(dgid, badHostUrls) {

			if networkErrCount == len(badHostUrls) {
				needUpdate = true
			}
			break
		}
	}

	if allocedErr != nil {
		err = allocedErr
		return
	}

	if needUpdate && !alreadyUpdate {
		// 尝试更新 gid->dgid 的映射
		dgid, isECed, err = c.GidStater.ForceUpdate(xl, egid)
		if err != nil {
			xl.Errorf("get file failed, state egid(%v) failed: %v\n while refresh gid and dgid relation", egid, err)
			return
		}
		alreadyUpdate = true
		goto tryget
	}
	return
}

func isNetworkError(err error) bool {

	if _, ok := err.(*rpc.ErrorInfo); ok {
		return false
	}
	return true
}

func (c *Client) selectGetNode(xl *xlog.Logger, dgid uint32, isECed bool,
	badHostUrls []string) (diskNode *DiskNode, err error) {

	diskNode, err = c.pfdNodeMgr.GetDownDiskNode(dgid, isECed, badHostUrls)

	if err == trackerapi.ErrGidECed {
		return
	}
	if err != nil {
		xl.Errorf("get disk node failed. dgid=%d err=%v", dgid, err)

		//缓存中没有该disk节点，重新到pfdcfg拉取
		if err == EDGNodeNotFind {
			var dgInfo *cfgapi.DiskGroupInfo = nil
			dgInfo, err = c.DgInfoer.GetDGInfo(xl, c.Guid, dgid)
			if err != nil {
				xl.Errorf("get hosts of dgid(%v) failed: %v\n", dgid, err)
				return nil, err
			}

			err = c.pfdNodeMgr.LoadDGInfo(dgInfo)
			if err != nil {
				xl.Errorf("pfd cache load dginfo failed. err=%v", err)
				return nil, err
			}

			diskNode, err = c.pfdNodeMgr.GetDownDiskNode(dgid, isECed, badHostUrls)
			if err != nil {
				xl.Errorf("get again failed. dgid=%d err=%v", dgid, err)
			}
		}
	}
	return
}

func (c *Client) get(xl *xlog.Logger, diskNode *DiskNode,
	fh []byte, from, to int64, proxyOk *bool) (rc io.ReadCloser, fsize int64, err error) {
	xl.Info("get host and dgid for file get:", diskNode.HostUrl, diskNode.Dgid)

	cli := stgapi.NewClientWithTimeout(diskNode.HostUrl)
	if c.pfdNodeMgr.idc != diskNode.Idc {
		xl.Infof("c.get: selfIdc(%v) != remoteIdc(%v), use proxy(ok=%v).", c.pfdNodeMgr.idc, diskNode.Idc, *proxyOk)
		if !*proxyOk {
			err = ErrProxyGetFails
			return
		}
		rc, fsize, err = cli.ProxyGet(xl, fh, from, to)
		if err != nil {
			// 所有代理都都读失败
			if e, ok := err.(rpc.RespError); (ok && e.HttpCode() == 503) || (!ok && isProxyError(err)) {
				*proxyOk = false
				xl.Error("all proxy fails", err)
			}

		}
	} else {
		rc, fsize, err = cli.Get(xl, fh, from, to)
	}
	return
}

func (c *Client) md5(xl *xlog.Logger, diskNode *DiskNode, fh []byte, proxyOk *bool) (md5 []byte, err error) {
	xl.Info("get host and dgid for file md5:", diskNode.HostUrl, diskNode.Dgid)

	cli := stgapi.NewClientWithTimeout(diskNode.HostUrl)
	if c.pfdNodeMgr.idc != diskNode.Idc {
		xl.Infof("c.md5: selfIdc(%v) != remoteIdc(%v), use proxy(ok=%v).", c.pfdNodeMgr.idc, diskNode.Idc, *proxyOk)
		if !*proxyOk {
			err = ErrProxyGetFails
			return
		}
		md5, err = cli.ProxyMd5(xl, fh)
		if err != nil {
			// 所有代理都都读失败
			if e, ok := err.(rpc.RespError); (ok && e.HttpCode() == 503) || (!ok && isProxyError(err)) {
				*proxyOk = false
				xl.Error("all proxy fails", err)
			}

		}
	} else {
		md5, err = cli.Md5(xl, fh)
	}
	return
}

func (c *Client) PutAt(l rpc.Logger, r io.Reader, fh []byte) (err error) {

	xl := xlog.NewWith(l)
	var (
		needUpdate    bool
		alreadyUpdate bool
		stgClient     stgapi.Client
	)

	egid, dgid, isECed, _, err := c.parseFhAndGetDGid(xl, fh)
	if err != nil {
		xl.Errorf("putat failed. parse fh failed.err:%v", err)
		return
	}
	if isECed {
		return trackerapi.ErrGidECed
	}

	var r2 io.Reader = ioutil.NopCloser(r)
	cr := &countReader{r: r2}

tryPutAt:
	masterHostUrl, _, err := c.selectMaster(xl, dgid)
	if err != nil {
		xl.Errorf("putat failed. select master failed. dgid=%u", dgid)
		return
	}

	stgClient = stgapi.NewClientWithTimeout(masterHostUrl)
	for i := 0; i < c.PutTryTimes; i++ {
		err = stgClient.PutAt(l, cr, fh)
		if err != nil {
			if cr.count != 0 {
				if rt, ok := r.(io.ReaderAt); ok {
					r2 = ioutil.NopCloser(&qio.Reader{ReaderAt: rt})
					cr = &countReader{r: r2}
				} else {
					xl.Errorf("need retry, but reader is not ReaderAt")
					return
				}
			}

			if httputil.DetectCode(err) == 612 {
				xl.Errorf("putat failed. ret 612. host:%v", masterHostUrl)
				needUpdate = true
				break
			}
			if isNetworkError(err) {
				needUpdate = true
				xl.Warnf("putat failed and going to try. network error. host:%v", masterHostUrl)
				continue // network error need retry
			}
			xl.Warnf("stgClient.putat failed, tryTime:%d, err:%v", i, err)
		}

		xl.Infof("file putat complete. host:%v dgid:%v err:%v", masterHostUrl, dgid, err)
		return
	}

	if needUpdate && !alreadyUpdate {
		// 尝试更新 gid->dgid 的映射
		dgid, isECed, err = c.GidStater.ForceUpdate(xl, egid)
		if err != nil {
			xl.Errorf("state egid(%v) failed: %v\n", egid, err)
			return
		}

		if isECed {
			return trackerapi.ErrGidECed
		}

		alreadyUpdate = true
		xl.Infof("Gid(%v) is refreshed, dgid=%v, try putat again...\n", egid, dgid)
		goto tryPutAt
	}

	xl.Errorf("putat file finally failed. err:%v", err)
	return
}

func (c *Client) selectMaster(xl *xlog.Logger, dgid uint32) (masterHostUrl string, hasBackup bool, err error) {
	masterHostUrl, _, hasBackup, err = c.selectMasterEx(xl, dgid)
	return
}

func (c *Client) selectMasterEx(xl *xlog.Logger, dgid uint32) (masterHostUrl string, idc string, hasBackup bool, err error) {

	masterHostUrl, idc, hasBackup, err = c.pfdNodeMgr.GetMasterHostUrlAndIdc(dgid)
	if err != nil {
		xl.Errorf("get disk node failed. dgid=%d err=%v", dgid, err)

		//缓存中没有该disk节点，重新到pfdcfg拉取
		if err == EDGNodeNotFind {
			var dgInfo *cfgapi.DiskGroupInfo = nil
			dgInfo, err = c.DgInfoer.GetDGInfo(xl, c.Guid, dgid)
			if err != nil {
				xl.Errorf("get hosts of dgid(%v) failed: %v\n", dgid, err)
				return "", "", false, err
			}

			err = c.pfdNodeMgr.LoadDGInfo(dgInfo)
			if err != nil {
				xl.Errorf("pfd cache load dginfo failed. err=%v", err)
				return "", "", false, err
			}

			masterHostUrl, idc, hasBackup, err = c.pfdNodeMgr.GetMasterHostUrlAndIdc(dgid)
		}
	}
	return
}

func (c *Client) Delete(l rpc.Logger, fh []byte) (err error) {

	xl := xlog.NewWith(l)
	var (
		needUpdate    bool
		alreadyUpdate bool = false
		stgClient     stgapi.Client
	)

	egid, dgid, isECed, ecing, err := c.parseFhAndGetDGid(xl, fh)
	if err != nil {
		xl.Errorf("delete failed. parse fh failed.err:%v", err)
		return
	}
	if isECed || ecing != 0 {
		return trackerapi.ErrGidECed
	}

tryDelete:
	var proxyOk = true
	masterHostUrl, idc, hasBackup, err := c.selectMasterEx(xl, dgid)
	if err != nil {
		xl.Errorf("delete failed. select host failed. err:%v", err)
		return
	}
	if hasBackup {
		return trackerapi.ErrGidECed
	}

	var deleteTimeout = 1000 * time.Millisecond
	if c.cfg != nil && c.cfg.Timeouts.DeleteClientTimeoutMs > 0 {
		deleteTimeout = time.Duration(c.cfg.Timeouts.DeleteClientTimeoutMs) * time.Millisecond
	}
	stgClient = stgapi.NewClientWithTimeoutEx(masterHostUrl, deleteTimeout) // master
	xl.Debugf("delete fh :%v from master host :%v", base64.URLEncoding.EncodeToString(fh), masterHostUrl)

	for i := 0; i < c.DeleteTryTimes; i++ {
		err = c.delete(xl, stgClient, idc, fh, &proxyOk)
		if err != nil {
			if httputil.DetectCode(err) == 612 {
				xl.Errorf("delete failed. ret 612. host:%v", masterHostUrl)
				needUpdate = true
				break
			}
			if isNetworkError(err) {
				xl.Warnf("delete failed and going to try. network error. host:%v err:%v", masterHostUrl, err)
				needUpdate = true
				continue // network error need retry
			}
			xl.Warnf("stgClient.Delete failed, tryTime:%d, err:%v", i, err)
		}
		xl.Infof("deleted: %v", dgid)
		return
	}

	if needUpdate && !alreadyUpdate {
		// 尝试更新 gid->dgid 的映射
		dgid, isECed, err = c.GidStater.ForceUpdate(xl, egid)
		if err != nil {
			xl.Errorf("state egid(%v) failed: %v\n", egid, err)
			return
		}
		if isECed {
			return trackerapi.ErrGidECed
		}

		alreadyUpdate = true
		xl.Infof("Gid(%v) is refreshed, dgid=%v, tryDelete again...\n", egid, dgid)
		goto tryDelete
	}
	xl.Errorf("delete file finally failed. err:%v", err)
	return
}

func (c *Client) delete(xl *xlog.Logger, stgClient stgapi.Client, masterIdc string, fh []byte, proxyOk *bool) (err error) {

	if c.pfdNodeMgr.idc != masterIdc {
		xl.Infof("c.delete: selfIdc(%v) != remoteIdc(%v),  use proxy(ok=%v).", c.pfdNodeMgr.idc, masterIdc, *proxyOk)
		if !*proxyOk {
			err = ErrProxyGetFails
			return
		}
		err = stgClient.ProxyDelete(xl, fh)
		if err != nil {
			// 所有代理都都失败
			if e, ok := err.(rpc.RespError); (ok && e.HttpCode() == 503) || (!ok && isProxyError(err)) {
				*proxyOk = false
				xl.Error("all proxy fails", err)
			}
		}
	} else {
		err = stgClient.Delete(xl, fh)
	}
	return
}

func (c *Client) GetType(l rpc.Logger, fh []byte) (typ cfgapi.DiskType, err error) {

	fhi, err := types.DecodeFh(fh)
	if err != nil {
		return
	}
	xl := xlog.NewWith(l)

	egid := types.EncodeGid(fhi.Gid)
	entry, err := c.GidStater.StateEntry(xl, egid)
	if err != nil {
		xl.Errorf("state egid(%v) failed: %v\n", egid, err)
		return
	}
	dgid := entry.Dgid

	typ, ok := c.pfdNodeMgr.GetType(dgid)
	if !ok {
		typ = cfgapi.DEFAULT
		return
	}

	xl.Debugf("get type ok. dgid:%v type:%v", dgid, typ)
	return
}

func (c *Client) AllBackup(l rpc.Logger, dgid uint32) (bool, error) {
	return c.pfdNodeMgr.AllBackup(dgid)
}

func (c *Client) refreshDgs() (err error) {

	xl := xlog.NewDummy()
	dgis, err := c.DgInfoer.ListDgs(xl, c.Guid)
	if err != nil {
		xl.Errorf("refreshDgs: ListDgs guid(%v) failed => %v\n", c.Guid, err)
		return
	}

	c.pfdNodeMgr.LoadPfdNodes(xl, dgis)
	return
}

func (c *Client) loopRefreshDgs() {

	var err error
	var interval int = c.DgsRefreshIntervalS
	for {
		<-time.After(time.Duration(interval) * time.Second)
		if interval != c.DgsRefreshIntervalS {
			interval = c.DgsRefreshIntervalS
		}
		err = c.refreshDgs()
		if err != nil {
			interval = fastRefreshIntervalS
		}
	}
}
