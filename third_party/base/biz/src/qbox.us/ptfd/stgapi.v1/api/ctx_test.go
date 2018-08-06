package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCtx(t *testing.T) {

	fp := &PositionCtx{
		Max:    1,
		Off:    12345678,
		Eblock: "0123456789012345",
	}

	ctx := EncodePositionCtx(fp)

	_, err := DecodePositionCtx(ctx[1:])
	assert.Error(t, err)

	fp1, err := DecodePositionCtx(ctx)
	assert.NoError(t, err)
	assert.Equal(t, *fp, *fp1)
}
