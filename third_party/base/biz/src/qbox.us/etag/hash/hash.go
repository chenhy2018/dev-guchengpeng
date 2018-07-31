/* 计算文件在七牛云存储上的 hash 值

算法大体如下：

    如果你能够确认文件 <= 4M，那么 hash = UrlsafeBase64([0x16, sha1(FileContent)])。也就是，文件的内容的sha1值（20个字节），前面加一个byte（值为0x16），构成 21 字节的二进制数据，然后对这 21 字节的数据做 urlsafe 的 base64 编码。
    如果文件 > 4M，则 hash = UrlsafeBase64([0x96, sha1([sha1(Block1), sha1(Block2), ...])])，其中 Block 是把文件内容切分为 4M 为单位的一个个块，也就是 BlockI = FileContent[I*4M:(I+1)*4M]。
*/
package hash

import (
	"crypto/sha1"
	"hash"
)

type qiniuHash struct {
	blockSha1 hash.Hash
	blockLeft int
	blockSize int

	sums [][]byte
}

func New() hash.Hash {
	return NewWith(1 << 22)
}

func NewWith(blockSize int) hash.Hash {
	h := new(qiniuHash)
	h.blockSha1 = sha1.New()
	h.blockSize = blockSize
	h.Reset()
	return h
}

// Write (via the embedded io.Writer interface) adds more data to the running hash.
// It never returns an error.
func (h *qiniuHash) Write(p []byte) (n int, err error) {
	for len(p) >= h.blockLeft {
		n2, _ := h.blockSha1.Write(p[:h.blockLeft])
		n += n2

		p = p[h.blockLeft:]

		h.sums = append(h.sums, h.blockSha1.Sum(nil))
		h.blockSha1.Reset()
		h.blockLeft = h.blockSize
	}
	if len(p) > 0 {
		n2, _ := h.blockSha1.Write(p)
		n += n2

		h.blockLeft -= len(p)
	}
	return
}

// Sum appends the current hash to b and returns the resulting slice.
// It does not change the underlying hash state.
func (h *qiniuHash) Sum(b []byte) (hash []byte) {
	if len(h.sums) == 0 { // < blockSize
		return h.blockSha1.Sum(b)
	} else if len(h.sums) == 1 && h.blockLeft == h.blockSize { // == blockSize
		return append(b, h.sums[0]...)
	}

	// > 4M
	h2 := sha1.New()
	for _, sha := range h.sums {
		h2.Write(sha)
	}
	if h.blockLeft < h.blockSize {
		sha := h.blockSha1.Sum(nil)
		h2.Write(sha)
	}
	return h2.Sum(b)
}

// Reset resets the Hash to its initial state.
func (h *qiniuHash) Reset() {
	h.blockSha1.Reset()
	h.blockLeft = h.blockSize
	h.sums = nil
}

// Size returns the number of bytes Sum will return.
func (h *qiniuHash) Size() int {
	return sha1.Size
}

// BlockSize returns the hash's underlying block size.
// The Write method must be able to accept any amount
// of data, but it may operate more efficiently if all writes
// are a multiple of the block size.
func (h *qiniuHash) BlockSize() int {
	return sha1.BlockSize
}
