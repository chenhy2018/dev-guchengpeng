package api

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/qiniu/http/bsonrpc.v1"
	"github.com/qiniu/http/httputil.v1"
	"github.com/qiniu/http/rpcutil.v1"
	"github.com/qiniu/http/webroute.v1"
	"github.com/qiniu/http/wsrpc.v1"
	"github.com/qiniu/log.v1"
	"github.com/qiniu/rpc.v1"
	"github.com/qiniu/xlog.v1"
	"github.com/stretchr/testify/assert"
	"qbox.us/ebd/api/types"
	ebdcfg "qbox.us/ebdcfg/api"
	"qbox.us/ebdcfg/api/qconf/stgapi"
	master "qbox.us/ebdmaster/api"
	ebdstg "qbox.us/ebdstg/api"
	ecb "qbox.us/ecb/api"
	"qbox.us/errors"
	"qbox.us/fh/fhver"
	"qbox.us/limit"
	pfdtypes "qbox.us/pfd/api/types"
)

const SUMaxLenForTest = 5

var PSECTORs = map[uint64]*struct {
	host        string
	data        []byte
	suid        uint64
	removed     bool
	broken      bool
	recycled    bool
	checkSumErr bool
}{
	0x100000001: {
		host: EstgHost,
		data: []byte{1, 2, 3, 4, 5},
		suid: 100,
	},
	0x200000002: {
		host: EstgHost,
		data: []byte{101, 102, 103, 104, 105},
		suid: 200,
	},
	0x300000003: {
		host: EstgHost,
		data: []byte{201, 202, 203, 204, 205},
		suid: 300,
	},

	0x2222222200000002: { // 0x200000002's son
		host: EstgHost,
		data: []byte{101, 102, 103, 104, 105},
		suid: 200,
	},

	0x400000004: {
		host: EstgHost,
		data: []byte{0, 0, 3, 4, 5},
		suid: 400,
	},
	0x500000005: {
		host: EstgHost,
		data: []byte{101, 102, 0, 0, 0},
		suid: 500,
	},
}

var FIDs = map[uint64]*master.FileInfo{
	11: &master.FileInfo{
		Soff:     0,
		Fsize:    int64(SUMaxLenForTest),
		Suids:    []uint64{100},
		Psectors: []uint64{0x100000001},
	},
	12: &master.FileInfo{
		Soff:     1,
		Fsize:    int64(SUMaxLenForTest) - 2,
		Suids:    []uint64{100},
		Psectors: []uint64{0x100000001},
	},
	21: &master.FileInfo{
		Soff:     0,
		Fsize:    2 * int64(SUMaxLenForTest),
		Suids:    []uint64{100, 200},
		Psectors: []uint64{0x100000001, 0x200000002},
	},
	22: &master.FileInfo{
		Soff:     2,
		Fsize:    int64(SUMaxLenForTest),
		Suids:    []uint64{100, 200},
		Psectors: []uint64{0x100000001, 0x200000002},
	},

	31: &master.FileInfo{
		Soff:     0,
		Fsize:    3 * int64(SUMaxLenForTest),
		Suids:    []uint64{100, 200, 300},
		Psectors: []uint64{0x100000001, 0x200000002, 0x300000003},
	},
	32: &master.FileInfo{
		Soff:     2,
		Fsize:    2 * int64(SUMaxLenForTest),
		Suids:    []uint64{100, 200, 300},
		Psectors: []uint64{0x100000001, 0x200000002, 0x300000003},
	},

	41: &master.FileInfo{
		Soff:     0,
		Fsize:    2 * int64(SUMaxLenForTest),
		Suids:    []uint64{200, 300},
		Psectors: []uint64{0x200000002, 0x300000003},
	},
	42: &master.FileInfo{
		Soff:     2,
		Fsize:    int64(SUMaxLenForTest),
		Suids:    []uint64{200, 300},
		Psectors: []uint64{0x200000002, 0x300000003},
	},
}

var (
	fh11 = pfdtypes.EncodeFh(&pfdtypes.FileHandle{Fid: 11})
	fh12 = pfdtypes.EncodeFh(&pfdtypes.FileHandle{Fid: 12})
	fh21 = pfdtypes.EncodeFh(&pfdtypes.FileHandle{Fid: 21})
	fh22 = pfdtypes.EncodeFh(&pfdtypes.FileHandle{Fid: 22})
)

var c *Client

var (
	MasterHost string
	EcbHost    string
	EstgHost   string
	EbdcfgHost string

	mockEcb *MockEcb
)

func init() {
	os.Remove("ecbs.conf")
	os.Remove("estgs.conf")

	log.SetOutputLevel(0)
	runtime.GOMAXPROCS(4)
	SUMaxLen = SUMaxLenForTest

	mockMaster := &MockMaster{}
	routerMaster := webroute.Router{Mux: http.NewServeMux(), Factory: wsrpc.Factory}
	svrMaster := httptest.NewServer(routerMaster.Register(mockMaster))
	MasterHost = svrMaster.URL
	log.Println("master:", MasterHost)

	mockEcb = &MockEcb{}
	routerEcb := webroute.Router{Mux: http.NewServeMux()}
	svrEcb := httptest.NewServer(routerEcb.Register(mockEcb))
	EcbHost = svrEcb.URL
	log.Println("ecb:", EcbHost)

	mockEStg := &MockEStg{}
	routerEStg := webroute.Router{Mux: http.NewServeMux(), Factory: wsrpc.Factory}
	svrEStg := httptest.NewServer(routerEStg.Register(mockEStg))
	EstgHost = svrEStg.URL

	mockEbdcfg := &MockEbdcfg{}
	routerEbdcfg := webroute.Router{Mux: http.NewServeMux(), Factory: wsrpc.Factory.Union(bsonrpc.Factory)}
	svrEbdcfg := httptest.NewServer(routerEbdcfg.Register(mockEbdcfg))
	EbdcfgHost = svrEbdcfg.URL

	cfg := &Config{
		Guid:        "0123456789012345",
		MasterHosts: []string{MasterHost},
		Memcached:   nil,
		EbdCfgHost:  EbdcfgHost,
		ReloadMs:    100,
		TimeoutMs:   0,
	}

	var err error
	c, err = New(cfg)
	if err != nil {
		log.Fatalln(err)
	}
	c.memcached = &MockMemCache{}
}

func doTestFid(t *testing.T, xl *xlog.Logger, fid uint64, from, to int64, data []byte) {
	fh := pfdtypes.EncodeFh(&pfdtypes.FileHandle{
		Ver:   fhver.FhPfdV2,
		Tag:   pfdtypes.FileHandle_ChunkBits,
		Fid:   fid,
		Fsize: int64(len(data)),
	})
	rc, fsize, err := c.Get(xl, fh, from, to)
	if !assert.NoError(t, err) {
		t.Log(errors.Detail(err))
		return
	}
	defer rc.Close()
	assert.Equal(t, to-from, fsize)
	b, err := ioutil.ReadAll(rc)
	if !assert.NoError(t, err) {
		t.Log(errors.Detail(err))
		return
	}
	assert.Equal(t, data[from:to], b)
	t.Log(fid, from, to)
}

func doTestNormalGet(t *testing.T, xl *xlog.Logger) {
	for fid, fi := range FIDs {
		xl := xlog.NewWith(xl.ReqId() + "/" + strconv.FormatUint(fid, 10))
		var data []byte
		for _, psector := range fi.Psectors {
			data = bytes.Join([][]byte{data, PSECTORs[psector].data}, nil)
		}
		xl.Debug(data, fi.Soff, int64(fi.Soff)+fi.Fsize)
		data = data[fi.Soff : int64(fi.Soff)+fi.Fsize]

		doTestFid(t, xl, fid, 0, int64(len(data)), data)
		doTestFid(t, xl, fid, 1, int64(len(data)), data)
		doTestFid(t, xl, fid, int64(len(data))/2, int64(len(data)), data)
		doTestFid(t, xl, fid, 0, int64(len(data))/2, data)
	}
}

func TestRgetLimit(t *testing.T) {
	cfg := &Config{
		Guid:        "0123456789012345",
		MasterHosts: []string{MasterHost},
		Memcached:   nil,
		EbdCfgHost:  EbdcfgHost,
		ReloadMs:    100,
		TimeoutMs:   0,
		RgetLimit:   1,
	}

	var err error
	c, err := New(cfg)
	if err != nil {
		log.Fatalln(err)
	}
	c.memcached = &MockMemCache{}

	mockEcb.Closed = false
	mockEcb.Block = make(chan struct{})
	xl := xlog.NewWith("RgetLimit")
	doTestNormalGet(t, xl)

	PSECTORs[0x200000002].broken = true
	defer func() {
		PSECTORs[0x200000002].broken = false
	}()

	wg := sync.WaitGroup{}
	wg.Add(2)
	var fid uint64 = 11
	fi := FIDs[fid]
	for _, psector := range fi.Psectors {
		PSECTORs[psector].broken = true
	}
	defer func() {
		for _, psector := range fi.Psectors {
			PSECTORs[psector].broken = false
		}
	}()

	var data []byte
	for _, psector := range fi.Psectors {
		data = bytes.Join([][]byte{data, PSECTORs[psector].data}, nil)
	}
	xl.Debug(data, fi.Soff, int64(fi.Soff)+fi.Fsize)
	data = data[fi.Soff : int64(fi.Soff)+fi.Fsize]
	fh := pfdtypes.EncodeFh(&pfdtypes.FileHandle{
		Ver:   fhver.FhPfdV2,
		Tag:   pfdtypes.FileHandle_ChunkBits,
		Fid:   fid,
		Fsize: int64(len(data)),
	})

	var errs uint32
	go func() {
		xl = xlog.NewWith("TestRgetLimit/first")
		rc, n, err := c.Get(xl, fh, 0, int64(len(data)))
		assert.NoError(t, err)
		assert.Equal(t, n, len(data))
		b, err := ioutil.ReadAll(rc)
		if err != nil {
			assert.Equal(t, err.Error(), limit.ErrLimit.Error())
			atomic.AddUint32(&errs, 1)
			close(mockEcb.Block)
		} else {
			assert.Equal(t, b, data)
		}
		wg.Done()

	}()
	go func() {
		xl = xlog.NewWith("TestRgetLimit/second")
		rc, n, err := c.Get(xl, fh, 0, int64(len(data)))
		assert.NoError(t, err)
		assert.Equal(t, n, len(data))
		b, err := ioutil.ReadAll(rc)
		if err != nil {
			assert.Equal(t, err.Error(), limit.ErrLimit.Error())
			atomic.AddUint32(&errs, 1)
			close(mockEcb.Block)
		} else {
			assert.Equal(t, b, data)
		}
		wg.Done()
	}()
	wg.Wait()
	assert.Equal(t, errs, 1)
	mockEcb.Block = nil
}

func TestNormalGet(t *testing.T) {
	mockEcb.Closed = true
	c.memcached.(Cleaner).Clear()
	xl := xlog.NewWith("TestNormalGet")
	doTestNormalGet(t, xl)
}

func TestFakeFsizeInFhi(t *testing.T) {
	fid := uint64(11)
	fsize := FIDs[fid].Fsize
	fakeFsize := fsize + 1
	fh := pfdtypes.EncodeFh(&pfdtypes.FileHandle{
		Ver:   fhver.FhPfdV2,
		Tag:   pfdtypes.FileHandle_ChunkBits,
		Fid:   fid,
		Fsize: fakeFsize,
	})
	xl := xlog.NewDummy()
	_, _, err := c.Get(xl, fh, 0, fakeFsize)
	if assert.Error(t, err) {
		assert.Equal(t, fmt.Sprintf("to(%v) is bigger than fsize(%v)", fakeFsize, fsize), err.Error())
	}
}

func replacePsectorInFids(old, new uint64) {
	for _, fi := range FIDs {
		for i := range fi.Psectors {
			if fi.Psectors[i] == old {
				fi.Psectors[i] = new
			}
		}
	}
}

func TestGetAfterRepair(t *testing.T) {
	for fid, fi := range FIDs {
		for i := float64(0); i < 1; i += .1 {
			for j := float64(1); j > i && (j-i) > 1e-7; j -= .1 {
				var data []byte
				for _, psector := range fi.Psectors {
					data = bytes.Join([][]byte{data, PSECTORs[psector].data}, nil)
				}
				data = data[fi.Soff : int64(fi.Soff)+fi.Fsize]

				mockEcb.Closed = true
				c.memcached.(Cleaner).Clear()
				xl := xlog.NewWith(fmt.Sprintf("%v/%v/%.1f/%.1f", "TestGetAfterRepair/prepare", fid, i, j))
				doTestNormalGet(t, xl)

				PSECTORs[0x200000002].removed = true
				replacePsectorInFids(0x200000002, 0x2222222200000002)

				from, to := int64(i*float64(len(data))), int64(j*float64(len(data)))

				xl = xlog.NewWith(fmt.Sprintf("%v/%v/%.1f/%.1f", "TestGetAfterRepair/first", fid, i, j))
				doTestFid(t, xl, fid, from, to, data)
				xl = xlog.NewWith(fmt.Sprintf("%v/%v/%.1f/%.1f", "TestGetAfterRepair/second", fid, i, j))
				doTestFid(t, xl, fid, from, to, data)

				PSECTORs[0x200000002].removed = false
				replacePsectorInFids(0x2222222200000002, 0x200000002)
			}
		}
	}
}

func TestGetRepairing(t *testing.T) {
	mockEcb.Closed = false
	c.memcached.(Cleaner).Clear()
	xl := xlog.NewWith("TestGetRepairing/prepare")
	doTestNormalGet(t, xl)

	PSECTORs[0x200000002].removed = true
	defer func() {
		PSECTORs[0x200000002].removed = false
	}()

	xl = xlog.NewWith("TestGetRepairing/first")
	doTestNormalGet(t, xl)
	xl = xlog.NewWith("TestGetRepairing/second")
	doTestNormalGet(t, xl)
}

func TestGetBroken(t *testing.T) {
	mockEcb.Closed = false
	c.memcached.(Cleaner).Clear()
	xl := xlog.NewWith("TestGetBroken/prepare")
	doTestNormalGet(t, xl)

	PSECTORs[0x200000002].broken = true
	defer func() {
		PSECTORs[0x200000002].broken = false
	}()

	xl = xlog.NewWith("TestGetBroken/first")
	doTestNormalGet(t, xl)
	xl = xlog.NewWith("TestGetBroken/second")
	doTestNormalGet(t, xl)
}

func TestGetRecycle(t *testing.T) {
	xl := xlog.NewWith("TestGetRecycle")
	doTestNormalGet(t, xl)

	var fid = uint64(22)
	oldfi := FIDs[fid]
	oldPsects := oldfi.Psectors

	newfi := &master.FileInfo{
		Soff:     2,
		Fsize:    int64(SUMaxLenForTest),
		Suids:    []uint64{400, 500},
		Psectors: []uint64{0x400000004, 0x500000005},
	}
	FIDs[fid] = newfi
	PSECTORs[0x200000002].recycled = true

	xl = xlog.NewWith(xl.ReqId() + "/" + strconv.FormatUint(fid, 10))
	var data []byte
	for _, psector := range oldPsects {
		data = bytes.Join([][]byte{data, PSECTORs[psector].data}, nil)
	}
	xl.Debug(data, oldfi.Soff, int64(oldfi.Soff)+oldfi.Fsize)
	data = data[oldfi.Soff : int64(oldfi.Soff)+oldfi.Fsize]

	doTestFid(t, xl, fid, 0, int64(len(data)), data)
	doTestFid(t, xl, fid, 1, int64(len(data)), data)
	doTestFid(t, xl, fid, int64(len(data))/2, int64(len(data)), data)
	doTestFid(t, xl, fid, 0, int64(len(data))/2, data)

	FIDs[fid] = oldfi
	PSECTORs[0x200000002].recycled = false
}

func TestGetCheckSumErrWithRecycle(t *testing.T) {
	xl := xlog.NewWith("TestGetCheckSumErrWithRecycle")
	doTestNormalGet(t, xl)

	var fid = uint64(22)
	oldfi := FIDs[fid]
	oldPsects := oldfi.Psectors

	newfi := &master.FileInfo{
		Soff:     2,
		Fsize:    int64(SUMaxLenForTest),
		Suids:    []uint64{400, 500},
		Psectors: []uint64{0x400000004, 0x500000005},
	}
	FIDs[fid] = newfi
	PSECTORs[0x200000002].recycled = true
	PSECTORs[0x500000005].checkSumErr = true

	xl = xlog.NewWith(xl.ReqId() + "/" + strconv.FormatUint(fid, 10))
	var data []byte
	for _, psector := range oldPsects {
		data = bytes.Join([][]byte{data, PSECTORs[psector].data}, nil)
	}
	xl.Debug(data, oldfi.Soff, int64(oldfi.Soff)+oldfi.Fsize)
	data = data[oldfi.Soff : int64(oldfi.Soff)+oldfi.Fsize]

	doTestFid(t, xl, fid, 0, int64(len(data)), data)
	doTestFid(t, xl, fid, 1, int64(len(data)), data)
	doTestFid(t, xl, fid, int64(len(data))/2, int64(len(data)), data)
	doTestFid(t, xl, fid, 0, int64(len(data))/2, data)
}

// ===================================================================

type MockMaster struct{}

type masterGetArgs struct {
	Fid uint64 `json:"fid"`
}

func (self *MockMaster) WsGet(args *masterGetArgs, env rpcutil.Env) {
	w := env.W
	if fi, ok := FIDs[args.Fid]; ok {
		httputil.Reply(w, 200, fi)
		return
	}
	httputil.ReplyWithCode(w, 612)
}

type masterGetsArgs struct {
	Sid uint64 `json:"sid"`
}

func (self *MockMaster) WsGets(args *masterGetsArgs, env rpcutil.Env) {
	w := env.W
	httputil.Reply(w, 200, map[string]interface{}{
		"psectors": [types.N + types.M]uint64{},
	})
}

// -------------------------------------------------------------------------------

type MockEcb struct {
	Closed bool
	Block  chan struct{}
}

func (self *MockEcb) DoRget(w http.ResponseWriter, req *http.Request) {

	xl := xlog.New(w, req)
	_ = xl
	if self.Block != nil {
		<-self.Block
	}

	if self.Closed {
		httputil.ReplyErr(w, 500, "should not run into rget")
		return
	}

	srgi, err := ecb.ReadStripeRgetInfo(req.Body)
	if err != nil {
		httputil.Error(w, err)
		return
	}
	for _, info := range PSECTORs {
		if info.suid == srgi.BadSuid {
			if info.recycled == false {
				if info.checkSumErr == true {
					xl.Infof("checksum error, load data by rget request, suid: %v", info.suid)
				}
				data := info.data[srgi.Soff : srgi.Soff+srgi.Bsize]
				httputil.ReplyWith(w, 200, "application/octet-stream", data)
				return
			}
			httputil.ReplyErr(w, 599, "too many error units")
			return
		}
	}

	panic("should not reach here")
}

// -------------------------------------------------------------------------------

type MockEStg struct{}

type estgGetArgs struct {
	Psect uint64 `flag:"_"`
	Off   uint32 `flag:"off"`
	Size  uint32 `flag:"n"`
	Suid  uint64 `flag:"suid"`
}

func (self *MockEStg) CmdGet_(args *estgGetArgs, env rpcutil.Env) {
	w, req := env.W, env.Req
	xl := xlog.New(w, req)
	_ = xl
	if info, ok := PSECTORs[args.Psect]; ok && !info.removed {
		if info.suid != args.Suid {
			xl.Errorf("suid not match, expect: %v, got: %v\n", info.suid, args.Suid)
			httputil.Error(w, ebdstg.ErrSuidNotMatched)
			return
		}
		data := info.data[args.Off : args.Off+args.Size]
		if info.recycled {
			xl.Errorf("suid recycled, old suid: %v, new suid: %v\n", info.suid, args.Suid)
			httputil.Error(w, ebdstg.ErrSuidNotMatched)
			return
		}
		if info.checkSumErr {
			xl.Errorf("psect checksume error, suid: %v\n", args.Psect, info.suid)
			httputil.Error(w, ebdstg.ErrChecksumError)
			return
		}
		if info.broken {
			body := &oneByteErrorReader{bytes.NewReader(data)}
			httputil.ReplyWithStream(w, 200, "application/octet-stream", body, int64(len(data)))
			return
		}
		httputil.ReplyWith(w, 200, "application/octet-stream", data)
		return
	}
	httputil.Error(w, ebdstg.ErrNoSuchDisk)
}

type oneByteErrorReader struct {
	r io.Reader
}

func (self *oneByteErrorReader) Read(p []byte) (int, error) {
	if len(p) > 1 {
		n, err := self.r.Read(p[:1])
		if err == nil {
			err = errors.New("time out")
		}
		return n, err
	}
	return self.r.Read(p)
}

// -------------------------------------------------------------------------------

type MockMemCache struct {
	l sync.Mutex
	m map[uint64]*master.FileInfo
}

// -------------------------------------------------------------------------------

func (self *MockMemCache) Set(l rpc.Logger, fid uint64, fi *master.FileInfo) error {
	self.l.Lock()
	defer self.l.Unlock()
	if self.m == nil {
		self.m = make(map[uint64]*master.FileInfo)
	}
	self.m[fid] = fi
	return nil
}

func (self *MockMemCache) Get(l rpc.Logger, fid uint64) (*master.FileInfo, error) {
	self.l.Lock()
	defer self.l.Unlock()
	if self.m == nil {
		self.m = make(map[uint64]*master.FileInfo)
	}
	if fi, ok := self.m[fid]; ok {
		return fi, nil
	}
	return nil, memcache.ErrCacheMiss
}

func (self *MockMemCache) Clear() {
	self.l.Lock()
	defer self.l.Unlock()
	self.m = make(map[uint64]*master.FileInfo)
}

type Cleaner interface {
	Clear()
}

// -------------------------------------------------------------------------------

type MockEbdcfg struct {
}
type getbArg struct {
	Id string `json:"id"`
}

func (self *MockEbdcfg) WbrpcGetb(args *getbArg, env *rpcutil.Env) (doc map[string]interface{}, err error) {

	_, diskId, err := stgapi.ParseDiskId(args.Id)
	if err != nil {
		return
	}
	for psector, pi := range PSECTORs {
		if pi.removed {
			continue
		}
		if diskId0, _ := types.DecodePsect(psector); diskId0 == diskId {
			hosts := [2]string{EstgHost, EstgHost}
			return map[string]interface{}{"hosts": hosts}, nil
		}
	}
	return nil, httputil.NewError(612, "not found")
}

type listArg struct {
	Guid string `json:"guid"`
}

func (self *MockEbdcfg) WsEcbList(args *listArg, env *rpcutil.Env) (ecbs []*ebdcfg.EcbInfo, err error) {

	ecbs = []*ebdcfg.EcbInfo{
		&ebdcfg.EcbInfo{
			Guid:  "0123456789012345",
			Hosts: [2]string{EcbHost, EcbHost},
		},
	}
	return
}

func (self *MockEbdcfg) WsDiskList(args *listArg, env *rpcutil.Env) (disks []*ebdcfg.DiskInfo, err error) {

	//只给一个不会换的，防止内存缓存
	psector := uint64(0x100000001)
	diskId, _ := types.DecodePsect(psector)
	disks = []*ebdcfg.DiskInfo{
		{
			Guid:   "0123456789012345",
			DiskId: diskId,
			Hosts:  [2]string{EstgHost, EstgHost},
		},
	}
	return
}

func TestClean(t *testing.T) {
	os.Remove("ecbs.conf")
	os.Remove("estgs.conf")
}
