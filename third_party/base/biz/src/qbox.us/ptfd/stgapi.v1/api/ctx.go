package api

import (
	"encoding/base64"
	"encoding/binary"
	"errors"
)

const (
	CtxSize        = 4 + 4 + 16
	EncodedCtxSize = (CtxSize + 3 - 1) / 3 * 4
)

type PositionCtx struct {
	Max    uint32
	Off    uint32
	Eblock string
}

func EncodePositionCtx(ctx *PositionCtx) string {

	b := make([]byte, CtxSize)
	binary.LittleEndian.PutUint32(b, ctx.Max)
	binary.LittleEndian.PutUint32(b[4:], ctx.Off)
	copy(b[8:], []byte(ctx.Eblock))
	return base64.URLEncoding.EncodeToString(b)
}

func DecodePositionCtx(ctx string) (*PositionCtx, error) {

	b, err := base64.URLEncoding.DecodeString(ctx)
	if err != nil {
		return nil, err
	}
	if len(b) != CtxSize {
		return nil, errors.New("invalid position")
	}
	return &PositionCtx{
		Max:    binary.LittleEndian.Uint32(b),
		Off:    binary.LittleEndian.Uint32(b[4:]),
		Eblock: string(b[8:]),
	}, nil
}
