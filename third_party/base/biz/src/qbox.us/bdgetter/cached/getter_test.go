package cached

import (
	"bytes"
	"crypto/rand"
	"crypto/sha1"
	"encoding/binary"
	"errors"
	"io"
	"io/ioutil"
	"sync"
	"testing"
	"time"

	"github.com/qiniu/log.v1"
	"github.com/qiniu/rpc.v1"
	"github.com/qiniu/xlog.v1"
	"github.com/stretchr/testify/assert"

	"qbox.us/api/dc"
	qfh "qbox.us/fh"
	"qbox.us/fh/proto"
)

var errHitFails = errors.New("hit fails")
var errNotFound = errors.New("not found")

// -----------------------------------------------------------------------------

func init() {
	log.SetOutputLevel(0)
}

type failedReader struct {
	data []byte
}

func (p *failedReader) Read(buf []byte) (n int, err error) {

	if len(p.data) == 0 {
		err = errHitFails
		return
	}
	n = copy(buf, p.data)
	p.data = p.data[n:]
	return
}

type mockStg struct {
	datas     map[string][]byte
	fails     []int64
	getFailed bool
}

func (p *mockStg) Cacheable(l rpc.Logger, fh []byte) bool {

	return true
}

func (p *mockStg) Get(xl rpc.Logger, fh []byte, from, to int64) (rc io.ReadCloser, n int64, err error) {

	data, ok := p.datas[string(fh)]
	if !ok {
		err = errNotFound
		return
	}
	if p.getFailed {
		err = errHitFails
		return
	}
	if to > int64(len(data)) {
		to = int64(len(data))
	}
	n = to - from
	r := io.Reader(bytes.NewReader(data[from:to]))
	if len(p.fails) > 0 {
		r = &failedReader{data[from : from+p.fails[0]]}
		p.fails = p.fails[1:]
	}
	rc = ioutil.NopCloser(r)
	return
}

// -----------------------------------------------------------------------------

type mockCachedGetter struct {
	ssdCached    *mockCached
	sataCached   *mockCached
	ssdCacheSize int64
	blocks       bool
}

func (p *mockCachedGetter) Cached(l rpc.Logger, fh []byte) (key []byte, cached Cached, blocks bool) {

	xl := xlog.NewWith(l)
	if fsize := qfh.Fsize(fh); fsize <= p.ssdCacheSize {
		xl.Println("ssdcached", fsize)
		// 4MB以下的缓存，直接拿hash作为key
		return qfh.Etag(fh)[1:], p.ssdCached, p.blocks
	}
	key1 := sha1.Sum(qfh.Etag(fh))
	return key1[:], p.sataCached, p.blocks
}

// -----------------------------------------------------------------------------

type mockCached struct {
	mu    sync.RWMutex
	datas map[string][]byte

	fails      []int64
	putFailed  bool
	getFailed  bool
	notHintKey bool
}

func (p *mockCached) count() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.datas)
}

func (p *mockCached) Set(xl *xlog.Logger, key []byte, r io.Reader, n int64) error {

	if p.putFailed {
		return errHitFails
	}
	// 放在io.ReadFull之前，为了让count方法更准确。
	p.mu.Lock()
	defer p.mu.Unlock()

	data := make([]byte, n)
	n1, err := io.ReadFull(r, data)
	if err != nil {
		xl.Error("mockCached.Set: ReadFull =>", err, n1, n)
		return err
	}
	p.datas[string(key)] = data
	return nil
}

func (p *mockCached) RangeGetHint(xl *xlog.Logger, key []byte, from, to int64) (rc io.ReadCloser, n int64, hint bool, err error) {

	p.mu.RLock()
	data, ok := p.datas[string(key)]
	p.mu.RUnlock()
	if !ok {
		hint = !p.notHintKey
		err = dc.ENoSuchEntry
		return
	}
	if p.getFailed {
		err = errHitFails
		return
	}
	if to > int64(len(data)) {
		to = int64(len(data))
	}
	n = to - from
	r := io.Reader(bytes.NewReader(data[from:to]))
	if len(p.fails) > 0 {
		r = &failedReader{data[from : from+p.fails[0]]}
		p.fails = p.fails[1:]
	}
	rc = ioutil.NopCloser(r)
	return
}

// -----------------------------------------------------------------------------

type testPfdCtx struct {
	fh   []byte
	data []byte
	from int64
	to   int64
}

func testCachedGetter(t *testing.T, p proto.CommonGetter, stg *mockStg, dc *mockCached, ctx *testPfdCtx) {

	xl := xlog.NewDummy()
	fh := ctx.fh
	data := ctx.data
	from := ctx.from
	to := ctx.to

	dc.datas = make(map[string][]byte)
	dc.fails = nil
	dc.putFailed = false
	dc.getFailed = false
	stg.datas = make(map[string][]byte)
	stg.fails = nil
	stg.getFailed = false

	w := bytes.NewBuffer(nil)
	n, err := p.Get(xl, fh, w, from, to)
	assert.Equal(t, err, errNotFound)
	assert.Equal(t, n, 0)
	assert.Equal(t, w.Len(), 0)

	dc.putFailed = true
	stg.getFailed = true
	stg.datas[string(fh)] = data
	w = bytes.NewBuffer(nil)
	n, err = p.Get(xl, fh, w, from, to)
	assert.Equal(t, err, errHitFails)
	assert.Equal(t, n, 0)
	assert.Equal(t, w.Len(), 0)

	stg.getFailed = false
	dc.putFailed = true
	stg.fails = []int64{2, 2, 3}
	w = bytes.NewBuffer(nil)
	n, err = p.Get(xl, fh, w, from, to)
	assert.Equal(t, err, errHitFails)
	assert.Equal(t, n, 7-from)
	assert.Equal(t, w.Bytes(), data[from:7])

	dc.putFailed = false
	stg.fails = []int64{2, 2, 3}
	w = bytes.NewBuffer(nil)
	n, err = p.Get(xl, fh, w, from, to)
	assert.Equal(t, err, errHitFails)
	assert.Equal(t, n, 7-from)
	assert.Equal(t, w.Bytes(), data[from:7])

	xl1 := xlog.NewWith("dc-stg")
	stg.fails = nil
	w = bytes.NewBuffer(nil)
	n, err = p.Get(xl1, fh, w, from, to)
	assert.NoError(t, err)
	assert.Equal(t, n, to-from)
	assert.Equal(t, w.Bytes(), data[from:to])
	time.Sleep(10 * 1e6)

	xl1 = xlog.NewWith("dc-2-stg-2-3")
	dc.fails = []int64{2}
	stg.fails = []int64{2, 3}
	w = bytes.NewBuffer(nil)
	n, err = p.Get(xl1, fh, w, from, to)
	assert.Equal(t, err, errHitFails)
	assert.Equal(t, n, 7)
	assert.Equal(t, w.Bytes(), data[from:from+7])

	dc.fails = []int64{2}
	stg.fails = []int64{3}
	w = bytes.NewBuffer(nil)
	n, err = p.Get(xl, fh, w, from, to)
	assert.NoError(t, err)
	assert.Equal(t, n, to-from)
	assert.Equal(t, w.Bytes(), data[from:to])

	dc.fails = []int64{2}
	stg.fails = []int64{}
	w = bytes.NewBuffer(nil)
	n, err = p.Get(xl, fh, w, from, to)
	assert.NoError(t, err)
	assert.Equal(t, n, to-from)
	assert.Equal(t, w.Bytes(), data[from:to])

	stg.fails = []int64{2, 2, 3}
	w = bytes.NewBuffer(nil)
	n, err = p.Get(xl, fh, w, from, to)
	assert.NoError(t, err)
	assert.Equal(t, n, to-from)
	assert.Equal(t, w.Bytes(), data[from:to])

	dc.fails = []int64{}
	dc.getFailed = true
	stg.fails = []int64{2, 3}
	w = bytes.NewBuffer(nil)
	n, err = p.Get(xl, fh, w, from, to)
	assert.Equal(t, err, errHitFails)
	assert.Equal(t, n, 5)
	assert.Equal(t, w.Bytes(), data[from:from+5])

	dc.fails = []int64{}
	dc.getFailed = true
	stg.fails = []int64{2}
	w = bytes.NewBuffer(nil)
	n, err = p.Get(xl, fh, w, from, to)
	assert.NoError(t, err)
	assert.Equal(t, n, to-from)
	assert.Equal(t, w.Bytes(), data[from:to])

	dc.getFailed = true
	stg.fails = nil
	w = bytes.NewBuffer(nil)
	n, err = p.Get(xl, fh, w, from, to)
	assert.NoError(t, err)
	assert.Equal(t, n, to-from)
	assert.Equal(t, w.Bytes(), data[from:to])

	dc.getFailed = false
	w = bytes.NewBuffer(nil)
	n, err = p.Get(xl, fh, w, from, to)
	assert.NoError(t, err)
	assert.Equal(t, n, to-from)
	assert.Equal(t, w.Bytes(), data[from:to])
}

func TestGetter(t *testing.T) {

	data := []byte("hello world")
	fh := make([]byte, 60)
	fh[0] = 5
	fh[1] = 0x96
	binary.LittleEndian.PutUint64(fh[24:], uint64(len(data)))

	froms := []int{0, 1, 1, 0}
	tos := []int{len(data), len(data), len(data) - 1, len(data) - 1}
	for i := range froms {
		cg := &mockCachedGetter{
			ssdCached: &mockCached{
				datas: make(map[string][]byte),
			},
			ssdCacheSize: 100,
		}
		g := &mockStg{
			datas: make(map[string][]byte),
		}
		ctx := &testPfdCtx{
			fh:   fh,
			data: data,
			from: int64(froms[i]),
			to:   int64(tos[i]),
		}
		p := New(g, cg, 1)
		testCachedGetter(t, p, g, cg.ssdCached, ctx)
	}

	for i := range froms {
		cg := &mockCachedGetter{
			sataCached: &mockCached{
				datas: make(map[string][]byte),
			},
			ssdCacheSize: int64(len(data) - 1),
		}
		g := &mockStg{
			datas: make(map[string][]byte),
		}
		ctx := &testPfdCtx{
			fh:   fh,
			data: data,
			from: int64(froms[i]),
			to:   int64(tos[i]),
		}
		p := New(g, cg, 1)
		testCachedGetter(t, p, g, cg.sataCached, ctx)
	}
}

func TestBlocksGetter(t *testing.T) {

	data := []byte("hello world")
	testBlocksGetter(t, data)
}

func TestBlocksGetterEx(t *testing.T) {
	sizes := []int64{
		// 2,
		chunkSize - 1,
		chunkSize,
		chunkSize + 1,
		2*chunkSize - 1,
		2 * chunkSize,
		2*chunkSize + 1,
	}
	for _, size := range sizes {
		data := make([]byte, size)
		rand.Read(data)
		testBlocksGetter(t, data)
	}
}

func testBlocksGetter(t *testing.T, data []byte) {
	fh := make([]byte, 60)
	fh[0] = 5
	fh[1] = 0x96
	binary.LittleEndian.PutUint64(fh[24:], uint64(len(data)))

	froms := []int{0, 1, 1, 0}
	tos := []int{len(data), len(data), len(data) - 1, len(data) - 1}

	for i := range froms {
		cg := &mockCachedGetter{
			sataCached: &mockCached{
				datas: make(map[string][]byte),
			},
			ssdCacheSize: 0,
			blocks:       true,
		}
		g := &mockStg{
			datas: make(map[string][]byte),
		}
		ctx := &testPfdCtx{
			fh:   fh,
			data: data,
			from: int64(froms[i]),
			to:   int64(tos[i]),
		}
		p := New(g, cg, 1)
		testCachedGetter(t, p, g, cg.sataCached, ctx)
	}
}

func TestHint(t *testing.T) {

	xl := xlog.NewWith("TestHint")

	data := []byte("hello world")
	fh := make([]byte, 60)
	fh[0] = 5
	fh[1] = 0x96
	binary.LittleEndian.PutUint64(fh[24:], uint64(len(data)))

	cg := &mockCachedGetter{
		sataCached: &mockCached{
			datas: make(map[string][]byte),
		},
		ssdCacheSize: int64(len(data) - 1),
	}
	g := &mockStg{
		datas: make(map[string][]byte),
	}
	p := New(g, cg, 1)

	var buf bytes.Buffer

	// 1. get no data
	_, err := p.Get(xl, fh, &buf, 0, 1)
	assert.Error(t, err)

	// 2. set not hint
	cg.sataCached.notHintKey = true
	g.datas[string(fh)] = data

	buf.Reset()
	n, err := p.Get(xl, fh, &buf, 0, 2)
	assert.NoError(t, err)
	assert.Equal(t, 2, n)
	assert.Equal(t, data[0:2], buf.Bytes())

	// should not set to cached
	assert.Equal(t, 0, cg.sataCached.count())

	// set hint on second get
	cg.sataCached.notHintKey = false

	buf.Reset()
	n, err = p.Get(xl, fh, &buf, 0, 2)
	assert.NoError(t, err)
	assert.Equal(t, 2, n)
	assert.Equal(t, data[0:2], buf.Bytes())

	// should set to cached
	assert.Equal(t, 1, cg.sataCached.count())
}

func TestMerge(t *testing.T) {
	{

		x1 := xlog.NewDummy()
		x1.Xprof2("mc.g;IO", 0, nil)
		x2 := x1.Spawn()
		for i := 0; i < 50; i++ {
			x2.Xlog("DC:1/404")
			x2.Xprof2("mc.g;EBDDN", 1*time.Millisecond, nil)
		}
		merge(x1, x2)
		for _, a := range x1.Xget() {
			f := false
			for _, b := range []string{"mc.g;IO", "DC*50:50/404", "mc.g*50", "EBDDN*50:50"} {
				if a == b {
					f = true
				}
			}
			assert.True(t, f)
		}
	}

	{

		x1 := xlog.NewDummy()
		x1.Xprof2("mc.g;IO", 0, nil)
		x2 := x1.Spawn()
		for i := 0; i < 10; i++ {
			x2.Xlog("DC:1/404")
			x2.Xprof2("mc.g;EBDDN", 1*time.Millisecond, nil)
		}
		merge(x1, x2)
		assert.Equal(t, x1.Xget(), []string{"mc.g;IO", "DC:1/404", "mc.g;EBDDN:1", "DC:1/404", "mc.g;EBDDN:1", "DC:1/404", "mc.g;EBDDN:1", "DC:1/404", "mc.g;EBDDN:1", "DC:1/404", "mc.g;EBDDN:1", "DC:1/404", "mc.g;EBDDN:1", "DC:1/404", "mc.g;EBDDN:1", "DC:1/404", "mc.g;EBDDN:1", "DC:1/404", "mc.g;EBDDN:1", "DC:1/404", "mc.g;EBDDN:1"})
	}

	{

		x1 := xlog.NewDummy()
		x1.Xprof2("mc.g;IO", 0, nil)
		x2 := x1.Spawn()
		for i := 0; i < 50; i++ {
			x2.Xprof2("mc.g:2/500;EBDMASTER;mc.s/500;m.Get:2;EBDDN:5", 1*time.Millisecond, nil)
		}
		merge(x1, x2)
		for _, a := range x1.Xget() {
			f := false
			for _, b := range []string{"mc.g;IO", "mc.g*50:100/500", "EBDMASTER*50", "mc.s*50/500", "m.Get*50:100", "EBDDN:5*50:50"} {
				if a == b {
					f = true
				}
			}
			assert.True(t, f)
		}
	}
}
