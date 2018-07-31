package qauth

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"
	"encoding/binary"
	"encoding/gob"
	"encoding/hex"
	"errors"
	"fmt"
	"qbox.us/cc/time"
	"syscall"
)

type encryptor struct {
	key []byte
	iv  []byte
}

func newEncryptor(saltKey string) *encryptor {
	// sha1 sums are 20 bytes long.  we use the first 16 bytes as
	// the aes key, and the last 16 bytes as the initialization
	// vector (understanding that they overlap, of course).
	keySha1 := sha1.New()
	keySha1.Write([]byte(saltKey))
	sum := keySha1.Sum(nil)
	return &encryptor{
		key: sum[:16],
		iv:  sum[4:],
	}
}

func (en *encryptor) encode(val interface{}, keyHint uint32) ([]byte, error) {

	var buffer [12]byte
	buf := bytes.NewBuffer(buffer[:])
	enc := gob.NewEncoder(buf)
	err := enc.Encode(val)
	if err != nil {
		return nil, err
	}

	key, iv := en.key, en.iv
	aesCipher, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	ns := time.Nanoseconds()
	rand := uint32(ns) ^ uint32(ns>>32)
	encodedGob := buf.Bytes()
	encodedLen := len(encodedGob)
	binary.LittleEndian.PutUint32(encodedGob, keyHint^rand)
	binary.LittleEndian.PutUint32(encodedGob[4:], rand)
	binary.LittleEndian.PutUint32(encodedGob[8:], uint32(encodedLen))

	padLen := aes.BlockSize - (encodedLen-8)%aes.BlockSize
	for padLen > 12 {
		buf.Write(buffer[:])
		padLen -= 12
	}
	buf.Write(buffer[:padLen])

	encodedBytes := buf.Bytes()
	//	fmt.Println("encodedBytes:", encodedBytes, len(encodedBytes) - 8, aes.BlockSize)
	encrypter := cipher.NewCBCEncrypter(aesCipher, iv)
	for i := 8; i < len(encodedBytes); i += 8 {
		for j := 0; j < 8; j++ {
			encodedBytes[i+j] ^= encodedBytes[j]
		}
	}
	encrypter.CryptBlocks(encodedBytes[8:], encodedBytes[8:])

	fmt.Println(hex.Dump(encodedBytes))
	return encodedBytes, nil
}

func getKeyHint(encodedBytes []byte) uint32 {
	keyHint := binary.LittleEndian.Uint32(encodedBytes)
	rand := binary.LittleEndian.Uint32(encodedBytes[4:])
	return keyHint ^ rand
}

var errDecode = errors.New("decode failed")

func (en *encryptor) decode(val interface{}, encodedBytes []byte) error {

	if (len(encodedBytes)-8)%aes.BlockSize != 0 {
		return syscall.EINVAL
	}

	key, iv := en.key, en.iv
	aesCipher, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	// decrypt in-place
	decrypter := cipher.NewCBCDecrypter(aesCipher, iv)
	decrypter.CryptBlocks(encodedBytes[8:], encodedBytes[8:])

	for i := 8; i < len(encodedBytes); i += 8 {
		for j := 0; j < 8; j++ {
			encodedBytes[i+j] ^= encodedBytes[j]
		}
	}

	encodedLen := binary.LittleEndian.Uint32(encodedBytes[8:])
	if encodedLen > uint32(len(encodedBytes)) {
		return errDecode
	}

	buf := bytes.NewBuffer(encodedBytes[12:encodedLen])
	dec := gob.NewDecoder(buf)
	return dec.Decode(val)
}
