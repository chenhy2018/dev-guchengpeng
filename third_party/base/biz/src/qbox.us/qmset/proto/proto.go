package proto

import (
	"github.com/qiniu/http/httputil.v1"
)

var ErrUnacceptable = httputil.NewError(401, "bad token: unacceptable")

var ErrInvalidKeyValues = httputil.NewError(400, "invalid key-values")
var ErrNotMsetGroup = httputil.NewError(400, "not mset group")
var ErrNotMbloom = httputil.NewError(400, "not mbloom")
var ErrValueNotUrlsafeBase64 = httputil.NewError(400, "param 'v' must be urlsafe base64 encoded")

// ------------------------------------------------------------------------

type MsetGrpCfg struct {
	Id  string `json:"c"`
	Max int    `json:"max"`
}

type MbloomGrpCfg struct {
	Id  string  `json:"c"`
	Max uint    `json:"max"`
	Fp  float64 `json:"fp"`
}

type FlipConfig struct {
	Msets   []MsetGrpCfg   `json:"msets"`
	Mblooms []MbloomGrpCfg `json:"mblooms"`
	Expires int            `json:"expires"`
}

type Flipper interface {
	Flip()
}

// ------------------------------------------------------------------------
