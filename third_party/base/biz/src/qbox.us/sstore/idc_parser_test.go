package sstore

import (
	"testing"
)

type testCase struct {
	fh        []byte
	expectIdc uint16
	expectErr error
}

var cases = []testCase{
	testCase{
		[]byte{0, 0, 255, 255, 255, 255, 0, 1, 0, 22},
		1,
		nil,
	},
	testCase{
		[]byte{1, 0, 255, 255, 255, 255, 0, 1, 0, 22},
		2,
		nil,
	},
	testCase{
		[]byte{2, 0, 255, 255, 255, 255, 0, 1, 0, 22},
		0,
		ErrUnknownBdInfo,
	},
	testCase{
		[]byte{0, 0, 22},
		1,
		nil,
	},
	testCase{
		[]byte{1, 0, 22},
		1,
		nil,
	},
	testCase{
		[]byte{2, 0, 22},
		1,
		nil,
	},
}

func TestParse(t *testing.T) {

	parser, err := NewIdcParser(map[string]uint16{
		"0": 1,
		"1": 2,
	})
	if err != nil {
		t.Fatal("NewIdcParser:", err)
	}

	for i, tc := range cases {
		idc, err := parser.Parse(tc.fh)
		if idc != tc.expectIdc || err != tc.expectErr {
			t.Fatal("unexpected result =>", i, idc, err)
		}
	}
}
