package rpc

import (
	//	"fmt"
	"testing"
)

func doTestParseRange(rg string, fsize int64, rgs [][2]int64, ok bool, t *testing.T) {

	rg2 := "bytes=" + rg
	rgs1, total1, ok1 := ParseRange(rg2, fsize)
	//	fmt.Println("ParseRange:", rg, rgs1, total1, ok1)
	if ok1 != ok {
		t.Error("ParseRange:", rg, rgs, rgs1, total1, ok1)
	}
	if ok {
		if len(rgs) != len(rgs1) {
			t.Error("ParseRange:", rg, rgs, ok1)
		} else {
			for i, rg1 := range rgs1 {
				if rg1[0] != rgs[i][0] || rg1[1] != rgs[i][1] {
					t.Error(i, "ParseRange:", rg, rgs, ok1)
					break
				}
			}
		}
	}
}

func TestParseRange(t *testing.T) {
	doTestParseRange("-", 100, nil, false, t)
	doTestParseRange("0-", 0, [][2]int64{{0, 0}}, false, t) // 请求0字节内容并无意义，故此服务端不支持。
	doTestParseRange("-1", 100, [][2]int64{{99, 100}}, true, t)
	doTestParseRange("1-", 100, [][2]int64{{1, 100}}, true, t)
	doTestParseRange("1-1", 100, [][2]int64{{1, 2}}, true, t)
	doTestParseRange("1-9", 100, [][2]int64{{1, 10}}, true, t)
	doTestParseRange("10-9", 100, nil, false, t)
	doTestParseRange("100-", 100, nil, false, t)
	doTestParseRange("99-", 100, [][2]int64{{99, 100}}, true, t)
	doTestParseRange("-1,100-", 100, nil, false, t)
	doTestParseRange("99- , 1-", 100, [][2]int64{{99, 100}, {1, 100}}, true, t)
	doTestParseRange("-99", 100, [][2]int64{{1, 100}}, true, t)
	doTestParseRange("-100", 100, [][2]int64{{0, 100}}, true, t)
	doTestParseRange("-101", 100, [][2]int64{{0, 100}}, true, t)
	doTestParseRange("50-99", 100, [][2]int64{{50, 100}}, true, t)
	doTestParseRange("50-100", 100, [][2]int64{{50, 100}}, true, t)
	doTestParseRange("50-101", 100, [][2]int64{{50, 100}}, true, t)
}
