// +build go1.7

package xlog

import (
	"net/http"
	"testing"

	"code.google.com/p/go.net/context"
	"github.com/stretchr/testify/assert"
)

func TestLogger_Basic(t *testing.T) {
	req, err := http.NewRequest("POST", "http://localhost:2601", nil)
	assert.NoError(t, err)

	// 不启用 req 的 ctx
	UseReqCtx = false
	xl := NewWithReq(req)
	assert.Nil(t, xl.ctx)
	assert.NotNil(t, xl.Context())

	xl2 := NewWith(xl)
	assert.Nil(t, xl2.ctx)

	// 启用 req 的 ctx
	UseReqCtx = true
	defer func() {
		UseReqCtx = false
	}()
	xl = NewWithReq(req)
	assert.Equal(t, context.Background(), xl.ctx)
	assert.NotNil(t, xl.Context())

	xl2 = NewWith(xl)
	assert.NotNil(t, xl2.ctx)
}
