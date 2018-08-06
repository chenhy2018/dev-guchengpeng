package dev

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"log"
	"math/big"
	"net"
	"syscall"
)

func getMaxInt() *big.Int {
	x := big.NewInt(1)
	return x.Lsh(x, 168) // 21 bytes * 8 = 168
}

var maxInt = getMaxInt()

func MakeRandId() string {
	var b [21]byte
	x, _ := rand.Int(rand.Reader, maxInt)
	copy(b[:], x.Bytes())
	return base64.URLEncoding.EncodeToString(b[:])
}

func GetNicSerialNumber() (sn []byte, err error) {
	itfs, err := net.Interfaces()
	if err != nil {
		log.Println("net.Interfaces failed:", err)
		return
	}
	for _, itf := range itfs {
		sn = itf.HardwareAddr
		if len(sn) != 0 {
			return
		}
	}
	log.Println("GetNicSerialNumber failed: not found")
	return nil, syscall.ENOENT
}

func MakeId() string {

	sn, err := GetNicSerialNumber()
	if err == nil {
		h := sha1.New()
		h.Write(sn)
		return base64.URLEncoding.EncodeToString(h.Sum(nil))
	}
	return MakeRandId()
}
