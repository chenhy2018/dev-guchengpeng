// +build !go1.8

package api

import "strings"

func isProxyError(err error) bool {
	return strings.Contains(err.Error(), "connecting to proxy")
}
