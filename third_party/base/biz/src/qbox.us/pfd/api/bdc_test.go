package api

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"qbox.us/errors"
	"qbox.us/fh/fhver"
	"qbox.us/pfd/api/types"
	cfgapi "qbox.us/pfdcfg/api"
	stgapi "qbox.us/pfdstg/api"
	trackerapi "qbox.us/pfdtracker/api"

	"github.com/qiniu/http/httputil.v1"
	"github.com/qiniu/http/rpcutil.v1"
	"github.com/qiniu/http/webroute.v1"
	"github.com/qiniu/http/wsrpc.v1"
	"github.com/qiniu/log.v1"
	"github.com/qiniu/mockhttp.v1"
	"github.com/qiniu/rpc.v1"
	"github.com/qiniu/xlog.v1"
	"github.com/stretchr/testify/assert"
	"qbox.us/pfdtracker/stater"
)

var (
	ErrDgidNotFound = httputil.NewError(612, "dgid not found")
	ErrGidNotFound  = httputil.NewError(612, "gid not found")
	ErrFileNotFound = httputil.NewError(404, "file not found")
	ErrAllocedEntry = httputil.NewError(stgapi.StatusAllocedEntry, "alloced entry")
	ErrNotAvail     = httputil.NewError(555, "not enough space")
)

const (
	SmallFileSize  = 8
	SmallFileLimit = 10
	BigFileSize    = 12
)

func init() {
	log.SetOutputLevel(0)
}

func randbytes(size int) ([]byte, string) {
	b := make([]byte, size)
	io.ReadFull(rand.Reader, b[:])
	return b, base64.URLEncoding.EncodeToString(b[:])
}

func TestDelete(t *testing.T) {
	xl := xlog.NewWith("TestDelete")

	mockDgInfoer := &MockDgInfoer{
		mutex: sync.RWMutex{},
	}
	mockStater := NewMockGidStater()

	// one gid 100
	mstg1 := StartDgidStgs(100, true) // start dgid 100's stg servers
	gid1 := types.NewGid(100)
	mstg1.AddGid(gid1)
	mockStater.AddGidDgid(gid1, 100)
	mockDgInfoer.AddDg(makeDgis([]uint32{100}, cfgapi.SSD))

	cli, err := NewClient(GUID, mockDgInfoer, mockStater, 3, 3, 1, SmallFileLimit, stgapi.TimeoutOption{}, "", nil)
	assert.NoError(t, err)
	stgapi.PostTimeoutTransport = http.DefaultTransport
	stgapi.GetTimeoutTransport = http.DefaultTransport
	stgapi.ProxyGetTimeoutTransport = stgapi.GetTimeoutTransport
	stgapi.ProxyPostTimeoutTransport = stgapi.PostTimeoutTransport

	fh, _, err := putGetSmall(t, cli)
	err = cli.Delete(xl, fh)
	assert.NoError(t, err)

	_, _, err = cli.Get(xl, fh, 0, SmallFileSize)
	assert.Equal(t, err.Error(), ErrFileNotFound.Error())

	mstg1.MakeDeleteFail(1)
	fh, _, err = putGetSmall(t, cli)
	assert.NoError(t, err)
	err = cli.Delete(xl, fh)
	assert.Error(t, err)
	err = cli.Delete(xl, fh)
	assert.NoError(t, err)

	// delete with isBackup
	fh, _, err = putGetSmall(t, cli)
	assert.NoError(t, err)
	mockDgInfoer.Dgis[0].IsBackup = []bool{false, false, true}
	time.Sleep(2 * time.Second)
	err = cli.Delete(xl, fh)
	assert.Equal(t, trackerapi.ErrGidECed, err)

	// put before transfer
	fh, _, err = putGetSmall(t, cli)

	xl.Info("transfer 100 to 200")
	// add stg group, dgid 200
	// transfer gid 100 to dgid 200
	mstg2 := StartDgidStgs(200, true)
	{
		mstg2.Stg = mstg1.Stg
		mstg1.Stg = nil
	}
	mockStater.RemoveGid(gid1)
	mockDgInfoer.RemoveDg(100)
	mockStater.AddGidDgid(gid1, 200)
	mockDgInfoer.AddDg(makeDgis([]uint32{200}, cfgapi.DEFAULT))
	mockStater.SetGidEcing(gid1, 1)

	err = cli.Delete(xl, fh)
	assert.Error(t, err)

	mockStater.SetGidEcing(gid1, 0)

	err = cli.Delete(xl, fh)
	assert.NoError(t, err)

	//_, _, err = cli.Get(xl, fh, 0, SmallFileSize)
	//assert.Equal(t, err.Error(), ErrFileNotFound.Error())
	//log.Fatal("delete end.")
}

func TestAllocPutat(t *testing.T) {
	xl := xlog.NewWith("TestAllocPutat")

	data := []byte("DEFvfgbvfd")
	fsize := int64(len(data))

	mockDgInfoer := &MockDgInfoer{
		mutex: sync.RWMutex{},
	}
	mockStater := NewMockGidStater()

	// first new gid
	mstg1 := StartDgidStgs(10, true) // start dgid 10's stg servers
	gid1 := types.NewGid(10)
	mstg1.AddGid(gid1)
	mockStater.AddGidDgid(gid1, 10)
	mockDgInfoer.AddDg(makeDgis([]uint32{10}, cfgapi.DEFAULT))

	cli, err := NewClient(GUID, mockDgInfoer, mockStater, 3, 3, 1, SmallFileLimit, stgapi.TimeoutOption{}, "", nil)
	assert.NoError(t, err)
	stgapi.PostTimeoutTransport = http.DefaultTransport
	stgapi.GetTimeoutTransport = http.DefaultTransport
	stgapi.ProxyGetTimeoutTransport = stgapi.GetTimeoutTransport
	stgapi.ProxyPostTimeoutTransport = stgapi.PostTimeoutTransport

	fh1, err := cli.Alloc(xl, 10, [20]byte{1, 2, 3})
	assert.NoError(t, err)

	fh2, err := cli.Alloc(xl, 10, [20]byte{1, 2, 3})
	assert.NoError(t, err)

	err = cli.PutAt(xl, bytes.NewReader(data), fh2)
	assert.NoError(t, err)

	for i := 0; i < 10; i++ {
		_, _, err := cli.Get(xl, fh1, 0, fsize)
		assert.Error(t, err)
		assert.Equal(t, stgapi.StatusAllocedEntry, httputil.DetectCode(err))
	}

	for i := 0; i < 10; i++ {
		rc, n, err := cli.Get(xl, fh2, 0, fsize)
		assert.NoError(t, err)
		assert.Equal(t, fsize, n)
		b, _ := ioutil.ReadAll(rc)
		rc.Close()
		assert.Equal(t, data, b)
	}

	// add stg group, dgid 40
	// transfer gid 10 to dgid 40
	mstg4 := StartDgidStgs(40, true)
	{
		mstg4.Stg = mstg1.Stg
		mstg1.Stg = nil
	}
	mockStater.RemoveGid(gid1)
	mockStater.AddGidDgid(gid1, 40)
	mockDgInfoer.AddDg(makeDgis([]uint32{40}, cfgapi.DEFAULT))
	mockDgInfoer.RemoveDg(10)
	xl.Info("transferred!")

	err = cli.PutAt(xl, bytes.NewReader(data), fh1)
	assert.NoError(t, err)

	for i := 0; i < 10; i++ {
		rc, n, err := cli.Get(xl, fh1, 0, fsize)
		assert.NoError(t, err)
		assert.Equal(t, fsize, n)
		b, _ := ioutil.ReadAll(rc)
		rc.Close()
		assert.Equal(t, data, b)
	}

	for i := 0; i < 10; i++ {
		rc, n, err := cli.Get(xl, fh2, 0, fsize)
		assert.NoError(t, err)
		assert.Equal(t, fsize, n)
		b, _ := ioutil.ReadAll(rc)
		rc.Close()
		assert.Equal(t, data, b)
	}
}

func TestReadonly(t *testing.T) {
	mockDgInfoer := &MockDgInfoer{
		mutex: sync.RWMutex{},
	}
	mockStater := NewMockGidStater()

	// first new gid
	mstg1 := StartDgidStgs(1, true) // start dgid 1's stg servers
	gid1 := types.NewGid(1)
	mstg1.AddGid(gid1)
	mockStater.AddGidDgid(gid1, 1)
	mockDgInfoer.AddDg(makeDgis([]uint32{1}, cfgapi.SSD))

	cli, err := NewClient(GUID, mockDgInfoer, mockStater, 3, 0, 1, SmallFileLimit, stgapi.TimeoutOption{}, "", nil)
	if err != nil {
		t.Fatal("new pfd client failed:", err)
	}
	stgapi.PostTimeoutTransport = http.DefaultTransport
	stgapi.GetTimeoutTransport = http.DefaultTransport
	stgapi.ProxyGetTimeoutTransport = stgapi.GetTimeoutTransport
	stgapi.ProxyPostTimeoutTransport = stgapi.PostTimeoutTransport

	fh, fsize, err := putGetSmall(t, cli)
	if err != nil {
		t.Fatal("put get failed ->", err)
	}

	for _, dgi := range mockDgInfoer.Dgis {
		dgi.ReadOnly = 1
	}
	// wait for refreshDg
	time.Sleep(2 * time.Second)

	rc, sz, err := cli.Get(nil, fh, 0, fsize)
	if err != nil {
		t.Fatal(err)
		return
	}
	if sz != fsize {
		t.Fatal(err)
	}
	n, err := io.Copy(ioutil.Discard, rc)
	assert.NoError(t, err)
	assert.Equal(t, fsize, n)

	_, _, err = cli.Put(nil, bytes.NewReader([]byte("a")), 1)
	if err == nil {
		t.Fatal("error should not be nil, because all dgs are readonly")
	}
	//	assert.Equal(t, "Put: no host for put", err.Error())
	assert.Equal(t, "disk matrix not found", err.Error())

	for _, dgi := range mockDgInfoer.Dgis {
		dgi.ReadOnly = 0
	}
	// wait for refreshDg
	time.Sleep(2 * time.Second)

	rc, sz, err = cli.Get(nil, fh, 0, fsize)
	if err != nil {
		t.Fatal(err)
		return
	}
	if sz != fsize {
		t.Fatal(err)
	}
	n, err = io.Copy(ioutil.Discard, rc)
	assert.NoError(t, err)
	assert.Equal(t, fsize, n)

	_, _, err = putGetSmall(t, cli)
	if err != nil {
		t.Fatal("put get failed ->", err)
	}
}

func TestPutGetNoSSD(t *testing.T) {
	mockDgInfoer := &MockDgInfoer{
		mutex: sync.RWMutex{},
	}
	mockStater := NewMockGidStater()

	// first new gid
	mstg1 := StartDgidStgs(1, true) // start dgid 1's stg servers
	gid1 := types.NewGid(1)
	mstg1.AddGid(gid1)
	mockStater.AddGidDgid(gid1, 1)
	mockDgInfoer.AddDg(makeDgis([]uint32{1}, cfgapi.DEFAULT))

	cli, err := NewClient(GUID, mockDgInfoer, mockStater, 3, 0, 1, SmallFileLimit, stgapi.TimeoutOption{}, "", nil)
	if err != nil {
		t.Fatal("new pfd client failed:", err)
	}
	stgapi.PostTimeoutTransport = http.DefaultTransport
	stgapi.GetTimeoutTransport = http.DefaultTransport
	stgapi.ProxyGetTimeoutTransport = stgapi.GetTimeoutTransport
	stgapi.ProxyPostTimeoutTransport = stgapi.PostTimeoutTransport

	// simple put get
	for i := 0; i < 100; i++ {
		_, _, err = putGetSmall(t, cli)
		if err != nil {
			t.Fatal("put get failed ->", err)
		}
	}
}

func TestPutGetProxy(t *testing.T) {
	{
		mockDgInfoer := &MockDgInfoer{
			mutex: sync.RWMutex{},
		}
		mockStater := NewMockGidStater()

		// first new gid
		mstg1 := StartDgidStgs(1, true) // start dgid 1's stg servers
		gid1 := types.NewGid(1)
		mstg1.AddGid(gid1)
		mockStater.AddGidDgid(gid1, 1)
		mockDgInfoer.AddDg(makeDgis([]uint32{1}, cfgapi.SSD))
		mockDgInfoer.Dgis[0].Idc = []string{"nb", "nb", "nb"}
		mockDgInfoer.Dgis[0].IsBackup = []bool{false, false, true}

		cli, err := NewClient(GUID, mockDgInfoer, mockStater, 3, 0, 1, SmallFileLimit, stgapi.TimeoutOption{}, "nb", nil)
		cli2, err := NewClient(GUID, mockDgInfoer, mockStater, 2, 0, 1, SmallFileLimit, stgapi.TimeoutOption{}, "xs", nil)
		if err != nil {
			t.Fatal("new pfd client failed:", err)
		}
		stgapi.PostTimeoutTransport = http.DefaultTransport
		stgapi.GetTimeoutTransport = http.DefaultTransport
		stgapi.ProxyPostTimeoutTransport = stgapi.PostTimeoutTransport
		var times int
		stgapi.ProxyGetTimeoutTransport = &http.Transport{
			Proxy: func(*http.Request) (*url.URL, error) {
				times++
				return url.Parse("http://2.2.2.2:80")
			},
			Dial: (&net.Dialer{Timeout: 100 * time.Millisecond}).Dial,
		}
		_, _, err = putGet2(t, SmallFileSize, cli, cli2)
		assert.Error(t, err)
		assert.Equal(t, 1, times)
	}
	{
		mockDgInfoer := &MockDgInfoer{
			mutex: sync.RWMutex{},
		}
		mockStater := NewMockGidStater()

		// first new gid
		mstg1 := StartDgidStgs(1, true) // start dgid 1's stg servers
		gid1 := types.NewGid(1)
		mstg1.AddGid(gid1)
		mockStater.AddGidDgid(gid1, 1)
		mockDgInfoer.AddDg(makeDgis([]uint32{1}, cfgapi.SSD))
		mockDgInfoer.Dgis[0].Idc = []string{"nb", "nb", "xs"}
		mockDgInfoer.Dgis[0].IsBackup = []bool{false, false, true}

		cli, err := NewClient(GUID, mockDgInfoer, mockStater, 3, 0, 1, SmallFileLimit, stgapi.TimeoutOption{}, "nb", nil)
		cli2, err := NewClient(GUID, mockDgInfoer, mockStater, 2, 0, 1, SmallFileLimit, stgapi.TimeoutOption{}, "xs", nil)
		if err != nil {
			t.Fatal("new pfd client failed:", err)
		}
		stgapi.PostTimeoutTransport = http.DefaultTransport
		stgapi.GetTimeoutTransport = http.DefaultTransport
		stgapi.ProxyPostTimeoutTransport = stgapi.PostTimeoutTransport
		var times int
		stgapi.ProxyGetTimeoutTransport = &http.Transport{
			Proxy: func(*http.Request) (*url.URL, error) {
				times++
				return url.Parse("http://2.2.2.2:80")
			},
			Dial: (&net.Dialer{Timeout: 100 * time.Millisecond}).Dial,
		}
		_, _, err = putGet2(t, SmallFileSize, cli, cli2)
		assert.NoError(t, err)
		assert.Equal(t, 1, times)
	}
}

var trytimes613 int

func TestPutGet(t *testing.T) {
	mockDgInfoer := &MockDgInfoer{
		mutex: sync.RWMutex{},
	}
	mockStater := NewMockGidStater()

	// first new gid
	mstg1 := StartDgidStgs(1, true) // start dgid 1's stg servers
	gid1 := types.NewGid(1)
	mstg1.AddGid(gid1)
	mockStater.AddGidDgid(gid1, 1)
	mockDgInfoer.AddDg(makeDgis([]uint32{1}, cfgapi.SSD))

	cli, err := NewClient(GUID, mockDgInfoer, mockStater, 3, 0, 1, SmallFileLimit, stgapi.TimeoutOption{}, "", nil)
	cli2, err := NewClient(GUID, mockDgInfoer, mockStater, 2, 0, 1, SmallFileLimit, stgapi.TimeoutOption{}, "", nil)
	if err != nil {
		t.Fatal("new pfd client failed:", err)
	}
	stgapi.PostTimeoutTransport = http.DefaultTransport
	stgapi.GetTimeoutTransport = http.DefaultTransport
	stgapi.ProxyGetTimeoutTransport = stgapi.GetTimeoutTransport
	stgapi.ProxyPostTimeoutTransport = stgapi.PostTimeoutTransport

	// simple put get
	for i := 0; i < 100; i++ {
		_, _, err = putGetSmall(t, cli)
		if err != nil {
			t.Fatal("put get failed ->", err)
		}
	}

	// no host for put big file
	_, _, err = putGetBig(t, cli)
	if err == nil {
		t.Fatal("big file should not put to SSD disk")
	}

	// make put fail
	mstg1.MakePutFail(1)
	_, _, err = putGetSmall(t, cli)
	if err == nil {
		t.Fatal("put get should failed, but not")
	}

	// make get fail
	mstg1.MakeGetFail(3)
	fh, fsize, err := putGetSmall(t, cli)
	if err == nil {
		t.Fatal("get should failed, but not")
	}

	mstg1.MakeGetFail(2)
	fh, fsize, err = putGetSmall(t, cli)
	if err != nil {
		t.Fatal("get should not failed, but not =>", err)
	}

	mstg1.MakeGetFail(3)
	mstg1.MakeGetFailError(httputil.NewError(613, "alloced header"))
	fh, fsize, err = putGetSmall(t, cli)
	if httputil.DetectCode(err) != 613 {
		t.Fatal("get should fail return 613")
	}
	assert.Equal(t, trytimes613, 3)
	trytimes613 = 0

	mstg1.MakeGetFail(2)
	mstg1.MakeGetFailError(httputil.NewError(613, "alloced header"))
	fh, fsize, err = putGetSmall(t, cli)
	if err != nil {
		t.Fatal("get should not failed, but not =>", err)
	}
	assert.Equal(t, trytimes613, 2)

	// add more stg groups, dgid 2, 3
	mstg2 := StartDgidStgs(2, true)
	gid2 := types.NewGid(2)
	mstg2.AddGid(gid2)
	mockStater.AddGidDgid(gid2, 2)
	mockDgInfoer.AddDg(makeDgis([]uint32{2}, cfgapi.DEFAULT))

	mstg3 := StartDgidStgs(3, true)
	gid3 := types.NewGid(3)
	mstg3.AddGid(gid3)
	mockStater.AddGidDgid(gid3, 3)
	mockDgInfoer.AddDg(makeDgis([]uint32{3}, cfgapi.SSD))

	// wait for refreshDg
	time.Sleep(2 * time.Second)

	// make put not avail
	mstg1.MakeNotAvailFail(1)
	mstg2.MakePutFail(1)
	_, _, err = putGetSmallReaderAt(t, cli2)
	if err != nil {
		t.Fatal("put should not fail, but failed")
	}

	mstg1.MakeNotAvailFail(1)
	mstg2.MakePutFail(1)
	_, _, err = putGetSmallReaderAt(t, cli2)
	if err != nil {
		t.Fatal("put should not fail, but not")
	}

	mstg1.MakeNotAvailFail(1)
	mstg2.MakePutFail(1)
	_, _, err = putGetSmallReaderAt(t, cli2)
	if err != nil {
		t.Fatal("put should not fail, but not")
	}

	mstg1.MakeNotAvailFail(0)
	mstg2.MakePutFail(0)

	// make put fail for small file
	mstg1.MakePutFail(1)
	mstg2.MakePutFail(1)
	_, _, err = putGetSmallReaderAt(t, cli)
	if err != nil {
		t.Fatal("put should not fail, but not")
	}

	// make put fail for small file
	mstg1.MakePutFail(1)
	mstg2.MakePutFail(0)
	mstg3.MakePutFail(1)
	fh, _, err = putGetSmallReaderAt(t, cli)
	if err != nil {
		t.Fatal("put should not fail, but not")
	}
	fhi, err := types.DecodeFh(fh)
	assert.NoError(t, err)
	if fhi.Gid.MotherDgid() != 2 {
		t.Fatal("not in sata", fhi.Gid.MotherDgid())
	}
	mstg3.MakePutFail(0)
	mstg1.MakePutFail(0)

	// make put fail for big file
	mstg1.MakePutFail(1)
	mstg2.MakePutFail(1)
	_, _, err = putGetBigReaderAt(t, cli)
	if err == nil {
		t.Fatal("put should fail, but not")
	}

	// test https://pm.qbox.me/issues/11037
	_, _, err = putGetSmallReaderAt(t, cli)
	if err != nil {
		t.Fatal("put shouldn't fail")
	}
	mstg1.MakePutFail(1)
	_, _, err = putGetSmallReaderAt(t, cli)
	if err != nil {
		t.Fatal("put shouldn't fail after retrying")
	}
	mstg1.MakePutFail(1)
	_, _, err = putGetSmallReaderAt(t, cli)
	if err != nil {
		t.Fatal("put shouldn't fail after retrying")
	}

	// add stg group, dgid 4
	// transfer gid 1 to dgid 4
	mstg4 := StartDgidStgs(4, true)
	{
		mstg4.Stg = mstg1.Stg
		mstg1.Stg = nil
	}
	mockStater.RemoveGid(gid1)
	mockStater.AddGidDgid(gid1, 4)
	mockDgInfoer.AddDg(makeDgis([]uint32{4}, cfgapi.DEFAULT))
	mockDgInfoer.RemoveDg(1)

	for _, dgi := range mockDgInfoer.Dgis {
		if dgi.Dgid == 3 {
			dgi.ReadOnly = 1
		}
	}
	time.Sleep(2 * time.Second)
	fh, _, err = putGetSmallReaderAt(t, cli)
	if err != nil {
		t.Fatal("put should not fail, but not")
	}
	fhi, err = types.DecodeFh(fh)
	assert.NoError(t, err)
	if fhi.Gid.MotherDgid() == 3 {
		t.Fatal("not in sata")
	}

	StopDgidStgs(1)
	StopDgidStgs(3)
	StopDgidStgs(4)
	fh, _, err = putGetSmallReaderAt(t, cli)
	if err != nil {
		t.Fatal("put should not fail, but not")
	}
	fhi, err = types.DecodeFh(fh)
	assert.NoError(t, err)
	if fhi.Gid.MotherDgid() == 3 {
		t.Fatal("not in sata")
	}

	rc, size, err := cli.Get(nil, fh, 0, fsize)
	if err != nil {
		t.Fatal("although data transfered, but should get ok, unfortunately, failed ->", err)
	}
	if size != fsize || rc == nil {
		t.Fatal("file size not math or content wrong")
	}

	mockStater.EC(gid1)
	rc, size, err = cli.Get(nil, fh, 0, fsize)
	if err != nil {
		t.Fatal("although data transfered, but should get ok, unfortunately, failed ->", err)
	}
	if size != fsize || rc == nil {
		t.Fatal("file size not math or content wrong")
	}
	time.Sleep(1.2e9)
	_, _, err = cli.Get(xlog.NewWith("TestPutGet-EC"), fh, 0, fsize)
	if err != nil {
		t.Fatal("although data transfered, but should get ok, unfortunately, failed ->", err)
	}
	if size != fsize || rc == nil {
		t.Fatal("file size not math or content wrong")
	}
}

func TestPutCannotDialRetry(t *testing.T) {
	mockDgInfoer := &MockDgInfoer{
		mutex: sync.RWMutex{},
	}
	mockStater := NewMockGidStater()

	mockDgInfoer.AddDg(makeDgis([]uint32{0}, cfgapi.SSD))
	// first new gid
	mstg1 := StartDgidStgs(1, true) // start dgid 1's stg servers
	gid1 := types.NewGid(1)
	mstg1.AddGid(gid1)
	mockStater.AddGidDgid(gid1, 1)
	mockDgInfoer.AddDg(makeDgis([]uint32{1}, cfgapi.SSD))

	cli, err := NewClient(GUID, mockDgInfoer, mockStater, 3, 0, 1, SmallFileLimit, stgapi.TimeoutOption{}, "", nil)
	if err != nil {
		t.Fatal("new pfd client failed:", err)
	}
	stgapi.PostTimeoutTransport = http.DefaultTransport
	stgapi.GetTimeoutTransport = http.DefaultTransport
	stgapi.ProxyGetTimeoutTransport = stgapi.GetTimeoutTransport
	stgapi.ProxyPostTimeoutTransport = stgapi.PostTimeoutTransport

	fh, fsize, err := putGetSmall(t, cli)
	assert.NoError(t, err)

	mstg1.MakePut612(1)
	err = putatGet2(t, fh, fsize, cli, cli)
	assert.NoError(t, err)
}

func TestGetBadDgid(t *testing.T) {
	runtime.GOMAXPROCS(8)
	mockDgInfoer := &MockDgInfoer{
		mutex: sync.RWMutex{},
	}
	mockStater := NewMockGidStater()

	// first new gid
	mstg1 := StartDgidStgs(1, true) // start dgid 1's stg servers
	mstg1.AddFullDgid([]uint32{2, 3, 4, 5})
	gid1 := types.NewGid(1)
	mstg1.AddGid(gid1)
	mockStater.AddGidDgid(gid1, 1)
	var dgis []*cfgapi.DiskGroupInfo
	for _, dgid := range []uint32{1, 2, 3, 4, 5} {
		dgi := &cfgapi.DiskGroupInfo{
			Guid:     GUID,
			Dgid:     dgid,
			Hosts:    makeStgHosts(1, true),
			DiskType: cfgapi.SSD,
		}
		dgis = append(dgis, dgi)
	}
	mockDgInfoer.AddDg(dgis)

	cli, err := NewClient(GUID, mockDgInfoer, mockStater, 3, 0, 1, SmallFileLimit, stgapi.TimeoutOption{}, "", nil)
	if err != nil {
		t.Fatal("new pfd client failed:", err)
	}
	stgapi.PostTimeoutTransport = http.DefaultTransport
	stgapi.GetTimeoutTransport = http.DefaultTransport
	stgapi.ProxyGetTimeoutTransport = stgapi.GetTimeoutTransport
	stgapi.ProxyPostTimeoutTransport = stgapi.PostTimeoutTransport

	for i := 0; i < 100; i++ {
		_, _, err = putGetSmall(t, cli)
		if err != nil {
			t.Fatal("put get failed ->", err)
		}
	}
}

func TestGetBackup(t *testing.T) {
	runtime.GOMAXPROCS(8)
	mockDgInfoer := &MockDgInfoer{
		mutex: sync.RWMutex{},
	}
	mockStater := NewMockGidStater()

	// first new gid
	mstg1 := StartDgidStgs(1, true) // start dgid 1's stg servers
	gid1 := types.NewGid(1)
	mstg1.AddGid(gid1)
	mockStater.AddGidDgid(gid1, 1)
	mockDgInfoer.AddDg(makeDgis([]uint32{1}, cfgapi.SSD))

	cli, err := NewClient(GUID, mockDgInfoer, mockStater, 3, 0, 1, SmallFileLimit, stgapi.TimeoutOption{}, "", nil)
	if err != nil {
		t.Fatal("new pfd client failed:", err)
	}
	stgapi.PostTimeoutTransport = http.DefaultTransport
	stgapi.GetTimeoutTransport = http.DefaultTransport
	stgapi.ProxyGetTimeoutTransport = stgapi.GetTimeoutTransport
	stgapi.ProxyPostTimeoutTransport = stgapi.PostTimeoutTransport

	fh, fsize, err := putGetSmall(t, cli)
	if err != nil {
		t.Fatal("put get failed ->", err)
	}
	size := int(fsize)

	N := 100
	mstg1.GetCountForHost = make(map[string]uint32)
	getConc(t, N, fh, size, cli)
	assert.True(t, mstg1.GetCountForHost["pfdstg1_0_0"] >= 0, "%v", mstg1.GetCountForHost)
	assert.True(t, mstg1.GetCountForHost["pfdstg1_1_0"] >= 0, "%v", mstg1.GetCountForHost)
	assert.True(t, mstg1.GetCountForHost["pfdstg1_2_0"] >= 0, "%v", mstg1.GetCountForHost)
	assert.Equal(t, N, mstg1.GetCountForHost["pfdstg1_0_0"]+mstg1.GetCountForHost["pfdstg1_1_0"]+mstg1.GetCountForHost["pfdstg1_2_0"], "%v", mstg1.GetCountForHost)

	for _, dgi := range mockDgInfoer.Dgis {
		dgi.IsBackup = make([]bool, len(dgi.Hosts))
		for j := range dgi.IsBackup {
			if j == 0 {
				continue
			}
			dgi.IsBackup[j] = true
		}
	}

	// wait for refreshDg
	time.Sleep(2 * time.Second)

	mstg1.FailForHost = make(map[string]bool)
	mstg1.FailForHost["pfdstg1_1_0"] = true
	mstg1.FailForHost["pfdstg1_2_0"] = true
	mstg1.GetCountForHost = make(map[string]uint32)
	getConc(t, N, fh, size, cli)
	assert.True(t, mstg1.GetCountForHost["pfdstg1_0_0"] == uint32(N), "%v", mstg1.GetCountForHost)
	assert.True(t, mstg1.GetCountForHost["pfdstg1_1_0"] == 0, "%v", mstg1.GetCountForHost)
	assert.True(t, mstg1.GetCountForHost["pfdstg1_2_0"] == 0, "%v", mstg1.GetCountForHost)

	mstg1.FailForHost = make(map[string]bool)
	mstg1.FailForHost["pfdstg1_0_0"] = true
	mstg1.GetCountForHost = make(map[string]uint32)
	getConc(t, N, fh, size, cli)
	assert.True(t, mstg1.GetCountForHost["pfdstg1_0_0"] == 0, "%v", mstg1.GetCountForHost)
	assert.True(t, mstg1.GetCountForHost["pfdstg1_1_0"] >= 0, "%v", mstg1.GetCountForHost)
	assert.True(t, mstg1.GetCountForHost["pfdstg1_2_0"] >= 0, "%v", mstg1.GetCountForHost)
	assert.Equal(t, N, mstg1.GetCountForHost["pfdstg1_1_0"]+mstg1.GetCountForHost["pfdstg1_2_0"], "%v", mstg1.GetCountForHost)
}

func TestGetBackupAll(t *testing.T) {
	runtime.GOMAXPROCS(8)
	mockDgInfoer := &MockDgInfoer{
		mutex: sync.RWMutex{},
	}
	mockStater := NewMockGidStater()

	// first new gid
	mstg1 := StartDgidStgs(1, true) // start dgid 1's stg servers
	gid1 := types.NewGid(1)
	mstg1.AddGid(gid1)
	mockStater.AddGidDgid(gid1, 1)
	mockDgInfoer.AddDg(makeDgis([]uint32{1}, cfgapi.SSD))
	// all is backup
	for _, dgi := range mockDgInfoer.Dgis {
		dgi.IsBackup = make([]bool, len(dgi.Hosts))
		for j := range dgi.IsBackup {
			dgi.IsBackup[j] = true
		}
	}

	cli, err := NewClient(GUID, mockDgInfoer, mockStater, 3, 0, 1, SmallFileLimit, stgapi.TimeoutOption{}, "", nil)
	if err != nil {
		t.Fatal("new pfd client failed:", err)
	}
	stgapi.PostTimeoutTransport = http.DefaultTransport
	stgapi.GetTimeoutTransport = http.DefaultTransport
	stgapi.ProxyGetTimeoutTransport = stgapi.GetTimeoutTransport
	stgapi.ProxyPostTimeoutTransport = stgapi.PostTimeoutTransport

	fh, fsize, err := putGetSmall(t, cli)
	if err != nil {
		t.Fatal("put get failed ->", err)
	}
	size := int(fsize)

	_, _, err = cli.Get(nil, fh, 0, int64(size))
	assert.NoError(t, err)

	mockStater.States[types.EncodeGid(gid1)].isECed = true
	_, _, err = cli.Get(nil, fh, 0, int64(size))
	assert.Error(t, err)
	assert.Equal(t, trackerapi.ErrGidECed, err)
}

func TestResNotExist(t *testing.T) {
	gid1 := types.NewGid(1)
	mstg1 := StartDgidStgs(1, false) // start dgid 1's stg servers
	mstg1.AddGid(gid1)
	mockStater := NewMockGidStater()
	mockStater.AddGidDgid(gid1, 1)
	var mockDgInfoer MockDgInfoer
	mockDgInfoer.AddDg(makeDgis2([]uint32{1}, cfgapi.SSD, []string{"nb", "nb", "hz"}, false))
	nbCli, err := NewClient(GUID, &mockDgInfoer, mockStater, 3, 0, 1, SmallFileLimit, stgapi.TimeoutOption{}, "nb", nil)
	assert.NoError(t, err)

	stgapi.PostTimeoutTransport = http.DefaultTransport
	stgapi.GetTimeoutTransport = http.DefaultTransport
	stgapi.ProxyGetTimeoutTransport = stgapi.GetTimeoutTransport
	stgapi.ProxyPostTimeoutTransport = stgapi.PostTimeoutTransport

	data, _ := randbytes(SmallFileSize)
	buf := bytes.NewBuffer(data)
	fsize := int64(buf.Len())

	fh, _, err := nbCli.Put(nil, buf, fsize)
	_, _, err = nbCli.Get(nil, fh, 0, fsize)

	mstg1.RemoveGid(gid1)

	_, _, err = nbCli.Get(nil, fh, 0, fsize)
	assert.Error(t, err)
}

func TestPreferLocal(t *testing.T) {
	xl := xlog.NewWith("TestPreferLocal")

	gid1 := types.NewGid(1)
	mstg1 := StartDgidStgs(1, true) // start dgid 1's stg servers
	mstg1.AddGid(gid1)
	mockStater := NewMockGidStater()
	mockStater.AddGidDgid(gid1, 1)
	var mockDgInfoer MockDgInfoer
	mockDgInfoer.AddDg(makeDgis2([]uint32{1}, cfgapi.SSD, []string{"nb", "nb", "hz"}, true))
	mockDgInfoer.AddDg(makeDgis2([]uint32{2}, cfgapi.SSD, []string{"hz", "hz", "nb"}, true))
	nbCli, err := NewClient(GUID, &mockDgInfoer, mockStater, 3, 0, 1, SmallFileLimit, stgapi.TimeoutOption{}, "nb", nil)
	assert.NoError(t, err)
	hzCli, err := NewClient(GUID, &mockDgInfoer, mockStater, 3, 0, 1, SmallFileLimit, stgapi.TimeoutOption{}, "hz", nil)
	assert.NoError(t, err)

	gid3 := types.NewGid(3)
	mstg3 := StartDgidStgs(3, true) // start dgid 3's stg servers
	mstg3.AddGid(gid3)
	mockStater.AddGidDgid(gid3, 3)
	mockDgInfoer.AddDg(makeDgis([]uint32{3}, cfgapi.DEFAULT))
	gid4 := types.NewGid(4)
	mstg4 := StartDgidStgs(4, true) // start dgid 4's stg servers
	mstg4.AddGid(gid4)
	mockStater.AddGidDgid(gid4, 4)
	mockDgInfoer.AddDg(makeDgis2([]uint32{4}, cfgapi.DEFAULT, []string{"nb", "nb", "hz"}, true))
	oldCli, err := NewClient(GUID, &mockDgInfoer, mockStater, 3, 0, 1, SmallFileLimit, stgapi.TimeoutOption{}, "", nil)
	assert.NoError(t, err)
	newCli, err := NewClient(GUID, &mockDgInfoer, mockStater, 3, 0, 1, SmallFileLimit, stgapi.TimeoutOption{}, "nb", nil)
	assert.NoError(t, err)

	stgapi.PostTimeoutTransport = http.DefaultTransport
	stgapi.GetTimeoutTransport = http.DefaultTransport
	stgapi.ProxyGetTimeoutTransport = stgapi.GetTimeoutTransport
	stgapi.ProxyPostTimeoutTransport = stgapi.PostTimeoutTransport

	var fhs [][]byte
	for i := 0; i < 10; i++ {
		fh, _, err := putGetSmall(t, nbCli)
		assert.NoError(t, err)
		fhs = append(fhs, fh)
		_, _, err = putGetSmall(t, hzCli)
		assert.Error(t, err) // 只能上传到本机房的 master
	}
	// 默认情况下只读本机房的节点
	assert.Equal(t, 5, mstg1.GetCountForHost["pfdstg1_0_0"])
	assert.Equal(t, 5, mstg1.GetCountForHost["pfdstg1_1_0"])
	assert.Equal(t, 0, mstg1.GetCountForHost["pfdstg1_2_0"])
	for i := 0; i < 10; i++ {
		fh, _, err := putGet2(t, SmallFileSize, nbCli, hzCli)
		assert.NoError(t, err)
		fhs = append(fhs, fh)
	}
	assert.Equal(t, 5, mstg1.GetCountForHost["pfdstg1_0_0"])
	assert.Equal(t, 5, mstg1.GetCountForHost["pfdstg1_1_0"])
	assert.Equal(t, 10, mstg1.GetCountForHost["pfdstg1_2_0"])

	// 跨机房也可以删除
	err = nbCli.Delete(xl, fhs[10])
	assert.NoError(t, err)
	err = hzCli.Delete(xl, fhs[11])
	assert.NoError(t, err)

	// 关闭主机房一个节点，还是从主机房另一个节点获取
	StopDgidStgs2(1, 0)
	for _, fh := range fhs[:5] {
		rc, _, err := nbCli.Get(xl, fh, 0, 1)
		assert.NoError(t, err)
		rc.Close()
	}
	assert.Equal(t, 10, mstg1.GetCountForHost["pfdstg1_1_0"])
	// 主机房的节点都关闭了，从另一个机房获取
	StopDgidStgs2(1, 1)
	for _, fh := range fhs[:5] {
		rc, _, err := nbCli.Get(xl, fh, 0, 1)
		assert.NoError(t, err)
		rc.Close()
	}
	assert.Equal(t, 15, mstg1.GetCountForHost["pfdstg1_2_0"])

	// 保证新老 API 兼容
	for i := 0; i < 6; i++ {
		_, _, err := putGetBig(t, oldCli)
		assert.NoError(t, err)
	}
	for j := 0; j <= 2; j++ {
		// 1, 1, 1
		assert.Equal(t, 1, mstg3.GetCountForHost[fmt.Sprintf("pfdstg3_%v_0", j)])
		// 1, 1, 1
		assert.Equal(t, 1, mstg4.GetCountForHost[fmt.Sprintf("pfdstg4_%v_0", j)])
	}
	for i := 0; i < 6; i++ {
		_, _, err := putGetBig(t, newCli)
		assert.NoError(t, err)
	}
	for j := 0; j <= 2; j++ {
		// 2, 2, 2
		assert.Equal(t, 2, mstg3.GetCountForHost[fmt.Sprintf("pfdstg3_%v_0", j)], strconv.FormatInt(int64(j), 10))
		// 3, 2, 1
		assert.Equal(t, 1+2-j, mstg4.GetCountForHost[fmt.Sprintf("pfdstg4_%v_0", j)], strconv.FormatInt(int64(j), 10))
	}
}

func TestPreferNoBackup(t *testing.T) {
	xl := xlog.NewWith("TestPreferNoBackup")

	gid1 := types.NewGid(1)
	mstg1 := StartDgidStgs(1, true) // start dgid 1's stg servers
	mstg1.AddGid(gid1)
	mockStater := NewMockGidStater()
	mockStater.AddGidDgid(gid1, 1)
	var mockDgInfoer MockDgInfoer
	mockDgInfoer.AddDg(makeDgis2([]uint32{1}, cfgapi.SSD, []string{"nb", "nb", "hz"}, true))
	mockDgInfoer.AddDg(makeDgis2([]uint32{2}, cfgapi.SSD, []string{"hz", "hz", "nb"}, true))
	for _, dgi := range mockDgInfoer.Dgis {
		dgi.IsBackup = make([]bool, len(dgi.Hosts))
		dgi.IsBackup[1] = true
	}
	nbCli, err := NewClient(GUID, &mockDgInfoer, mockStater, 3, 0, 1, SmallFileLimit, stgapi.TimeoutOption{}, "nb", nil)
	assert.NoError(t, err)
	hzCli, err := NewClient(GUID, &mockDgInfoer, mockStater, 3, 0, 1, SmallFileLimit, stgapi.TimeoutOption{}, "hz", nil)
	assert.NoError(t, err)

	stgapi.PostTimeoutTransport = http.DefaultTransport
	stgapi.GetTimeoutTransport = http.DefaultTransport
	stgapi.ProxyGetTimeoutTransport = stgapi.GetTimeoutTransport
	stgapi.ProxyPostTimeoutTransport = stgapi.PostTimeoutTransport

	var fhs [][]byte
	for i := 0; i < 10; i++ {
		fh, _, err := putGetSmall(t, nbCli)
		assert.NoError(t, err)
		fhs = append(fhs, fh)
		_, _, err = putGetSmall(t, hzCli)
		assert.Error(t, err) // 只能上传到本机房的 master
	}
	// 默认情况下只读本机房的节点
	assert.Equal(t, 10, mstg1.GetCountForHost["pfdstg1_0_0"])
	assert.Equal(t, 0, mstg1.GetCountForHost["pfdstg1_1_0"])
	assert.Equal(t, 0, mstg1.GetCountForHost["pfdstg1_2_0"])

	for i := 0; i < 10; i++ {
		fh, _, err := putGet2(t, SmallFileSize, nbCli, hzCli)
		assert.NoError(t, err)
		fhs = append(fhs, fh)
	}
	assert.Equal(t, 10, mstg1.GetCountForHost["pfdstg1_0_0"])
	assert.Equal(t, 0, mstg1.GetCountForHost["pfdstg1_1_0"])
	assert.Equal(t, 10, mstg1.GetCountForHost["pfdstg1_2_0"])

	// 关闭主机房一个节点，另一个节点是备份的, 从另一个机房获取
	StopDgidStgs2(1, 0)
	mstg1.GetCountForHost = make(map[string]uint32)
	for _, fh := range fhs[:5] {
		rc, _, err := nbCli.Get(xl, fh, 0, 1)
		assert.NoError(t, err)
		rc.Close()
	}
	assert.Equal(t, 5, mstg1.GetCountForHost["pfdstg1_2_0"])

}

func TestPreferNoBroken(t *testing.T) {
	xl := xlog.NewWith("TestPreferNoBroken")

	gid1 := types.NewGid(1)
	mstg1 := StartDgidStgs(1, true) // start dgid 1's stg servers
	mstg1.AddGid(gid1)
	mockStater := NewMockGidStater()
	mockStater.AddGidDgid(gid1, 1)
	var mockDgInfoer MockDgInfoer
	mockDgInfoer.AddDg(makeDgis2([]uint32{1}, cfgapi.SSD, []string{"nb", "nb", "hz"}, true))
	mockDgInfoer.AddDg(makeDgis2([]uint32{2}, cfgapi.SSD, []string{"hz", "hz", "nb"}, true))
	for _, dgi := range mockDgInfoer.Dgis {
		dgi.IsBackup = make([]bool, len(dgi.Hosts))
		dgi.IsBackup[2] = true
	}
	nbCli, err := NewClient(GUID, &mockDgInfoer, mockStater, 3, 0, 1, SmallFileLimit, stgapi.TimeoutOption{}, "nb", nil)
	assert.NoError(t, err)
	hzCli, err := NewClient(GUID, &mockDgInfoer, mockStater, 3, 0, 1, SmallFileLimit, stgapi.TimeoutOption{}, "hz", nil)
	assert.NoError(t, err)

	stgapi.PostTimeoutTransport = http.DefaultTransport
	stgapi.GetTimeoutTransport = http.DefaultTransport
	stgapi.ProxyGetTimeoutTransport = stgapi.GetTimeoutTransport
	stgapi.ProxyPostTimeoutTransport = stgapi.PostTimeoutTransport

	var fhs [][]byte
	for i := 0; i < 10; i++ {
		fh, _, err := putGetSmall(t, nbCli)
		assert.NoError(t, err)
		fhs = append(fhs, fh)
		_, _, err = putGetSmall(t, hzCli)
		assert.Error(t, err) // 只能上传到本机房的 master
	}
	// 默认情况下只读本机房的节点
	assert.Equal(t, 5, mstg1.GetCountForHost["pfdstg1_0_0"])
	assert.Equal(t, 5, mstg1.GetCountForHost["pfdstg1_1_0"])
	assert.Equal(t, 0, mstg1.GetCountForHost["pfdstg1_2_0"])
	for _, dgi := range mockDgInfoer.Dgis {
		dgi.Repair = make([]bool, len(dgi.Hosts))
		dgi.Repair[0] = true
		dgi.Repair[1] = true
	}

	hzCli2, err := NewClient(GUID, &mockDgInfoer, mockStater, 3, 0, 1, SmallFileLimit, stgapi.TimeoutOption{}, "hz", nil)
	assert.NoError(t, err)
	stgapi.PostTimeoutTransport = http.DefaultTransport
	stgapi.GetTimeoutTransport = http.DefaultTransport
	stgapi.ProxyGetTimeoutTransport = stgapi.GetTimeoutTransport
	stgapi.ProxyPostTimeoutTransport = stgapi.PostTimeoutTransport

	for _, fh := range fhs[:5] {
		rc, _, err := hzCli2.Get(xl, fh, 0, 1)
		assert.NoError(t, err)
		rc.Close()
	}

	assert.Equal(t, 5, mstg1.GetCountForHost["pfdstg1_0_0"])
	assert.Equal(t, 5, mstg1.GetCountForHost["pfdstg1_1_0"])
	assert.Equal(t, 5, mstg1.GetCountForHost["pfdstg1_2_0"])
}

func TestAnotherIdc(t *testing.T) {
	xl := xlog.NewWith("TestAnotherIdc")

	proxyHit := 0
	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		proxyHit++
		w.Write([]byte("haha"))
	}))
	defer proxy.Close()

	gid1 := types.NewGid(1)
	mstg1 := StartDgidStgs(1, true) // start dgid 1's stg servers
	mstg1.AddGid(gid1)
	mockStater := NewMockGidStater()
	mockStater.AddGidDgid(gid1, 1)
	var mockDgInfoer MockDgInfoer
	mockDgInfoer.AddDg(makeDgis2([]uint32{1}, cfgapi.SSD, []string{"nb", "nb", "hz"}, true))
	nbCli, err := NewClient(GUID, &mockDgInfoer, mockStater, 3, 0, 1, SmallFileLimit, stgapi.TimeoutOption{}, "nb", []string{proxy.URL})
	assert.NoError(t, err)

	stgapi.PostTimeoutTransport = http.DefaultTransport
	stgapi.GetTimeoutTransport = http.DefaultTransport

	fh, _, err := putGetSmall(t, nbCli)
	assert.NoError(t, err)

	mstg1.FailForHost["pfdstg1_0_0"] = true
	mstg1.FailForHost["pfdstg1_1_0"] = true
	delete(mockhttp.Route, "pfdstg1_2_0")

	rc, _, err := nbCli.Get(xl, fh, 0, SmallFileSize)
	assert.NoError(t, nil)
	body, _ := ioutil.ReadAll(rc)
	assert.Equal(t, []byte("haha"), body)
	rc.Close()
	assert.Equal(t, 1, proxyHit)
}

func putGetSmall(t *testing.T, cli *Client) (fh []byte, fsize int64, err error) {
	return putGet(t, SmallFileSize, cli)
}

func putGetBig(t *testing.T, cli *Client) (fh []byte, fsize int64, err error) {
	return putGet(t, BigFileSize, cli)
}

func putGetSmallReaderAt(t *testing.T, cli *Client) (fh []byte, fsize int64, err error) {
	return putGetReaderAt(t, SmallFileSize, cli)
}

func putGetBigReaderAt(t *testing.T, cli *Client) (fh []byte, fsize int64, err error) {
	return putGetReaderAt(t, BigFileSize, cli)
}

func putGetReaderAt(t *testing.T, size int, cli *Client) (fh []byte, fsize int64, err error) {

	data, mark := randbytes(size)
	tmpfile := path.Join(os.TempDir(), mark)
	f, err := os.Create(tmpfile)
	if err != nil {
		t.Fatal("create file failed =>", err)
	}
	defer f.Close()
	defer os.Remove(tmpfile)

	f.Write(data)
	f.Seek(0, 0)
	fsize = int64(len(data))
	fh, _, err = cli.Put(nil, f, fsize)
	if err != nil {
		return
	}
	rc, sz, err := cli.Get(nil, fh, 0, fsize)
	if err != nil {
		return
	}
	if sz != fsize || rc == nil {
		err = errors.New("size or content not match")
		return
	}
	return
}

func putGet(t *testing.T, size int, cli *Client) (fh []byte, fsize int64, err error) {
	return putGet2(t, size, cli, cli)
}

func putGet2(t *testing.T, size int, putter, getter *Client) (fh []byte, fsize int64, err error) {
	data, _ := randbytes(size)
	buf := bytes.NewBuffer(data)
	fsize = int64(buf.Len())
	fh, _, err = putter.Put(nil, buf, fsize)
	if err != nil {
		return
	}

	r, sz, err := getter.Get(nil, fh, 0, fsize)
	if err != nil {
		return
	}
	if sz != fsize {
		err = errors.New("size not match")
		return
	}
	data2, _ := ioutil.ReadAll(r)
	if !bytes.Equal(data, data2) {
		return nil, 0, errors.New("data not match")
	}
	return
}

func putatGet2(t *testing.T, fh []byte, size int64, putter, getter *Client) (err error) {
	data, _ := randbytes(int(size))
	buf := bytes.NewBuffer(data)
	err = putter.PutAt(nil, buf, fh)
	if err != nil {
		return
	}

	r, sz, err := getter.Get(nil, fh, 0, size)
	if err != nil {
		return
	}
	if sz != size {
		err = errors.New("size not match")
		return
	}
	data2, _ := ioutil.ReadAll(r)
	if !bytes.Equal(data, data2) {
		return errors.New("data not match")
	}
	return
}

func getConc(t *testing.T, N int, fh []byte, size int, cli *Client) {
	wg := sync.WaitGroup{}
	wg.Add(N)
	for i := 0; i < N; i++ {
		go func() {
			defer wg.Done()
			_, sz, err := cli.Get(nil, fh, 0, int64(size))
			assert.NoError(t, err)
			assert.Equal(t, size, sz)
		}()
	}
	wg.Wait()
}

//---------------------------------------------

const GUID = "1234567890"

func makeStgHosts(dgid uint32, group bool) [][2]string {

	var hostPrefix = "pfdstg"
	dgstr := strconv.Itoa(int(dgid))
	if group {
		return [][2]string{
			[2]string{
				"http://" + hostPrefix + dgstr + "_0_0",
				"http://" + hostPrefix + dgstr + "_0_1",
			},
			[2]string{
				"http://" + hostPrefix + dgstr + "_1_0",
				"http://" + hostPrefix + dgstr + "_1_1",
			},
			[2]string{
				"http://" + hostPrefix + dgstr + "_2_0",
				"http://" + hostPrefix + dgstr + "_2_1",
			},
		}

	} else {
		return [][2]string{
			[2]string{
				"http://" + hostPrefix + dgstr + "_0_0",
				"http://" + hostPrefix + dgstr + "_0_1",
			},
		}

	}

}

func makeDgi(dgid uint32, dt cfgapi.DiskType) *cfgapi.DiskGroupInfo {
	return &cfgapi.DiskGroupInfo{
		Guid:     GUID,
		Dgid:     dgid,
		Hosts:    makeStgHosts(dgid, true),
		DiskType: dt,
	}
}

func makeDgis(dgids []uint32, dt cfgapi.DiskType) (dgis []*cfgapi.DiskGroupInfo) {
	for _, id := range dgids {
		dgis = append(dgis, makeDgi(id, dt))
	}
	return
}

func makeDgis2(dgids []uint32, dt cfgapi.DiskType, idc []string, group bool) (dgis []*cfgapi.DiskGroupInfo) {

	for _, dgid := range dgids {
		dgis = append(dgis, &cfgapi.DiskGroupInfo{
			Guid:     GUID,
			Dgid:     dgid,
			Hosts:    makeStgHosts(dgid, group),
			DiskType: dt,
			Idc:      idc,
		})
	}
	return
}

func StartDgidStgs(dgid uint32, group bool) *MockStg {

	mstg := NewMockStg()
	stgHosts := makeStgHosts(dgid, group)
	for _, hosts := range stgHosts {
		mux := http.NewServeMux()
		router := &webroute.Router{Factory: wsrpc.Factory, Mux: mux}
		stgHost := httpUnPrefix(hosts[0])
		mockhttp.Bind(stgHost, router.Register(mstg))
	}

	return mstg
}

func StopDgidStgs(dgid uint32) {

	stgHosts := makeStgHosts(dgid, true)
	for _, hosts := range stgHosts {
		stgHost := httpUnPrefix(hosts[0])
		delete(mockhttp.Route, stgHost)
	}
}

func StopDgidStgs2(dgid uint32, ith int) {

	stgHosts := makeStgHosts(dgid, true)
	for i, hosts := range stgHosts {
		if i == ith {
			stgHost := httpUnPrefix(hosts[0])
			delete(mockhttp.Route, stgHost)
		}
	}
}

func httpUnPrefix(host string) string {
	if strings.HasPrefix(host, "http://") {
		return host[7:]
	}
	return host
}

//---------------------------------------------

type MockDgInfoer struct {
	Dgis  []*cfgapi.DiskGroupInfo
	mutex sync.RWMutex
}

func (mgi *MockDgInfoer) ListDgs(l rpc.Logger, guid string) ([]*cfgapi.DiskGroupInfo, error) {
	return mgi.Dgis, nil
}

func (mgi *MockDgInfoer) GetDGInfo(l rpc.Logger, guid string, dgid uint32) (dgInfo *cfgapi.DiskGroupInfo, err error) {
	for _, dgi := range mgi.Dgis {
		if dgi.Dgid == dgid {
			return dgi, nil
		}
	}
	return nil, errors.New("Not find dgi")
}

/**
func (mgi *MockDgInfoer) Hosts(l rpc.Logger,
	guid string, dgid uint32) (hosts []dgapi.Host, err error) {

	mgi.mutex.RLock()
	defer mgi.mutex.RUnlock()

	for _, dgi := range mgi.Dgis {
		if dgi.Dgid == dgid {
			for j := range dgi.Hosts {
				h := dgapi.Host{
					Url: dgi.Hosts[j][0],
				}
				if len(dgi.IsBackup) > j {
					h.IsBackup = dgi.IsBackup[j]
				}
				hosts = append(hosts, h)
			}
			return
		}
	}
	return nil, ErrDgidNotFound
}
*/
func (mgi *MockDgInfoer) AddDg(dgis []*cfgapi.DiskGroupInfo) {
	mgi.mutex.Lock()
	defer mgi.mutex.Unlock()

	mgi.Dgis = append(mgi.Dgis, dgis...)
}

func (mgi *MockDgInfoer) RemoveDg(dgid uint32) {
	mgi.mutex.Lock()
	defer mgi.mutex.Unlock()

	for i := range mgi.Dgis {
		if mgi.Dgis[i].Dgid == dgid {
			mgi.Dgis = append(mgi.Dgis[:i], mgi.Dgis[i+1:]...)
			return
		}
	}
}

//---------------------------------------------

type gidState struct {
	dgid   uint32
	isECed bool
	ecing  int32
}

type MockGidStater struct {
	States map[string]*gidState
	mutex  sync.RWMutex
}

func NewMockGidStater() *MockGidStater {
	return &MockGidStater{
		States: make(map[string]*gidState),
		mutex:  sync.RWMutex{},
	}
}

func (mg *MockGidStater) State(l rpc.Logger, egid string) (dgid uint32, isECed bool, err error) {
	return mg.ForceUpdate(l, egid)
}

func (mg *MockGidStater) StateWithGroup(l rpc.Logger, egid string) (group string, dgid uint32, isECed bool, err error) {
	dgid, isECed, err = mg.ForceUpdate(l, egid)
	return
}

func (mg *MockGidStater) StateEntry(l rpc.Logger, egid string) (entry stater.Entry, err error) {
	mg.mutex.RLock()
	defer mg.mutex.RUnlock()

	state, ok := mg.States[egid]
	if !ok {
		err = ErrGidNotFound
		return
	}
	entry.Dgid = state.dgid
	entry.EC = state.isECed
	entry.Ecing = state.ecing
	return
}

func (mg *MockGidStater) ForceUpdate(l rpc.Logger, egid string) (dgid uint32, isECed bool, err error) {
	mg.mutex.RLock()
	defer mg.mutex.RUnlock()

	state, ok := mg.States[egid]
	if !ok {
		err = ErrGidNotFound
		return
	}
	return state.dgid, state.isECed, nil
}

func (mg *MockGidStater) AddGidDgid(gid types.Gid, dgid uint32) {
	mg.mutex.Lock()
	defer mg.mutex.Unlock()

	egid := types.EncodeGid(gid)
	mg.States[egid] = &gidState{dgid, false, 0}
}

func (mg *MockGidStater) SetGidEcing(gid types.Gid, ecing int32) {
	mg.mutex.Lock()
	defer mg.mutex.Unlock()

	egid := types.EncodeGid(gid)
	mg.States[egid].ecing = ecing
}

func (mg *MockGidStater) RemoveGid(gid types.Gid) {
	mg.mutex.Lock()
	defer mg.mutex.Unlock()

	delete(mg.States, types.EncodeGid(gid))
}

func (mg *MockGidStater) EC(gid types.Gid) {
	mg.mutex.Lock()
	defer mg.mutex.Unlock()

	egid := types.EncodeGid(gid)
	mg.States[egid].isECed = true
}

//---------------------------------------------

type mockData map[uint64][]byte // fid -> data

type MockStg struct {
	Stg             map[string]mockData // gid->data
	ActiveGid       string
	CurrentFid      uint64
	PutFailTimes    int
	Put612          int
	GetFailTimes    int
	NotAvailTimes   int
	DeleteFailTimes int
	FailForHost     map[string]bool
	GetCountForHost map[string]uint32
	mutex           sync.Mutex
	GetErr          error
	fulldgid        map[uint32]bool
}

func NewMockStg() *MockStg {
	return &MockStg{
		Stg:             make(map[string]mockData),
		FailForHost:     make(map[string]bool),
		GetCountForHost: make(map[string]uint32),
		mutex:           sync.Mutex{},
	}
}

func (ms *MockStg) AddFullDgid(dgids []uint32) {
	ms.fulldgid = make(map[uint32]bool)
	for _, dgid := range dgids {
		ms.fulldgid[dgid] = true
	}
}

func (ms *MockStg) AddGid(gid types.Gid) {
	egid := types.EncodeGid(gid)

	ms.mutex.Lock()
	defer ms.mutex.Unlock()
	ms.Stg[egid] = make(mockData)
	ms.ActiveGid = egid
}

func (ms *MockStg) RemoveGid(gid types.Gid) {
	egid := types.EncodeGid(gid)

	ms.mutex.Lock()
	defer ms.mutex.Unlock()
	delete(ms.Stg, egid)
	log.Infof("delete gid :%v", gid)
	ms.ActiveGid = ""
}

func (ms *MockStg) allocFid() uint64 {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	fid := ms.CurrentFid
	ms.CurrentFid++
	return fid
}

func (ms *MockStg) MakePutFail(times int) {
	ms.PutFailTimes = times
}

func (ms *MockStg) MakePut612(times int) {
	ms.Put612 = times
}

func (ms *MockStg) MakeGetFailError(err error) {
	ms.GetErr = err
}

func (ms *MockStg) MakeNotAvailFail(times int) {
	ms.NotAvailTimes = times
}

func (ms *MockStg) MakeGetFail(times int) {
	ms.GetFailTimes = times
}

func (ms *MockStg) MakeDeleteFail(times int) {
	ms.DeleteFailTimes = times
}

type putArgs struct {
	Dgid uint32 `flag:"_"`
}

type putRet struct {
	Efh string `json:"fh"`
}

func (ms *MockStg) CmdpPut(env *rpcutil.Env) (ret putRet, err error) {
	return ms.CmdpPut_(nil, env)
}
func (ms *MockStg) CmdpPut_(args *putArgs, env *rpcutil.Env) (ret putRet, err error) {

	if _, ok := ms.fulldgid[args.Dgid]; ok {
		err = ErrNotAvail
		return
	}

	if ms.NotAvailTimes > 0 {
		ms.NotAvailTimes--
		io.Copy(ioutil.Discard, env.Req.Body)
		err = ErrNotAvail
		return
	}

	if ms.PutFailTimes > 0 {
		ms.PutFailTimes--
		io.Copy(ioutil.Discard, env.Req.Body)
		err = errors.New("Fail for fun.")
		return
	}
	if ms.Put612 > 0 {
		ms.Put612--
		err = httputil.NewError(612, "612 for fun.")
		return
	}

	fsize := env.Req.ContentLength

	buf := bytes.NewBuffer(nil)
	n, err := io.Copy(buf, env.Req.Body)
	if err != nil {
		log.Println("io.Copy failed =>", n, err)
		return
	}

	fid := ms.allocFid()
	ms.Stg[ms.ActiveGid][fid] = buf.Bytes()

	gid, _ := types.DecodeGid(ms.ActiveGid)
	fhi := &types.FileHandle{
		Ver:    fhver.FhPfdV2,
		Tag:    types.FileHandle_ChunkBits,
		Gid:    gid,
		Offset: 0,
		Fsize:  fsize,
		Fid:    fid,
	}
	fh := types.EncodeFh(fhi)
	ret.Efh = base64.URLEncoding.EncodeToString(fh)
	return
}

type allocArgs struct {
	Fsize int64  `flag:"_"`
	Hash  []byte `flag:"hash,base64"`
}

func (ms *MockStg) CmdAlloc_(args *allocArgs, env *rpcutil.Env) (ret putRet, err error) {

	if ms.PutFailTimes > 0 {
		ms.PutFailTimes--
		err = errors.New("Fail for fun.")
		return
	}
	if ms.Put612 > 0 {
		ms.Put612--
		err = httputil.NewError(612, "612 for fun.")
		return
	}

	fid := ms.allocFid()
	ms.Stg[ms.ActiveGid][fid] = []byte{}

	gid, _ := types.DecodeGid(ms.ActiveGid)
	fhi := &types.FileHandle{
		Ver:    fhver.FhPfdV2,
		Tag:    types.FileHandle_ChunkBits,
		Gid:    gid,
		Offset: 0,
		Fsize:  args.Fsize,
		Fid:    fid,
	}
	copy(fhi.Hash[:], args.Hash)
	fh := types.EncodeFh(fhi)
	ret.Efh = base64.URLEncoding.EncodeToString(fh)
	return
}

type putatArgs struct {
	Fh []byte `flag:"_,base64"`
}

func (ms *MockStg) CmdpPutat_(args *putatArgs, env *rpcutil.Env) (err error) {

	if ms.PutFailTimes > 0 {
		ms.PutFailTimes--
		err = errors.New("Fail for fun.")
		return
	}
	if ms.Put612 > 0 {
		ms.Put612--
		err = httputil.NewError(612, "612 for fun.")
		return
	}

	buf := bytes.NewBuffer(nil)
	n, err := io.Copy(buf, env.Req.Body)
	if err != nil {
		log.Println("io.Copy failed =>", n, err)
		return
	}

	fhi, err := types.DecodeFh(args.Fh)
	if err != nil {
		err = errors.Info(err, "DecodeFh").Detail(err)
		return
	}

	egid := types.EncodeGid(fhi.Gid)
	stg, ok := ms.Stg[egid]
	if !ok {
		return ErrGidNotFound
	}

	stg[fhi.Fid] = buf.Bytes()
	return
}

type getArgs struct {
	Fh []byte `flag:"_,base64"`
}

func (ms *MockStg) CmdGet_(args *getArgs, env *rpcutil.Env) {

	w := env.W
	if ms.GetFailTimes > 0 {
		ms.GetFailTimes--
		if ms.GetErr != nil {
			if httputil.DetectCode(ms.GetErr) == 613 {
				trytimes613++
			}
			httputil.Error(w, ms.GetErr)
			return
		}
		httputil.Error(w, errors.New("Fail for fun."))
		return
	}
	if ms.FailForHost[env.Req.Host] {
		httputil.Error(w, errors.New("Fail for host"))
		return
	}
	ms.mutex.Lock()
	if _, ok := ms.GetCountForHost[env.Req.Host]; !ok {
		ms.GetCountForHost[env.Req.Host] = 0
	}
	ms.GetCountForHost[env.Req.Host] += 1
	ms.mutex.Unlock()

	fhi, err := types.DecodeFh(args.Fh)
	if err != nil {
		err = errors.Info(err, "DecodeFh").Detail(err)
		httputil.Error(w, err)
		return
	}

	egid := types.EncodeGid(fhi.Gid)
	stg, ok := ms.Stg[egid]
	if !ok {
		httputil.Error(w, ErrGidNotFound)
		return
	}

	data, ok := stg[fhi.Fid]
	if !ok {
		httputil.Error(w, ErrFileNotFound)
		return
	}
	if len(data) == 0 && fhi.Fsize > 0 {
		httputil.Error(w, ErrAllocedEntry)
		return
	}

	h := w.Header()
	h.Set("Content-Length", strconv.Itoa(len(data)))
	r := bytes.NewReader(data)
	n, err := io.Copy(w, r)
	if err != nil {
		log.Fatal("io.Copy: failed =>", n, errors.Detail(err))
	}
	return
}

type deleteArgs struct {
	Fh []byte `flag:"_,base64"`
}

func (ms *MockStg) CmdDelete_(args *deleteArgs, env *rpcutil.Env) {

	w := env.W
	if ms.DeleteFailTimes > 0 {
		ms.DeleteFailTimes--
		httputil.Error(w, errors.New("Delete: Fail for fun."))
		return
	}

	fhi, err := types.DecodeFh(args.Fh)
	if err != nil {
		err = errors.Info(err, "DecodeFh").Detail(err)
		httputil.Error(w, err)
		return
	}

	egid := types.EncodeGid(fhi.Gid)
	stg, ok := ms.Stg[egid]
	if !ok {
		httputil.Error(w, ErrGidNotFound)
		return
	}

	delete(stg, fhi.Fid)
	httputil.ReplyWithCode(w, 200)
	return
}
