package cc

import (
	"encoding/hex"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"github.com/qiniu/xlog.v1"
	"strconv"
	"testing"
)

func clearCacheFile(prefix string) error {

	err := os.Remove(prefix + "_list")
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	for i := 0; i < 10; i++ {
		err := os.Remove(prefix + "_append_" + strconv.Itoa(i))
		if err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

func createCacheFile(prefix string, idxs []int) error {

	if err := ioutil.WriteFile(prefix+"_list", []byte{}, 0777); err != nil {
		return err
	}

	for _, i := range idxs {
		err := ioutil.WriteFile(prefix+"_append_"+strconv.Itoa(i), []byte{}, 0777)
		if err != nil {
			return err
		}
	}
	return nil
}

func testLoad(idx int, idxs []int) error {

	if err := clearCacheFile(""); err != nil {
		return err
	}

	if err := createCacheFile("", idxs); err != nil {
		return err
	}

	cache := &SimpleKeyCacheEx{
		maxBuf:         100,
		SimpleKeyCache: NewSimpleKeyCache(nil),
	}

	if err := cache.load(); err != nil {
		return err
	}

	if cache.idx != idx {
		return errors.New("unexpect idx")
	}

	return nil
}

type loadCase struct {
	idx  int
	idxs []int
}

func TestLoad(t *testing.T) {

	defer clearCacheFile("")

	cases := []loadCase{
		loadCase{0, []int{0}},
		loadCase{1, []int{1}},
		loadCase{1, []int{0, 1}},
		loadCase{2, []int{1, 2}},
		loadCase{9, []int{9}},
		loadCase{9, []int{8, 9}},
		loadCase{9, []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}},
		loadCase{0, []int{9, 0}},
		loadCase{0, []int{7, 8, 9, 0}},
		loadCase{1, []int{7, 8, 9, 0, 1}},
		loadCase{7, []int{1, 7}},
		loadCase{1, []int{1, 9}},
	}

	for _, c := range cases {
		if err := testLoad(c.idx, c.idxs); err != nil {
			t.Fatalf("testLoad: idx:%v idxs:%v failed => %v\n", c.idx, c.idxs, err)
		}
	}
}

func TestKeyCacheEx(t *testing.T) {

	defer clearCacheFile("cache")
	xl := xlog.NewDummy()

	// clean
	dir := "." //os.TempDir()
	os.Remove(filepath.Join(dir, "cache_list"))
	for i := 0; i < 10; i++ {
		os.Remove(filepath.Join(dir, "cache_append_"+strconv.Itoa(i)))
	}

	// keys
	keys := make([]string, 256*256)
	for i := 0; i < 256; i++ {
		for j := 0; j < 256; j++ {
			keys[i*256+j] = hex.EncodeToString([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6, 7, 8, byte(i), byte(j)})
		}
	}

	flag := false
	outOfLimit := func(space int64, count int) bool {
		return flag
	}

	cache, err := NewSimpleKeyCacheEx(filepath.Join(dir, "cache"), 1e9*3600, 1, outOfLimit)
	if err != nil {
		t.Fatal("NewSimpleKeyCacheEx:", err)
	}

	for i := 0; i < 8; i++ {
		cache.Set(xl, keys[i], int64(1))
	}

	cache.Shutdown()

	cache, err = NewSimpleKeyCacheEx(filepath.Join(dir, "cache"), 1e9*3600, 1, outOfLimit)
	if cache == nil || err != nil {
		t.Fatal("init cache failed:", err)
		return
	}
	for i := 0; i < 8; i++ {
		if cache.Get(xl, keys[i]) != keys[i] {
			t.Fatal("can not get after save", keys[i])
		}
	}

	for i := 8; i < 16; i++ {
		cache.Set(xl, keys[i], int64(1))
	}
	flag = true
	for i := 16; i < 32; i++ {
		cache.Set(xl, keys[i], int64(1))
	}

	cache.Shutdown()

	flag = false
	cache, err = NewSimpleKeyCacheEx(filepath.Join(dir, "cache"), 1e9*3600, 1, outOfLimit)
	if err != nil {
		t.Fatal("NewSimpleKeyCacheEx:", err)
	}
	for i := 0; i < 16; i++ {
		if cache.Get(xl, keys[i]) != "" {
			t.Fatal("wrong get after save", keys[i])
		}
	}
	for i := 16; i < 32; i++ {
		if cache.Get(xl, keys[i]) != keys[i] {
			t.Fatal("can not get after save", keys[i])
		}
	}
	for i := 16; i < 24; i++ {
		cache.Get(xl, keys[i])
	}
	flag = true
	for i := 0; i < 8; i++ {
		cache.Set(xl, keys[i], int64(1))
	}
	for i := 0; i < 8; i++ {
		if cache.Get(xl, keys[i]) != keys[i] {
			t.Fatal("can not get", keys[i])
		}
	}
	for i := 16; i < 24; i++ {
		if cache.Get(xl, keys[i]) != keys[i] {
			t.Fatal("can not get", keys[i])
		}
	}
	for i := 24; i < 32; i++ {
		if cache.Get(xl, keys[i]) != "" {
			t.Fatal("wrong get", keys[i])
		}
	}

	cache.Shutdown()

	for i := 2; i < 16; i++ {
		flag = false
		cache, err = NewSimpleKeyCacheEx(filepath.Join(dir, "cache"), 1e7, 1, outOfLimit)
		if err != nil {
			t.Fatal("NewSimpleKeyCacheEx:", err)
		}
		flag = true
		for j := 0; j < 16; j++ {
			cache.Set(xl, keys[i*16+j], int64(1))
		}
		for j := 0; j < 16; j++ {
			if cache.Get(xl, keys[i*16+j]) != keys[i*16+j] {
				t.Fatal("cat not get", keys[i*16+j])
			}
		}
		cache.done1 <- true
		<-cache.done2
		{
			flag := false
			outOfLimit := func(space int64, count int) bool {
				return flag
			}
			cache, err := NewSimpleKeyCacheEx(filepath.Join(dir, "cache"), 1e9*3600, 1, outOfLimit)
			if err != nil {
				t.Fatal("NewSimpleKeyCacheEx:", err)
			}

			for j := 0; j < 16; j++ {
				if cache.Get(xl, keys[i*16+j]) != keys[i*16+j] {
					t.Fatal("cat not get", keys[i*16+j])
				}
			}

			cache.done1 <- true
			<-cache.done2
		}
	}
}
