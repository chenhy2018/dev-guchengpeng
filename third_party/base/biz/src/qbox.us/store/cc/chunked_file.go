package cc

import (
	"crypto/sha1"
	"hash"
	"sync"
)

type ChunkedFile struct {
	Buf       []byte
	From      int
	ChunkBits uint
	Pos       int64
	Sha1      hash.Hash
	Pool      *ChunkPool
	RW        ReadWriterAt
	Commit    func(key []byte, w *ChunkedFile, chunkNo int32, chunkSize int) error
	*sync.Mutex
}

func NopCommit(key []byte, w *ChunkedFile, chunkNo int32, chunkSize int) error {
	return nil
}

func FreeChunk(w *ChunkedFile, chunkNo int32, chunkSize int) error {
	w.Pool.Free(chunkNo)
	return nil
}

func NewChunkedFile(rw ReadWriterAt, pool *ChunkPool, chunkBits uint, bufBits uint) *ChunkedFile {
	if bufBits > chunkBits {
		return nil
	}
	buf := make([]byte, 1<<bufBits)
	pos := int64(pool.Alloc()) << chunkBits
	return &ChunkedFile{buf, 0, chunkBits, pos, sha1.New(), pool, rw, NopCommit, new(sync.Mutex)}
}

func (p *ChunkedFile) ChunkNo() int32 { // 这个函数假设由外部来锁定
	return int32(p.Pos >> p.ChunkBits)
}

func (p *ChunkedFile) Chunk() (chunkNo int32, chunkSize int) { // 这个函数假设由外部来锁定
	chunkBits := p.ChunkBits
	chunkNo = int32(p.Pos >> chunkBits)
	chunkMask := (1 << chunkBits) - 1
	chunkSize = (int(p.Pos) & chunkMask) + p.From
	return
}

func (p *ChunkedFile) Sync() error {
	p.Lock()
	defer p.Unlock()
	return p.DirtySync()
}

func (p *ChunkedFile) DirtySync() error {
	_, err := p.RW.WriteAt(p.Buf[:p.From], p.Pos)
	if err != nil {
		return err
	}
	_, err = p.Sha1.Write(p.Buf[:p.From])
	return err
}

func (p *ChunkedFile) Close() error {
	p.Lock()
	defer p.Unlock()
	return p.DirtyClose()
}

func (p *ChunkedFile) DirtyClose() error {
	chunkNo, chunkSize := p.Chunk()
	if chunkSize > 0 {
		err := p.DirtySync()
		if err != nil {
			return err
		}
		return p.Commit(p.Sha1.Sum(nil), p, chunkNo, chunkSize)
	}
	p.Pool.Free(chunkNo)
	return nil
}

func (p *ChunkedFile) Write(val []byte) (n int, err error) {
	p.Lock()
	defer p.Unlock()
	return p.DirtyWrite(val)
}

func (p *ChunkedFile) DirtyWrite(val []byte) (n int, err error) {
	for len(val) > 0 {
		cnt := copy(p.Buf[p.From:], val)
		p.From += cnt
		if p.From < len(p.Buf) {
			break
		}

		_, err = p.RW.WriteAt(p.Buf, p.Pos)
		if err != nil {
			return
		}
		_, err = p.Sha1.Write(p.Buf)
		if err != nil {
			return
		}
		p.From = 0
		p.Pos += int64(len(p.Buf))

		chunkBits := p.ChunkBits
		chunkMaxSize := 1 << chunkBits
		if (int(p.Pos) & (chunkMaxSize - 1)) == 0 {
			chunkNo := int32(p.Pos>>chunkBits) - 1
			p.Pos = int64(p.Pool.Alloc()) << chunkBits
			err = p.Commit(p.Sha1.Sum(nil), p, chunkNo, chunkMaxSize)
			p.Sha1.Reset()
		}
		n += cnt
		val = val[cnt:]
	}
	n += len(val)
	return
}
