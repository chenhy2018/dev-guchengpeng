package fusionec

import (
	"net/http"

	"golang.org/x/net/context"

	"github.com/qiniu/xlog.v1"

	lbc "github.com/qiniu/rpc.v3/lb.v2"
)

const (
	fusionecAdminPrefix = "/v1/fec/admin"
)

type FusionecAdminClient struct {
	client *lbc.Client
}

// Admin 鉴权 Bearer Token
func NewFusionecAdminClient(hosts []string, tr http.RoundTripper) *FusionecAdminClient {
	httpClient := &http.Client{Transport: tr}
	cfg := lbc.Config{Http: httpClient, TryTimes: uint32(len(hosts))}
	client, _ := lbc.New(hosts, &cfg)

	return &FusionecAdminClient{client: client}
}

type FreezeNotificationReq struct {
	Uid uint32 `json:"uid"`
}

// 延时冻结时候的通知
func (f *FusionecAdminClient) FreezeNotification(xl *xlog.Logger, req *FreezeNotificationReq) (err error) {
	ctx := xlog.NewContextWith(context.TODO(), xl)

	err = f.client.CallWithJson(ctx, nil, "POST", fusionecAdminPrefix+"/freeze/notification", req)
	return
}

type FreezeType string

const (
	FreezeTypeAuto   FreezeType = "auto"   // 自动冻结
	FreezeTypeManual FreezeType = "manual" // 手动冻结
)

func (ft FreezeType) Valid() bool {
	switch ft {
	case FreezeTypeAuto, FreezeTypeManual:
		return true
	default:
		return false
	}
}

type FreezeReq struct {
	Uid        uint32     `json:"uid"`
	FreezeType FreezeType `json:"freezeType"`
}

// 冻结通知
func (f *FusionecAdminClient) Freeze(xl *xlog.Logger, req *FreezeReq) (err error) {
	ctx := xlog.NewContextWith(context.TODO(), xl)

	err = f.client.CallWithJson(ctx, nil, "POST", fusionecAdminPrefix+"/freeze", req)
	return
}

type UnFreezeReq struct {
	Uid uint32 `json:"uid"`
}

// 解冻通知
func (f *FusionecAdminClient) UnFreeze(xl *xlog.Logger, req *UnFreezeReq) (err error) {
	ctx := xlog.NewContextWith(context.TODO(), xl)

	err = f.client.CallWithJson(ctx, nil, "POST", fusionecAdminPrefix+"/unfreeze", req)
	return
}
