package fusiondomain

import (
	"fmt"
	"net/url"

	"github.com/qiniu/api/auth/digest"
	lb "github.com/qiniu/rpc.v1/lb.v2.1"
	"github.com/qiniu/xlog.v1"
	"qbox.us/api/fusion/fusion"
)

type ClientV2 struct {
	lbClient *lb.Client
}

func NewClientV2(hosts []string, ak, sk string) (c *ClientV2, err error) {
	mac := &digest.Mac{
		AccessKey: ak,
		SecretKey: []byte(sk),
	}
	t := digest.NewTransport(mac, nil)

	lbClient := lb.New(&lb.Config{Hosts: hosts, TryTimes: uint32(len(hosts))}, t)
	return &ClientV2{lbClient}, nil
}

func (c *ClientV2) Create(xl *xlog.Logger, domain string, req *fusion.CreateDomainReq) (res *fusion.CommonRes, err error) {
	//path := "/v2/domains/" + domain
	var path string
	if req.Uid > 0 {
		path = "/v2/admin/domains/" + domain
	} else {
		path = "/v2/domains/" + domain
	}
	xl.Info("Create path:", path)
	if err = c.lbClient.CallWithJson(xl, &res, path, req); err != nil {
		xl.Error(err)
		return nil, err
	}
	return
}

func (c *ClientV2) Delete(xl *xlog.Logger, domain string) (res *fusion.CommonRes, err error) {
	path := "/v2/admin/domains/" + domain
	xl.Info("Delete path:", path)
	if err = c.lbClient.DeleteCall(xl, &res, path); err != nil {
		xl.Error(err)
		return nil, err
	}
	return
}

func (c *ClientV2) ModifySource(xl *xlog.Logger, domain string, req *fusion.ModifyDomainSourceReq) (res *fusion.CommonRes, err error) {
	path := "/v2/domains/" + domain + "/source"
	xl.Info("ModifySource path:", path)
	if err = c.lbClient.CallWithJson(xl, &res, path, req); err != nil {
		xl.Error(err)
		return nil, err
	}
	return
}

func (c *ClientV2) AdminGet(xl *xlog.Logger, domain string) (res *fusion.GetDomains_Res, err error) {
	path := "/v2/admin/domains/" + domain
	err = c.lbClient.GetCall(xl, &res, path)
	return

}

func (c *ClientV2) Get(xl *xlog.Logger, domain string, uid uint32) (res *fusion.GetDomains_Res, err error) {
	var path string
	if uid == 0 { //for admin
		path = "/v2/admin/domains/" + domain
	} else { //for user
		path = fmt.Sprintf("/v2/domains/%s?uid=%d", domain, uid)
	}

	if err = c.lbClient.GetCall(xl, &res, path); err != nil {
		xl.Error(err)
		return nil, err
	}

	return
}

func (c *ClientV2) QueryUidByDomain(xl *xlog.Logger, domain string) (uint32, error) {
	var (
		res struct {
			Uid uint32 `json:"uid"`
		}
		path = fmt.Sprintf("/v2/domains/%s/uid", domain)
	)
	err := c.lbClient.GetCall(xl, &res, path)

	return res.Uid, err
}

func (c *ClientV2) List(xl *xlog.Logger, req *fusion.GetDomainsReq) (res *fusion.GetDomainsRes, err error) {
	var path string
	if req.Uid > 0 {
		path = "/v2/admin/domains?"
	} else {
		path = "/v2/domains?"
	}

	vs := url.Values{}
	if req.Marker != "" {
		vs.Set("marker", req.Marker)
	}
	if req.Limit > 0 {
		vs.Set("limit", fmt.Sprintf("%d", req.Limit))
	}
	if req.Uid > 0 {
		vs.Set("uid", fmt.Sprintf("%d", req.Uid))
	}
	if req.DomainPrefix != "" {
		vs.Set("domainPrefix", req.DomainPrefix)
	}
	if req.OperatingState != "" {
		vs.Set("operatingState", string(req.OperatingState))
	}
	if req.OperationType != "" {
		vs.Set("operationType", string(req.OperationType))

	}
	if req.RefererTypes != "" {
		vs.Set("refererTypes", req.RefererTypes)
	}
	if req.SourceQiniuBucket != "" {
		vs.Set("sourceQiniuBucket", req.SourceQiniuBucket)
	}
	if req.SourceType != "" {
		vs.Set("sourceType", string(req.SourceType))
	}
	path = path + vs.Encode()

	xl.Info("List path:", path)
	if err = c.lbClient.GetCall(xl, &res, path); err != nil {
		xl.Error(err)
		return nil, err
	}
	return
}

func (c *ClientV2) Freeze(xl *xlog.Logger, req *fusion.FreezeUserDomainsArgs) (ret fusion.IdRet, err error) {
	path := "/v2/admin/user/domains/freeze"
	err = c.lbClient.CallWithJson(xl, &ret, path, req)
	return
}

func (c *ClientV2) UnFreeze(xl *xlog.Logger, req *fusion.FreezeUserDomainsArgs) (ret fusion.IdRet, err error) {
	path := "/v2/admin/user/domains/unfreeze"
	err = c.lbClient.CallWithJson(xl, &ret, path, req)
	return
}

func (c *ClientV2) GetFreezeTaskState(xl *xlog.Logger, taskId string) (ret fusion.GetTaskDomainsRes, err error) {
	path := "/v2/task/domains?taskId=" + taskId
	err = c.lbClient.GetCall(xl, &ret, path)
	return
}
