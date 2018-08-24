package rtutil

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test570(t *testing.T) {

	ast := assert.New(t)
	c := &http.Client{
		Transport: New570RT(http.DefaultTransport),
	}

	resp, err := c.Get("http://127.0.0.1:12366")
	ast.Nil(err)
	ast.Equal(570, resp.StatusCode)
	ast.Equal("application/json", resp.Header["Content-Type"][0])
}
