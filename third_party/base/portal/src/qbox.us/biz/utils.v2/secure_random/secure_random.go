package secure_random

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
)

func Hex(n int) string {
	return hex.EncodeToString(ByteArray(n))
}

func Base64(n int) string {
	return base64.StdEncoding.EncodeToString(ByteArray(n))
}

func UrlsafeBase64(n int) string {
	return base64.URLEncoding.EncodeToString(ByteArray(n))
}

func ByteArray(n int) []byte {
	b := make([]byte, n)
	n2, err := io.ReadFull(rand.Reader, b)
	if err != nil {
		panic("io.ReadFull from rand.Reader failed: " + err.Error())
	}
	if n2 != n {
		panic(fmt.Sprintf("io.ReadFull from rand.Reader failed: expect %d bytes, got %d bytes", n, n2))
	}
	return b
}
