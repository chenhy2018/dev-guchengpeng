package fusionmq

import (
	"net/http"

	"strconv"

	"code.google.com/p/go.net/context"
	"github.com/qiniu/log.v1"
	"github.com/qiniu/rpc.v3"
)

type Client struct {
	Client *rpc.Client
	Host   string
}

func NewClient(host string, client *http.Client) (c *Client, err error) {
	return &Client{&rpc.Client{client}, host}, nil
}

type MqArgs struct {
	Capacity int    `json:"capacity"`
	Expires  string `json:"expires"`
	TryTimes int    `json:"tryTimes"`
}

func (c Client) Make(ctx context.Context, mqName string, ret interface{}, args MqArgs) error {
	url := c.Host + "/mq/" + mqName
	log.Info(url)
	return c.Client.CallWithJson(ctx, ret, "POST", url, args)
}

func (c Client) DeleteMq(ctx context.Context, mqName string, ret interface{}) error {
	url := c.Host + "/mq/" + mqName
	log.Info(url)
	return c.Client.Call(ctx, ret, "DELETE", url)
}

type BodyArgs struct {
	Body []byte `json:"body"`
}

type IdRet struct {
	MsgId string `json:"msgId"`
}

type ListRet struct {
	Marker   string   `json:"marker"`
	Messages []MsgRet `json:"items"`
}

type MsgRet struct {
	Id   string `json:"id"`
	Body []byte `json:"body"`
}

type MsgError struct {
	Error string `json:"error"`
}

func (c Client) Put(ctx context.Context, mqName string, ret interface{}, args BodyArgs) error {
	url := c.Host + "/mq/" + mqName + "/message"
	log.Info(url)
	return c.Client.CallWithJson(ctx, ret, "POST", url, args)
}

func (c Client) PutBytes(ctx context.Context, mqName string, ret interface{}, msg []byte) error {
	url := c.Host + "/mq/" + mqName + "/message"
	log.Info(url)
	return c.Client.CallWithJson(ctx, ret, "POST", url, BodyArgs{msg})
}

func (c Client) PutString(ctx context.Context, mqName string, ret interface{}, msg string) error {
	url := c.Host + "/mq/" + mqName + "/message"
	log.Info(url)
	return c.Client.CallWithJson(ctx, ret, "POST", url, BodyArgs{[]byte(msg)})
}

func (c Client) Pop(ctx context.Context, mqName string, ret interface{}) error {
	url := c.Host + "/mq/" + mqName + "/message"
	log.Info(url)
	return c.Client.Call(ctx, ret, "GET", url)
}

func (c Client) GetById(ctx context.Context, mqName string, msgId string, ret interface{}) (err error) {
	url := c.Host + "/mq/" + mqName + "/message/" + msgId
	return c.Client.Call(ctx, ret, "GET", url)
}

func (c Client) PopN(ctx context.Context, mqName string, n int) (messages []MsgRet, err error) {
	for i := 0; i < n; i++ {
		var message MsgRet
		errPopOne := c.Pop(ctx, mqName, &message)
		if i == 0 && errPopOne != nil {
			err = errPopOne
			break
		}
		if errPopOne != nil {
			break
		}
		messages = append(messages, message)
	}
	return
}

func (c Client) Delete(ctx context.Context, mqName string, msgId string, ret interface{}) error {
	url := c.Host + "/mq/" + mqName + "/message/" + msgId
	log.Info(url)
	return c.Client.Call(ctx, ret, "DELETE", url)
}

func (c Client) Touch(ctx context.Context, mqName string, msgId string, ret interface{}) error {
	url := c.Host + "/mq/" + mqName + "/message/" + msgId + "/touch"
	log.Info(url)
	return c.Client.Call(ctx, ret, "PUT", url)
}

func (c Client) Stat(ctx context.Context, mqName string, ret interface{}) error {
	url := c.Host + "/mq/" + mqName + "/stat"
	log.Info(url)
	return c.Client.Call(ctx, ret, "GET", url)
}

func (c Client) List(ctx context.Context, mqName string, marker string, limit int, ret interface{}) error {
	url := c.Host + "/mq/" + mqName + "/messages" + "?marker=" + marker + "&limit=" + strconv.Itoa(limit)
	log.Info(url)
	return c.Client.Call(ctx, ret, "GET", url)
}
