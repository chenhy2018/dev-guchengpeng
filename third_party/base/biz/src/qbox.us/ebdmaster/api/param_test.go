package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPsectsEncoding(t *testing.T) {

	psects := [N + M]uint64{1, 2, 1234567890}
	strPsects := EncodePsects(&psects)

	_, err := DecodePsects(strPsects[1:])
	assert.Error(t, err)

	_, err = DecodePsects(strPsects[4:])
	assert.Equal(t, ErrInvalidLength, err)

	news, err := DecodePsects(strPsects)
	assert.NoError(t, err)
	assert.Equal(t, psects, *news)
}

func TestCrc32sEncoding(t *testing.T) {

	crc32s := [N + M]uint32{1, 2, 12345678}
	strCrc32s := EncodeCrc32s(&crc32s)

	_, err := DecodeCrc32s(strCrc32s[1:])
	assert.Error(t, err)

	_, err = DecodeCrc32s(strCrc32s[4:])
	assert.Equal(t, ErrInvalidLength, err)

	news, err := DecodeCrc32s(strCrc32s)
	assert.NoError(t, err)
	assert.Equal(t, crc32s, *news)
}

func TestBadisEncoding(t *testing.T) {

	badis := []BadInfo{
		{1, 10},
		{2, 20},
	}
	strBadis := EncodeBadis(badis)

	_, err := DecodeBadis(strBadis[1:])
	assert.Error(t, err)

	_, err = DecodeBadis(strBadis[4:])
	assert.Equal(t, ErrInvalidLength, err)

	news, err := DecodeBadis(strBadis)
	assert.NoError(t, err)
	assert.Equal(t, badis, news)
}

func TestBadsEncoding(t *testing.T) {

	bads := [M]int8{1, 2, 3, 4}
	strBads := EncodeBads(bads)

	_, err := DecodeBads(strBads[1:])
	assert.Error(t, err)

	_, err = DecodeBads(strBads[4:])
	assert.Equal(t, ErrInvalidLength, err)

	news, err := DecodeBads(strBads)
	assert.NoError(t, err)
	assert.Equal(t, bads, news)
}
