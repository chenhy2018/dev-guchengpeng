package wallet

import (
	"net/url"
	"strconv"
)

import (
	"time"

	"github.com/qiniu/rpc.v1"
	. "qbox.us/api/pay/pay"
	P "qbox.us/api/pay/price/v3"
	. "qbox.us/zone"
)

type HandleResPack struct {
	Host   string
	Client *rpc.Client
}

func NewHandleResPack(host string, client *rpc.Client) *HandleResPack {
	return &HandleResPack{host, client}
}

type ModelResPackQuota struct {
	Id           string         `json:"id"`
	Uid          uint32         `json:"uid,omitempty"`
	Zone         Zone           `json:"zone,omitempty"`
	ResPackId    string         `json:"respack_id"`
	BindId       string         `json:"bind_id"`
	Item         Item           `json:"item"`
	DataType     P.ItemDataType `json:"data_type"`
	Quota        int64          `json:"quota"`
	RelatedMonth time.Time      `json:"related_month"`
	ExpireAt     time.Time      `json:"expire_at"`
	CreateAt     time.Time      `json:"create_at"`

	Transactions []ModelResPackTransaction `json:"transactions,omitempty"`
}

type ModelResPackTransaction struct {
	QuotaId  string    `json:"quota_id"`
	BillId   string    `json:"bill_id,omitempty"`
	Used     int64     `json:"used"`
	CreateAt time.Time `json:"create_at"`
}

func (r HandleResPack) UserQuotaListEffect(logger rpc.Logger, req ReqUidAndSecond) (resp []ModelResPackQuota, err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	if req.Zone != nil {
		value.Add("zone", (*req.Zone).String())
	}
	value.Add("time", req.Time.ToString())
	err = r.Client.Call(logger, &resp, r.Host+"/v3/respack/user/quota/list/effect?"+value.Encode())
	return
}

func (r HandleResPack) RevertByBillId(logger rpc.Logger, req ReqID) (err error) {
	value := url.Values{}
	value.Add("id", req.ID)
	err = r.Client.Call(logger, nil, r.Host+"/v3/respack/revert/by/bill/id?"+value.Encode())
	return
}
