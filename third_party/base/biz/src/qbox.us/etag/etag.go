package etag

import (
	"hash"

	qhash "qbox.us/etag/hash"
)

type Digest struct {
	hash.Hash
	chunkSize int
}

func New(chunkSize int) *Digest {

	return &Digest{qhash.NewWith(chunkSize), chunkSize}
}

func (r *Digest) Sum() []byte {

	return r.Hash.Sum(nil)
}

func (r *Digest) ChunkSize() int {
	return r.chunkSize
}
