package dev

import (
	"fmt"
	"testing"
)

func TestMakeDevid(t *testing.T) {

	s0 := MakeId()
	fmt.Println(s0)

	for i := 0; i < 32; i++ {
		s := MakeId()
		if s != s0 {
			fmt.Println(s)
		}
	}
}
