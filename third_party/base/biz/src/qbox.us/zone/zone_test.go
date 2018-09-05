package zone

import (
	"testing"
)

type testCase struct {
	input   string
	str     string
	name    string
	display string
	isValid bool
}

var testCases = []testCase{
	{"0", "0", "z0", "宁波机房", true},
	{"1", "1", "z1", "昌平机房", true},
	{"2", "2", "na0", "北美机房", true},
	{"3", "3", "z2", "华南机房", true},
	{"1000", "1000", "bq", "容器云华北1区", true},
	{"1001", "1001", "nq", "容器云华东1区", true},
	{"1002", "1002", "lac", "容器云北美1区", true},
	{"1003", "1003", "gq", "容器云华南1区", true},
	{"1004", "1004", "xq", "容器云华东2区", true},
	{"2001", "2001", "kbj1", "北京一区", true},
	{"2011", "2011", "ksh1", "上海一区", true},
	{"2020", "2020", "kgz0", "广州零区", true},
	{"2021", "2021", "kgz1", "广州一区", true},
	{"2022", "2022", "kgz2", "广州二区", true},
	{"2031", "2031", "khk1", "香港一区", true},
	{"2041", "2041", "kto1", "北美一区", true},
	{"2051", "2051", "ksg1", "新加坡一区", true},

	{"z0", "0", "z0", "宁波机房", true},
	{"z1", "1", "z1", "昌平机房", true},
	{"na0", "2", "na0", "北美机房", true},
	{"z2", "3", "z2", "华南机房", true},
	{"bq", "1000", "bq", "容器云华北1区", true},
	{"nq", "1001", "nq", "容器云华东1区", true},
	{"lac", "1002", "lac", "容器云北美1区", true},
	{"gq", "1003", "gq", "容器云华南1区", true},
	{"xq", "1004", "xq", "容器云华东2区", true},
	{"kbj1", "2001", "kbj1", "北京一区", true},
	{"ksh1", "2011", "ksh1", "上海一区", true},
	{"kgz0", "2020", "kgz0", "广州零区", true},
	{"kgz1", "2021", "kgz1", "广州一区", true},
	{"kgz2", "2022", "kgz2", "广州二区", true},
	{"khk1", "2031", "khk1", "香港一区", true},
	{"kto1", "2041", "kto1", "北美一区", true},
	{"ksg1", "2051", "ksg1", "新加坡一区", true},

	{"-10", "", "", "", false},
	{"100", "", "", "", false},
	{"abc", "", "", "", false},
}

func TestZone(t *testing.T) {
	for _, tCase := range testCases {

		zone, err := NewZone(tCase.input)
		if tCase.isValid {
			if err != nil {
				t.Fatalf("NewZone(%s) should not return error, but get: %s\n", tCase.input, err)
			}

			if tCase.isValid != zone.IsValid() {
				t.Fatalf("The zone %d IsValid() should return %v, but get %v\n", zone, tCase.isValid, zone.IsValid())
			}

			if zone.Name() != tCase.name {
				t.Fatalf("The zone %d Name() should return %s, but get %s\n", zone, tCase.name, zone.Name())
			}

			if zone.String() != tCase.str {
				t.Fatalf("The zone %d String() should return %s, but get %s\n", zone, tCase.str, zone.String())
			}

			if zone.Humanize() != tCase.display {
				t.Fatalf("The zone %d Humanize() should be %s, but get %s\n", zone, tCase.display, zone.Humanize())
			}
		} else if err == nil {
			t.Fatalf("NewZone(%s) should return error, but get nil", tCase.input)
		}
	}
}
