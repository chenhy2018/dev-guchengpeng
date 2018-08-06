// 相比v1，bucketInfo不会返回BindDomains
package uc

import (
	"net/http"

	"github.com/qiniu/rpc.v1"
	"github.com/qiniu/rpc.v1/lb.v2.1"
	"qbox.us/api/uc"
)

type BucketInfo uc.BucketInfoWithoutBindDomains

type BucketInfos []struct {
	Name string     `json:"name"`
	Info BucketInfo `json:"info"`
}

// ------------------------------------------------------------------------------------------

type Service struct {
	Conn       *lb.Client
	uc.Service // 功能废弃，兼容保留
}

func New(host string, t http.RoundTripper) *Service {
	cfg := &lb.Config{
		Hosts:    []string{host},
		TryTimes: 1,
	}
	client := lb.New(cfg, t)
	return &Service{
		Conn:    client,
		Service: *uc.New(host, t),
	}
}

func NewWithMultiHosts(hosts []string, t http.RoundTripper) *Service {
	cfg := &lb.Config{
		Hosts:    hosts,
		TryTimes: uint32(len(hosts)),
	}
	client := lb.New(cfg, t)
	return &Service{
		Conn: client,
	}
}

// ------------------------------------------------------------------------------------------

func (r Service) BucketInfo(l rpc.Logger, bucket string) (info BucketInfo, err error) {

	params := map[string][]string{
		"bucket": {bucket},
	}
	err = r.Conn.CallWithForm(l, &info, "/v2/bucketInfo", params)
	return
}

func (r Service) BucketInfos(l rpc.Logger, zone string) (infos BucketInfos, err error) {

	params := map[string][]string{}
	if zone != "" {
		params["zone"] = []string{zone}
	}
	err = r.Conn.CallWithForm(l, &infos, "/v2/bucketInfos", params)
	return
}

func (r Service) GlbBucketInfos(l rpc.Logger, region, global string) (infos BucketInfos, err error) {

	params := map[string][]string{}
	if region != "" {
		params["region"] = []string{region}
	}
	if global != "" {
		params["global"] = []string{global}
	}
	err = r.Conn.CallWithForm(l, &infos, "/v2/bucketInfos", params)
	return
}
