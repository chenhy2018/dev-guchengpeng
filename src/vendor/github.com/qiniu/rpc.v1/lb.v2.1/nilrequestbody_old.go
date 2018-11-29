// +build !go1.8

package lb

import "io"

func isNilRequestBody(body io.ReadCloser) bool {
	return body == nil
}
