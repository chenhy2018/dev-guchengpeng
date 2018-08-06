package ustack

import (
	"net/http"

	"github.com/qiniu/rpc.v2/failover"
)

// --------------------------------------------------

func defaultNewConn(hosts []string, rt http.RoundTripper) Conn {

	return failover.New(hosts, &failover.Config{
		Http: &http.Client{
			Transport: rt,
		},
	})
}

// --------------------------------------------------

