package v3

import (
	"net/url"
	"strconv"
)

import (
	"github.com/qiniu/rpc.v1"
	. "qbox.us/api/pay/pay"
	. "qbox.us/zone"
)

type ReqBaseEvmPriceSetter struct {
	Uid          uint32 `json:"uid"`
	Zones        []Zone `json:"zones"`
	EffectTime   Day    `json:"effect_time"`
	DeadTime     Day    `json:"dead_time"`
	ModelBaseEvm        // ModelBase.ID 默认填空
}

func (r HandleUserEvm) BaseSet(logger rpc.Logger, req ReqBaseEvmPriceSetter) (err error) {
	err = r.Client.CallWithJson(logger, nil, r.Host+"/v3/user/evm/base/set", req)
	return
}

func (r HandleUserEvm) PackageSet(logger rpc.Logger, req ReqEntrySetter) (err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	for _, v := range req.Zones {
		value.Add("zones", v.String())
	}
	value.Add("id", req.ID)
	value.Add("effect_time", req.EffectTime.ToString())
	value.Add("dead_time", req.DeadTime.ToString())
	err = r.Client.CallWithForm(logger, nil, r.Host+"/v3/user/evm/package/set", map[string][]string(value))
	return
}

func (r HandleUserEvm) PackageDel(logger rpc.Logger, req ReqEntryDel) (err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	for _, v := range req.Zones {
		value.Add("zones", v.String())
	}
	value.Add("ops", req.Ops)
	err = r.Client.CallWithForm(logger, nil, r.Host+"/v3/user/evm/package/del", map[string][]string(value))
	return
}

func (r HandleUserEvm) DiscountSet(logger rpc.Logger, req ReqEntrySetter) (err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	for _, v := range req.Zones {
		value.Add("zones", v.String())
	}
	value.Add("id", req.ID)
	value.Add("effect_time", req.EffectTime.ToString())
	value.Add("dead_time", req.DeadTime.ToString())
	err = r.Client.CallWithForm(logger, nil, r.Host+"/v3/user/evm/discount/set", map[string][]string(value))
	return
}

func (r HandleUserEvm) DiscountDel(logger rpc.Logger, req ReqEntryDel) (err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	for _, v := range req.Zones {
		value.Add("zones", v.String())
	}
	value.Add("ops", req.Ops)
	err = r.Client.CallWithForm(logger, nil, r.Host+"/v3/user/evm/discount/del", map[string][]string(value))
	return
}

func (r HandleUserEvm) RebateSet(logger rpc.Logger, req ReqEntrySetter) (err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	for _, v := range req.Zones {
		value.Add("zones", v.String())
	}
	value.Add("id", req.ID)
	value.Add("effect_time", req.EffectTime.ToString())
	value.Add("dead_time", req.DeadTime.ToString())
	err = r.Client.CallWithForm(logger, nil, r.Host+"/v3/user/evm/rebate/set", map[string][]string(value))
	return
}

func (r HandleUserEvm) RebateDel(logger rpc.Logger, req ReqEntryDel) (err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	for _, v := range req.Zones {
		value.Add("zones", v.String())
	}
	value.Add("ops", req.Ops)
	err = r.Client.CallWithForm(logger, nil, r.Host+"/v3/user/evm/rebate/del", map[string][]string(value))
	return
}

func (r HandleUserEvm) LifecycleChange(logger rpc.Logger, req ReqLifeCycle) (err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	for _, v := range req.Zones {
		value.Add("zones", v.String())
	}
	value.Add("op", req.OP)
	value.Add("effect_time", req.EffectTime.ToString())
	value.Add("dead_time", req.DeadTime.ToString())
	err = r.Client.CallWithForm(logger, nil, r.Host+"/v3/user/evm/lifecycle/change", map[string][]string(value))
	return
}
