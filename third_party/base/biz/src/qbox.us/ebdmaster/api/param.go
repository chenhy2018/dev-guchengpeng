package api

import (
	"encoding/base64"
	"encoding/binary"
	"errors"
)

var (
	ErrInvalidLength = errors.New("invalid length")
)

// -----------------------------------------------------------------------------

func EncodeSuids(suids *[N + M]uint64) string {

	n := len(suids) * 8
	b := make([]byte, n)
	for i, suid := range suids {
		binary.LittleEndian.PutUint64(b[i*8:], suid)
	}
	return base64.URLEncoding.EncodeToString(b)
}

func DecodeSuids(strSuids string) (*[N + M]uint64, error) {

	b, err := base64.URLEncoding.DecodeString(strSuids)
	if err != nil {
		return nil, err
	}
	if len(b) != (M+N)*8 {
		return nil, ErrInvalidLength
	}
	var psuids [N + M]uint64
	for i := 0; i < N+M; i++ {
		psuids[i] = binary.LittleEndian.Uint64(b[i*8:])
	}
	return &psuids, nil
}

// -----------------------------------------------------------------------------

func EncodePsects(psects *[N + M]uint64) string {

	n := len(psects) * 8
	b := make([]byte, n)
	for i, psect := range psects {
		binary.LittleEndian.PutUint64(b[i*8:], psect)
	}
	return base64.URLEncoding.EncodeToString(b)
}

func DecodePsects(strPsects string) (*[N + M]uint64, error) {

	b, err := base64.URLEncoding.DecodeString(strPsects)
	if err != nil {
		return nil, err
	}
	if len(b) != (M+N)*8 {
		return nil, ErrInvalidLength
	}
	var psects [N + M]uint64
	for i := 0; i < N+M; i++ {
		psects[i] = binary.LittleEndian.Uint64(b[i*8:])
	}
	return &psects, nil
}

// -----------------------------------------------------------------------------

func EncodeCrc32s(crc32s *[N + M]uint32) string {

	n := len(crc32s) * 4
	b := make([]byte, n)
	for i, crc32 := range crc32s {
		binary.LittleEndian.PutUint32(b[i*4:], crc32)
	}
	return base64.URLEncoding.EncodeToString(b)
}

func DecodeCrc32s(strCrc32s string) (*[N + M]uint32, error) {

	b, err := base64.URLEncoding.DecodeString(strCrc32s)
	if err != nil {
		return nil, err
	}
	if len(b) != (M+N)*4 {
		return nil, ErrInvalidLength
	}
	var crc32s [N + M]uint32
	for i := 0; i < N+M; i++ {
		crc32s[i] = binary.LittleEndian.Uint32(b[i*4:])
	}
	return &crc32s, nil
}

// -----------------------------------------------------------------------------

func EncodeBadis(badis []BadInfo) string {

	n := len(badis) * 8
	b := make([]byte, n)
	for i, info := range badis {
		binary.LittleEndian.PutUint32(b[i*8:], info.Idx)
		binary.LittleEndian.PutUint32(b[i*8+4:], info.Reason)
	}
	return base64.URLEncoding.EncodeToString(b)
}

func DecodeBadis(strBadis string) ([]BadInfo, error) {

	b, err := base64.URLEncoding.DecodeString(strBadis)
	if err != nil {
		return nil, err
	}
	if len(b)%8 != 0 {
		return nil, ErrInvalidLength
	}
	n := len(b) / 8
	badis := make([]BadInfo, n)
	for i := 0; i < n; i++ {
		badis[i] = BadInfo{
			Idx:    binary.LittleEndian.Uint32(b[i*8:]),
			Reason: binary.LittleEndian.Uint32(b[i*8+4:]),
		}
	}
	return badis, nil
}

// -----------------------------------------------------------------------------

func EncodeBads(bads [M]int8) string {

	b := make([]byte, len(bads))
	for i, bad := range bads {
		b[i] = byte(bad)
	}
	return base64.URLEncoding.EncodeToString(b)
}

func DecodeBads(strBads string) ([M]int8, error) {

	b, err := base64.URLEncoding.DecodeString(strBads)
	if err != nil {
		return [M]int8{}, err
	}
	if len(b) != M {
		return [M]int8{}, ErrInvalidLength
	}
	var bads [M]int8
	for i, v := range b {
		bads[i] = int8(v)
	}
	return bads, nil
}
