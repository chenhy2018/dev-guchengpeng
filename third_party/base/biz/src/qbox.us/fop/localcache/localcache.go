package localcache

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"github.com/qiniu/errors"
	"github.com/qiniu/log.v1"
	"github.com/qiniu/xlog.v1"
)

var (
	ErrNotFound     = errors.New("not found")
	ErrSizeMismatch = errors.New("mismatched fsize with saved")
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type LocalCacheConfig struct {
	DiskDir      string `json:"disk_dir"`
	RamDir       string `json:"ram_dir"`
	MemLimitByte int64  `json:"mem_limit_byte"` // 默认 0 不开启内存缓存
	ExpireS      int64  `json:"expire_s"`
}

type LocalCache struct {
	LocalCacheConfig
	memCache  *MemCache
	fnameBase int64
}

func NewLocalCache(cfg *LocalCacheConfig) (lc *LocalCache, err error) {
	if cfg == nil {
		cfg = &LocalCacheConfig{}
	}

	if cfg.DiskDir == "" {
		cfg.DiskDir = os.TempDir()
	}
	if err = os.MkdirAll(cfg.DiskDir, 0775); err != nil {
		err = errors.Info(err, "os.MkdirAll", cfg.DiskDir).Detail(err)
		return
	}

	var mc *MemCache
	if cfg.RamDir != "" {
		if err = os.MkdirAll(cfg.RamDir, 0775); err != nil {
			err = errors.Info(err, "os.MkdirAll", cfg.RamDir).Detail(err)
			return
		}
	} else {
		mc, err = NewMemCache(time.Duration(cfg.ExpireS) * time.Second)
		if err != nil {
			err = errors.Info(err, "NewMemCache").Detail(err)
			return
		}
	}

	go cleanFileDir(cfg.DiskDir, cfg.ExpireS)
	if cfg.RamDir != "" {
		go cleanFileDir(cfg.RamDir, cfg.ExpireS)
	}

	lc = &LocalCache{
		LocalCacheConfig: *cfg,
		memCache:         mc,
	}
	return
}

const (
	typeMem     = "mem"
	typeRamDisk = "ramdisk"
	typeDisk    = "disk"
)

func (lc *LocalCache) encodeFid(t, fname string) string {
	return t + "-" + fname
}

func (lc *LocalCache) decodeFid(fid string) (t, fname string) {

	l := strings.SplitN(fid, "-", 2)
	if len(l) != 2 {
		// default typeDisk
		return typeDisk, fid
	}
	return l[0], l[1]

}

func (lc *LocalCache) TempDir(fsize int64) (dir string, err error) {
	if fsize > lc.MemLimitByte || lc.MemLimitByte == 0 || fsize == -1 {
		dir = lc.DiskDir
	} else {
		if lc.RamDir == "" {
			err = errors.New("empty ram disk dir")
			return
		}
		dir = lc.RamDir
	}
	return
}

func (lc *LocalCache) TempFile(prefix string, fsize int64) (f *os.File, err error) {
	base, err := lc.TempDir(fsize)
	if err != nil {
		return
	}
	fpath := filepath.Join(base, prefix+lc.newFname())
	return os.Create(fpath)
}

func (lc *LocalCache) Save(xl *xlog.Logger, body io.Reader, fsize int64) (fid string, n int64, err error) {

	fname := lc.newFname()

	// disk
	if fsize > lc.MemLimitByte || lc.MemLimitByte == 0 || fsize == -1 {
		fpath := filepath.Join(lc.DiskDir, fname)
		n, err = lc.saveReaderToFile(xl, fpath, body, fsize)
		if err != nil {
			err = errors.Info(err, "lc.saveReaderToFile").Detail(err)
			return
		}
		fid = lc.encodeFid(typeDisk, fname)
		return
	}

	// ramdisk
	if lc.RamDir != "" {
		fpath := filepath.Join(lc.RamDir, fname)
		n, err = lc.saveReaderToFile(xl, fpath, body, fsize)
		if err != nil {
			err = errors.Info(err, "lc.saveReaderToFile").Detail(err)
			return
		}
		fid = lc.encodeFid(typeRamDisk, fname)
		return
	}

	// mem
	data, err := ioutil.ReadAll(body)
	if err != nil {
		err = errors.Info(err, "ioutil.ReadAll").Detail(err)
		return
	}
	if int64(len(data)) != fsize {
		xl.Warnf("mismatched readerSize(%d) and fsize(%d)", len(data), fsize)
		err = ErrSizeMismatch
		return
	}
	lc.memCache.Set(fname, data)
	n = int64(len(data))
	fid = lc.encodeFid(typeMem, fname)

	return
}

func (lc *LocalCache) newFname() string {

	id := atomic.AddInt64(&lc.fnameBase, 1)
	return fmt.Sprintf("%x_%x_%x_%x", os.Getpid(), time.Now().UnixNano()/100, id, rand.Int()%1000)
}

func (lc *LocalCache) saveReaderToFile(xl *xlog.Logger, fpath string, body io.Reader, fsize int64) (n int64, err error) {
	f, err := os.Create(fpath)
	if err != nil {
		return
	}
	defer func() {
		f.Close()
		if err != nil {
			os.Remove(fpath)
		}
	}()

	n, err = io.Copy(f, body)
	if err != nil {
		err = errors.Info(err, "io.Copy").Detail(err)
		return
	}
	if fsize != n && fsize != -1 {
		xl.Warnf("saveReader: mismatched readerSize(%d) and fsize(%d)", n, fsize)
		err = ErrSizeMismatch
		return
	}
	return
}

func (lc *LocalCache) Remove(fid string) (err error) {

	t, fname := lc.decodeFid(fid)

	switch t {
	case typeDisk:
		fpath := filepath.Join(lc.DiskDir, fname)
		return os.Remove(fpath)
	case typeRamDisk:
		fpath := filepath.Join(lc.RamDir, fname)
		return os.Remove(fpath)
	case typeMem:
		lc.memCache.Remove(fname)
		return
	default:
		err = errors.New("invalid type")
	}
	return
}

func (lc *LocalCache) Get(fid string) (rc ReaderAtSeekCloser, err error) {

	t, fname := lc.decodeFid(fid)

	switch t {
	case typeDisk:
		fpath := filepath.Join(lc.DiskDir, fname)
		rc, err = os.Open(fpath)
		if err != nil {
			if strings.Contains(err.Error(), "no such file or directory") {
				err = ErrNotFound
				return
			}
			err = errors.Info(err, "os.Open").Detail(err)
			return
		}
		return
	case typeRamDisk:
		fpath := filepath.Join(lc.RamDir, fname)
		rc, err = os.Open(fpath)
		if err != nil {
			if strings.Contains(err.Error(), "no such file or directory") {
				err = ErrNotFound
				return
			}
			err = errors.Info(err, "os.Open").Detail(err)
			return
		}
		return
	case typeMem:
		var data []byte
		data, err = lc.memCache.Get(fname)
		if err == ErrNoSuchEntry {
			err = ErrNotFound
			return
		}
		if err != nil {
			return
		}
		rc = rsc{Reader: bytes.NewReader(data)}
		return
	default:
		err = errors.New("invalid type")
		return
	}

	return
}

func cleanFileDir(dirName string, deleteAfterS int64) {
	if deleteAfterS <= 0 {
		return
	}

	tick := time.Tick(time.Second * time.Duration(deleteAfterS) / 10)
	for _ = range tick {
		dir, err := os.Open(dirName)
		if err != nil {
			log.Error("os.Open: ", dirName, err)
			continue
		}
		fis, err := dir.Readdir(-1)
		dir.Close()
		if err != nil {
			log.Warn("dir.Readdir: ", err)
			continue
		}

		outOfDate := time.Now().Add(-time.Duration(deleteAfterS) * time.Second)
		for _, fi := range fis {
			if fi.IsDir() {
				continue
			}
			// Mac OS X ModTime 精确度是秒
			if fi.ModTime().Before(outOfDate) {
				log.Info("[schedule] rm file:", fi.Name())
				os.Remove(filepath.Join(dirName, fi.Name()))
			}
		}
	}
}

// -----------------------

type ReaderAtSeeker interface {
	io.ReaderAt
	io.ReadSeeker
}

type ReaderAtSeekCloser interface {
	ReaderAtSeeker
	io.Closer
}

type rsc struct {
	*bytes.Reader
}

func (r rsc) Close() error {
	return nil
}
