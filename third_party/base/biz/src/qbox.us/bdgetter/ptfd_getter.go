package bdgetter

import (
	"qbox.us/ptfd/masterapi.v1"
	"qbox.us/ptfd/stgapi.v1"
)

type PtfdConfig struct {
	stgapi.Config
	Master masterapi.Config `json:"master"`
}
