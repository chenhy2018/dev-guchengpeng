package rewrite

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"qbox.us/qconf/qconfapi"
)

func TestRemoveLeftSlash(t *testing.T) {
	testCases := map[string]string{
		"/":               "/",
		"/abcd":           "/abcd",
		"//abcd":          "/abcd",
		"///////abcd/efg": "/abcd/efg",
		"":                "/",
	}
	for k, v := range testCases {
		assert.Equal(t, RemoveLeftSlash(k), v)
	}
}

func TestKODO4270(t *testing.T) {
	QconfCli = &qconfapi.Client{}
	testCases := map[string]string{
		"/videos/abc":         "/videos/abc",
		"/videos/abc/bcd.ts":  "/videos/abc/bcd.ts",
		"/videos/ots/bcd.ts":  "/videos/ots/bcd.ts",
		"/qpdxv/ots/bcd.ts":   "/qpdxv/ots/bcd.ts",
		"/videos/vts/bcd.ts":  "/videos/vts/bcd.ts",
		"/qpdxv/vts/bcd.ts":   "/qpdxv/vts/bcd.ts",
		"/qpdxvx/vts/bcd.ts1": "/qpdxvx/vts/bcd.ts1",
		"/qpdxv/vts/bcd.tsx":  "/qpdxv/vts/bcd.tsx",
	}
	for k, v := range testCases {
		assert.Equal(t, KODO4270(k), v)
	}
}
