// 计算文件在七牛云存储上的 qetag 值
package qetag

import (
	"encoding/base64"
	"errors"
	"hash"

	qhash "qbox.us/etag/hash"
)

const Size = 21

type digest struct {
	hash.Hash
	writeSize int64
}

func New() hash.Hash {
	return &digest{qhash.NewWith(1 << 22), 0}
}

// Write (via the embedded io.Writer interface) adds more data to the running hash.
// It never returns an error.
func (h *digest) Write(p []byte) (n int, err error) {
	n, err = h.Hash.Write(p)
	h.writeSize += int64(n)
	return
}

// Sum appends the current hash to b and returns the resulting slice.
// It does not change the underlying hash state.
func (r *digest) Sum(b []byte) (hash []byte) {
	var prefix byte = 0x16
	if r.writeSize > 1<<22 {
		prefix = 0x96
	}
	b = append(b, prefix)
	return r.Hash.Sum(b)
}

// Reset resets the Hash to its initial state.
func (h *digest) Reset() {
	h.Hash.Reset()
	h.writeSize = 0
}

// Size returns the number of bytes Sum will return.
func (h *digest) Size() int {
	return Size
}

type Qetag []byte

// Sum returns the qetag checksum of the data.
func Sum(data []byte) Qetag {
	d := New()
	d.Write(data)
	return d.Sum(nil)
}

// Encode returns the string of Sum result
func (e Qetag) String() string {
	return base64.URLEncoding.EncodeToString(e)
}

func ParseQetag(e string) (qetag Qetag, err error) {
	qetag, err = base64.URLEncoding.DecodeString(e)
	if err != nil {
		return
	}
	if len(qetag) != Size {
		qetag, err = nil, errors.New("qetag size not match")
		return
	}
	if qetag[0] != 0x16 && qetag[0] != 0x96 {
		qetag, err = nil, errors.New("invalid first byte of qetag")
		return
	}
	return
}
