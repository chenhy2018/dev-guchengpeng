package iomc

import (
	"testing"

	"github.com/stretchr/testify.v2/assert"
)

func TestMessage(t *testing.T) {
	msg := Message{
		Uid:        1,
		Bucket:     "ABC",
		Key:        "ABCD",
		Itbl:       123,
		Op:         2,
		Fh:         []byte{1, 2},
		PutTime:    111,
		ChangeTime: 222,
		IoIgnored:  true,
	}
	data, err := Encode(&msg)
	assert.Equal(t, err, nil)
	msgN, err := Decode(data)
	assert.Equal(t, err, nil)
	assert.Equal(t, msgN.Uid, msg.Uid)
	assert.Equal(t, msgN.Bucket, msg.Bucket)
	assert.Equal(t, msgN.Key, msg.Key)
	assert.Equal(t, msgN.Itbl, msg.Itbl)
	assert.Equal(t, msgN.Op, msg.Op)
	assert.Equal(t, msgN.Fh, msg.Fh)
	assert.Equal(t, msgN.PutTime, msg.PutTime)
	assert.Equal(t, msgN.ChangeTime, msg.ChangeTime)
	assert.Equal(t, msgN.IoIgnored, msg.IoIgnored)
}
