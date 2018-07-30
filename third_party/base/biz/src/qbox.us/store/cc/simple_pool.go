package cc

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"github.com/qiniu/xlog.v1"
	"strconv"
	"syscall"
	"time"
)

type SimplePool struct {
	root string
}

func NewSimplePool(root string) SimplePool {
	return SimplePool{root}
}

func (p SimplePool) getPath(dir, key string) (string, string) {
	path := filepath.Join(dir, key[:2], key[2:4])
	return path, filepath.Join(path, key)
}

func (p SimplePool) createFile(dir, file string) (*os.File, error) {
	err := os.MkdirAll(dir, 0777)
	if err != nil {
		return nil, err
	}
	return os.Create(file)
}

func (p SimplePool) Put(xl *xlog.Logger, key string, reader io.Reader, length int64) (n int64, err error) {

	d, f := p.getPath(p.root, key)
	f1 := f + "." + strconv.FormatInt(time.Now().UnixNano(), 10)
	file, err := p.createFile(d, f1)
	if err != nil {
		return
	}
	defer func() {
		os.Remove(f1)
	}()
	defer file.Close()
	if length >= 0 {
		n, err = io.CopyN(file, reader, length)
	} else {
		n, err = io.Copy(file, reader)
	}
	if err == nil {
		err = os.Rename(f1, f)
	}
	return
}

type SimpleWriter struct {
	f1, f2 string
	err    error
	writer io.WriteCloser
}

func newSimpleWriter(f1, f2 string, writer io.WriteCloser) *SimpleWriter {
	return &SimpleWriter{f1, f2, nil, writer}
}
func (w *SimpleWriter) SetErr(err error) {
	w.err = err
}
func (w *SimpleWriter) Write(p []byte) (n int, err error) {
	n, err = w.writer.Write(p)
	w.err = err
	return
}
func (w *SimpleWriter) Close() error {
	w.writer.Close()
	if w.err == nil {
		return os.Rename(w.f2, w.f1)
	}
	os.Remove(w.f2)
	return errors.New("failed")
}

func (p SimplePool) GetWriter(xl *xlog.Logger, key string) (writer *SimpleWriter, err error) {

	d, f := p.getPath(p.root, key)
	f1 := f + "." + strconv.FormatInt(time.Now().UnixNano(), 10) + ".tmp"
	file, err := p.createFile(d, f1)
	if err != nil {
		return
	}
	return newSimpleWriter(f, f1, file), nil
}

func (p SimplePool) Get(xl *xlog.Logger, key string, pos int64) (reader io.ReadCloser, length int64, err error) {

	_, path := p.getPath(p.root, key)
	file, err := os.Open(path)
	if err == nil {
		_, err = file.Seek(pos, 0)
		if err != nil {
			file.Close()
			return
		}
		stat, _ := file.Stat()
		return file, stat.Size() - pos, nil
	}

	if os.IsNotExist(err) {
		return nil, 0, syscall.ENOENT
	}
	return
}

func (p SimplePool) Delete(xl *xlog.Logger, key string) (err error) {

	dir, path := p.getPath(p.root, key)
	err = os.Remove(path)
	if err != nil {
		return
	}
	return p.deleteEmptyDir(dir)
}

func (p SimplePool) deleteEmptyDir(dir string) (err error) {
	return
}

type SimpleFileInfo struct {
	Name string
	Size int64
}

func (p SimplePool) Keys() (keys []SimpleFileInfo, err error) {
	file, err := os.Open(p.root)
	if err != nil {
		return
	}
	defer file.Close()

	m := make(map[string][]SimpleFileInfo)
	sum := 0

	infos, err := file.Readdir(-1)
	if err != nil {
		return
	}
	type item struct {
		key  string
		keys []SimpleFileInfo
	}
	ch := make(chan item, 10)
	for _, info := range infos {
		go func(info os.FileInfo) {
			if !info.IsDir() {
				ch <- item{"", nil}
			}
			keys_, err := p.keys1(p.root, info.Name())
			if err != nil {
				ch <- item{"", nil}
			}
			sum2 := 0
			for _, ks := range keys_ {
				sum2 += len(ks)
			}
			keys2 := make([]SimpleFileInfo, sum2)
			off := 0
			for _, ks := range keys_ {
				copy(keys2[off:], ks)
				off += len(ks)
			}
			ch <- item{info.Name(), keys2}
		}(info)
	}
	for i := 0; i < len(infos); i++ {
		item_ := <-ch
		if item_.key == "" {
			continue
		}
		m[item_.key] = item_.keys
		sum += len(item_.keys)
	}
	keys = make([]SimpleFileInfo, sum)
	off := 0
	for _, ks := range m {
		copy(keys[off:], ks)
		off += len(ks)
	}
	return
}

func (p SimplePool) keys1(dir, base string) (keys map[string][]SimpleFileInfo, err error) {
	file, err := os.Open(filepath.Join(dir, base))
	if err != nil {
		return
	}
	defer file.Close()
	infos, err := file.Readdir(-1)
	if err != nil {
		return
	}
	keys = make(map[string][]SimpleFileInfo)
	type item struct {
		key   string
		infos []SimpleFileInfo
	}
	ch := make(chan item, 10)
	for _, info := range infos {
		go func(info os.FileInfo) {
			if !info.IsDir() {
				ch <- item{"", nil}
			}
			keys_, err := p.keys2(filepath.Join(dir, base, info.Name()))
			if err != nil {
				ch <- item{"", nil}
			}
			ch <- item{filepath.Join(base, info.Name()), keys_}
		}(info)
	}
	for i := 0; i < len(infos); i++ {
		item_ := <-ch
		if item_.key == "" {
			continue
		}
		keys[item_.key] = item_.infos
	}
	return
}

func (p SimplePool) keys2(dir string) (keys []SimpleFileInfo, err error) {
	file, err := os.Open(dir)
	if err != nil {
		return
	}
	defer file.Close()
	infos, err := file.Readdir(-1)
	if err != nil {
		return
	}
	keys = make([]SimpleFileInfo, len(infos))
	idx := 0
	for _, info := range infos {
		if filepath.Ext(info.Name()) == ".tmp" {
			continue
		}
		keys[idx] = SimpleFileInfo{info.Name(), info.Size()}
		idx++
	}
	return keys[:idx], nil
}
