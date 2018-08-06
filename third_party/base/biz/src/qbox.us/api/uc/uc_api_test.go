package uc

import (
	"testing"
)

func TestPatternRegex(t *testing.T) {
	tcs := map[string]bool{
		"":               false,
		"*":              true,
		"*.":             false,
		".*":             false,
		"*.163.com":      true,
		"163.com":        true,
		"*.163qq.com":    true,
		"163qq.com":      true,
		"*.qq.com":       true,
		"qq.com":         true,
		"*.qq163.com":    true,
		"qq163.com":      true,
		"*.163-x.com":    true,
		"163-x.com":      true,
		"163.com:90000":  true,
		"*.163-x.com:9":  true,
		"163.com:900000": false,
		"163.com:":       false,
	}

	for k, v := range tcs {
		if PatternRegex.MatchString(k) != v {
			t.Fatal("match failed:", k, v)
		}
	}
}
