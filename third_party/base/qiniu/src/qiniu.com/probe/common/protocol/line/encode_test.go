package line

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify.v1/assert"
)

func TestEscape(t *testing.T) {

	assert.Equal(t,
		[]byte(`1234567890abcdefghijklmnopgrstuvwxyz`),
		escape([]byte(`1234567890abcdefghijklmnopgrstuvwxyz`),
			escapeCodeK, escapeCodeV),
		"escape normal")

	assert.Equal(t,
		[]byte(`\ \,\=\"`),
		escape([]byte(` ,="`),
			escapeCodeK, escapeCodeV),
		"escape all")

	assert.Equal(t,
		[]byte(`x\ x\,x\=x\"`),
		escape([]byte(`x x,x=x"`),
			escapeCodeK, escapeCodeV),
		"escape half")
}

func BenchmarkEscape1(b *testing.B) {
	for i := 0; i < b.N; i++ {
		escape(
			[]byte(`sfwefjloj"wef=fsf,sfwe sfwefwegregr`),
			escapeCodeK, escapeCodeV)
	}
}

func BenchmarkEscape2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		in := []byte(`sfwefjloj"wef=fsf,sfwe sfwefwegregr`)
		for b, esc := range escapeCodes {
			in = bytes.Replace(in, []byte{b}, esc, -1)
		}
	}
}
