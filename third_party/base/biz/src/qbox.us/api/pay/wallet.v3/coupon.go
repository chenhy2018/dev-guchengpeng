package wallet

import (
	"net/url"
	"strconv"
)

import (
	"github.com/qiniu/rpc.v1"
	. "qbox.us/api/pay/pay"
)

type HandleCoupon struct {
	Host   string
	Client *rpc.Client
}

func NewHandleCoupon(host string, client *rpc.Client) *HandleCoupon {
	return &HandleCoupon{host, client}
}

type ModelCoupon struct {
	Quota      Money             `json:"quota"`
	Balance    Money             `json:"balance"`
	CreateAt   HundredNanoSecond `json:"create_at"`
	UpdateAt   HundredNanoSecond `json:"update_at"`
	EffectTime HundredNanoSecond `json:"effecttime"`
	DeadTime   HundredNanoSecond `json:"deadtime"`
	Uid        uint32            `json:"uid"` //用户ID
	Day        int               `json:"day"`
	Id         string            `json:"id"`
	Title      string            `json:"title"`
	Desc       string            `json:"desc"`
	Type       CouponType        `json:"type"`
	Status     CouponStatus      `json:"status"`
	Scope      Scope             `json:"scope"`
}

type ReqCouponNew struct {
	Quota    Money      `json:"quota"`
	Day      int        `json:"day"`
	DeadTime int64      `json:"deadtime"`
	Type     CouponType `json:"type"`
	Desc     string     `json:"desc"`
	Title    string     `json:"title"`
	Scope    Scope      `json:"scope"`
}

type ReqCouponActive struct {
	Excode string `json:"excode"`
	Uid    uint32 `json:"uid"`
	Id     string `json:"id"`
	Desc   string `json:"desc"`
}

type ReqCouponHistory struct {
	Uid    uint32 `json:"uid"`
	Offset int    `json:"offset"`
	Limit  int    `json:"limit"`
}

type ReqCouponAdminList struct {
	Offset *int    `json:"offset"`
	Limit  *int    `json:"limit"`
	Uid    int64   `json:"uid"`
	Type   string  `json:"type"`
	Status int     `json:"status"`
	Title  *string `json:"title"`
}

type ReqCouponAdminCount struct {
	Uid    int64  `json:"uid"`
	Type   string `json:"type"`
	Status int    `json:"status"`
	Title  string `json:"title"`
}

type ReqUidAndID struct {
	Uid uint32 `json:"uid"`
	ID  string `json:"id"`
}

func (r HandleCoupon) Get(logger rpc.Logger, req ReqID) (resp ModelCoupon, err error) {
	value := url.Values{}
	value.Add("id", req.ID)
	err = r.Client.Call(logger, &resp, r.Host+"/v3/coupon/get?"+value.Encode())
	return
}

func (r HandleCoupon) List(logger rpc.Logger, req ReqUid) (out []ModelCoupon, err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	err = r.Client.Call(logger, &out, r.Host+"/v3/coupon/list?"+value.Encode())
	return
}

type ReqListByScope struct {
	ID     string `json:"id"`
	Offset int    `json:"offset"`
	Limit  int    `json:"limit"`
}

func (r HandleCoupon) ListByscope(logger rpc.Logger, req ReqListByScope) (out []ModelCoupon, err error) {
	value := url.Values{}
	value.Add("id", req.ID)
	value.Add("offset", strconv.FormatInt(int64(req.Offset), 10))
	value.Add("limit", strconv.FormatInt(int64(req.Limit), 10))
	err = r.Client.Call(logger, &out, r.Host+"/v3/coupon/list/byscope?"+value.Encode())
	return
}

func (r HandleCoupon) History(logger rpc.Logger, req ReqCouponHistory) (out []ModelCoupon, err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	value.Add("offset", strconv.FormatInt(int64(req.Offset), 10))
	value.Add("limit", strconv.FormatInt(int64(req.Limit), 10))
	err = r.Client.Call(logger, &out, r.Host+"/v3/coupon/history?"+value.Encode())
	return
}

func (r HandleCoupon) AdminCount(logger rpc.Logger, req ReqCouponAdminCount) (count int, err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatInt(int64(req.Uid), 10))
	value.Add("type", req.Type)
	value.Add("status", strconv.FormatInt(int64(req.Status), 10))
	value.Add("title", req.Title)
	err = r.Client.Call(logger, &count, r.Host+"/v3/coupon/admin/count?"+value.Encode())
	return
}

func (r HandleCoupon) AdminList(logger rpc.Logger, req ReqCouponAdminList) (out []ModelCoupon, err error) {
	value := url.Values{}
	if req.Offset != nil {
		value.Add("offset", strconv.FormatInt(int64(*req.Offset), 10))
	}
	if req.Limit != nil {
		value.Add("limit", strconv.FormatInt(int64(*req.Limit), 10))
	}
	value.Add("uid", strconv.FormatInt(int64(req.Uid), 10))
	value.Add("type", req.Type)
	value.Add("status", strconv.FormatInt(int64(req.Status), 10))
	if req.Title != nil {
		value.Add("title", *req.Title)
	}
	err = r.Client.Call(logger, &out, r.Host+"/v3/coupon/admin/list?"+value.Encode())
	return
}

func (r HandleCoupon) New(logger rpc.Logger, req ReqCouponNew) (id string, err error) {
	err = r.Client.CallWithJson(logger, &id, r.Host+"/v3/coupon/new", req)
	return
}

func (r HandleCoupon) Active(logger rpc.Logger, req ReqCouponActive) (out ModelCoupon, err error) {
	value := url.Values{}
	value.Add("excode", req.Excode)
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	value.Add("id", req.Id)
	value.Add("desc", req.Desc)
	err = r.Client.CallWithForm(logger, &out, r.Host+"/v3/coupon/active", map[string][]string(value))
	return
}

func (r HandleCoupon) Cancel(logger rpc.Logger, req ReqUidAndID) (err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	value.Add("id", req.ID)
	err = r.Client.CallWithForm(logger, nil, r.Host+"/v3/coupon/cancel", map[string][]string(value))
	return
}
