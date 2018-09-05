package dnspodapi

import (
	"bytes"
	"crypto/cipher"
	"crypto/des"
	"encoding/base64"
	"github.com/qiniu/log.v1"
)

func Base64DesEncryt(inputStr, desKey string) (string, error) {
	//<1> 为原始数据进行DES加密
	crypted, err := DesEncrypt([]byte(inputStr), []byte(desKey))
	if err != nil {
		log.Debug(err)
		return "", err
	}
	//这里加密后的串可能不可显
	//<2> 进行base64转码加密，输出可显字串
	// Go支持标准的和兼容URL的base64编码。我们这里使用标准的base64编码，这个函数需要一个`[]byte`参数
	b64cryptedStr := base64.StdEncoding.EncodeToString(crypted)

	return b64cryptedStr, nil
}

func Base64DesDecrypt(inputStr, desKey string) (string, error) {
	//<1> 进行base64转码解密
	// 解码一个base64编码可能返回一个错误，如果你不知道输入是否是正确的base64编码，你需要检测一些解码错误
	b64origData, err := base64.StdEncoding.DecodeString(inputStr)
	if err != nil {
		log.Debug(err)
		return "", err
	}

	//<2> 进行DES解密，解析出原始数据
	origData, err2 := DesDecrypt(b64origData, []byte(desKey))
	if err2 != nil {
		log.Debug(err2)
		return "", err2
	}

	return string(origData), nil
}

//des加密
func DesEncrypt(origData, key []byte) ([]byte, error) {
	block, err := des.NewCipher(key)
	if err != nil {
		return nil, err
	}
	origData = PKCS5Padding(origData, block.BlockSize())
	// origData = ZeroPadding(origData, block.BlockSize())
	blockMode := cipher.NewCBCEncrypter(block, key)
	crypted := make([]byte, len(origData))
	// 根据CryptBlocks方法的说明，如下方式初始化crypted也可以
	// crypted := origData
	blockMode.CryptBlocks(crypted, origData)
	return crypted, nil
}

//des解密
func DesDecrypt(crypted, key []byte) ([]byte, error) {
	block, err := des.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockMode := cipher.NewCBCDecrypter(block, key)
	//origData := make([]byte, len(crypted))
	origData := crypted
	blockMode.CryptBlocks(origData, crypted)
	//origData = PKCS5UnPadding(origData)

	origData = ZeroUnPadding(origData)
	return origData, nil
}

func PKCS5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func ZeroUnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}
