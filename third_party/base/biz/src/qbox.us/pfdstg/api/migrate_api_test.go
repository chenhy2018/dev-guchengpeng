package api

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"hash/crc32"
	"io"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"

	"qbox.us/pfd/api/types"
)

func TestMigrateInfo(t *testing.T) {
	m := &MigrateInfo{
		Flag:      1,
		GidCount:  3,
		FidCounts: []uint32{1, 0, 2},
		Gids:      []types.Gid{types.NewGid(1), types.NewGid(2), types.NewGid(3)},
		FidInfos: []FileInfo{
			{Fid: 11, Off: 0, Fsize: 1024},
			{Fid: 31, Off: 0, Fsize: 1024},
			{Fid: 32, Off: 1321, Fsize: 101},
		},
		FirstFidDataBegin: 123,
		LastFidDataEnd:    321,
	}
	b := EncodeMigrateInfoWithCrc(m)

	m2, err := DecodeMigrateInfoWithCrc(bytes.NewReader(b), int64(len(b)))
	assert.NoError(t, err)
	assert.Equal(t, m, m2)
	b2 := EncodeMigrateInfoWithCrc(m2)
	assert.Equal(t, b, b2)
}

func TestAppendCrc32DecodeSuccess(t *testing.T) {
	tcs := []int64{0, 1, 2, 3, 4, 5, 16, 32*1024 - 1, 32 * 1024, 32*1024 + 1, 64 * 1024}
	for _, length := range tcs {
		data := make([]byte, length)
		rand.Read(data)
		crc := crc32.ChecksumIEEE(data)
		crcb := make([]byte, 4)
		binary.LittleEndian.PutUint32(crcb, crc)
		body := append(data, crcb...)

		r, n := appendCrc32Decode(bytes.NewReader(body), int64(len(body)))
		assert.Equal(t, len(data), n)
		b, err := ioutil.ReadAll(r)
		assert.NoError(t, err)
		assert.Equal(t, data, b)

		r2, n2 := appendCrc32Decode(bytes.NewReader(body), int64(len(body)))
		assert.Equal(t, len(data), n2)
		b2 := make([]byte, n2)
		_, err2 := io.ReadFull(r2, b2)
		assert.NoError(t, err2)
		assert.Equal(t, data, b2)
		n3, err3 := r2.Read(make([]byte, 1))
		assert.Equal(t, 0, n3)
		assert.Equal(t, io.EOF, err3)
	}
}

func TestAppendCrc32DecodeFail(t *testing.T) {
	N := int64(1)

	b := make([]byte, N)
	rand.Read(b)
	body := append(b, 0, 0, 0, 0)
	b2 := make([]byte, N)

	var (
		r   io.Reader
		n   int64
		err error
	)

	r, n = appendCrc32Decode(bytes.NewReader(body), int64(len(body)))
	assert.Equal(t, len(b), n)
	defer func() {
		v := recover()
		err = v.(error)
		if assert.Error(t, err) {
			assert.Contains(t, err.Error(), "unmatched crc32")
		}
	}()
	_, err = ioutil.ReadAll(r)

	r, n = appendCrc32Decode(bytes.NewReader(body), int64(len(body)))
	assert.Equal(t, len(b), n)
	_, err = io.ReadFull(r, b2)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(), "unmatched crc32")
	}

	r, n = appendCrc32Decode(io.LimitReader(bytes.NewReader(body), N), int64(len(body)))
	assert.Equal(t, len(b), n)
	_, err = io.ReadFull(r, b2)
	if assert.Error(t, err) {
		assert.Equal(t, io.ErrUnexpectedEOF, err)
	}
}
