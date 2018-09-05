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

type ReqBasePriceSetter struct {
	Uid        uint32 `json:"uid"`
	Zones      []Zone `json:"zones"`
	EffectTime Day    `json:"effect_time"`
	DeadTime   Day    `json:"dead_time"`
	ModelBase         // ModelBase.ID 默认填空
}

func (r HandleUser) BaseSetZones(logger rpc.Logger, req ReqBasePriceSetter) (err error) {
	err = r.Client.CallWithJson(logger, nil, r.Host+"/v3/user/base/set/zones", req)
	return
}

func (r HandleUser) BaseSet(logger rpc.Logger, req ReqBasePriceSetter) (err error) {
	err = r.Client.CallWithJson(logger, nil, r.Host+"/v3/user/base/set", req)
	return
}

type ReqCustomgroup struct {
	Uid         uint32 `json:"uid"`
	Customgroup int    `json:"customgroup"`
	Day         *Day   `json:"day"` // 生效起始时间。如果不填，默认以调用当天为起始点。
}

func (r HandleUser) CustomgroupUpdate(logger rpc.Logger, req ReqCustomgroup) (err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	value.Add("customgroup", strconv.FormatInt(int64(req.Customgroup), 10))
	if req.Day != nil {
		value.Add("day", (*req.Day).ToString())
	}
	err = r.Client.CallWithForm(logger, nil, r.Host+"/v3/user/customgroup/update", map[string][]string(value))
	return
}

type ReqEntrySetter struct {
	Uid        uint32 `json:"uid"`
	Zones      []Zone `json:"zones"`
	ID         string `json:"id"`
	EffectTime Day    `json:"effect_time"`
	DeadTime   Day    `json:"dead_time"`
}

type ReqEntryDel struct {
	Uid   uint32 `json:"uid"`
	Zones []Zone `json:"zones"`
	Ops   string `json:"ops"` // split by ","
}

func (r HandleUser) PackageSetZones(logger rpc.Logger, req ReqEntrySetter) (err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	for _, v := range req.Zones {
		value.Add("zones", v.String())
	}
	value.Add("id", req.ID)
	value.Add("effect_time", req.EffectTime.ToString())
	value.Add("dead_time", req.DeadTime.ToString())
	err = r.Client.CallWithForm(logger, nil, r.Host+"/v3/user/package/set/zones", map[string][]string(value))
	return
}

func (r HandleUser) PackageSet(logger rpc.Logger, req ReqEntrySetter) (err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	for _, v := range req.Zones {
		value.Add("zones", v.String())
	}
	value.Add("id", req.ID)
	value.Add("effect_time", req.EffectTime.ToString())
	value.Add("dead_time", req.DeadTime.ToString())
	err = r.Client.CallWithForm(logger, nil, r.Host+"/v3/user/package/set", map[string][]string(value))
	return
}

func (r HandleUser) PackageDelZones(logger rpc.Logger, req ReqEntryDel) (err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	for _, v := range req.Zones {
		value.Add("zones", v.String())
	}
	value.Add("ops", req.Ops)
	err = r.Client.CallWithForm(logger, nil, r.Host+"/v3/user/package/del/zones", map[string][]string(value))
	return
}

func (r HandleUser) PackageDel(logger rpc.Logger, req ReqEntryDel) (err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	for _, v := range req.Zones {
		value.Add("zones", v.String())
	}
	value.Add("ops", req.Ops)
	err = r.Client.CallWithForm(logger, nil, r.Host+"/v3/user/package/del", map[string][]string(value))
	return
}

func (r HandleUser) DiscountSetZones(logger rpc.Logger, req ReqEntrySetter) (err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	for _, v := range req.Zones {
		value.Add("zones", v.String())
	}
	value.Add("id", req.ID)
	value.Add("effect_time", req.EffectTime.ToString())
	value.Add("dead_time", req.DeadTime.ToString())
	err = r.Client.CallWithForm(logger, nil, r.Host+"/v3/user/discount/set/zones", map[string][]string(value))
	return
}

func (r HandleUser) DiscountSet(logger rpc.Logger, req ReqEntrySetter) (err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	for _, v := range req.Zones {
		value.Add("zones", v.String())
	}
	value.Add("id", req.ID)
	value.Add("effect_time", req.EffectTime.ToString())
	value.Add("dead_time", req.DeadTime.ToString())
	err = r.Client.CallWithForm(logger, nil, r.Host+"/v3/user/discount/set", map[string][]string(value))
	return
}

func (r HandleUser) DiscountDelZones(logger rpc.Logger, req ReqEntryDel) (err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	for _, v := range req.Zones {
		value.Add("zones", v.String())
	}
	value.Add("ops", req.Ops)
	err = r.Client.CallWithForm(logger, nil, r.Host+"/v3/user/discount/del/zones", map[string][]string(value))
	return
}

func (r HandleUser) DiscountDel(logger rpc.Logger, req ReqEntryDel) (err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	for _, v := range req.Zones {
		value.Add("zones", v.String())
	}
	value.Add("ops", req.Ops)
	err = r.Client.CallWithForm(logger, nil, r.Host+"/v3/user/discount/del", map[string][]string(value))
	return
}

func (r HandleUser) RebateSetZones(logger rpc.Logger, req ReqEntrySetter) (err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	for _, v := range req.Zones {
		value.Add("zones", v.String())
	}
	value.Add("id", req.ID)
	value.Add("effect_time", req.EffectTime.ToString())
	value.Add("dead_time", req.DeadTime.ToString())
	err = r.Client.CallWithForm(logger, nil, r.Host+"/v3/user/rebate/set/zones", map[string][]string(value))
	return
}

func (r HandleUser) RebateSet(logger rpc.Logger, req ReqEntrySetter) (err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	for _, v := range req.Zones {
		value.Add("zones", v.String())
	}
	value.Add("id", req.ID)
	value.Add("effect_time", req.EffectTime.ToString())
	value.Add("dead_time", req.DeadTime.ToString())
	err = r.Client.CallWithForm(logger, nil, r.Host+"/v3/user/rebate/set", map[string][]string(value))
	return
}

func (r HandleUser) RebateDelZones(logger rpc.Logger, req ReqEntryDel) (err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	for _, v := range req.Zones {
		value.Add("zones", v.String())
	}
	value.Add("ops", req.Ops)
	err = r.Client.CallWithForm(logger, nil, r.Host+"/v3/user/rebate/del/zones", map[string][]string(value))
	return
}

func (r HandleUser) RebateDel(logger rpc.Logger, req ReqEntryDel) (err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	for _, v := range req.Zones {
		value.Add("zones", v.String())
	}
	value.Add("ops", req.Ops)
	err = r.Client.CallWithForm(logger, nil, r.Host+"/v3/user/rebate/del", map[string][]string(value))
	return
}

type ReqLifeCycle struct {
	Uid        uint32 `json:"uid"`
	Zones      []Zone `json:"zones"`
	OP         string `json:"op"`
	EffectTime Day    `json:"effect_time"`
	DeadTime   Day    `json:"dead_time"`
}

func (r HandleUser) LifecycleChangeZones(logger rpc.Logger, req ReqLifeCycle) (err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	for _, v := range req.Zones {
		value.Add("zones", v.String())
	}
	value.Add("op", req.OP)
	value.Add("effect_time", req.EffectTime.ToString())
	value.Add("dead_time", req.DeadTime.ToString())
	err = r.Client.CallWithForm(logger, nil, r.Host+"/v3/user/lifecycle/change/zones", map[string][]string(value))
	return
}

func (r HandleUser) LifecycleChange(logger rpc.Logger, req ReqLifeCycle) (err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	for _, v := range req.Zones {
		value.Add("zones", v.String())
	}
	value.Add("op", req.OP)
	value.Add("effect_time", req.EffectTime.ToString())
	value.Add("dead_time", req.DeadTime.ToString())
	err = r.Client.CallWithForm(logger, nil, r.Host+"/v3/user/lifecycle/change", map[string][]string(value))
	return
}
