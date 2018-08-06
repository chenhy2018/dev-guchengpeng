package cc

import (
	"github.com/qiniu/rpc.v2"
)

// ------------------------------------------------------------------------------------------

type ListFlavorsRet []ServerFlavor

func (r *Service) ListFlavors(l rpc.Logger) (ret ListFlavorsRet, err error) {

	err = r.Conn.Call(l, &ret, "GET", r.Host+r.ApiPrefix+"/flavors")
	return
}

type FlavorInfo struct {
	Id    string `json:"id"`
	Name  string `json:"name"`
	Ram   int    `json:"ram"`
	Vcpus int    `json:"vcpus"`
	Disk  int    `json:"disk"`
}

func (r *Service) FlavorInfo(l rpc.Logger, flavorId string) (ret FlavorInfo, err error) {

	err = r.Conn.Call(l, &ret, "GET", r.Host+r.ApiPrefix+"/flavors/"+flavorId)
	return
}

type ListFlavorsDetailRet []FlavorInfo

func (r *Service) ListFlavorsDetail(l rpc.Logger) (ret ListFlavorsDetailRet, err error) {

	err = r.Conn.Call(l, &ret, "GET", r.Host+r.ApiPrefix+"/flavors/detail")
	return
}

// ------------------------------------------------------------------------------------------
