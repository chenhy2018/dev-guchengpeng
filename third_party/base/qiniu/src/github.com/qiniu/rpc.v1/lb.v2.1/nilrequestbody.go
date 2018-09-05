// +build go1.8

package lb

import (
	"io"
	"net/http"
)

func isNilRequestBody(body io.ReadCloser) bool {
	return body == nil || body == http.NoBody
}
