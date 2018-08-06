package types

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGidEncodeDecode(t *testing.T) {
	gid := NewGid(3)
	assert.True(t, time.Now().UnixNano()-gid.UnixTime() < int64(time.Millisecond))
	assert.Equal(t, gid.MotherDgid(), uint32(3))

	egid := EncodeGid(gid)
	gid2, err := DecodeGid(egid)
	assert.Nil(t, err)
	assert.Equal(t, gid2, gid)

	egid2 := EncodeGid(gid2)
	assert.Equal(t, egid2, egid)
}

func TestGidMarshalUnmarshal(t *testing.T) {
	gid := NewGid(4)
	b, err := gid.MarshalJSON()
	assert.NoError(t, err)
	assert.Equal(t, `"`+EncodeGid(gid)+`"`, string(b))

	var gid2 Gid
	err = gid2.UnmarshalJSON(b)
	assert.NoError(t, err)
	assert.Equal(t, gid, gid2)
}
