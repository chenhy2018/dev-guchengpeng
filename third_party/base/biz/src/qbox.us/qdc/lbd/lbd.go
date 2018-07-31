package lbd

import (
	"encoding/binary"
	"errors"
	"io"

	"github.com/qiniu/log.v1"
	"qbox.us/store"
	"qbox.us/store/cc"
	"github.com/qiniu/xlog.v1"

	. "qbox.us/api/bd/errors"
)

const MaxCacheSize = 1024*1024*4 - 4

type StgInterface interface {
	store.Getter
	store.Putter
}

type LocalBd struct {
	cache     *cc.SimpleIntCache
	Pool      *cc.ChunkPool
	Storage   StgInterface
	ChunkBits uint
	RW        cc.ReadWriterAt
	Info      *CacheInfo
	OtherLbds map[int]store.MultiBdInterface
	MyId      int
}

func NewLocalBd(rw cc.ReadWriterAt, pool *cc.ChunkPool, chunkBits uint, limitCacheItems int, stg StgInterface, info *CacheInfo, otherLbds map[int]store.MultiBdInterface, myId int) *LocalBd {
	cache := cc.NewSimpleIntCache(limitCacheItems)
	log.Warn("loading cache index...")
	err := info.Load(cache, pool)
	log.Warn("load done.err :", err)
	if err != nil {
		log.Warn("load cache index:", err)
		return nil
	}
	return &LocalBd{cache, pool, stg, chunkBits, rw, info, otherLbds, myId}
}

func (p *LocalBd) Get(xl *xlog.Logger, key []byte, w io.Writer, from, to int, bds [4]uint16) error {

	chunkBits := p.ChunkBits
	key1 := string(key)

	if idx := p.cache.Get(key1); idx >= 0 {
		pos := int64(idx) << chunkBits
		r1 := cc.NewReader(p.RW, pos+int64(from))
		_, err := io.CopyN(w, r1, int64(to-from))
		xl.Info("LocalBd.Get: local cache hit, err :", err)
		if err == nil {
			return nil
		}
		return p.Storage.Get(xl, key, w, from, to, bds)
	}

	idx := p.Pool.Alloc()
	xl.Debug("LocalBd.Get: Pool.Alloc:", idx)

	var err error
	finished := false
	defer func() {
		if !finished {
			xl.Warn("LocalBd.Get: not finish, err :", err, ", idx :", idx)
			p.Pool.Free(idx)
		}
	}()

	// transaction(update index) -- begin
	err = p.Info.Clear(idx)
	if err != nil {
		xl.Warn("LocalBd.Get: clear index err:", err)
		return err
	}
	xl.Debug("LocalBd.Get: clear idx:", idx)
	pos := int64(idx) << chunkBits
	w1 := &cc.Writer{p.RW, pos}
	err = p.Storage.Get(xl, key, w1, 0, 1<<chunkBits, bds)
	if err != nil {
		if err != EKeyNotFound {
			xl.Warn("LocalBd.Get failed:", err)
			return err
		}
		idc := int(bds[3])
		xl.Warn("LocalBd.Get: get from backend err, try other idc... :", err, idc)
		other, ok := p.OtherLbds[idc]
		if !ok || idc == p.MyId {
			xl.Warn("LocalBd.Get: can not find other idc(or hit myself) :", idc, p.MyId)
			return err
		}
		w1 := &cc.Writer{p.RW, pos}
		err = other.Get(xl, key, w1, 0, 1<<chunkBits, bds)
		if err != nil {
			xl.Warn("LocalBd.Get: try other idc err :", err)
			return err
		}
	}

	r1 := cc.NewReader(p.RW, pos+int64(from))
	_, err = io.CopyN(w, r1, int64(to-from))

	oldIdx := p.cache.Set(key1, idx)
	finished = true

	xl.Debug("LocalBd.Get: oldIdx:", oldIdx)
	err = p.Info.Set(key, idx) // atomic
	xl.Debug("LocalBd.Get: update cache index", err)
	// transaction(update index) -- end

	if oldIdx >= 0 {
		p.Pool.Free(oldIdx)
	}
	return err
}

func (p *LocalBd) Put(xl *xlog.Logger, r io.Reader, n int, key []byte, doCache bool, bds [3]uint16) (err error) {
	if doCache {
		idx := p.Pool.Alloc()
		xl.Info("LocalBd.Put: do cache, key :", key, n, idx)
		finished := false
		defer func() {
			if !finished {
				xl.Warn("LocalBd.Put: not finish, err :", err, ", idx :", idx)
				p.Pool.Free(idx)
			}
		}()

		err = p.Info.Clear(idx)
		if err != nil {
			xl.Warn("LocalBd.Put: clear index err :", err)
			return err
		}
		xl.Debug("LocalBd.Put: clear idx:", idx)
		pos := int64(idx) << p.ChunkBits
		w1 := &cc.Writer{p.RW, pos}
		_, err = io.CopyN(w1, r, int64(n))
		if err != nil {
			xl.Warn("LocalBd.Put: copy to local err :", err)
			return err
		}

		oldIdx := p.cache.Set(string(key), idx)
		finished = true
		xl.Debug("LocalBd.Put: oldIdx:", oldIdx)
		if oldIdx >= 0 {
			p.Pool.Free(oldIdx)
		}
		err = p.Info.Set(key, idx) // atomic
		if err != nil {
			xl.Warn("LocalBd.Put: update cache index err :", err)
			return err
		}
		r = cc.NewReader(p.RW, pos)
	}

	xl.Debug("LocalBd.Put: n =", n)
	return p.Storage.Put(xl, key, r, n, bds)
}

func (p *LocalBd) GetLocal(xl *xlog.Logger, key []byte) (r io.Reader, size int, err error) {

	idx := p.cache.Get(string(key))
	if idx < 0 {
		err = EKeyNotFound
		return
	}

	r1 := cc.NewReader(p.RW, int64(idx)<<p.ChunkBits)
	header := make([]byte, 4)
	_, err = io.ReadFull(r1, header)
	if err != nil {
		xl.Warn("read header err :", err)
		return
	}

	size = int(binary.LittleEndian.Uint32(header))
	if size > MaxCacheSize {
		xl.Warn("too large size")
		err = errors.New("to large size")
		return
	}

	r = &io.LimitedReader{r1, int64(size)}

	xl.Debug("LocalBd.GetLocal done,size :", size)
	return
}

func (p *LocalBd) PutLocal(xl *xlog.Logger, r io.Reader, n int, key []byte) (err error) {

	if n > MaxCacheSize {
		xl.Warn("too large size")
		return errors.New("to large size")
	}
	idx := p.Pool.Alloc()
	xl.Info("put local, key :", key, n, idx)

	finished := false
	defer func() {
		if !finished {
			xl.Warn("lbd.PutLocal : not finish, err :", err, ", idx :", idx)
			p.Pool.Free(idx)
		}
	}()

	err = p.Info.Clear(idx)
	if err != nil {
		xl.Warn("clear index err:", err)
		return err
	}
	xl.Debug("clear idx:", idx)
	pos := int64(idx) << p.ChunkBits
	w1 := &cc.Writer{p.RW, pos}

	header := make([]byte, 4)
	binary.LittleEndian.PutUint32(header, uint32(n))
	_, err = w1.Write(header)
	if err != nil {
		xl.Warn("write header err :", err)
		return err
	}

	_, err = io.CopyN(w1, r, int64(n))
	if err != nil {
		return err
	}

	oldIdx := p.cache.Set(string(key), idx)
	finished = true
	xl.Debug("oldIdx:", oldIdx)
	if oldIdx >= 0 {
		p.Pool.Free(oldIdx)
	}
	err = p.Info.Set(key, idx) // atomic
	if err != nil {
		xl.Warn("put local, update cache index err :", err)
	}
	return err
}
