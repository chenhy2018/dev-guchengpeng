package mockable

import (
	"github.com/qiniu/http/rpcutil.v1"
	"qiniupkg.com/mockable.v1/net"
)

type statService struct {
}

type getStatArg struct {
}

type GetIpsRet struct {
	IpInfos []IpInfo `json:"ipInfos"`
}

type IpInfo struct {
	Ip     string `json:"ip"`
	InBps  int    `json:"inBps"`
	OutBps int    `json:"outBps"`
}

func (s *statService) GetStatIps(arg *getStatArg, env *rpcutil.Env) (ret GetIpsRet, err error) {

	for i := range net.MockingIPs {
		inBps, outBps := net.MockingIPInfos[i].GetBps()
		ret.IpInfos = append(ret.IpInfos, IpInfo{
			Ip:     net.MockingIPs[i],
			InBps:  inBps,
			OutBps: outBps,
		})
	}

	return
}
