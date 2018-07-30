package configs

import (
	"net/http"
	"syscall"

	"github.com/teapots/params"
	"github.com/teapots/render"
	"github.com/teapots/teapot"
)

var (
	ErrNotExists = syscall.ENOENT
)

type Base struct {
	Req *http.Request       `inject`
	Rw  http.ResponseWriter `inject`

	Params *params.Params `inject`
	Render render.Render  `inject`
	Log    teapot.Logger  `inject`
}
