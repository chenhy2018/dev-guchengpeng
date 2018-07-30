package largefile

import (
	"qbox.us/largefile"
	"sync/atomic"
)

const eor = '\n'

// ----------------------------------------------------------
// NOTE: 这个包已经迁移到 github.com/qiniu/largefile/log

type Logger struct {
	off int64
	f   *largefile.Instance
}

func New(f *largefile.Instance) (r *Logger, err error) {
	fsize, err := f.Size()
	if err != nil {
		return
	}
	return &Logger{fsize, f}, nil
}

func Open(name string, chunkBits uint) (r *Logger, err error) {
	f, err := largefile.Open(name, chunkBits)
	if err != nil {
		return
	}
	return New(f)
}

func (r *Logger) Close() (err error) {
	return r.f.Close()
}

func (r *Logger) Log(msg []byte) {
	msg = append(msg, eor)
	n := int64(len(msg))
	off := atomic.AddInt64(&r.off, n)
	r.f.WriteAt(msg, off-n)
}

// ----------------------------------------------------------
