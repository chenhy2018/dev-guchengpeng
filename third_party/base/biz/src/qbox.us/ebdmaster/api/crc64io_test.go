package api

import (
	"bytes"
	"math/rand"
	"testing"

	"github.com/qiniu/encoding/binary"
	"github.com/stretchr/testify/assert"
	"io"
)

func TestCrc64(t *testing.T) {
	fail(t)
	success(t)
	limitRead(t)
}

func fail(t *testing.T) {
	b := make([]byte, 1)
	rand.Read(b)
	body := append(b, 0, 0, 0, 0, 0, 0, 0, 0)
	r := NewCrc64StreamReader(bytes.NewReader(body))
	b1 := make([]byte, 1)
	err := binary.Read(r, binary.LittleEndian, b1)
	assert.NoError(t, err)
	assert.Equal(t, b, b1)
	var actual uint64
	expect := r.(*Crc64StreamReader).Sum64()
	err = binary.Read(r, binary.LittleEndian, &actual)
	assert.NoError(t, err)
	assert.NotEqual(t, expect, actual)
}

func success(t *testing.T) {
	buf := bytes.NewBuffer(make([]byte, 0))
	w := NewCrc64StreamWriter(buf)
	b := make([]byte, 8)
	rand.Read(b)
	w.Write(b)
	expectCrc64, err := w.(*Crc64StreamWriter).WriteSum64()
	assert.NoError(t, err)
	b1 := make([]byte, 8)
	r := NewCrc64StreamReader(buf)
	_, err = r.Read(b1)
	assert.NoError(t, err)
	assert.Equal(t, b, b1)
	actualCrc64 := r.(*Crc64StreamReader).Sum64()
	assert.Equal(t, expectCrc64, actualCrc64)
}

func limitRead(t *testing.T) {
	buf := bytes.NewBuffer(make([]byte, 0))
	w := NewCrc64StreamWriter(buf)
	b := make([]byte, 100)
	rand.Read(b)
	w.Write(b)
	expectCrc64, err := w.(*Crc64StreamWriter).WriteSum64()
	assert.NoError(t, err)

	r := newReader(buf)
	r = NewCrc64StreamReader(r)

	c := make([]byte, 50)
	n, err := r.Read(c)
	assert.NoError(t, err)
	assert.Equal(t, 20, n)
	c = make([]byte, 80)
	N := 0
	for {
		n, err = r.Read(c[N:])
		assert.NoError(t, err)
		assert.Equal(t, 20, n)
		N += n
		if N == len(c) {
			break
		}
	}
	actualCrc64 := r.(*Crc64StreamReader).Sum64()
	assert.Equal(t, expectCrc64, actualCrc64)
}

type fixSizeReader struct {
	r 	io.Reader
}
func newReader(r io.Reader) io.Reader {
	return &fixSizeReader{r}
}
func (self *fixSizeReader) Read(p []byte) (n int, err error) {
	return io.ReadFull(self.r, p[:20])
}
