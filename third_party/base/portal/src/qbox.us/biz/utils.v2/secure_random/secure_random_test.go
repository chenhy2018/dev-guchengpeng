package secure_random

import (
	"testing"
)

import (
	"github.com/stretchr/testify/assert"
)

func TestHex(t *testing.T) {
	assert.Equal(t, len(Hex(12)), 24)
}

func TestBase64(t *testing.T) {
	assert.Equal(t, len(Base64(12)), 16)
	assert.Equal(t, len(Base64(13)), 20)
}

func TestUrlsafeBase64(t *testing.T) {
	assert.Equal(t, len(UrlsafeBase64(12)), 16)
	assert.Equal(t, len(UrlsafeBase64(13)), 20)
}

func BenchmarkHex(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Hex(12)
	}
}
func BenchmarkBase64(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Base64(12)
	}
}
func BenchmarkUrlsafeBase64(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = UrlsafeBase64(12)
	}
}
