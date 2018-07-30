package cc

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"github.com/qiniu/log.v1"
	"github.com/qiniu/xlog.v1"
	"strconv"
	"time"
)

const (
	_SET    = 1
	_DELETE = 2
	_GET    = 3
)

type SimpleKeyCacheEx struct {
	idx        int
	maxBuf     int
	done1      chan bool
	done2      chan bool
	fileName   string
	appendFile *_BuffFile
	*SimpleKeyCache
}

func NewSimpleKeyCacheEx(fileName string, duration int64, maxBuf int,
	outOfLimit func(space int64, count int) bool) (*SimpleKeyCacheEx, error) {
	cache := &SimpleKeyCacheEx{0, maxBuf, make(chan bool), make(chan bool),
		fileName, nil, NewSimpleKeyCache(outOfLimit)}
	err := cache.load()
	if err != nil {
		return nil, err
	}
	cache.appendFile, err = _CreateBuffFile(fileName+"_append_"+strconv.Itoa(cache.idx), 29, maxBuf)
	if err != nil {
		return nil, err
	}
	go func() {
		for {
			cache.save()
			ch := time.After(time.Duration(duration))
			select {
			case <-ch:
			case <-cache.done1:
				cache.done2 <- true
				return
			}
		}
	}()
	return cache, nil
}

func (r *SimpleKeyCacheEx) Shutdown() {
	r.save()

	r.done1 <- true
	<-r.done2
}

func (r *SimpleKeyCacheEx) Get(xl *xlog.Logger, key string) string {

	r.mutex.Lock()
	defer r.mutex.Unlock()

	old := r.SimpleKeyCache.DirtyGet(key)
	if old != "" {
		data := make([]byte, 29)
		keyB, _ := hex.DecodeString(key)
		copy(data[:20], keyB)
		data[28] = _GET
		r.appendFile.Write(data)
	}
	return old
}

func (r *SimpleKeyCacheEx) Set(xl *xlog.Logger, key string, size int64) string {

	r.mutex.Lock()
	defer r.mutex.Unlock()

	old := r.SimpleKeyCache.DirtySet(key, size)

	if old == "" {
		data := make([]byte, 29)
		keyB, _ := hex.DecodeString(key)
		copy(data[:20], keyB)
		binary.LittleEndian.PutUint64(data[20:28], uint64(size))
		data[28] = _SET
		r.appendFile.Write(data)
	} else {
		data := make([]byte, 29*2)
		keyB, _ := hex.DecodeString(old)
		copy(data[:20], keyB)
		data[28] = _DELETE
		keyB, _ = hex.DecodeString(key)
		copy(data[29:49], keyB)
		binary.LittleEndian.PutUint64(data[49:57], uint64(size))
		data[57] = _SET
		r.appendFile.Write(data)
	}
	return old
}

func (r *SimpleKeyCacheEx) Delete(xl *xlog.Logger, key string) string {

	r.mutex.Lock()
	defer r.mutex.Unlock()

	old := r.SimpleKeyCache.DirtyDelete(key)
	if old != "" {
		data := make([]byte, 29)
		keyB, _ := hex.DecodeString(old)
		copy(data[:20], keyB)
		data[28] = _DELETE
		r.appendFile.Write(data)
	}
	return old
}

func (r *SimpleKeyCacheEx) load() error {
	file, err := os.Open(r.fileName + "_list")
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()

	fi, err := file.Stat()
	if err != nil {
		return err
	}
	data := make([]byte, fi.Size())
	_, err = io.ReadFull(file, data)
	if err != nil {
		return err
	}
	num := len(data) / 29
	for i := 0; i < num; i++ {
		if data[i*29+28] != 0 {
			r.SimpleKeyCache.Set(
				hex.EncodeToString(data[i*29:i*29+20]),
				int64(binary.LittleEndian.Uint64(data[i*29+20:i*29+28])))
		}
	}

	files := make([]*os.File, 10)
	begin := -1
	end := -1
	last := false
	for i := 0; i < 10; i++ {
		file, err := os.Open(r.fileName + "_append_" + strconv.Itoa(i))
		if err == nil {
			defer file.Close()
			if !last {
				begin = i
			}
			files[i] = file
			last = true
		} else {
			if i != 0 && last {
				end = i - 1
			}
			last = false
		}
	}
	if end == -1 && begin == -1 {
		r.idx = 0
		return nil
	}

	if end == -1 {
		end = 9
	}

	if begin <= end {
		files = files[begin : end+1]
	} else {
		files_ := make([]*os.File, 10-begin+end+1)
		copy(files_[:10-begin], files[begin:])
		copy(files_[10-begin:], files[:end+1])
		files = files_
	}

	for _, file := range files {
		if file == nil {
			continue
		}
		fi, err := file.Stat()
		if err != nil {
			return err
		}
		if fi.Size() == 0 {
			continue
		}
		data := make([]byte, fi.Size())
		_, err = io.ReadFull(file, data)
		if err != nil {
			return err
		}
		num := len(data) / 29
		for i := 0; i < num; i++ {
			if data[i*29+28] == _SET {
				r.SimpleKeyCache.Set(
					hex.EncodeToString(data[i*29:i*29+20]),
					int64(binary.LittleEndian.Uint64(data[i*29+20:i*29+28])))
			} else if data[i*29+28] == _DELETE {
				key := hex.EncodeToString(data[i*29 : i*29+20])
				if e, ok := r.cache[key]; ok {
					r.chunks.Remove(e)
					delete(r.cache, key)
				}
			} else if data[i*29+28] == _GET {
				r.SimpleKeyCache.Get(hex.EncodeToString(data[i*29 : i*29+20]))
			}
		}
	}

	r.idx = end

	return nil
}

func (r *SimpleKeyCacheEx) sw() (ok bool, items []simpleKeyCacheItem) {

	file, err := _CreateBuffFile(r.fileName+"_append_"+strconv.Itoa((r.idx+1)%10), 29, r.maxBuf)
	if err != nil {
		log.Error("sw create bufferfile failed", err)
		return
	}

	r.mutex.Lock()
	defer r.mutex.Unlock()

	begin := time.Now().UnixNano()

	r.idx = (r.idx + 1) % 10
	r.appendFile.Close()
	r.appendFile = file
	log.Info("sw key cache 1", time.Now().UnixNano()-begin)

	items = make([]simpleKeyCacheItem, r.chunks.Len())
	i := 0
	for e := r.chunks.Front(); e != nil; e = e.Next() {
		v := e.Value.(*simpleKeyCacheItem)
		items[i] = simpleKeyCacheItem{v.key, v.size}
		i++
	}
	log.Info("sw key cache", time.Now().UnixNano()-begin)
	ok = true
	return
}

func (r *SimpleKeyCacheEx) save() {

	defer func() {
		p := recover()
		if p != nil {
			log.Error("save panic", p)
		}
	}()

	begin := time.Now().UnixNano()
	ok, items := r.sw()
	if !ok {
		return
	}

	data := make([]byte, len(items)*29)
	for i, item := range items {
		keyB, _ := hex.DecodeString(item.key)
		copy(data[i*29:i*29+20], keyB)
		binary.LittleEndian.PutUint64(data[i*29+20:i*29+28], uint64(item.size))
		data[i*29+28] = 1
	}

	file, err := os.Create(r.fileName + "_list" + "_tmp")
	if err != nil {
		log.Warn("save key cache: Create tmp failed =>", err)
		return
	}
	finish := false
	defer func() {
		file.Close()
		if finish {
			os.Rename(r.fileName+"_list"+"_tmp", r.fileName+"_list")
			for i := 0; i < 10; i++ {
				if i == r.idx {
					continue
				}
				fileName := r.fileName + "_append_" + strconv.Itoa(i)
				if err := os.Remove(fileName); err != nil {
					log.Warn("save key cache: Remove", fileName, "failed =>", err)
				}
			}
		} else {
			os.Remove(r.fileName + "_list" + "_tmp")
		}

		log.Info("save key cache", len(items), time.Now().UnixNano()-begin)
	}()

	if _, err = file.Write(data); err != nil {
		log.Warn("save key cache: Write data failed =>", err)
	}

	finish = err == nil

}

//---------------------------------------------------------------------------//

func (r *SimpleKeyCacheEx) sw2() (data []byte) {

	r.mutex.Lock()
	defer r.mutex.Unlock()

	begin := time.Now().UnixNano()

	r.idx = (r.idx + 1) % 10
	r.appendFile.Close()
	r.appendFile, _ = _CreateBuffFile(r.fileName+"_append_"+strconv.Itoa(r.idx), 29, r.maxBuf)
	log.Info("sw key cache 1", time.Now().UnixNano()-begin)

	data = make([]byte, r.chunks.Len()*29)
	log.Info("sw key cache 2", time.Now().UnixNano()-begin)
	idx := 0
	for e := r.chunks.Front(); e != nil; e = e.Next() {
		item := e.Value.(*simpleKeyCacheItem)
		keyB, _ := hex.DecodeString(item.key)
		copy(data[idx*29:idx*29+20], keyB)
		binary.LittleEndian.PutUint64(data[idx*29+20:idx*29+28], uint64(item.size))
		data[idx*29+28] = 1
		idx++
	}
	log.Info("sw key cache", time.Now().UnixNano()-begin)
	return
}

func (r *SimpleKeyCacheEx) save2() {

	begin := time.Now().UnixNano()
	data := r.sw2()

	file, err := os.Create(r.fileName + "_list" + "_tmp")
	if err != nil {
		return
	}
	finish := false
	defer func() {
		file.Close()
		if finish {
			os.Rename(r.fileName+"_list"+"_tmp", r.fileName+"_list")
			for i := 0; i < 10; i++ {
				if i == r.idx {
					continue
				}
				os.Remove(r.fileName + "_append_" + strconv.Itoa(i))
			}
		} else {
			os.Remove(r.fileName + "_list" + "_tmp")
		}

		log.Info("save key cache", time.Now().UnixNano()-begin)
	}()

	_, err = file.Write(data)

	finish = err == nil
}

//---------------------------------------------------------------------------//

type _BuffFile struct {
	From int
	Size int
	Buf  []byte
	File *os.File
}

func _CreateBuffFile(filename string, itemSize, itemCount int) (*_BuffFile, error) {
	buf := make([]byte, itemSize*itemCount)
	file, err := os.Create(filename)
	if err != nil {
		return nil, err
	}
	return &_BuffFile{0, itemSize, buf, file}, nil
}

func (p *_BuffFile) Sync() (err error) {
	if p.From != 0 {
		_, err = p.File.Write(p.Buf[:p.From])
	}
	return
}

func (p *_BuffFile) Close() (err error) {
	err = p.Sync()
	if err != nil {
		p.File.Close()
	} else {
		err = p.File.Close()
	}
	return
}

func (p *_BuffFile) Write(val []byte) (n int, err error) {

	if len(val)%p.Size != 0 {
		return 0, errors.New("wrong value")
	}

	if len(p.Buf) == 0 {
		return p.File.Write(val)
	}

	for len(val) > 0 {
		cnt := copy(p.Buf[p.From:], val)
		p.From += cnt
		if p.From < len(p.Buf) {
			break
		}

		_, err = p.File.Write(p.Buf)
		if err != nil {
			return
		}
		p.From = 0
		n += cnt
		val = val[cnt:]
	}
	n += len(val)
	return
}
