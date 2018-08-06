package entity

import (
	"crypto/rand"
	"encoding/base64"
	"io"
	"strings"
)

const iamKeyPrefix = "IAM-"

// KeyLen Access Key 和 Secret Key 的字节数
const KeyLen = 33

// IsIAMKey 判断一个key是否为IAM的key
func IsIAMKey(key string) bool {
	return len(key) == KeyLen*4/3 &&
		strings.HasPrefix(key, iamKeyPrefix)
}

func makeKey(n int) string {
	if n < 1 {
		return ""
	}
	b := make([]byte, n)
	io.ReadFull(rand.Reader, b)
	return base64.URLEncoding.EncodeToString(b)
}

// MakeAccessKey 生成 Access Key
func MakeAccessKey() string {
	return iamKeyPrefix + makeKey(KeyLen-3)
}

// MakeSecretKey 生成 Secret Key
func MakeSecretKey() string {
	return makeKey(KeyLen)
}
