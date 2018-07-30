package store

import (
	"bytes"
	"github.com/qiniu/xlog.v1"
	"testing"
)

func Get(p Getter, key []byte, val []byte) error {
	b := bytes.NewBuffer(nil)
	err := p.Get(xlog.NewDummy(), key, b, 0, len(val), [4]uint16{0, 0xffff})
	if err == nil {
		copy(val, b.Bytes())
	}
	return err
}

func Put(p Putter, key []byte, val []byte) error {
	b := bytes.NewBuffer(val)
	return p.Put(xlog.NewDummy(), key, b, len(val), [3]uint16{0, 0xffff})
}

func TestSimpleStorage(t *testing.T) {
}
