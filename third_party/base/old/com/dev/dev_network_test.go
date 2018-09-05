package dev

import (
	"encoding/hex"
	"fmt"
	"testing"
)

func TestGetNicSerialNumber(t *testing.T) {

	sn, err := GetNicSerialNumber()
	if err != nil {
		t.Fatal(err)
	}
	s := hex.EncodeToString(sn)
	fmt.Println(s)
}
