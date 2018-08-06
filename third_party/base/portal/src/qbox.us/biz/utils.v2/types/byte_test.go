package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestByteString(t *testing.T) {
	a := []int64{
		57,
		1024,
		1476,
		1048576,
		1970520,
		1073741824,
		3023152624,
		1099511627776,
		94598095939840,
		1319413953331200,
	}
	e := []string{
		"57.00 B",
		"1.00 KB",
		"1.44 KB",
		"1.00 MB",
		"1.88 MB",
		"1.00 GB",
		"2.82 GB",
		"1.00 TB",
		"86.04 TB",
		"1200.00 TB",
	}
	for index, value := range a {
		actual := Byte(value).String()
		expect := e[index]
		assert.Equal(t, expect, actual)
	}
}
