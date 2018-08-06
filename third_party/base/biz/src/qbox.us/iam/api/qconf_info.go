package api

import (
	"fmt"

	rpc "qbox.us/api/fusion/netpkg/rpc.v2"
	"qbox.us/iam/entity"
)

func (c *Client) GetQConfInfo(l rpc.Logger, accessKey string) (entity.QConfInfo, error) {
	reqPath := fmt.Sprintf("/iam/qconf/%s", accessKey)
	var output struct {
		Data entity.QConfInfo `json:"data"`
	}
	err := c.client.GetCall(l, &output, reqPath)
	return output.Data, err
}
