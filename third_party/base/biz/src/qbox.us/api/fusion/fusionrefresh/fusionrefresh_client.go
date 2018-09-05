package fusionrefresh

import (
	"github.com/qiniu/api/auth/digest"
	"github.com/qiniu/log.v1"
	"github.com/qiniu/rpc.v3/lb"
	"golang.org/x/net/context"
	"net/http"
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

// type GetRefreshSettingArgs struct {
// 	Uid uint32 `json:"uid"`
// }

// func (c *Client) GetRefreshSetting(ctx context.Context, uid uint32) (res *fusion.GetRefreshSettingRes, err error) {
// 	var path string = "/refresh/setting"

// 	req := GetRefreshSettingArgs{uid}
// 	log.Info("GetRefreshSetting path:", path)
// 	if err = c.lbClient.CallWithJson(ctx, &res, "GET", path, req); err != nil {
// 		log.Error(err)
// 		return nil, err
// 	}

// 	return
// }

// type CreateRefreshSettingArgs struct {
// 	fusion.CreateRefreshSettingReq
// }

// func (c *Client) PostRefreshSetting(ctx context.Context, req *fusion.CreateRefreshSettingReq) (res *fusion.CommonRes, err error) {
// 	path := "/refresh/setting"

// 	log.Info("PostRefreshSetting path:", path)
// 	if err = c.lbClient.CallWithJson(ctx, &res, "POST", path, req); err != nil {
// 		log.Error(err)
// 		return nil, err
// 	}
// 	return
// }

// type GetPrefetchSettingArgs struct {
// 	Uid uint32 `json:"uid"`
// }

// func (c *Client) GetPrefetchSetting(ctx context.Context, uid uint32) (res *fusion.GetPrefetchSettingRes, err error) {
// 	var path string = "/prefetch/setting"

// 	req := GetPrefetchSettingArgs{uid}
// 	log.Info("GetPrefetchSetting path:", path)
// 	if err = c.lbClient.CallWithJson(ctx, &res, "GET", path, req); err != nil {
// 		log.Error(err)
// 		return nil, err
// 	}

// 	return
// }

// type CreatePrefetchSettingArgs struct {
// 	fusion.CreatePrefetchSettingReq
// }

// func (c *Client) PostPrefetchSetting(ctx context.Context, req *fusion.CreatePrefetchSettingReq) (res *fusion.CommonRes, err error) {
// 	path := "/prefetch/setting"

// 	log.Info("PostPrefetchSetting path:", path)
// 	if err = c.lbClient.CallWithJson(ctx, &res, "POST", path, req); err != nil {
// 		log.Error(err)
// 		return nil, err
// 	}
// 	return
// }

type PostRefreshArgs struct {
	fusion.PostRefreshReq
}

func (c *Client) PostRefresh(ctx context.Context, req *fusion.PostRefreshReq) (res *fusion.PostRefreshRes, err error) {
	path := "/refresh"

	log.Info("PostRefresh path:", path)
	if err = c.lbClient.CallWithJson(ctx, &res, "POST", path, req); err != nil {
		log.Error(err)
		return nil, err
	}
	return
}

type PostPrefetchArgs struct {
	fusion.PostPrefetchReq
}

func (c *Client) PostPrefetch(ctx context.Context, req *fusion.PostPrefetchReq) (res *fusion.PostPrefetchRes, err error) {
	path := "/prefetch"

	log.Info("PostPrefetch path:", path)
	if err = c.lbClient.CallWithJson(ctx, &res, "POST", path, req); err != nil {
		log.Error(err)
		return nil, err
	}
	return
}

func (c *Client) GetPrefetchQuery(ctx context.Context, req *fusion.GetPrefetchQueryReq) (res *fusion.GetPrefetchQueryRes, err error) {
	var path string = "/prefetch/query?requestId=" + req.RequestId

	log.Info("GetPrefetchQuery path:", path)
	if err = c.lbClient.Call(ctx, &res, "GET", path); err != nil {
		log.Error(err)
		return nil, err
	}
	// if err = c.lbClient.CallWithJson(ctx, &res, "GET", path, req); err != nil {
	// 	log.Error(err)
	// 	return nil, err
	// }

	return
}
