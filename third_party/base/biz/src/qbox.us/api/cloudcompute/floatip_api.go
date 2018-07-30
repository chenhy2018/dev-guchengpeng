package cc

import (
	"strconv"

	"github.com/qiniu/rpc.v2"
)

type FloatingipInfo struct {
	Id                string `json:"id"`
	Name              string `json:"name"`
	Status            string `json:"status"`
	CreatedAt         int64  `json:"created_at"`
	RateLimit         int    `json:"rate_limit"`
	PortId            string `json:"port_id"`
	FloatingipAddress string `json:"floating_ip_address"`
	Provider          string `json:"provider"`
}

type CreateFloatingipArgs struct {
	Name      string
	RateLimit int
	NetworkId string
	Provider  string // CHINATELECOM or CHINAUNICOM
}

func (r *Service) CreateFloatingip(l rpc.Logger, args *CreateFloatingipArgs) (ret FloatingipInfo, err error) {

	params := map[string][]string{
		"name":       {args.Name},
		"rate_limit": {strconv.Itoa(args.RateLimit)},
		"network_id": {args.NetworkId},
		"provider":   {args.Provider},
	}

	err = r.Conn.CallWithForm(l, &ret, "POST", r.Host+r.ApiPrefix+"/floatingips", params)
	return
}

// ------------------------------------------------------------

func (r *Service) BindFloatingip(l rpc.Logger, floatingipId, portId string) (err error) {

	params := map[string][]string{
		"port_id": {portId},
	}

	return r.Conn.CallWithForm(l, nil, "POST",
		r.Host+r.ApiPrefix+"/floatingips/"+floatingipId+"/bind", params)
}

func (r *Service) UnbindFloatingip(l rpc.Logger, floatingipId string) (err error) {

	return r.Conn.Call(l, nil, "POST",
		r.Host+r.ApiPrefix+"/floatingips/"+floatingipId+"/unbind")
}

// ------------------------------------------------------------

func (r *Service) DeleteFloatingip(l rpc.Logger, floatingipId string) (err error) {

	return r.Conn.Call(l, nil, "DELETE", r.Host+r.ApiPrefix+"/floatingips/"+floatingipId)
}

// ------------------------------------------------------------

type ListFloatingipsRet []FloatingipInfo

func (r *Service) ListFloatingips(l rpc.Logger) (ret ListFloatingipsRet, err error) {

	err = r.Conn.Call(l, &ret, "GET", r.Host+r.ApiPrefix+"/floatingips")
	return
}

// ------------------------------------------------------------

func (r *Service) GetFloatingipInfo(l rpc.Logger, floatingipId string) (ret FloatingipInfo, err error) {

	err = r.Conn.Call(l, &ret, "GET", r.Host+r.ApiPrefix+"/floatingips/"+floatingipId)
	return
}

// ------------------------------------------------------------

type AdjustFloatingipArgs struct {
	RateLimit int
}

func (r *Service) AdjustFloatingip(l rpc.Logger, floatingipId string, args AdjustFloatingipArgs) (err error) {

	params := map[string][]string{
		"rate_limit": {strconv.Itoa(args.RateLimit)},
	}

	return r.Conn.CallWithForm(l, nil, "POST", r.Host+r.ApiPrefix+"/floatingips/"+floatingipId+"/adjust", params)
}
