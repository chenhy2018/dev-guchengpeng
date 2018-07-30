package qauth

import (
	"github.com/qiniu/log.v1"
	"sync"
	"syscall"
)

type Instance struct {
	key, keyPrev         string
	keyHint, keyHintPrev uint32
	encr, encrPrev       *encryptor
	m                    sync.Mutex
}

func New(keyHint uint32, key string) *Instance {
	encr := newEncryptor(key)
	return &Instance{key: key, keyHint: keyHint, encr: encr}
}

func (r *Instance) UpdateKey(keyHint uint32, key string) (keyHintPrev uint32, keyPrev string) {
	r.m.Lock()
	defer r.m.Unlock()
	r.keyHintPrev, r.keyPrev = r.keyHint, r.key
	r.keyHint, r.key = keyHint, key
	return r.keyHintPrev, r.keyPrev
}

func (r *Instance) Encode(val interface{}) (b []byte, err error) {
	r.m.Lock()
	defer r.m.Unlock()
	return r.encr.encode(val, r.keyHint)
}

func (r *Instance) Decode(val interface{}, b []byte) (err error) {
	if len(b) < 8 {
		return syscall.EINVAL
	}
	keyHint := getKeyHint(b)

	r.m.Lock()
	defer r.m.Unlock()

	var encr *encryptor
	if r.keyHint == keyHint {
		encr = r.encr
	} else if r.encrPrev != nil && r.keyHintPrev == keyHint {
		encr = r.encrPrev
	} else {
		log.Println("invalid keyHint:", keyHint)
		return syscall.EINVAL
	}
	return encr.decode(val, b)
}
