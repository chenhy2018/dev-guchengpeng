package hash

import (
	"bytes"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"io"
	"log"
	"testing"
	"testing/iotest"
)

const (
	BLOCK_BITS = 22
	BLOCK_SIZE = 1 << BLOCK_BITS
)

func blockCount(fsize int64) int {
	return int((fsize + (BLOCK_SIZE - 1)) >> BLOCK_BITS)
}

func calSha1(b []byte, r io.Reader) ([]byte, error) {
	h := sha1.New()
	_, err := io.Copy(h, r)
	if err != nil {
		return nil, err
	}
	return h.Sum(b), nil
}

func getHash(f io.Reader, fsize int64) (sum []byte, err error) {

	blockCnt := blockCount(fsize)
	sha1Buf := make([]byte, 0, 20)
	if blockCnt <= 1 { // file size <= 4M
		sha1Buf, err = calSha1(sha1Buf, f)
		if err != nil {
			return
		}
	} else { // file size > 4M
		sha1BlockBuf := make([]byte, 0, blockCnt*20)
		for i := 0; i < blockCnt; i++ {
			body := io.LimitReader(f, BLOCK_SIZE)
			sha1BlockBuf, err = calSha1(sha1BlockBuf, body)
			if err != nil {
				return
			}
		}
		sha1Buf, _ = calSha1(sha1Buf, bytes.NewReader(sha1BlockBuf))
	}
	return sha1Buf, nil
}

func TestHash(t *testing.T) {

	sizes := []int{
		0,
		1,
		BLOCK_SIZE - 1,
		BLOCK_SIZE,
		BLOCK_SIZE + 1,
		BLOCK_SIZE + BLOCK_SIZE - 1,
		BLOCK_SIZE + BLOCK_SIZE,
		BLOCK_SIZE + BLOCK_SIZE + 1,
	}

	for _, size := range sizes {
		data := make([]byte, size)
		io.ReadFull(rand.Reader, data)

		a, _ := getHash(bytes.NewReader(data), int64(len(data)))
		log.Printf("size: %v\thash: %v\n", size, base64.URLEncoding.EncodeToString(a))

		h := New()
		{
			n, _ := h.Write(data)
			b := h.Sum(nil)
			if !bytes.Equal(a, b) {
				t.Errorf("sumA != sumB, len(sumA): %v, len(sumB): %v, a: %v, b: %v, size: %v", len(a), len(b), base64.URLEncoding.EncodeToString(a), base64.URLEncoding.EncodeToString(b), size)
			}
			if n != len(data) {
				t.Error("h.Size() != len(data)", h.Size(), len(data))
			}
		}
		{
			h.Reset()
			n, _ := io.Copy(h, bytes.NewReader(data))
			b := h.Sum(nil)
			if !bytes.Equal(a, b) {
				t.Errorf("sumA != sumB, len(sumA): %v, len(sumB): %v, a: %v, b: %v, size: %v", len(a), len(b), base64.URLEncoding.EncodeToString(a), base64.URLEncoding.EncodeToString(b), size)
			}
			if int(n) != len(data) {
				t.Error("h.Size() != len(data)", h.Size(), len(data))
			}
		}
		{
			h.Reset()
			n, _ := io.Copy(h, iotest.OneByteReader(bytes.NewReader(data)))
			b := h.Sum(nil)
			if !bytes.Equal(a, b) {
				t.Errorf("sumA != sumB, len(sumA): %v, len(sumB): %v, a: %v, b: %v, size: %v", len(a), len(b), base64.URLEncoding.EncodeToString(a), base64.URLEncoding.EncodeToString(b), size)
			}
			if int(n) != len(data) {
				t.Error("h.Size() != len(data)", h.Size(), len(data))
			}
		}
		{
			h.Reset()
			n, _ := io.Copy(h, iotest.HalfReader(bytes.NewReader(data)))
			b := h.Sum(nil)
			if !bytes.Equal(a, b) {
				t.Errorf("sumA != sumB, len(sumA): %v, len(sumB): %v, a: %v, b: %v, size: %v", len(a), len(b), base64.URLEncoding.EncodeToString(a), base64.URLEncoding.EncodeToString(b), size)
			}
			if int(n) != len(data) {
				t.Error("h.Size() != len(data)", h.Size(), len(data))
			}
		}
		{
			h.Reset()
			n, _ := h.Write(data)
			b := h.Sum([]byte("xxx"))
			a := append([]byte("xxx"), a...)
			if !bytes.Equal(a, b) {
				t.Errorf("sumA != sumB, len(sumA): %v, len(sumB): %v, a: %v, b: %v, size: %v", len(a), len(b), base64.URLEncoding.EncodeToString(a), base64.URLEncoding.EncodeToString(b), size)
			}
			if n != len(data) {
				t.Error("h.Size() != len(data)", h.Size(), len(data))
			}
		}
	}
}
