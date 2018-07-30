package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncodeFids(t *testing.T) {
	fids := []uint64{2, 3, 4, 5, 6}
	str := EncodeFids(fids)
	assert.Equal(t, str, "AgAAAAAAAAADAAAAAAAAAAQAAAAAAAAABQAAAAAAAAAGAAAAAAAAAA==")
}

func TestDecodeFids(t *testing.T) {
	str := "AgAAAAAAAAADAAAAAAAAAAQAAAAAAAAABQAAAAAAAAAGAAAAAAAAAA=="
	fids, err := DecodeFids(str)
	assert.NoError(t, err)
	assert.Equal(t, fids, []uint64{2, 3, 4, 5, 6})
}
