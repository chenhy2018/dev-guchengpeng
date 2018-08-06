package client

import (
	"qbox.us/biz/component/client"

	"github.com/teapots/teapot"
)

func Transport() interface{} {
	return func(log teapot.ReqLogger) *client.TransportWithReqLogger {
		return client.NewTransportWithReqLogger(log)
	}
}
