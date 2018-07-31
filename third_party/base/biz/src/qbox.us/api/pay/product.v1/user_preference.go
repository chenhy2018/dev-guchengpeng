package product

import (
	"net/url"
	"strconv"
	"time"

	"qbox.us/api/pay/pay"
)

import (
	"github.com/qiniu/rpc.v1"
)

type HandleUserPreference struct {
	Host   string
	Client *rpc.Client
}

func NewHandleUserPreference(host string, client *rpc.Client) *HandleUserPreference {
	return &HandleUserPreference{host, client}
}

type ReqUserPreferenceGet struct {
	Uid uint32 `json:"uid"`
}

type ItemPreference struct {
	Enabled bool `json:"enabled"`
}

type UserPreference struct {
	Uid       uint32                      `json:"uid"`
	Items     map[pay.Item]ItemPreference `json:"items"`
	UpdatedAt time.Time                   `json:"updated_at"`
}

func (r HandleUserPreference) Get(logger rpc.Logger, req ReqUserPreferenceGet) (resp UserPreference, err error) {
	value := url.Values{}
	value.Add("uid", strconv.FormatUint(uint64(req.Uid), 10))
	err = r.Client.Call(logger, &resp, r.Host+"/v1/user_preference/get?"+value.Encode())
	return
}

type ReqUserPreferenceSet struct {
	Preference UserPreference `json:"preference"`
}

func (r HandleUserPreference) Set(logger rpc.Logger, req ReqUserPreferenceSet) (err error) {
	err = r.Client.CallWithJson(logger, nil, r.Host+"/v1/user_preference/set", req)
	return
}
