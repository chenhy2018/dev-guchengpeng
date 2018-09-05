package line

import (
	"strconv"

	"golang.org/x/net/context"

	"github.com/qiniu/log.v1"
	"github.com/qiniu/rpc.v3/lb"

	"qbox.us/api/fusion/fusion"
	"qbox.us/api/fusion/fusion/fusionline"
)

type Client struct {
	lbClient *lb.Client
}

func NewClient(hosts []string, lbConf *lb.Config) (c *Client, err error) {
	lbClient, err := lb.New(hosts, lbConf)
	if err != nil {
		return nil, err
	}
	return &Client{lbClient}, nil
}

func (c *Client) Create(ctx context.Context, l *fusion.Line) (err error) {
	path := "/v2/lines/" + l.Id
	return c.lbClient.CallWithJson(ctx, nil, "POST", path, l)
}

func (c *Client) Get(ctx context.Context, id string) (l *fusion.Line, err error) {
	path := "/v2/lines/" + id
	err = c.lbClient.Call(ctx, &l, "GET", path)
	if err != nil {
		log.Error(err)
	}
	return
}

func (c *Client) Update(ctx context.Context, l *fusion.Line) error {
	path := "/v2/lines/" + l.Id
	return c.lbClient.CallWithJson(ctx, nil, "PUT", path, l)
}

func (c *Client) Delete(ctx context.Context, id string) error {
	path := "/v2/lines/" + id
	return c.lbClient.Call(ctx, nil, "DELETE", path)
}

func (c *Client) List(ctx context.Context, listArgs fusionline.ListArgs) (ls []*fusion.Line, err error) {
	path := "/v2/lines"
	params := map[string][]string{
		"hide": []string{strconv.FormatBool(listArgs.Hide)},
	}
	if listArgs.Type != "" {
		params["type"] = []string{listArgs.Type}
	}
	if listArgs.Default != "" {
		params["default"] = []string{listArgs.Default}
	}
	if listArgs.QiniuPrivate != "" {
		params["qiniuPrivate"] = []string{listArgs.QiniuPrivate}
	}
	if listArgs.GeoCover != "" {
		params["geoCover"] = []string{listArgs.GeoCover}
	}
	if listArgs.Protocol != "" {
		params["protocol"] = []string{listArgs.Protocol}
	}
	if listArgs.Platform != "" {
		params["platform"] = []string{listArgs.Platform}
	}
	if listArgs.Features != "" {
		params["features"] = []string{listArgs.Features}
	}
	err = c.lbClient.CallWithForm(ctx, &ls, "GET", path, params)
	if err != nil {
		log.Error(err)
	}
	return
}
