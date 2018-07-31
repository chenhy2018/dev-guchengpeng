package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDigitString2Bytes(t *testing.T) {
	type testCase struct {
		Str   string
		Bytes []byte
	}

	testCases := []testCase{
		{"123", []byte{1, 2, 3}},
		{"", []byte{}},
		{"1a2b3", []byte{1, 2, 3}},
	}

	for _, c := range testCases {
		assert.Equal(t,
			c.Bytes,
			DigitString2Bytes(c.Str),
		)
	}
}
