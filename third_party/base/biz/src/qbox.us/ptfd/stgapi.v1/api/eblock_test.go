package api

import (
	"encoding/base64"
	"math"
	"strings"
	"testing"
	"time"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

func TestEblock(t *testing.T) {

	addr := &BlockAddr{
		Dgid:  1,
		Round: 2,
		Tidx:  3,
	}
	eblock := EncodeEblock(addr)

	// check compatibility
	b := ((*[1 << 27]byte)(unsafe.Pointer(addr)))[:addrSize]
	eblock1 := base64.URLEncoding.EncodeToString(b)
	assert.Equal(t, eblock, eblock1)

	_, err := DecodeEblock(eblock[1:])
	assert.Error(t, err)

	addr1, err := DecodeEblock(eblock)
	assert.NoError(t, err)
	assert.Equal(t, *addr, *addr1)

	// check compatibility
	var addr2 BlockAddr
	b, err = base64.URLEncoding.DecodeString(eblock)
	assert.NoError(t, err)
	copy(((*[1 << 27]byte)(unsafe.Pointer(&addr2)))[:], b)
	assert.Equal(t, addr1, &addr2)

	addr = &BlockAddr{
		Dgid:  132,
		Round: uint32(time.Now().Unix()),
		Tidx:  4096 * 13,
	}
	eblock = EncodeEblock(addr)

	prefix1 := EncodeDgRoundPrefix(addr.Dgid, addr.Round)
	assert.Equal(t, DgRoundPrefixSize, len(prefix1))
	assert.True(t, strings.HasPrefix(eblock, prefix1))

	addr.Tidx = 0
	eblock = EncodeEblock(addr)
	assert.True(t, strings.HasPrefix(eblock, prefix1))

	addr.Tidx = math.MaxUint32
	assert.True(t, strings.HasPrefix(eblock, prefix1))

	prefix4 := EncodeDgRoundPrefix(addr.Dgid, 0xF0FFffff)
	assert.NotEqual(t, prefix1, prefix4)

	// have same prefix
	for i := 1; i <= 0xf; i++ {
		prefix5 := EncodeDgRoundPrefix(addr.Dgid, 0xF0FFffff|(uint32(i)<<24))
		assert.Equal(t, prefix4, prefix5)
	}

	prefix6 := EncodeDgRoundPrefix(678, addr.Round)
	assert.NotEqual(t, prefix1, prefix6)
	assert.NotEqual(t, prefix4, prefix6)
}
