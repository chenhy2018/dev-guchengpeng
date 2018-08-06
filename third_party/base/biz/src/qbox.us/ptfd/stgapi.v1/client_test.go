package stgapi

import (
	"bytes"
	"fmt"
	"hash/crc32"
	"io"
	"io/ioutil"
	"math/rand"
	"testing"
	"testing/iotest"

	"qbox.us/ptfd/stgapi.v1/api"

	"github.com/qiniu/errors"
	"github.com/qiniu/rpc.v1"
	"github.com/qiniu/xlog.v1"
	"github.com/stretchr/testify/assert"
)

func randData(size uint32) []byte {

	data := make([]byte, size)
	for i := range data {
		data[i] = byte(rand.Intn(256))
	}
	return data
}

type mockStgInfo struct {
	datas [][]byte
}

type mockStg struct {
	cfg           *mockCfg
	stgs          map[string]*mockStgInfo
	maxTrys       int
	createCount   int
	putCount      int
	getCount      int
	proxyGetCount int
	proxyFails    int
	proxyErr      error
	failedGet     bool
}

func newMockStg(cfg *mockCfg, maxTrys int) *mockStg {

	stgs := make(map[string]*mockStgInfo)
	for _, hosts := range cfg.hosts {
		info := &mockStgInfo{
			datas: make([][]byte, 0),
		}
		for _, host := range hosts {
			stgs[host] = info
		}
	}
	return &mockStg{cfg: cfg, stgs: stgs, maxTrys: maxTrys}
}

func (p *mockStg) Create(l rpc.Logger, host string, max uint32, r io.Reader, length uint32) (ret api.StgRet, err error) {

	buf := bytes.NewBuffer(nil)
	trys := p.createCount % p.maxTrys
	if trys != p.maxTrys-1 {
		r = iotest.TimeoutReader(r)
	}
	p.createCount++
	_, err = io.Copy(buf, r)
	if err != nil {
		err = errors.New("mockStg.Create: ioCopy => " + err.Error())
		return
	}
	if uint32(buf.Len()) != length {
		err = errors.New("mockStg.Create: length check")
		return
	}
	fp := &api.PositionCtx{Max: max, Off: length, Eblock: api.ZeroEblock}
	if length == 0 {
		ret = api.StgRet{
			Ctx: api.EncodePositionCtx(fp),
		}
		return
	}
	data := buf.Bytes()
	info := p.stgs[host]
	round := uint32(len(info.datas))
	info.datas = append(info.datas, data)
	addr := &api.BlockAddr{Dgid: p.cfg.dgids[host], Round: round}
	fp.Eblock = api.EncodeEblock(addr)
	ret = api.StgRet{
		Ctx: api.EncodePositionCtx(fp),
	}
	return
}

func (p *mockStg) Put(l rpc.Logger, host string, ctx string, off uint32, r io.Reader, length uint32) (ret api.StgRet, err error) {

	buf := bytes.NewBuffer(nil)
	trys := p.putCount % p.maxTrys
	if trys != p.maxTrys-1 && off == 0 {
		r = iotest.TimeoutReader(r)
	}
	p.putCount++
	_, err = io.Copy(buf, r)
	if err != nil {
		err = errors.New("mockStg.Put: ioCopy => " + err.Error())
		return
	}
	if uint32(buf.Len()) != length {
		err = errors.New("mockStg.Put: length check")
		return
	}
	data := buf.Bytes()
	fp, _ := api.DecodePositionCtx(ctx)
	if fp.Off != off {
		err = errors.New(fmt.Sprint("Put: fp.Off != off, ", fp.Off, off))
		return
	}
	info := p.stgs[host]
	if fp.Eblock != api.ZeroEblock {
		addr, _ := api.DecodeEblock(fp.Eblock)
		info.datas[addr.Round] = append(info.datas[addr.Round], data...)
	} else {
		round := uint32(len(info.datas))
		addr := &api.BlockAddr{Dgid: p.cfg.dgids[host], Round: round}
		fp.Eblock = api.EncodeEblock(addr)
		info.datas = append(info.datas, data)
	}
	fp.Off += length
	ret = api.StgRet{
		Ctx: api.EncodePositionCtx(fp),
	}
	return
}

func (p *mockStg) Get(l rpc.Logger, host string, eblock string, from, to uint32) (rc io.ReadCloser, err error) {

	info := p.stgs[host]
	a, _ := api.DecodeEblock(eblock)
	data := info.datas[a.Round][from:to]
	if !p.failedGet {
		rc = ioutil.NopCloser(bytes.NewReader(data))
		return
	}
	p.getCount++
	if p.getCount%p.maxTrys == 0 {
		err = errors.New("Get: random failed")
		return
	}
	if p.getCount%p.maxTrys == 1 {
		n := rand.Intn(len(data)) + 1
		rc = ioutil.NopCloser(bytes.NewReader(data[:n]))
		return
	}
	rc = ioutil.NopCloser(bytes.NewReader(data))
	return
}

func (p *mockStg) ProxyGet(l rpc.Logger, host string, eblock string, from, to uint32) (rc io.ReadCloser, err error) {
	p.proxyGetCount++
	if p.proxyFails > 0 {
		p.proxyFails--
		return nil, p.proxyErr
	}
	return p.Get(l, host, eblock, from, to)
}

type mockCfg struct {
	dgids       map[string]uint32
	hosts       map[uint32][]string
	idcs        map[uint32]string
	index       int
	activeHosts map[string][]string
	activeIndex int
}

func newMockCfg(hosts map[uint32][]string) *mockCfg {

	dgids := make(map[string]uint32)
	hosts1 := make(map[uint32][]string)
	idcs := make(map[uint32]string)
	actives := make(map[string][]string)
	for dgid, host := range hosts {
		dgids[host[1]] = dgid
		hosts1[dgid] = host[1:]
		idcs[dgid] = host[0]
		actives[host[0]] = append(actives[host[0]], host[1])
	}
	return &mockCfg{dgids: dgids, hosts: hosts1, idcs: idcs, activeHosts: actives}
}

func (p *mockCfg) HostsIdc(xl *xlog.Logger, dgid uint32) (hosts []string, idx int, idc string, err error) {

	p.index++
	return p.hosts[dgid], p.index % len(p.hosts[dgid]), p.idcs[dgid], nil
}

func (p *mockCfg) Actives(xl *xlog.Logger, idc string) ([]string, int, error) {
	p.activeIndex++
	return p.activeHosts[idc], p.activeIndex % len(p.activeHosts[idc]), nil
}

var g_hosts = map[uint32][]string{
	1: {"nb", "11", "12", "13"},
	2: {"nb", "21", "22", "23"},
	3: {"nb", "31", "32", "33"},
	4: {"nb", "41", "42", "43"},
	5: {"hz", "51", "52", "53"},
	6: {"hz", "61", "62", "63"},
}

func TestClient(t *testing.T) {

	cfg := newMockCfg(g_hosts)
	stg := newMockStg(cfg, 3)

	p := &Client{
		stg:        stg,
		cfg:        cfg,
		maxPutTrys: 3,
		idc:        "nb",
	}
	p2 := &Client{
		stg:        stg,
		cfg:        cfg,
		maxPutTrys: 3,
		idc:        "hz",
	}

	xl := xlog.NewWith("create1")
	max := uint32(api.MaxDataSize)
	size := uint32(3*1024 + 324)
	b := randData(size)
	ra := bytes.NewReader(b)
	data := b
	cret, err := p.Create(xl, max, ra, size)
	assert.NoError(t, err)

	xl = xlog.NewWith("put1")
	off := size
	size = 63*1024 + 4456
	b = randData(size)
	ra = bytes.NewReader(b)
	data = append(data, b...)
	pret, err := p.Put(xl, cret, off, ra, size)
	assert.NoError(t, err)

	xl = xlog.NewWith("put2")
	off += size
	size = 132*1024 - off
	b = randData(size)
	ra = bytes.NewReader(b)
	data = append(data, b...)
	pret, err = p.Put(xl, pret, off, ra, size)
	assert.NoError(t, err)

	xl = xlog.NewWith("put3")
	off += size
	size = 148*1024 - off
	b = randData(size)
	ra = bytes.NewReader(b)
	data = append(data, b...)
	pret, err = p.Put(xl, pret, off, ra, size)
	assert.NoError(t, err)

	xl = xlog.NewWith("put4")
	off += size
	size = 155*1024 - off
	b = randData(size)
	ra = bytes.NewReader(b)
	data = append(data, b...)
	pret, err = p.Put(xl, pret, off, ra, size)
	assert.NoError(t, err)

	xl = xlog.NewWith("put5")
	off += size
	size = 179*1024 - off
	b = randData(size)
	ra = bytes.NewReader(b)
	data = append(data, b...)
	pret, err = p.Put(xl, pret, off, ra, size)
	assert.NoError(t, err)

	xl = xlog.NewWith("put6")
	off += size
	size = 1024 + 79
	b = randData(size)
	ra = bytes.NewReader(b)
	data = append(data, b...)
	pret, err = p.Put(xl, pret, off, ra, size)
	assert.NoError(t, err)

	xl = xlog.NewWith("put7")
	off += size
	size = max - off
	b = randData(size)
	ra = bytes.NewReader(b)
	data = append(data, b...)
	pret, err = p.Put(xl, pret, off, ra, size)
	assert.NoError(t, err)

	fp, _ := api.DecodePositionCtx(pret)
	eblocks := []string{fp.Eblock}

	max64 := int64(max)
	froms := []int64{0, 0, 1, 1, 0}
	tos := []int64{1, max64, max64, max64 - 1, max64 - 1}
	for i := range froms {
		xl = xlog.NewWith(fmt.Sprintf("get-succes%v", i))
		from, to := froms[i], tos[i]
		rc, err := p.Get(xl, eblocks, from, to)
		assert.NoError(t, err, "%v", i)
		buf := bytes.NewBuffer(nil)
		n, err := io.Copy(buf, rc)
		assert.NoError(t, err, "%v", i)
		assert.Equal(t, to-from, n, "%v", i)
		assert.Equal(t, n, buf.Len(), "%v", i)
		rc.Close()
		crc := crc32.ChecksumIEEE(data[from:to])
		assert.Equal(t, crc, crc32.ChecksumIEEE(buf.Bytes()), "%v", i)
	}
	assert.Equal(t, 0, stg.proxyGetCount)

	rc, err := p2.Get(xl, eblocks, 0, 1)
	assert.NoError(t, err)
	buf, _ := ioutil.ReadAll(rc)
	assert.Equal(t, data[:1], buf)
	rc.Close()
	assert.Equal(t, 1, stg.proxyGetCount)

	stg.proxyFails = 2
	stg.proxyGetCount = 0
	stg.proxyErr = &rpc.ErrorInfo{Code: 503, Err: "test"}
	rc, err = p2.Get(xl, eblocks, 0, 1)
	assert.NoError(t, err)
	buf, err = ioutil.ReadAll(rc)
	assert.Error(t, err)
	rc.Close()
	assert.Equal(t, 1, stg.proxyGetCount)
	assert.Equal(t, 1, stg.proxyFails)

	stg.proxyFails = 1
	stg.proxyGetCount = 0
	stg.proxyErr = &rpc.ErrorInfo{Code: 400, Err: "test"}
	rc, err = p2.Get(xl, eblocks, 0, 1)
	assert.NoError(t, err)
	buf, _ = ioutil.ReadAll(rc)
	assert.Equal(t, data[:1], buf)
	rc.Close()
	assert.Equal(t, 2, stg.proxyGetCount)
	assert.Equal(t, 0, stg.proxyFails)

	stg.proxyFails = 1
	stg.proxyGetCount = 0
	stg.proxyErr = &rpc.ErrorInfo{Code: 400, Err: "connecting to proxy"}
	rc, err = p2.Get(xl, eblocks, 0, 1)
	assert.NoError(t, err)
	buf, _ = ioutil.ReadAll(rc)
	assert.Equal(t, data[:1], buf)
	rc.Close()
	assert.Equal(t, 2, stg.proxyGetCount)
	assert.Equal(t, 0, stg.proxyFails)

	stg.proxyFails = 2
	stg.proxyGetCount = 0
	stg.proxyErr = errors.New("connecting to proxy")
	rc, err = p2.Get(xl, eblocks, 0, 1)
	assert.NoError(t, err)
	buf, err = ioutil.ReadAll(rc)
	assert.Error(t, err)
	rc.Close()
	assert.Equal(t, 1, stg.proxyGetCount)
	assert.Equal(t, 1, stg.proxyFails)

	stg.failedGet = true
	for i := range froms {
		xl = xlog.NewWith(fmt.Sprintf("get-failed%v", i))
		from, to := froms[i], tos[i]
		rc, err := p.Get(xl, eblocks, from, to)
		assert.NoError(t, err, "%v", i)
		buf := bytes.NewBuffer(nil)
		n, err := io.Copy(buf, rc)
		assert.NoError(t, err, "%v", i)
		assert.Equal(t, to-from, n, "%v", i)
		assert.Equal(t, n, buf.Len(), "%v", i)
		rc.Close()
		crc := crc32.ChecksumIEEE(data[from:to])
		assert.Equal(t, crc, crc32.ChecksumIEEE(buf.Bytes()), "%v", i)
	}

	for i := 0; i < len(g_hosts); i++ {
		max := uint32(api.MaxDataSize)
		size := uint32(3*1024 + 324)
		b := randData(size)
		ra := bytes.NewReader(b)
		_, err := p.Create(xl, max, ra, size)
		assert.NoError(t, err)
	}
	for _, host := range g_hosts {
		if host[0] != "nb" {
			for _, h := range host[1:] {
				assert.Equal(t, 0, len(stg.stgs[h].datas))
			}
		}
	}

}
