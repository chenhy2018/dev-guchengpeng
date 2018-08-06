package dns

import (
	"net/http"

	"github.com/qiniu/rpc.v3"
	"golang.org/x/net/context"
	"qbox.us/api/fusion/fusion/fusiondns"
)

type Client struct {
	Host      string
	rpcClient rpc.Client
}

func NewClient(host string, httpClient *http.Client) (c *Client, err error) {
	return &Client{host, rpc.Client{httpClient}}, nil
}

func (c *Client) SetRecord(ctx context.Context, args *fusiondns.SetRecordArgs) (task fusiondns.TaskResult, err error) {
	u := c.Host + "/v2/domain/" + args.SubDomain
	err = c.rpcClient.CallWithJson(ctx, &task, "POST", u, args)
	return
}

func (c *Client) DnsGrey(ctx context.Context, args *fusiondns.DnsGreyArgs) (err error) {
	u := c.Host + "/v2/dns/grey"
	var ret interface{}
	return c.rpcClient.CallWithJson(ctx, ret, "POST", u, args)
}

func (c *Client) QueryRecord(ctx context.Context, args *fusiondns.QueryRecordArgs) (ret *fusiondns.QueryRecordRes, err error) {
	url := c.Host + "/v2/domain/" + args.SubDomain + "?dnsProvider=" + string(args.DNSProvider)
	err = c.rpcClient.Call(ctx, &ret, "GET", url)
	return
}

func (c *Client) DeleteRecord(ctx context.Context, args *fusiondns.DeleteRecordArgs) (task fusiondns.TaskResult, err error) {
	url := c.Host + "/v2/domain/" + args.SubDomain
	err = c.rpcClient.CallWithJson(ctx, &task, "DELETE", url, args)
	return
}
