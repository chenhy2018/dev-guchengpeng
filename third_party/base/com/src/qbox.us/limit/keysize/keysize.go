package keysize

import (
	"io"
	"sync"

	"qbox.us/limit"
)

type KeySizeLimit struct {
	mutex   sync.Mutex
	current map[string]uint64
	limit   uint64
}

func New(n uint64) *KeySizeLimit {
	return &KeySizeLimit{current: make(map[string]uint64), limit: uint64(n)}
}

func (l *KeySizeLimit) Running() uint64 {

	l.mutex.Lock()
	defer l.mutex.Unlock()

	all := uint64(0)
	for _, v := range l.current {
		all += v
	}
	return uint64(all)
}

func (l *KeySizeLimit) Acquire(key2 []byte, value uint64) error {

	key := string(key2)

	l.mutex.Lock()
	defer l.mutex.Unlock()

	n := l.current[key]
	if n+value > l.limit {
		return limit.ErrLimit
	}
	l.current[key] = n + value

	return nil
}

func (l *KeySizeLimit) Release(key2 []byte, value uint64) {

	key := string(key2)

	l.mutex.Lock()
	n := l.current[key]
	if n <= value {
		delete(l.current, key)
	} else {
		l.current[key] = n - value
	}
	l.mutex.Unlock()
}

func (l *KeySizeLimit) AcquireRemains(key2 []byte) (value uint64) {
	key := string(key2)
	l.mutex.Lock()
	defer l.mutex.Unlock()
	n := l.current[key]
	value = l.limit - n
	l.current[key] = l.limit
	return
}

//使用KeySizeLimit进行限制，边写入边申请空间，如果可用空间不足则报错
type KeySizeWriter struct {
	Key      []byte
	Limit    *KeySizeLimit
	Wrc      io.WriteCloser
	C        uint64 //当前已申请字节数
	isClosed bool
	sync.RWMutex
}

func (lw *KeySizeWriter) Write(p []byte) (n int, err error) {
	n, err = lw.Wrc.Write(p)
	if err != nil {
		return
	}
	lw.Lock()
	defer lw.Unlock()
	if lw.isClosed {
		return
	}
	err = lw.Limit.Acquire(lw.Key, uint64(n))
	if err != nil {
		return
	}
	lw.C += uint64(n)
	return
}

func (lw *KeySizeWriter) Close() (err error) {
	lw.Lock()
	defer lw.Unlock()
	if lw.isClosed {
		return
	}
	lw.isClosed = true
	lw.Limit.Release(lw.Key, lw.C)
	return
}
