package vline

import (
	"net/http"
	"net/url"
	"strconv"

	"golang.org/x/net/context"

	"github.com/qiniu/xlog.v1"
	"qbox.us/api/fusion/fusion"
	"qbox.us/api/fusion/fusion/fusionvline"
	rpcv2 "qbox.us/api/fusion/netpkg/rpc.v2"
)

type Client struct {
	Host      string
	rpcClient rpcv2.Client
}

func NewClient(host string, httpClient *http.Client) (c *Client, err error) {
	return &Client{host, rpcv2.Client{httpClient}}, nil
}

func (c *Client) Create(ctx context.Context, args *fusionvline.CreateArgs) (task fusionvline.TaskResult, err error) {
	l := xlog.FromContextSafe(ctx)
	u := c.Host + "/v2/domains/" + args.Domain
	err = c.rpcClient.CallWithJson(l, &task, "POST", u, args)
	return
}

func (c *Client) GetTask(ctx context.Context, taskId string) (task fusionvline.Task, err error) {
	l := xlog.FromContextSafe(ctx)
	u := c.Host + "/v2/tasks/" + taskId
	err = c.rpcClient.Call(l, &task, "GET", u)
	return
}

func (c *Client) DeleteTask(ctx context.Context, taskId string) (err error) {
	l := xlog.FromContextSafe(ctx)
	u := c.Host + "/v2/tasks/" + taskId
	return c.rpcClient.Call(l, nil, "DELETE", u)
}

func (c *Client) GetAllTasks(ctx context.Context) (tasks []*fusionvline.Task, err error) {
	l := xlog.FromContextSafe(ctx)
	u := c.Host + "/v2/tasks"
	err = c.rpcClient.Call(l, &tasks, "GET", u)
	return
}

type CallbackArgs struct {
	Cmd        string                `json:"-"`
	Domain     string                `json:"domain"`
	TaskId     string                `json:"taskId"`
	Cname      string                `json:"cname"`
	Result     fusion.OperatingState `json:"result"`
	ResultDesc string                `json:"resultDesc"`
}

func (c *Client) Callback(ctx context.Context, args *CallbackArgs) (err error) {
	l := xlog.FromContextSafe(ctx)
	u := c.Host + "/v2/callback/" + args.Cmd
	return c.rpcClient.CallWithJson(l, nil, "POST", u, args)
}

func (c *Client) SyncToDomainDB(ctx context.Context, args *fusionvline.SyncToDomainDBArgs) (err error) {
	l := xlog.FromContextSafe(ctx)
	u := c.Host + "/v2/domaindb/" + args.Domain
	l.Info("u", u)
	return c.rpcClient.CallWithJson(l, nil, "POST", u, args)
}

func (c *Client) Buckets(ctx context.Context, uid uint32) (buckets []string, err error) {
	l := xlog.FromContextSafe(ctx)
	u := c.Host + "/v2/utils/buckets?uid=" + strconv.Itoa(int(uid))
	err = c.rpcClient.Call(l, &buckets, "GET", u)
	return
}

func (c *Client) Domains(ctx context.Context, uid uint32, bucket string) (ds []*fusionvline.DomainExt, err error) {
	l := xlog.FromContextSafe(ctx)
	query := url.Values{}
	query.Set("uid", strconv.Itoa(int(uid)))
	query.Set("bucket", bucket)
	u := c.Host + "/v2/utils/domains?" + query.Encode()
	err = c.rpcClient.Call(l, &ds, "GET", u)
	return
}
