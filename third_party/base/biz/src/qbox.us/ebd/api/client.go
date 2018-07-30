package api

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/qiniu/errors"
	"github.com/qiniu/http/httputil.v1"
	"github.com/qiniu/rpc.v1"
	"github.com/qiniu/xlog.v1"
	"qbox.us/ebd/api/types"
	master "qbox.us/ebdmaster/api"
	ebdstg "qbox.us/ebdstg/api"
	ecb "qbox.us/ecb/api"
	"qbox.us/limit"
	"qbox.us/limit/count"
	"qbox.us/memcache/mcg"
	pfdcfg "qbox.us/pfdcfg/api"
)

var SUMaxLen uint32 = types.SUMaxLen

const (
	Same = iota
	Diff_Repair
	Diff_Recycle

	DefaultRgetLimit = 1000
)

// 状态跳转标记
const (
	SKey_RefreshGlobal = iota
	SKey_GetHost
	SKey_RefreshInfo
	SKey_DirectGet
	SKey_RGet
	SKey_LastCheck
	SKey_Exit
)

type Config struct {
	Guid              string     `json:"guid"`
	MasterHosts       []string   `json:"master_hosts"`
	MasterBackupHosts []string   `json:"master_backup_hosts"`
	MasterConn        int        `json:"master_conn"`
	Memcached         []mcg.Node `json:"memcached"`
	McExpires         int32      `json:"mc_expires_s"` // memcache item的失效时间，最多一个月，小于ebd的延迟回收时间(见ebdmaster配置中的delay_realloc_day)。以秒为单位，0为不失效
	EbdCfgHost        string     `json:"cfg_host"`
	ReloadMs          int        `json:"reload_ms"`
	TimeoutMs         int        `json:"timeout_ms"`
	NoBlocks          bool       `json:"no_blocks"`
	NoCached          bool       `json:"no_cached"`
	RgetLimit         int        `json:"rget_limit"`
}

type MemCache interface {
	Set(l rpc.Logger, fid uint64, fi *master.FileInfo) error
	Get(l rpc.Logger, fid uint64) (*master.FileInfo, error)
}

type EStgs interface {
	Host(l rpc.Logger, guid string, diskId uint32) (host string, err error)
}

type Client struct {
	guid      string
	master    *master.LbClient
	memcached MemCache
	stgs      EStgs
	rget      *rgetClient
	rgetlimit limit.Limit
}

func New(cfg *Config) (c *Client, err error) {
	timeout := time.Duration(cfg.TimeoutMs) * time.Millisecond
	ebdstgTr := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		Dial: (&net.Dialer{
			Timeout:   timeout,
			KeepAlive: 30 * time.Second,
		}).Dial,
		ResponseHeaderTimeout: timeout,
	}
	ebdmasterTr := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		Dial: (&net.Dialer{
			Timeout:   timeout,
			KeepAlive: 30 * time.Second,
		}).Dial,
		ResponseHeaderTimeout: timeout,
		MaxIdleConnsPerHost:   cfg.MasterConn,
	}
	rgetTr := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		Dial: (&net.Dialer{
			Timeout:   timeout,
			KeepAlive: 30 * time.Second,
		}).Dial,
	}
	ebdstg.DefaultClient = rpc.Client{&http.Client{Transport: ebdstgTr}}

	oneMonth := int32(30 * 24 * 60 * 60)
	if cfg.McExpires > oneMonth {
		panic("cfg.McExpires should be no more than 1 month")
	}

	c = &Client{
		guid: cfg.Guid,
	}
	c.master = master.NewLbClient(cfg.MasterHosts, cfg.MasterBackupHosts, ebdmasterTr)
	c.memcached = nilMemcached{}
	if len(cfg.Memcached) > 0 {
		c.memcached, err = newMemcachedClient(cfg.Memcached, cfg.McExpires)
		if err != nil {
			return
		}
	}
	c.stgs, err = newEstgsClient(cfg.Guid, cfg.EbdCfgHost, cfg.ReloadMs)
	if err != nil {
		return
	}
	c.rget, err = newRgetClient(cfg.Guid, cfg.EbdCfgHost, cfg.ReloadMs, rgetTr)
	if err != nil {
		return
	}
	if cfg.RgetLimit <= 0 {
		c.rgetlimit = count.New(DefaultRgetLimit)
	} else {
		c.rgetlimit = count.New(cfg.RgetLimit)
	}
	return
}

func NewWithHost(cfg *Config, stgs EStgs) (c *Client, err error) {
	c, err = New(cfg)
	c.stgs = stgs
	return
}

func (c *Client) GetType(l rpc.Logger, fh []byte) (typ pfdcfg.DiskType, err error) {
	return pfdcfg.DEFAULT, nil
}

type nullReadCloser struct{}

func (self nullReadCloser) Read(p []byte) (int, error) {
	return 0, io.EOF
}

func (self nullReadCloser) Close() error {
	return nil
}

func (self *Client) Get(l rpc.Logger,
	fh []byte, from, to int64) (rc io.ReadCloser, fsize int64, err error) {

	xl := xlog.NewWith(l)
	fhi, err := types.DecodeFh(fh)
	if err != nil {
		err = errors.Info(err, "types.DecodeFh")
		return
	}
	if fhi.Fsize == 0 {
		return nullReadCloser{}, 0, nil
	}
	if from-to == 0 {
		return nullReadCloser{}, 0, nil
	}
	fid := fhi.Fid

	rc, fsize, err = self.getFromEbd(xl, fid, from, to, true)
	return
}

func (self *Client) GetFid(xl *xlog.Logger, fid uint64, from, to int64) (rc io.ReadCloser, fsize int64, err error) {
	return self.getFromEbd(xl, fid, from, to, false)
}

func (self *Client) getFromEbd(xl *xlog.Logger, fid uint64, from, to int64, cache bool) (rc io.ReadCloser, fsize int64, err error) {

	var fi *master.FileInfo
	var errMc error

	if cache {
		fi, errMc = self.memcached.Get(xl, fid)
		if errMc != nil && errMc != memcache.ErrCacheMiss {
			errors.Info(errMc, "memcached.Get").LogError(xl.ReqId())
		}
	}

	if errMc != nil || !cache {
		fi, err = self.getFileInfo(xl, fid, cache)
		if err != nil {
			err = errors.Info(err, "getFileInfo, fid:", fid).Detail(err)
			return
		}
	}

	if to > fi.Fsize {
		// 这里得用ebdmaster里存的fi.Fsize,因为fhi.Fsize还是有可能被伪造的
		return nil, 0, errors.New(fmt.Sprintf("to(%v) is bigger than fsize(%v)", to, fi.Fsize))
	}

	ch := make(chan *xlog.Logger, 1)
	pr, pw := io.Pipe()
	go func(xl *xlog.Logger) {
		_, err = self.WriteTo(xl, pw, fi, fid, from, to)
		pw.CloseWithError(err)
		ch <- xl
	}(xl.Spawn())

	return &xlReadCloser{pr, ch, xl}, to - from, nil
}

func (self *Client) WriteTo(xl *xlog.Logger, w io.Writer, fi *master.FileInfo, fid uint64, from, to int64) (n int64, err error) {

	defer xl.Xtrack("e.Write", time.Now(), &err)
	xl.Debugf("fid %v, WriteTo(direct read), from: %v, to: %v, fi: %+v\n", fid, from, to, fi)

	context := newContext(xl, self, w, fi, fid, from, to)
	stateMgr := NewStateMgr()

	n, err = stateMgr.readFile(context)
	return
}

func (self *Client) LoadForRecycle(xl *xlog.Logger, fi *master.FileInfo, fid uint64, sfrom, sto int64) (rc io.ReadCloser, fsize int64, err error) {

	xl.Debugf("read for recycle: fid %v, from: %v, to: %v, fsize: %v, fi: %v\n", fid, sfrom, sto, fi.Fsize, fi)

	from := sfrom
	to := sto

	if to > fi.Fsize {
		return nil, 0, errors.New(fmt.Sprintf("to(%v) is bigger than fsize(%v)", to, fi.Fsize))
	}

	ch := make(chan *xlog.Logger, 1)
	pr, pw := io.Pipe()
	go func(xl *xlog.Logger) {
		_, err = self.WriteTo(xl, pw, fi, fid, from, to)
		pw.CloseWithError(err)
		ch <- xl
	}(xl.Spawn())

	return &xlReadCloser{pr, ch, xl}, to - from, nil
}

// ---------------------------------------------------------------------
// StateMgr 控制状态之间的转换，见 https://cf.qiniu.io/pages/viewpage.action?pageId=18488084
type StateMgr struct {
	stateMap map[int]State
}

func NewStateMgr() *StateMgr {
	stateMap := map[int]State{
		SKey_RefreshGlobal: &StateRefreshGlobal{},
		SKey_GetHost:       &StateGetHost{},
		SKey_RefreshInfo:   &StateRefreshInfo{},
		SKey_DirectGet:     &StateDirectGet{},
		SKey_RGet:          &StateRget{},
		SKey_LastCheck:     &StateLastCheck{},
	}
	return &StateMgr{
		stateMap: stateMap,
	}
}

func (s *StateMgr) readFile(c *Context) (n int64, err error) {

	var stateKey int
	stateKey = SKey_RefreshGlobal
	for {
		next := s.stateMap[stateKey].Handle(c)
		if next == SKey_Exit {
			break
		}
		stateKey = next
	}
	n, err = c.readN, c.errC
	return
}

// ---------------------------------------------------------------------
type State interface {
	Handle(c *Context) (stateKey int) // 处理当前状态, 并根据context获取下一个状态的key
}

type StateRefreshGlobal struct {
}

func (s *StateRefreshGlobal) Handle(c *Context) int {

	c.globalFrom += c.readN
	from := c.globalFrom + int64(c.newfi.Soff)
	to := c.globalTo + int64(c.newfi.Soff)
	sfrom, pfrom := uint32(from%int64(SUMaxLen)), from/int64(SUMaxLen)
	sto, pto := uint32(to%int64(SUMaxLen)), to/int64(SUMaxLen)
	if sto == 0 {
		sto, pto = SUMaxLen, to/int64(SUMaxLen)-1
	}
	c.localFrom = sfrom
	c.localIdx = int(pfrom)
	c.localTo = SUMaxLen
	if pfrom == pto {
		c.localTo = sto
	}
	return SKey_GetHost
}

// StateRefreshInfo 刷新fileInfo
type StateRefreshInfo struct {
}

func (s *StateRefreshInfo) Handle(c *Context) int {

	err := c.refreshInfo()
	if err != nil {
		c.errC = errors.Info(err, "refreshInfo").Detail(err)
		return SKey_Exit
	}

	c.xl.Debugf("fid %v, WriteTo(refresh file info), from: %v, to: %v, fi: %+v\n", c.fid, c.globalFrom, c.globalTo, c.newfi)

	ret := checkFileInfo(c.oldfi, c.newfi)
	switch ret {
	case Same:
		return SKey_RGet
	case Diff_Repair:
		return SKey_GetHost
	case Diff_Recycle:
		return SKey_RefreshGlobal
	default:
		panic("unrecognized state")
	}
}

// StateGetHost 从ebdcfg中获取磁盘的host
type StateGetHost struct {
}

func (s *StateGetHost) Handle(c *Context) int {

	psector := c.newfi.Psectors[c.localIdx]
	diskId, _ := types.DecodePsect(psector)
	host, err := c.client.stgs.Host(c.xl, c.client.guid, diskId)
	if err != nil {
		if code := httputil.DetectCode(err); code != 612 && code != 801 {
			c.errC = errors.Info(err, "ebdcfg.Host failed", diskId).Detail(err)
			return SKey_Exit
		}
		c.xl.Infof("diskId(%v)/psector(%v) not exist in ebdcfg\n", diskId, psector)

		return SKey_RefreshInfo // 获取host失败, 需要刷新fileInfo, 如果没有改变, 则rget。
	}
	c.localHost = host
	return SKey_DirectGet
}

// StateDirectGet 直接从ebdstg中读取
type StateDirectGet struct {
}

func (s *StateDirectGet) Handle(c *Context) int {

	psector := c.newfi.Psectors[c.localIdx]
	suid := c.newfi.Suids[c.localIdx]
	from := c.localFrom
	to := c.localTo
	host := c.localHost
	wr := &errWriteRecoder{c.w, nil}
	xl := c.xl
	n, err := c.client.estgGet(xl, wr, host, psector, suid, from, to)
	if err == nil {
		return readNextBlock(c, n)
	}
	if wr.errW != nil || err == io.ErrShortWrite {
		c.errC = errors.Info(err, "ReadBlock: get failed(client error), host", host).Detail(err)
		return SKey_Exit
	}

	xl.Warnf("ReadBlock: psector(%v), get failed => n: %v, err: %v\n", psector, n, errors.Detail(err))

	c.readN += n
	c.localFrom += uint32(n)
	errStatus := ebdstg.Reason(err)
	switch errStatus {
	case ebdstg.ErrReasonSuidNotMatch:
		return SKey_RefreshInfo
	case ebdstg.ErrReasonNetworkError,
		ebdstg.ErrReasonDiskIoError,
		ebdstg.ErrReasonNoSuchDisk,
		ebdstg.ErrReasonBrokenDisk,
		ebdstg.ErrReasonChecksum:
		return SKey_RGet
	default:
		c.errC = errors.Info(err, "ReadBlock: get failed, host", host).Detail(err)
		return SKey_Exit
	}
}

// StateRget 进行修复读
type StateRget struct {
}

func (s *StateRget) Handle(c *Context) int {

	if e := c.client.rgetlimit.Acquire(RgetKey); e != nil {
		c.errC = errors.Info(e, "ReadBlock: get failed, host", c.localHost).Detail(c.errC)
		c.xl.Error("limit error, too many rget")
		return SKey_Exit
	}
	defer c.client.rgetlimit.Release(RgetKey)

	n, err := c.client.rGet(c.xl, c.w, c.newfi.Suids[c.localIdx], c.localFrom, c.localTo)
	if err != nil {
		c.errC = err
		return SKey_LastCheck
	}
	return readNextBlock(c, n)
}

type StateLastCheck struct {
}

// StateLastCheck 最后一次检查是否读取出错
func (s *StateLastCheck) Handle(c *Context) int {

	err := c.refreshInfo()
	if err != nil {
		c.errC = errors.Info(err, "refreshInfo").Detail(err)
		return SKey_Exit
	}
	ret := checkFileInfo(c.oldfi, c.newfi)
	switch ret {
	case Same:
		return SKey_Exit
	case Diff_Repair:
		return SKey_GetHost
	case Diff_Recycle:
		return SKey_RefreshGlobal
	default:
		panic("unrecognized state")
	}
}

func readNextBlock(c *Context, n int64) int {

	c.readN += n
	c.localFrom = 0
	c.localTo = SUMaxLen
	to := int64(c.newfi.Soff) + c.globalTo
	sto, pto := uint32(to%int64(SUMaxLen)), to/int64(SUMaxLen)
	if sto == 0 {
		sto, pto = SUMaxLen, to/int64(SUMaxLen)-1
	}
	if c.localIdx == int(pto) {
		return SKey_Exit
	}
	if c.localIdx == int(pto-1) {
		c.localTo = sto
	}
	c.localIdx++
	return SKey_GetHost
}

// ---------------------------------------------------------------------

var (
	RgetKey = []byte("rget")
)

type Context struct {
	client     *Client
	w          io.Writer // 记录读的数据
	xl         *xlog.Logger
	fid        uint64
	oldfi      *master.FileInfo // 旧的fileInfo
	newfi      *master.FileInfo // 新的fileInfo
	globalFrom int64            // 文件级
	globalTo   int64            // 文件级
	localFrom  uint32           // block级
	localTo    uint32           // block级
	localIdx   int              // 正在读的文件块的偏移个数
	localHost  string           // localIdx的host
	readN      int64            // 已读的字节长度
	errC       error
}

func newContext(xl *xlog.Logger, client *Client, w io.Writer, fi *master.FileInfo, fid uint64, from, to int64) *Context {
	return &Context{
		client:     client,
		w:          w,
		xl:         xl,
		fid:        fid,
		newfi:      fi,
		globalFrom: from,
		globalTo:   to,
	}
	return nil
}

func (c *Context) refreshInfo() error {
	fi, err := c.client.getFileInfo(c.xl, c.fid, true)
	if err != nil {
		return errors.Info(err, "getFileInfo, fid:", c.fid).Detail(err)
	}
	c.oldfi = c.newfi
	c.newfi = fi
	return nil
}

func (self *Client) rGet(xl *xlog.Logger,
	w io.Writer, suid uint64, from, to uint32) (n int64, err error) {

	defer xl.Xtrack("rGet", time.Now(), &err)

	xl.Debugf("rGet, suid: %v, from: %v, to: %v", suid, from, to)
	sid := suid / (types.N + types.M)
	psectors, err := self.master.Gets(xl, sid)
	if err != nil {
		err = errors.Info(err, "master.gets Failed", sid).Detail(err)
		return
	}
	srgi := &ecb.StripeRgetInfo{
		Soff:    from,
		Bsize:   uint32(to - from),
		BadSuid: suid,
		Psects:  psectors,
	}

	return self.rget.WriteTo(xl, w, srgi)
}

func (self *Client) estgGet(xl *xlog.Logger,
	w io.Writer, host string, psector, suid uint64, from, to uint32) (n int64, err error) {

	defer xl.Xtrack("eGet", time.Now(), &err)

	r, err := ebdstg.Get(host, xl, psector, from, to-from, suid)
	if err != nil {
		return
	}
	defer r.Close()
	n, err = io.Copy(w, r)
	return
}

func (self *Client) getFileInfo(xl *xlog.Logger, fid uint64, cache bool) (fi *master.FileInfo, err error) {
	defer xl.Xtrack("m.Get", time.Now(), &err)

	fi, err = self.master.Get(xl, fid)
	if err != nil {
		errors.Info(err, "self.master.Get").Detail(err)
		return
	}
	if cache {
		if errMc := self.memcached.Set(xl, fid, fi); errMc != nil {
			errors.Info(errMc, "memcached.Set").LogError(xl.ReqId())
		}
	}
	return
}

type xlReadCloser struct {
	io.ReadCloser
	ch chan *xlog.Logger
	xl *xlog.Logger
}

func (self *xlReadCloser) Close() error {
	err := self.ReadCloser.Close()
	xl1 := <-self.ch
	self.xl.Xput(xl1.Xget())
	return err
}

type errWriteRecoder struct {
	w    io.Writer
	errW error
}

func (self *errWriteRecoder) Write(p []byte) (n int, err error) {
	n, err = self.w.Write(p)
	self.errW = err
	return
}

func checkFileInfo(old, new *master.FileInfo) int {
	if old.Soff != new.Soff {
		return Diff_Recycle
	}
	if !equal(old.Suids, new.Suids) {
		return Diff_Recycle
	}
	if !equal(old.Psectors, new.Psectors) {
		return Diff_Repair
	}
	return Same
}

func equal(x, y []uint64) bool {
	if len(x) != len(y) {
		return false
	}
	for i := range x {
		if x[i] != y[i] {
			return false
		}
	}
	return true
}
