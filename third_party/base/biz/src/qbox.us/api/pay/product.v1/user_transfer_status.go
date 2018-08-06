package product

import (
	"net/url"
	"strconv"
)

import (
	"time"

	"github.com/qiniu/rpc.v1"
)

type HandleUserTransferStatus struct {
	Host   string
	Client *rpc.Client
}

func NewHandleUserTransferStatus(host string, client *rpc.Client) *HandleUserTransferStatus {
	return &HandleUserTransferStatus{host, client}
}

type ReqUserTransferStatus struct {
	Uid              uint32  `json:"uid"`
	Month            int     `json:"month"` // format: 201604
	TransferDiscount float64 `json:"transfer_discount"`
	BandwidthPrice   float64 `json:"bandwidth_price"`
}

type ReqUserTransferStatusList struct {
	Uid   *uint32 `json:"uid"`
	Month *int    `json:"month"`
	// e.p: 0.8 means greater than or equal than 0.8
	//  -0.8 means smaller than  or equal than 0.8
	TransferDiscount *float64 `json:"transfer_discount"`
	// same as TransferDiscount
	BandwidthPrice *float64 `json:"bandwidth_price"`
	Page           int      `json:"page"`
	PageSize       int      `json:"page_size"`
}

type RespUserTransferStatus struct {
	Uid              uint32    `json:"uid"`
	Month            int       `json:"month"`
	TransferDiscount float64   `json:"transfer_discount"`
	BandwidthPrice   float64   `json:"bandwidth_price"`
	CreatedAt        time.Time `json:"created_at"`
}

type ReqUserTransferStatusGet struct {
	Uid   uint32 `json:"uid"`
	Month int    `json:"month"`
}

type ReqUserTransferStatusQuarterList struct {
	Uid     uint32 `json:"uid"`
	Year    int    `json:"year"`
	Quarter int    `json:"quarter"`
}

type ReqUserTransferStatusExport struct {
	Year    int `json:"year"`
	Quarter int `json:"quarter"`
}

func (r HandleUserTransferStatus) Save(logger rpc.Logger, params ReqUserTransferStatus) (err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(params.Uid), 10))
	value.Add("month", strconv.FormatInt(int64(params.Month), 10))
	value.Add("transfer_discount", strconv.FormatFloat(params.TransferDiscount, 'f', -1, 64))
	value.Add("bandwidth_price", strconv.FormatFloat(params.BandwidthPrice, 'f', -1, 64))
	err = r.Client.CallWithForm(logger, nil, r.Host+"/v1/user_transfer_status/save", map[string][]string(value))
	return
}

func (r HandleUserTransferStatus) List(logger rpc.Logger, params ReqUserTransferStatusList) (data []RespUserTransferStatus, err error) {
	value := url.Values{}
	if params.Uid != nil {
		value.Add("uid", strconv.FormatUint(uint64(*params.Uid), 10))
	}
	if params.Month != nil {
		value.Add("month", strconv.FormatInt(int64(*params.Month), 10))
	}
	if params.TransferDiscount != nil {
		value.Add("transfer_discount", strconv.FormatFloat(*params.TransferDiscount, 'f', -1, 64))
	}
	if params.BandwidthPrice != nil {
		value.Add("bandwidth_price", strconv.FormatFloat(*params.BandwidthPrice, 'f', -1, 64))
	}
	value.Add("page", strconv.FormatInt(int64(params.Page), 10))
	value.Add("page_size", strconv.FormatInt(int64(params.PageSize), 10))
	err = r.Client.Call(logger, &data, r.Host+"/v1/user_transfer_status/list?"+value.Encode())
	return
}

func (r HandleUserTransferStatus) Get(logger rpc.Logger, params ReqUserTransferStatusGet) (data RespUserTransferStatus, err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(params.Uid), 10))
	value.Add("month", strconv.FormatInt(int64(params.Month), 10))
	err = r.Client.Call(logger, &data, r.Host+"/v1/user_transfer_status/get?"+value.Encode())
	return
}

func (r HandleUserTransferStatus) QuarterList(logger rpc.Logger, params ReqUserTransferStatusQuarterList) (data []RespUserTransferStatus, err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(params.Uid), 10))
	value.Add("year", strconv.FormatInt(int64(params.Year), 10))
	value.Add("quarter", strconv.FormatInt(int64(params.Quarter), 10))
	err = r.Client.Call(logger, &data, r.Host+"/v1/user_transfer_status/quarter/list?"+value.Encode())
	return
}

func (r HandleUserTransferStatus) Export(logger rpc.Logger, params ReqUserTransferStatusExport) (err error) {
	value := url.Values{}
	value.Add("year", strconv.FormatInt(int64(params.Year), 10))
	value.Add("quarter", strconv.FormatInt(int64(params.Quarter), 10))
	err = r.Client.Call(logger, nil, r.Host+"/v1/user_transfer_status/export?"+value.Encode())
	return
}
