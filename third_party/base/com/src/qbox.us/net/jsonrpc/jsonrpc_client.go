package jsonrpc

import (
	"bytes"
	"encoding/json"
	"qbox.us/errors"
	"github.com/qiniu/log.v1"
)

type Client struct {
	BaseUrl string
}

func NewClient(baseUrl string) Client {
	return Client{baseUrl}
}

// Call invokes the named function, waits for it to complete, and returns its error status.
func (client Client) Call(serviceMethod string, args interface{}, reply interface{}) (err error) {

	log.Debug("jsonrpc.Call:", serviceMethod, args)

	b := bytes.NewBuffer(nil)
	err = json.NewEncoder(b).Encode(args)
	if err != nil {
		err = errors.Info(errors.EINVAL, "jsonrpc.Call", serviceMethod, args).Detail(err).Warn()
		return
	}

	resp, err := PostEx(client.BaseUrl+serviceMethod, "application/json", b, int64(b.Len()))
	if err != nil {
		err = errors.Info(err, "jsonrpc.Call", serviceMethod, args).Detail(err).Warn()
		return err
	}

	_, err = callRet(reply, resp)
	if err != nil {
		err = errors.Info(err, "jsonrpc.Call", serviceMethod, args).Detail(err).Warn()
		return
	}

	log.Debug("jsonrpc.Call:", serviceMethod, args, reply)
	return
}
