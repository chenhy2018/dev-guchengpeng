package store

import (
	"io"
	"qbox.us/store/cc"
	"github.com/qiniu/xlog.v1"
)

// -----------------------------------------------------------

type CachedStorage struct {
	Pool      *cc.ChunkPool
	RW        cc.ReadWriterAt
	Storage   MultiBdInterface
	ChunkBits uint
}

func NewCachedStorage(rw cc.ReadWriterAt, pool *cc.ChunkPool, chunkBits uint, stg MultiBdInterface) *CachedStorage {
	return &CachedStorage{pool, rw, stg, chunkBits}
}

func (p *CachedStorage) Put(xl *xlog.Logger, key []byte, chunkNo int32, chunkSize int, doCache bool, bds [3]uint16) (err error) {

	r1 := cc.NewReader(p.RW, int64(chunkNo)<<p.ChunkBits)
	err = p.Storage.Put(xl, key, r1, chunkSize, doCache, bds)

	xl.Debug("CachedStorage.Put: p.Pool.Free:", chunkNo)
	p.Pool.Free(chunkNo)
	return
}

func (p *CachedStorage) Get(xl *xlog.Logger, key []byte, w io.Writer, from, to int, bds [4]uint16) error {
	xl.Debug("CachedStorage.Get:", key, from, to, bds)
	err := p.Storage.Get(xl, key, w, from, to, bds) // 这要求 Getter.Get 接口在 to 参数超出范围时不认为是错误
	return err
}

// -----------------------------------------------------------
