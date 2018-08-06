package gaea

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/qiniu/rpc.v1"

	"qbox.us/biz/api/gaea/enums"
)

type DeveloperService struct {
	host   string
	client rpc.Client
}

func NewDeveloperService(host string, t http.RoundTripper) *DeveloperService {
	return &DeveloperService{
		host: host,
		client: rpc.Client{
			&http.Client{Transport: t},
		},
	}
}

type DeveloperInfo struct {
	Uid   uint32 `json:"Uid"`
	Email string `json:"Email"`

	FullName        string       `json:"Fullname"`
	PhoneNumber     string       `json:"PhoneNumber"`
	ImCategory      int          `json:"ImCategory"`
	ImNumber        string       `json:"ImNumber"`
	CompanyCategory int          `json:"CompanyCategory"`
	CompanySize     int          `json:"CompanySize"`
	CompanyName     string       `json:"CompanyName"`
	Website         string       `json:"Website"`
	Gender          enums.Gender `json:"Gender"`
	Location        []string     `json:"Location"`

	MobileBinded   bool                 `json:"MobileBinded"`
	LicenseVersion enums.LicenseVersion `json:"LicenseVersion"`
	RegisterIp     string               `json:"RegisterIp"`
	InviterUid     uint32               `json:"InviterUid"`
	IsActived      bool                 `json:"IsActived"`
	SalesInvite    string               `json:"SalesInvite"`

	// 各类时间
	CreatedAt              int64     `json:"CreatedAt"`
	CreatedAtTime          time.Time `json:"CreatedAtTime"`
	UpdatedAt              time.Time `json:"UpdatedAt"`
	UpgradeStdAt           time.Time `json:"UpgradeStdAt"`
	UpgradeVipAt           time.Time `json:"UpgradeVipAt"`
	LastPasswordModifyTime time.Time `json:"LastPasswordModifyTime"`
	LastEmailModifyTime    time.Time `json:"LastEmailModifyTime"`

	// Two-factor authentication
	IsTwoFactorAuthEnabled bool `josn:"IsTwoFactorAuthEnabled"`
	IsTotpEnabled          bool `json:"IsTotpEnabled"`

	// 待废弃的
	ZendeskId  int64  `bson:"ZendeskId"`
	MemoCounts int    `bson:"MemoCounts"`
	CrmId      string `bson:"CrmId"`
	CrmLeadId  string `bson:"CrmLeadId"`
}

type developerInfoResp struct {
	CommonResponse

	Info *DeveloperInfo `json:"data"`
}

// DeveloperService.Info provides detailed developer info from `db.developers` table.
// Some fields are omitted for security reasons.
func (s *DeveloperService) Info(l rpc.Logger) (info *DeveloperInfo, err error) {
	var resp developerInfoResp

	err = s.client.GetCall(l, &resp, s.host+"/api/oauth/user")
	if err != nil {
		return
	}

	if resp.Code != CodeOK {
		err = errors.New(fmt.Sprintf("DeveloperService::Info() failed with code: %d", resp.Code))
		return
	}

	info = resp.Info

	return
}
