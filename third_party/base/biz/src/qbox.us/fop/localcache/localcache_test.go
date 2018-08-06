package localcache

import (
	"bytes"
	"io/ioutil"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/qiniu/xlog.v1"
	"github.com/stretchr/testify/assert"
)

func randRead(p []byte) (n int, err error) { // for go1.4
	for i := 0; i < len(p); i += 7 {
		val := rand.Int63()
		for j := 0; i+j < len(p) && j < 7; j++ {
			p[i+j] = byte(val)
			val >>= 8
		}
	}
	return len(p), nil
}

func TestLocalCacheWithMem(t *testing.T) {
	testLocalCache(t, false)
}

func TestLocalCacheWithRamDisk(t *testing.T) {
	testLocalCache(t, true)
}

func testLocalCache(t *testing.T, useRamDisk bool) {

	diskDir, err := ioutil.TempDir("", "localcache-disk")
	assert.NoError(t, err)
	defer os.RemoveAll(diskDir)

	ramDir := ""
	if useRamDisk {
		ramDir, err = ioutil.TempDir("", "localcache-ram")
		assert.NoError(t, err)
		defer os.RemoveAll(ramDir)
	}

	cfg := &LocalCacheConfig{
		DiskDir:      diskDir,
		RamDir:       ramDir,
		MemLimitByte: 1024,
		ExpireS:      2, // Mac OS X ModTime 精确度是秒
	}

	ast := assert.New(t)

	lc, err := NewLocalCache(cfg)
	ast.Nil(err)

	type Case struct {
		size int
		data []byte
	}
	cases := []Case{
		{size: 1},
		{size: 8},
		{size: 512},
		{size: 1024},
		{size: 1024 + 1},
		{size: 1024 * 2},
		{size: 1024 * 10},
	}
	for i, cs := range cases {
		b := make([]byte, cs.size)
		_, err := randRead(b)
		ast.Nil(err)
		cases[i].data = b
	}

	xl := xlog.NewDummy()
	for _, cs := range cases {
		_, _, err = lc.Save(xl, bytes.NewReader(cs.data), int64(cs.size+1))
		ast.Equal(ErrSizeMismatch, err)

		fid, fsize, err := lc.Save(xl, bytes.NewReader(cs.data), -1)
		ast.Nil(err)
		ast.Equal(cs.size, fsize)

		b, err := get(lc, fid)
		ast.Nil(err)
		ast.Equal(cs.data, b)

		// get twice
		b, err = get(lc, fid)
		ast.Nil(err)
		ast.Equal(cs.data, b)

		ast.Nil(lc.Remove(fid))

		_, err = lc.Get(fid)
		ast.Equal(ErrNotFound, err)
	}

	var fids []string
	for _, cs := range cases {
		fid, fsize, err := lc.Save(xl, bytes.NewReader(cs.data), int64(cs.size))
		ast.Nil(err)
		ast.Equal(cs.size, fsize)

		b, err := get(lc, fid)
		ast.Nil(err)
		ast.Equal(cs.data, b)

		// get twice
		b, err = get(lc, fid)
		ast.Nil(err)
		ast.Equal(cs.data, b)

		fids = append(fids, fid)
	}

	time.Sleep(500 * time.Millisecond)
	for i, cs := range cases {
		fid := fids[i]
		b, err := get(lc, fid)
		ast.Nil(err, fid)
		ast.Equal(cs.data, b)

		ast.Nil(lc.Remove(fid))
		_, err = lc.Get(fid)
		ast.Equal(ErrNotFound, err)
	}

	// remove after
	fids = []string{}
	for _, cs := range cases {
		fid, fsize, err := lc.Save(xl, bytes.NewReader(cs.data), int64(cs.size))
		ast.Nil(err)
		ast.Equal(cs.size, fsize)
		fids = append(fids, fid)
	}
	time.Sleep(time.Millisecond * 2200)
	for _, fid := range fids {
		_, err = lc.Get(fid)
		ast.Equal(ErrNotFound, err)
	}
}

func get(lc *LocalCache, fid string) (b []byte, err error) {
	r, err := lc.Get(fid)
	if err != nil {
		return
	}
	b, err = ioutil.ReadAll(r)
	if err != nil {
		return
	}
	err = r.Close()
	return
}
