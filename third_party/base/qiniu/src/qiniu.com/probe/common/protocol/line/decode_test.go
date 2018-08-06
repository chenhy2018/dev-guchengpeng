package line

import (
	"testing"

	"github.com/stretchr/testify.v1/assert"
)

func TestParseMeasurement(t *testing.T) {

	f := func(in string) string {
		_, out := ParseMeasurement([]byte(in))
		return string(out)
	}

	assert.Equal(t, "abcdefg", f("abcdefg "), "")
	assert.Equal(t, "abcdefg", f("abcdefg,ss "), "")
	assert.Equal(t, `abcde,fg`, f(`abcde\,fg,ss `), "")
	assert.Equal(t, `abcde fg`, f(`abcde\ fg,ss `), "")
	assert.Equal(t, `abcde fg`, f(`abcde\ fg,ss `), "")
	assert.Equal(t, `abcde\ fg`, f(`abcde\\ fg `), "")
}
