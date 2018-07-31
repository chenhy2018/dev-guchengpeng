package lbd

import (
	"crypto/sha1"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"io"
	"io/ioutil"
	"syscall"
	"time"

	"github.com/qiniu/log.v1"
	"qbox.us/store"
	"qbox.us/store/cc"
	"qbox.us/timeio"
	"github.com/qiniu/xlog.v1"

	qio "github.com/qiniu/io"
	qioutil "github.com/qiniu/io/ioutil"
	. "qbox.us/api/bd/errors"
)

//const MaxCacheSize = 1024*1024*4 - 4

//var EKeyNotFound = errors.New("key not found")

//type StgInterface interface {
//	store.Getter
//	store.Putter
//	store.Querier
//}

type SimpleLocalBd struct {
	Pool      cc.SimplePool
	cache     *cc.SimpleKeyCacheEx
	Storage   StgInterface
	OtherLbds map[int]store.MultiBdInterface
	ChunkBits uint
	MyId      int
}

func NewSimpleLocalBd(root, filename string, duration int64, maxBuf int,
	chunkBits uint, limitCount int, limitSpace int64,
	stg StgInterface, otherLbds map[int]store.MultiBdInterface, myId int) (*SimpleLocalBd, error) {

	cache, err := cc.NewSimpleKeyCacheEx(filename, duration, maxBuf, func(space int64, count int) bool {
		return limitCount <= count || limitSpace <= space
	})
	if err != nil {
		return nil, err
	}

	log.Warn("loading cache index...")
	lbd := &SimpleLocalBd{cc.NewSimplePool(root), cache, stg, otherLbds, chunkBits, myId}
	//	err := lbd.load()
	//	log.Warn("load done.err :", err)
	//	if err != nil {
	//		log.Warn("load cache index:", err)
	//		return nil
	//	}
	return lbd, nil
}

/*
func (p *SimpleLocalBd) load() error {
	keys, err := p.Pool.Keys()
	if err != nil {
		return err
	}
	for _, key := range keys {
		p.cache.Set(key.Name, key.Size)
	}
	return nil
}
*/

func (p *SimpleLocalBd) put1(xl *xlog.Logger, key string, r io.Reader, length int64) (n int64, err error) {

	w, err := p.Pool.GetWriter(xl, key)
	if err != nil {
		xl.Warn("SimpleLocalBd.put1: get writer from pool failed", key, err)
		return
	}
	finish := false
	defer func() {
		if finish {
			err = w.Close()
		} else {
			w.SetErr(err)
			w.Close()
		}
	}()

	hash := sha1.New()

	tw := timeio.NewWriter(w)
	w1 := io.MultiWriter(tw, hash)
	now := time.Now()
	if length >= 0 {
		n, err = io.CopyN(w1, r, length)
	} else {
		n, err = io.Copy(w1, r)
	}
	xl.Debugf("put1: io.Copy time: %v(100ns), poolW: %v(100ns)\n", int64(time.Since(now)/100), tw.Time()/100)

	h := hash.Sum(nil)
	if key != hex.EncodeToString(h) {
		err = errors.New("check failed")
		w.SetErr(err)
		return
	}

	finish = true
	return
}

type sizedWriter struct {
	size int64
	w    io.Writer
}

func (r *sizedWriter) Write(p []byte) (n int, err error) {
	n, err = r.w.Write(p)
	r.size += int64(n)
	return
}

func (p *SimpleLocalBd) put2(xl *xlog.Logger, key string, writer func(io.Writer) error) (n int64, err error) {

	now := time.Now()
	w, err := p.Pool.GetWriter(xl, key)
	xl.Debugf("SimpleLocalBd.put2: pool getwriter %v(100ns)", int64(time.Since(now)/100))
	if err != nil {
		log.Warn("get writer from pool failed", key, err)
		return
	}
	w1 := &sizedWriter{0, w}
	err = writer(w1)
	if err == nil {
		err = w.Close()
	} else {
		w.SetErr(err)
		w.Close()
	}
	if err != nil {
		log.Warn("write local failed:", key, err)
		return
	}

	//	p.setCache(key)
	return w1.size, nil
}

func (p *SimpleLocalBd) put2Check(xl *xlog.Logger, key string, length int64, writer func(*xlog.Logger, io.Writer) error) (
	n int64, err error) {

	now := time.Now()
	w, err := p.Pool.GetWriter(xl, key)
	xl.Debugf("SimpleLocalBd.put2Check: pool getwriter %v(100ns)", int64(time.Since(now)/100))
	if err != nil {
		xl.Warn("SimpleLocalBd.put2Check: get writer from pool failed", key, err)
		return
	}
	finish := false
	defer func() {
		if finish {
			err = w.Close()
		} else {
			w.SetErr(err)
			w.Close()
		}
	}()

	w2 := &sizedWriter{0, w}
	hash := sha1.New()

	w1 := io.MultiWriter(hash, w2)
	err = writer(xl, w1)
	if err != nil {
		return
	}

	h := hash.Sum(nil)
	if key != hex.EncodeToString(h) {
		err = errors.New("check failed")
		w.SetErr(err)
		return
	}

	n = w2.size
	finish = true
	return
}

func (p *SimpleLocalBd) setCache(xl *xlog.Logger, key string, size int64) {

	var swapped bool
	var cacheSet, poolDel int64
	now := time.Now()
	oldKey := p.cache.Set(xl, key, size)
	cacheSet = int64(time.Since(now) / 100)
	if oldKey != "" {
		swapped = true
		now := time.Now()
		p.Pool.Delete(xl, oldKey)
		poolDel = int64(time.Since(now) / 100)
	}
	log.Debugf("SimpleLocalBd.setCache: swapped %v, cache.set %v(100ns), pool.del %v(100ns)", swapped, cacheSet, poolDel)
	return
}

func (p *SimpleLocalBd) keyToString(key []byte) string {
	return hex.EncodeToString(key)
}

func (p *SimpleLocalBd) Get(xl *xlog.Logger, key []byte, out io.Writer, from, to int, bds [4]uint16) error {

	key1 := p.keyToString(key)

	now := time.Now()
	key2 := p.cache.Get(xl, key1)
	cacheGet := int64(time.Since(now) / 100)
	if key2 != "" {
		now := time.Now()
		r, n, err := p.Pool.Get(xl, key2, int64(from))
		poolGet := int64(time.Since(now) / 100)
		if err == nil {
			defer r.Close()
			// 临时添加，为了清理bug造成的脏数据
			if n+int64(from) == 0 && key1 != "da39a3ee5e6b4b0d3255bfef95601890afd80709" {
				p.cache.Delete(xl, key1)
				p.Pool.Delete(xl, key2)
				goto ReadRbd
			}
			if n > int64(to-from) {
				n = int64(to - from)
			}
			tw := timeio.NewWriter(out)
			now := time.Now()
			_, err = io.CopyN(tw, r, n)
			ioCopy := int64(time.Since(now) / 100)
			ioWrite := tw.Time() / 100
			xl.Debugf("SimpleLocalBd.Get: hit, cache.get %v(100ns), pool.get %v(100ns), io.write %v(100ns), io.copy %v(100ns), err %v", cacheGet, poolGet, ioWrite, ioCopy, err)
			return err
		} else {
			if err == syscall.ENOENT {
				p.cache.Delete(xl, key1)
				goto ReadRbd
			}
		}
		tw := timeio.NewWriter(out)
		now = time.Now()
		err = p.Storage.Get(xl, key, out, from, to, bds)
		stgGet := int64(time.Since(now) / 100)
		ioWrite := tw.Time() / 100
		xl.Debugf("SimpleLocalBd.Get: hit half, cache.get %v(100ns), pool.get %v(100ns), io.write %v(100ns), stg.get %v(100ns), err %v", cacheGet, poolGet, ioWrite, stgGet, err)
		return err
	}

ReadRbd:
	xl.Info("SimpleLocalBd.Get: read from rbd", key1)
	xl.Xlogf("(rbd/myidc:%d/idc:%d/ibd:%d)", p.MyId, bds[3], bds[0])

	var ioWrite1, ioWrite2, stgGet int64
	n, err := p.put2Check(xl, key1, 1<<p.ChunkBits,
		func(xl *xlog.Logger, w io.Writer) error {
			tw1 := timeio.NewWriter(w)
			tw2 := timeio.NewWriter(out)
			writer := streamWriter(tw1, tw2, from, to)
			now := time.Now()
			err := p.Storage.Get(xl, key, writer, 0, 1<<p.ChunkBits, bds)
			stgGet = int64(time.Since(now) / 100)
			ioWrite1 = tw1.Time() / 100
			ioWrite2 = tw2.Time() / 100
			return streamWriterErr(err, writer)
		})
	xl.Debugf("SimpleLocalBd.Get: miss1, %v bytes, io.write1 %v(100ns), io.write2 %v(100ns), stg.get %v(100ns), err %v", n, ioWrite1, ioWrite2, stgGet, err)

	if err != nil {
		if err != EKeyNotFound {
			xl.Warn("LocalBd.Get failed:", err)
			return err
		}
		idc := int(bds[3])
		xl.Warn("SimpleLocalBd.Get: get from backend err, try other idc... :", key1, err, idc)
		other, ok := p.OtherLbds[idc]
		if !ok || idc == p.MyId {
			xl.Warn("SimpleLocalBd.Get: can not find other idc(or hit myself) :", idc, p.MyId, ok)
			return err
		}
		xl.Xlogf("(olbd/myidc:%d/idc:%d/ibd:%d)", p.MyId, idc, bds[0])
		var ioWrite1, ioWrite2, otherGet int64
		n, err = p.put2Check(xl, key1, 1<<p.ChunkBits,
			func(xl *xlog.Logger, w io.Writer) error {
				tw1 := timeio.NewWriter(w)
				tw2 := timeio.NewWriter(out)
				writer := streamWriter(tw1, tw2, from, to)
				now := time.Now()
				err := other.Get(xl, key, writer, 0, 1<<p.ChunkBits, bds)
				otherGet = int64(time.Since(now) / 100)
				ioWrite1 = tw1.Time() / 100
				ioWrite2 = tw2.Time() / 100
				return streamWriterErr(err, writer)
			})
		xl.Debugf("SimpelLocalBd.Get: miss2, %v bytes, io.write1 %v(100ns), io.write2 %v(100ns), other.get %v(100ns), err %v", n, ioWrite1, ioWrite2, otherGet, err)
		if err != nil {
			xl.Warn("SimpleLocalBd.Get: try other idc err :", key1, err)
			return err
		}
	}

	p.setCache(xl, key1, n)

	/*	r, n, err := p.Pool.Get(key1, int64(from))
		if err != nil {
			xl.Warn("SimpleLocalBd.Get: get from pool failed", key1, err)
			return err
		}
		if n > int64(to-from) {
			n = int64(to-from)
		}
		_, err = io.CopyN(out, r, n)
	*/return err
}

func streamWriterErr(err error, mw *qio.OptimisticMultiWriter) error {

	if err != nil {
		return err
	}
	return mw.Writers[0].Err
}

func streamWriter(fw io.Writer, pw io.Writer, from, to int) *qio.OptimisticMultiWriter {

	fw2 := qioutil.SeqWriter(ioutil.Discard, from, pw, to-from)
	return &qio.OptimisticMultiWriter{
		Writers: []qio.OptimisticWriter{
			{Writer: fw},
			{Writer: fw2},
		},
	}
}

func (p *SimpleLocalBd) Put(xl *xlog.Logger, r io.Reader, n int, key []byte, doCache bool, bds [3]uint16) (err error) {

	key1 := p.keyToString(key)
	var n1 int64
	if doCache {
		now := time.Now()
		n1, err = p.put1(xl, key1, r, int64(n))
		xl.Info("SimpleLocalBd.Put: p.put1:", int64(time.Since(now)/100))
		if err != nil {
			xl.Warn("SimpleLocalBd.Put: copy to local err", key1, err)
			return err
		}
		p.setCache(xl, key1, n1)
		reader, _, err := p.Pool.Get(xl, key1, 0)
		if err != nil {
			return err
		}
		defer reader.Close()
		r = reader
	}

	tr := timeio.NewReader(r)
	now := time.Now()
	err = p.Storage.Put(xl, key, tr, n, bds)
	stgPut := int64(time.Since(now) / 100)
	ioRead := tr.Time() / 100
	xl.Debugf("SimpleLocalBd.Put: %v bytes, stg.put %v(100ns), io.read %v(100ns), err %v", n, stgPut, ioRead, err)
	return
}

func (p *SimpleLocalBd) GetLocal(xl *xlog.Logger, key []byte) (r io.ReadCloser, size int, err error) {

	var cacheGet, poolGetw, read4b int64
	defer func() {
		xl.Debugf("SimpleLocalBd.GetLocal: %v bytes, cache.get %v(100ns), pool.getw %v(100ns), read.4b %v(100ns), err %v", size, cacheGet, poolGetw, read4b, err)
	}()

	key1 := p.keyToString(key)
	now := time.Now()
	key2 := p.cache.Get(xl, key1)
	cacheGet = int64(time.Since(now) / 100)
	if key2 == "" {
		err = EKeyNotFound
		return
	}

	now = time.Now()
	r1, n, err := p.Pool.Get(xl, key1, 0)
	poolGetw = int64(time.Since(now) / 100)
	if err != nil {
		if err == syscall.ENOENT {
			p.cache.Delete(xl, key1)
			err = EKeyNotFound
		}
		return
	}
	finish := false
	defer func() {
		if !finish {
			r1.Close()
		}
	}()

	header := make([]byte, 4)
	now = time.Now()
	_, err = io.ReadFull(r1, header)
	read4b = int64(time.Since(now) / 100)
	if err != nil {
		xl.Warn("read header err :", key1, err)
		return
	}

	size = int(binary.LittleEndian.Uint32(header))

	if int64(size)+4 != n {
		xl.Warn("read wrong data", key1, n, size)
		p.Pool.Delete(xl, key1)
		p.cache.Delete(xl, key1)
		err = EKeyNotFound
		return
	}

	if size > MaxCacheSize {
		xl.Warn("too large size", key1)
		err = errors.New("to large size")
		return
	}

	type _readCloser struct {
		io.Reader
		io.Closer
	}

	r = _readCloser{&io.LimitedReader{r1, int64(size)}, r1}

	finish = true
	return
}

func (p *SimpleLocalBd) PutLocal(xl *xlog.Logger, r io.Reader, n int, key []byte) (err error) {

	key1 := p.keyToString(key)

	if n > MaxCacheSize {
		xl.Warn("too large size", key1)
		return errors.New("to large size")
	}

	var write4b, ioWrite, ioCopy int64
	n1, err := p.put2(xl, key1,
		func(writer io.Writer) error {
			header := make([]byte, 4)
			binary.LittleEndian.PutUint32(header, uint32(n))
			now := time.Now()
			_, err := writer.Write(header)
			write4b = int64(time.Since(now) / 100)
			if err != nil {
				xl.Warn("write header err :", key1, err)
				return err
			}

			tw := timeio.NewWriter(writer)
			now = time.Now()
			_, err = io.CopyN(tw, r, int64(n))
			ioCopy = int64(time.Since(now) / 100)
			ioWrite = tw.Time() / 100
			return err
		})
	xl.Debugf("SimpleLocalBd.PutLocal: %v bytes, write.4b %v(100ns), io.write %v(100ns), io.copy %v(100ns), err %v", n1, write4b, ioWrite, ioCopy, err)
	if err != nil {
		xl.Warn("LocalBd.PutLocal put2 err:", err)
		return
	}

	p.setCache(xl, key1, n1)
	return
}
