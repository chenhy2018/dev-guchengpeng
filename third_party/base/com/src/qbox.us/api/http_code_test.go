package api

import (
	"errors"
	"testing"

	"github.com/qiniu/http/httputil.v1"
	"github.com/qiniu/rpc.v1"
	"github.com/stretchr/testify/assert"
)

func TestHttpCode(t *testing.T) {
	assert.Equal(t, HttpCode(EInvalidArgs), InvalidArgs)

	err := NewError(570)
	assert.Equal(t, HttpCode(err), 570)

	err = &rpc.ErrorInfo{Code: 612}
	assert.Equal(t, HttpCode(err), 612)

	err = &httputil.ErrorInfo{Code: 631}
	assert.Equal(t, HttpCode(err), 631)

	err = errors.New("unknown error")
	assert.Equal(t, HttpCode(err), 599)

	err = RegisterError(500, "register error")
	assert.Equal(t, HttpCode(err), 500)
}
