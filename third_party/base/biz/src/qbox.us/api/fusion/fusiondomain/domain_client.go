package fusiondomain

import (
	"fmt"
	"net/http"

	"github.com/qiniu/api/auth/digest"
	"github.com/qiniu/log.v1"
	"github.com/qiniu/rpc.v3/lb"
	"golang.org/x/net/context"
	"qbox.us/api/fusion/fusion"
)

type Client struct {
	lbClient *lb.Client
}

func NewClient(hosts []string, ak, sk string) (c *Client, err error) {
	mac := &digest.Mac{
		AccessKey: ak,
		SecretKey: []byte(sk),
	}
	t := digest.NewTransport(mac, nil)

	httpClient := &http.Client{Transport: t}
	lbConf := &lb.Config{Http: httpClient}

	lbClient, err := lb.New(hosts, lbConf)
	if err != nil {
		return nil, err
	}
	return &Client{lbClient}, nil
}

func (c *Client) Create(ctx context.Context, domain string, req *fusion.CreateDomainReq) (res *fusion.CommonRes, err error) {
	//path := "/v2/domains/" + domain
	var path string
	if req.Uid > 0 {
		path = "/v2/admin/domains/" + domain
	} else {
		path = "/v2/domains/" + domain
	}
	log.Info("Create path:", path)
	if err = c.lbClient.CallWithJson(ctx, &res, "POST", path, req); err != nil {
		log.Error(err)
		return nil, err
	}
	return
}

func (c *Client) Delete(ctx context.Context, domain string) (res *fusion.CommonRes, err error) {
	path := "/v2/admin/domains/" + domain
	log.Info("Delete path:", path)
	if err = c.lbClient.Call(ctx, &res, "DELETE", path); err != nil {
		log.Error(err)
		return nil, err
	}
	return
}

func (c *Client) ModifySource(ctx context.Context, domain string, req *fusion.ModifyDomainSourceReq) (res *fusion.CommonRes, err error) {
	path := "/v2/domains/" + domain + "/source"
	log.Info("ModifySource path:", path)
	if err = c.lbClient.CallWithJson(ctx, &res, "POST", path, req); err != nil {
		log.Error(err)
		return nil, err
	}
	return
}

type GetDomainArgs struct {
	Uid uint32 `json:"uid"`
}

func (c *Client) AdminGet(ctx context.Context, domain string) (res *fusion.GetDomains_Res, err error) {
	path := "/v2/admin/domains/" + domain
	err = c.lbClient.Call(ctx, &res, "GET", path)
	return

}

func (c *Client) Get(ctx context.Context, domain string, uid uint32) (res *fusion.GetDomains_Res, err error) {
	var path string
	if uid == 0 { //for admin
		path = "/v2/admin/domains/" + domain
	} else { //for user
		path = "/v2/domains/" + domain
	}

	req := GetDomainArgs{uid}
	if err = c.lbClient.CallWithJson(ctx, &res, "GET", path, req); err != nil {
		log.Error(err)
		return nil, err
	}

	return
}

func (c *Client) QueryUidByDomain(ctx context.Context, domain string) (uint32, error) {
	var (
		res struct {
			Uid uint32 `json:"uid"`
		}
		path = fmt.Sprintf("/v2/domains/%s/uid", domain)
	)
	err := c.lbClient.Call(ctx, &res, "GET", path)

	return res.Uid, err
}

func (c *Client) List(ctx context.Context, req *fusion.GetDomainsReq) (res *fusion.GetDomainsRes, err error) {
	var path string
	if req.Uid > 0 {
		path = "/v2/admin/domains"
	} else {
		path = "/v2/domains"
	}

	log.Info("List path:", path)
	if err = c.lbClient.CallWithJson(ctx, &res, "GET", path, req); err != nil {
		log.Error(err)
		return nil, err
	}
	return
}

func (c *Client) Freeze(ctx context.Context, req *fusion.FreezeUserDomainsArgs) (ret fusion.IdRet, err error) {
	path := "/v2/admin/user/domains/freeze"
	err = c.lbClient.CallWithJson(ctx, &ret, "POST", path, req)
	return
}

func (c *Client) UnFreeze(ctx context.Context, req *fusion.FreezeUserDomainsArgs) (ret fusion.IdRet, err error) {
	path := "/v2/admin/user/domains/unfreeze"
	err = c.lbClient.CallWithJson(ctx, &ret, "POST", path, req)
	return
}

func (c *Client) GetFreezeTaskState(ctx context.Context, taskId string) (ret fusion.GetTaskDomainsRes, err error) {
	path := "/v2/task/domains?taskId=" + taskId
	err = c.lbClient.Call(ctx, &ret, "GET", path)
	return
}
