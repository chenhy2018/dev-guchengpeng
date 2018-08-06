// +build !go1.7

package xlog

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogger_Basic(t *testing.T) {
	req, err := http.NewRequest("POST", "http://localhost:2601", nil)
	assert.NoError(t, err)

	xl := NewWithReq(req)
	assert.Nil(t, xl.ctx)
	assert.NotNil(t, xl.Context())

	xl2 := NewWith(xl)
	assert.Nil(t, xl2.ctx)
}
