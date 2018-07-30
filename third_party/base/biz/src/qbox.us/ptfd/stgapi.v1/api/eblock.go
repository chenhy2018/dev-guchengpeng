package api

import (
	"encoding/base64"
	"encoding/binary"
	"errors"
)

const (
	addrSize   = 12
	EblockSize = (addrSize + 3 - 1) / 3 * 4

	DgRoundPrefixSize = 10
)

var (
	ZeroEblock = EncodeEblock(&BlockAddr{})
)

type BlockAddr struct {
	Dgid  uint32
	Round uint32
	Tidx  uint32
}

func EncodeEblock(addr *BlockAddr) string {
	b := make([]byte, addrSize)
	binary.LittleEndian.PutUint32(b, addr.Dgid)
	binary.LittleEndian.PutUint32(b[4:], addr.Round)
	binary.LittleEndian.PutUint32(b[8:], addr.Tidx)
	return base64.URLEncoding.EncodeToString(b)
}

func DecodeEblock(eblock string) (*BlockAddr, error) {

	b, err := base64.URLEncoding.DecodeString(eblock)
	if err != nil {
		return nil, err
	}
	if len(b) != addrSize {
		return nil, errors.New("invalid eblock")
	}
	var addr BlockAddr
	addr.Dgid = binary.LittleEndian.Uint32(b)
	addr.Round = binary.LittleEndian.Uint32(b[4:])
	addr.Tidx = binary.LittleEndian.Uint32(b[8:])
	return &addr, nil
}

// dgid 4B + round 4B = 64b, choose first 60b as prefix
// eblock[:10] represents 60b, still lost 4b info
// Round [0x?0???_????, 0x?F???_????] have the same prefix
func EncodeDgRoundPrefix(dgid, round uint32) string {
	var addr BlockAddr
	addr.Dgid = dgid
	addr.Round = round
	return EncodeEblock(&addr)[:DgRoundPrefixSize]
}
