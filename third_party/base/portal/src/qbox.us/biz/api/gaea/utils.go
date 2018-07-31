package gaea

import (
	"encoding/json"
	"net/http"

	"github.com/qiniu/rpc.v1"
)

func CallWithCookie(req *http.Request, ret interface{}, l rpc.Logger, client rpc.Client) (err error) {
	resp, err := client.Do(l, req)
	if err != nil {
		return
	}

	defer func() {
		resp.Body.Close()
	}()

	if resp.StatusCode/100 == 2 {
		if ret != nil && resp.ContentLength != 0 {
			err = json.NewDecoder(resp.Body).Decode(ret)
			if err != nil {
				return
			}
		}
		if resp.StatusCode == http.StatusOK {
			return nil
		}
	}
	return rpc.ResponseError(resp)
}
