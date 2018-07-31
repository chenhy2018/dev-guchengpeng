package portal_account

import (
	"fmt"
	"github.com/qiniu/rpc.v1"
	"net/http"
)

type Client struct {
	Host string
	Conn rpc.Client
}

func New(host string, t http.RoundTripper) Client {
	client := &http.Client{Transport: t}
	return Client{Host: host, Conn: rpc.Client{client}}
}

func (c Client) SendMessage(l rpc.Logger, uid uint32, channelId int, templateName string, data map[string]interface{}) (oid string, err error) {
	data = map[string]interface{}{
		"uid":           uid,
		"channel_id":    channelId,
		"template_name": templateName,
		"data":          data,
	}
	ret := map[string]interface{}{}
	err = c.Conn.CallWithJson(l, &ret, c.Host+"/api/message/send", data)
	if err != nil {
		return
	}
	resp_data := ""
	if tmp, ok := ret["data"].(string); ok {
		resp_data = tmp
	} else {
		err = fmt.Errorf("call api/message/send failed, missing data")
	}
	if tmp, ok := ret["code"].(float64); ok {
		code := int(tmp)
		if code != 200 {
			err = fmt.Errorf("call api/message/send failed, code %d, message %s", code, resp_data)
		}
	} else {
		err = fmt.Errorf("call api/message/send failed, missing code")
	}
	oid = resp_data
	return
}
