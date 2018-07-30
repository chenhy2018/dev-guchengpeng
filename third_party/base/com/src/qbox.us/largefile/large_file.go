package largefile

import (
	"os"
	"qbox.us/errors"
	"strconv"
	"sync"
	"sync/atomic"
	"syscall"
)

// --------------------------------------------------------------------

type File struct {
	*os.File
	fno uint32
	ref int32
}

func (r *File) Release() {
	ref := atomic.AddInt32(&r.ref, -1)
	if ref == 0 {
		//log.Debug("FileClose:", r.fno)
		//@bugfix: 因为largefile有可能被作为log的Writer，所以这里不能调用log相关的方法，否则会导致log Mutex重入挂起
		r.File.Close()
	}
}

func (r *File) Acquire() {
	atomic.AddInt32(&r.ref, 1)
}

// --------------------------------------------------------------------
// NOTE: 这个包已经迁移到 github.com/qiniu/largefile

const (
	BucketCount      = 8
	DefaultChunkBits = 22
)

type Instance struct {
	files     [BucketCount]*File
	mutex     sync.RWMutex
	base      string
	chunkBits uint
}

// NOTE: 这个包已经迁移到 github.com/qiniu/largefile
//
func Open(name string, chunkBits uint) (r *Instance, err error) {

	err = syscall.Mkdir(name, 0777)
	if err != nil {
		if err != syscall.EEXIST {
			err = errors.Info(err, "largefile.OpenInstance failed", name).Detail(err)
			return
		}
		err = nil
	}

	if chunkBits > 32 {
		err = errors.Info(errors.EINVAL, "largefile.OpenInstance failed: invalid argument")
		return
	} else if chunkBits == 0 {
		chunkBits = DefaultChunkBits
	}
	return &Instance{base: name + "/", chunkBits: chunkBits}, nil
}

func (r *Instance) Close() (err error) {

	r.mutex.Lock()
	defer r.mutex.Unlock()

	for i, f := range r.files {
		if f != nil {
			f.Release()
		}
		r.files[i] = nil
	}
	return nil
}

func (r *Instance) Size() (fsize int64, err error) {

	n := len(r.base)
	fis, err := readDir(r.base[:n-1])
	if err != nil {
		return
	}

	var ibase, fsize1 int64
	for _, fi := range fis {
		//log.Debug("File:", fi.Name(), fi.Size())
		//@bugfix: 因为largefile有可能被作为log的Writer，所以这里不能调用log相关的方法，否则会导致log Mutex重入挂起
		idx, err2 := strconv.ParseInt(fi.Name(), 36, 64)
		if err2 != nil {
			continue
		}
		if idx >= ibase {
			ibase = idx
			fsize1 = fi.Size()
		}
	}
	return (ibase << r.chunkBits) + fsize1, nil
}

func readDir(dirname string) ([]os.FileInfo, error) {

	f, err := os.Open(dirname)
	if err != nil {
		return nil, err
	}
	list, err := f.Readdir(-1)
	f.Close()
	return list, err
}

func (r *Instance) getFile(off int64) (f *File, offNew int64, sizeLeft int, err error) {

	chunkBits := r.chunkBits
	idx := off >> chunkBits
	fno := uint32(idx)
	bno := fno % BucketCount

	offNew = off - (idx << chunkBits)
	sizeLeft = (1 << chunkBits) - int(offNew)

	r.mutex.RLock()
	f = r.files[bno]
	if f != nil && f.fno == fno {
		f.Acquire()
		r.mutex.RUnlock()
		return
	}
	r.mutex.RUnlock()

	fp, err := os.OpenFile(r.base+strconv.FormatInt(idx, 36), os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return
	}

	f = &File{fp, fno, 2}

	r.mutex.Lock()
	of := r.files[bno]
	r.files[bno] = f
	r.mutex.Unlock()

	if of != nil {
		of.Release()
	}

	return
}

func (r *Instance) Truncate(size int64) (err error) {

	f, sizeNew, _, err := r.getFile(size)
	if err != nil {
		err = errors.Info(err, "largefile.Instance.Truncate failed").Detail(err)
		return
	}
	defer f.Release()

	err = f.Truncate(sizeNew)
	if err != nil {
		err = errors.Info(err, "largefile.Instance.Truncate failed").Detail(err)
		return
	}

	idx := size >> r.chunkBits
	for {
		idx++
		err = syscall.Unlink(r.base + strconv.FormatInt(idx, 36))
		if err != nil {
			if err == syscall.ENOENT {
				err = nil
				break
			}
			err = errors.Info(err, "largefile.Instance.Truncate failed").Detail(err)
			return
		}

		fno := uint32(idx)
		bno := fno % BucketCount

		r.mutex.Lock()
		of := r.files[bno]
		if of != nil && fno == of.fno {
			of.Release()
			r.files[bno] = nil
		}
		r.mutex.Unlock()
	}
	return
}

func (r *Instance) ReadAt(buf []byte, off int64) (n int, err error) {

	f, offNew, sizeLeft, err := r.getFile(off)
	if err != nil {
		err = errors.Info(err, "largefile.Instance.ReadAt failed").Detail(err)
		return
	}
	defer f.Release()

	if len(buf) <= sizeLeft {
		return f.ReadAt(buf, offNew)
	}

	n, err = f.ReadAt(buf[:sizeLeft], offNew)
	if err != nil {
		return
	}

	n2, err := r.ReadAt(buf[sizeLeft:], off+int64(sizeLeft))
	n += n2
	return
}

func (r *Instance) WriteAt(buf []byte, off int64) (n int, err error) {

	f, offNew, sizeLeft, err := r.getFile(off)
	if err != nil {
		err = errors.Info(err, "largefile.Instance.ReadAt failed").Detail(err)
		return
	}
	defer f.Release()

	if len(buf) <= sizeLeft {
		return f.WriteAt(buf, offNew)
	}

	n, err = f.WriteAt(buf[:sizeLeft], offNew)
	if err != nil {
		return
	}

	n2, err := r.WriteAt(buf[sizeLeft:], off+int64(sizeLeft))
	n += n2
	return
}

// --------------------------------------------------------------------
