package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSuid(t *testing.T) {

	suid := EncodeSuid(1000031312, 10)
	sid, idx := DecodeSuid(suid)
	assert.Equal(t, 1000031312, sid)
	assert.Equal(t, 10, idx)
}
