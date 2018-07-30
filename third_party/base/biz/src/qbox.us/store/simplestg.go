package store

import (
	"bytes"
	"io"
	ioadmin "qbox.us/admin_api/io"
	"qbox.us/cc/sha1"
	"github.com/qiniu/xlog.v1"
	"syscall"
)

type SimpleStorage struct {
	stg map[string][]byte
}

func NewSimpleStorage() *SimpleStorage {
	stg := make(map[string][]byte)
	return &SimpleStorage{stg}
}

// 要求在 to 参数超出范围时不认为是错误
func (p *SimpleStorage) Get(xl *xlog.Logger, key []byte, w io.Writer, from, to int, bds [4]uint16) error {
	if v, ok := p.stg[string(key)]; ok {
		n := len(v)
		if from > to {
			return syscall.EINVAL
		}
		if to > n {
			to = n
		}
		_, err := w.Write(v[from:to])
		return err
	}
	return syscall.ENOENT
}

func (p *SimpleStorage) Put(xl *xlog.Logger, key []byte, r io.Reader, n int) error {
	val := make([]byte, n)
	n2, err := io.ReadFull(r, val)
	if n != n2 || err != nil || !bytes.Equal(key, sha1.Hash(val)) {
		return EVerifyFailed
	}
	p.stg[string(key)] = val
	return nil
}

/*func (p *SimpleStorage) Put(r io.Reader, n int) (key []byte, err os.Error) {
	val := make([]byte, n)
	n2, err := r.Read(val)
	if err != nil { return }
	if n != n2 { err = os.EIO; return }
	key = sha1.Hash(val)
	p.stg[string(key)] = val
	return key, nil
}*/

func NewSimpleStorage2() *SimpleStorage2 {
	stg := NewSimpleStorage()
	return &SimpleStorage2{stg}
}

type SimpleStorage2 struct {
	S *SimpleStorage
}

func (p *SimpleStorage2) ServiceStat() (info []*ioadmin.CacheInfoEx, err error) {
	return
}

func (p *SimpleStorage2) Get(xl *xlog.Logger, key []byte, w io.Writer, from, to int, bds [4]uint16) error {
	return p.S.Get(xl, key, w, from, to, bds)
}

func (p *SimpleStorage2) Put(xl *xlog.Logger, key []byte, r io.Reader, n int, doCache bool, bds [3]uint16) error {
	return p.S.Put(xl, key, r, n)
}
