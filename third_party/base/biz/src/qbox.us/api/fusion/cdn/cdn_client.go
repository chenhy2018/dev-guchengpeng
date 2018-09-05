package cdn

import (
	"net/http"

	rpcv2 "qbox.us/api/fusion/netpkg/rpc.v2"

	"golang.org/x/net/context"

	"github.com/qiniu/log.v1"
	"github.com/qiniu/xlog.v1"
	"qbox.us/api/fusion/fusion/fusioncdn"
)

type Client struct {
	Host      string
	rpcClient rpcv2.Client
}

func NewClient(host string, httpClient *http.Client) (c *Client, err error) {
	return &Client{host, rpcv2.Client{httpClient}}, nil
}

func (c *Client) Query(ctx context.Context, args *fusioncdn.QueryArgs) (domainInfo fusioncdn.DomainInfo, err error) {
	l := xlog.FromContextSafe(ctx)
	u := c.Host + "/v2/domains/" + args.Domain + "?cdnProvider=" + string(args.CdnProvider)
	err = c.rpcClient.CallWithJson(l, &domainInfo, "GET", u, args)
	return
}

func (c *Client) Create(ctx context.Context, args *fusioncdn.CreateArgs) (task fusioncdn.TaskResult, err error) {
	l := xlog.FromContextSafe(ctx)
	u := c.Host + "/v2/domains/" + args.Domain
	err = c.rpcClient.CallWithJson(l, &task, "POST", u, args)
	return
}

func (c *Client) Modify(ctx context.Context, args *fusioncdn.ModifyArgs) (task fusioncdn.TaskResult, err error) {
	l := xlog.FromContextSafe(ctx)
	u := c.Host + "/v2/domains/" + args.Domain
	log.Info("modify url", u)
	err = c.rpcClient.CallWithJson(l, &task, "PUT", u, args)
	return
}

func (c *Client) Delete(ctx context.Context, args *fusioncdn.DeleteArgs) (task fusioncdn.TaskResult, err error) {
	l := xlog.FromContextSafe(ctx)
	u := c.Host + "/v2/domains/" + args.Domain
	err = c.rpcClient.CallWithJson(l, &task, "DELETE", u, args)
	return
}
