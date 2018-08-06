package client

import (
	"code.google.com/p/go.net/context"
	"github.com/qiniu/rpc.v3/lb"
	_ "github.com/qiniu/version"
	"github.com/qiniu/xlog.v1"
	cm "qbox.us/api/domainunify"
)

type Client struct {
	client *lb.Client
}

func NewClient(hosts []string, lbConf *lb.Config) (c *Client, err error) {
	client, err := lb.New(hosts, lbConf)
	if err != nil {
		return nil, err
	}

	return &Client{client}, nil
}

func (c *Client) Create(ctx context.Context, domain string, uid uint32, nocheck bool) (res cm.RetRes, err error) {
	path := "/v1/domain/create/" + domain
	xl := xlog.FromContextSafe(ctx)
	args := cm.CreateArgs{
		UID:     uid,
		NoCheck: nocheck,
	}

	if err = c.client.CallWithJson(ctx, &res, "POST", path, args); err != nil {
		xl.Error(err)
		return
	}

	xl.Info("domainunify Create success")
	return
}

func (c *Client) DomainQuery(ctx context.Context, domain string, uid uint32, starttime, endtime int64) (res cm.TimeRes, err error) {
	xl := xlog.FromContextSafe(ctx)
	path := "/v1/domain/query/" + domain
	args := cm.QueryArgs{
		UID:       uid,
		StartTime: starttime,
		EndTime:   endtime,
	}

	if err = c.client.CallWithJson(ctx, &res, "POST", path, args); err != nil {
		xl.Error(err)
		return
	}

	xl.Info("domainunify DomainQuery success")
	return
}

func (c *Client) UpdateUID(ctx context.Context, domain string, uid, targetUID uint32) (res cm.RetRes, err error) {
	xl := xlog.FromContextSafe(ctx)
	path := "/v1/domain/update/uid/" + domain
	args := cm.UpdateUIDArgs{
		UID:       uid,
		TargetUID: targetUID,
	}

	if err = c.client.CallWithJson(ctx, &res, "POST", path, args); err != nil {
		xl.Error(err)
		return
	}

	xl.Info("domainunify UpateUID success")
	return
}

func (c *Client) UpdateProd(ctx context.Context, domain string, uid uint32, targetProd string) (res cm.RetRes, err error) {
	xl := xlog.FromContextSafe(ctx)
	path := "/v1/domain/update/prod/" + domain
	args := cm.UpdateProdArgs{
		UID:        uid,
		TargetProd: targetProd,
	}

	if err = c.client.CallWithJson(ctx, &res, "POST", path, args); err != nil {
		xl.Error(err)
		return
	}

	xl.Info("domainunify UpdateProd success")
	return
}

func (c *Client) Delete(ctx context.Context, domain string) (res cm.RetRes, err error) {
	xl := xlog.FromContextSafe(ctx)
	path := "/v1/domain/" + domain

	if err = c.client.Call(ctx, &res, "DELETE", path); err != nil {
		xl.Error(err)
		return
	}

	xl.Info("domainunify Delete success")
	return
}

func (c *Client) GetLog(ctx context.Context, domain string) (res cm.LogRes, err error) {
	xl := xlog.FromContextSafe(ctx)
	path := "/v1/log/" + domain

	if err = c.client.Call(ctx, &res, "GET", path); err != nil {
		xl.Error(err)
		return
	}

	xl.Info("domainunify GetLog success")
	return
}

func (c *Client) CheckICP(ctx context.Context, domain string) (res cm.RetRes, err error) {
	xl := xlog.FromContextSafe(ctx)
	path := "/v1/icp/" + domain

	if err = c.client.Call(ctx, &res, "GET", path); err != nil {
		xl.Error(err)
		return
	}

	xl.Info("domainunify CheckICP success")
	return
}

func (c *Client) DownloadVerifyFile(ctx context.Context, domain string) (res cm.RetRes, err error) {
	xl := xlog.FromContextSafe(ctx)
	path := "/v1/downloadverifyfile/" + domain

	if err = c.client.Call(ctx, &res, "GET", path); err != nil {
		xl.Error(err)
		return
	}

	xl.Info("domainunify DownloadVerifyFile success")
	return
}

func (c *Client) Verify(ctx context.Context, domain string, newcname bool, targetProd string) (res cm.RetRes, err error) {
	xl := xlog.FromContextSafe(ctx)
	path := "/v1/verify"
	args := cm.VerifyArgs{
		Domain:     domain,
		NewCname:   newcname,
		TargetProd: targetProd,
	}

	if err = c.client.CallWithJson(ctx, &res, "POST", path, args); err != nil {
		xl.Error(err)
		return
	}

	xl.Info("domainunify Verify success")
	return
}

func (c *Client) Domainstate(ctx context.Context, domain string) (res cm.DomainstateRes, err error) {
	xl := xlog.FromContextSafe(ctx)
	path := "/v1/domainstate/" + domain

	if err = c.client.Call(ctx, &res, "GET", path); err != nil {
		xl.Error(err)
		return
	}

	xl.Info("domainunify Domainstate success")
	return
}

func (c *Client) Verifystate(ctx context.Context) (res cm.VerifystateRes, err error) {
	xl := xlog.FromContextSafe(ctx)
	path := "/v1/verifystate"

	if err = c.client.Call(ctx, &res, "GET", path); err != nil {
		xl.Error(err)
		return
	}

	xl.Info("domainunify Verifystate success")
	return
}

func (c *Client) AdminHumanverify(ctx context.Context) (res cm.HumanVerifyRes, err error) {
	xl := xlog.FromContextSafe(ctx)
	path := "/v1/admin/humanverify"

	if err = c.client.Call(ctx, &res, "GET", path); err != nil {
		xl.Error(err)
		return
	}

	xl.Info("domainunify AdminHumanverify success")
	return
}

func (c *Client) PostAdminVerify(ctx context.Context, taskId string, isAccept bool, msg string) (res cm.RetRes, err error) {
	xl := xlog.FromContextSafe(ctx)
	path := "/v1/admin/humanverify/" + taskId
	args := cm.HumanVerifyArgs{
		IsAccept: isAccept,
		Msg:      msg,
	}

	if err = c.client.CallWithJson(ctx, &res, "POST", path, args); err != nil {
		xl.Error(err)
		return
	}

	xl.Info("domainunify PostAdminVerify success")
	return
}
