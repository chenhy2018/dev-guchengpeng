package iputil

import (
	"testing"

	"github.com/stretchr/testify.v1/require"
)

func TestIpChecker(t *testing.T) {
	checker := NewDefaultIdcIpChecker()
	myIdc := checker.GetMyIDC()
	require.NotEqual(t, "", myIdc)
	require.NotEqual(t, "", checker.myIDC)
	cases := map[string]string{
		"192.168.35.26":  "nb",
		"10.30.25.23":    "bc",
		"10.44.34.26":    "fs",
		"10.200.20.23":   "cs_dev",
		"10.200.30.23":   "cs_dev_xs",
		"10.34.33.46":    "xs",
		"10.40.34.43":    "lac",
		"10.42.34.22":    "gz",
		"10.32.30.21":    "tc",
		"10.36.33.23":    "ns",
		"183.136.139.24": "183.136",
		"qiniu.com":      "",
		"127.0.0.2":      "local",
	}
	for ips, idc := range cases {
		require.Equal(t, idc, checker.GetIDCByIP(ips), ips)
	}
	checker.SetMyIDC("nb")
	require.False(t, checker.IsNotInSameIDC("192.168.22.33"))
	require.True(t, checker.IsNotInSameIDC("10.32.30.21"))
	require.True(t, checker.IsNotInSameIDC("qiniu.com:80"))
	require.True(t, checker.IsNotInSameIDC("qiniu.com"))
	checker, err := NewIpCheckerWithIpMap(map[string]string{
		"bc": "10.30.0.0/15",
	})
	require.NoError(t, err)
	require.Equal(t, "bc", checker.GetIDCByIP("10.30.1.1"))
	require.Equal(t, "10.42", checker.GetIDCByIP("10.42.1.1"))
}
