package lbd

import (
	"errors"
	"io"
	"os"
	"qbox.us/store/cc"
	//"github.com/qiniu/log.v1"
)

type CacheInfo struct {
	file *os.File
}

func NewCacheInfo(filename string, limit int32) (*CacheInfo, error) {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}
	err = file.Truncate(int64(limit) * 21)
	if err != nil {
		return nil, err
	}
	return &CacheInfo{file}, err
}

func (p *CacheInfo) Set(key []byte, index int32) (err error) {
	if len(key) != 20 {
		return errors.New("len of key must be 20")
	}
	var data [21]byte
	copy(data[:20], key)
	data[20] = 1
	_, err = p.file.WriteAt(data[:], int64(index*21))
	return
}

func (p *CacheInfo) Clear(index int32) (err error) {
	_, err = p.file.WriteAt([]byte{0}, int64(index*21+20))
	return
}

func (p *CacheInfo) Load(cache *cc.SimpleIntCache, pool *cc.ChunkPool) (err error) {
	fi, err := p.file.Stat()
	if err != nil {
		return
	}
	data := make([]byte, fi.Size())
	_, err = io.ReadFull(p.file, data)
	if err != nil {
		return
	}
	num := len(data) / 21
	pool.UseAll()
	for i := 0; i < num; i++ {
		if data[i*21+20] == 0 {
			pool.Free(int32(i))
		} else {
			cache.Set(string(data[i*21:i*21+20]), int32(i))
		}
	}
	return
}
