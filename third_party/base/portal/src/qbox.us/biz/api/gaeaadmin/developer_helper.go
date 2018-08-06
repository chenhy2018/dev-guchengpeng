package gaeaadmin

import (
	"net/url"
	"strconv"
	"time"

	"qbox.us/biz/services.v2/account"
)

type LicenseVersion string

const (
	LicenseNoFound LicenseVersion = ""
	LicenseActived LicenseVersion = "0.1"
)

type Gender int

const (
	GenderMale   Gender = 0
	GenderFemale Gender = 1
)

type ImCategory int

const (
	QQ ImCategory = iota
	MSN
	GTalk
	Skype
	Other
)

type InternalCategory int

const (
	_InternalCategoryMin InternalCategory = -1
	NormalUser           InternalCategory = 0
	InternalUser         InternalCategory = 1
	TestUser             InternalCategory = 2
	_InternalCategoryMax InternalCategory = 3
)

type TotpStatus int

const (
	TotpStatusDisabled TotpStatus = iota
	TotpStatusEnabled
)

type TotpType int

const (
	TotpTypeAuthenticator TotpType = iota
	TotpTypeMobile
)

type Developer struct {
	Uid             uint32         `json:"uid"`
	Email           string         `json:"email"`
	FullName        string         `json:"fullName"`
	Gender          Gender         `json:"gender"`
	PhoneNumber     string         `json:"phoneNumber"`
	ImNumber        string         `json:"imNumber"`
	ImCategory      ImCategory     `json:"imCategory"`
	WebSite         string         `json:"webSite"`
	CompanyName     string         `json:"companyName"`
	ContractAddress string         `json:"contractAddress"`
	MobileBinded    bool           `json:"mobileBinded"`
	LicenseVersion  LicenseVersion `json:"licenseVersion"`

	Tags []string `json:"tags"`

	RegisterIp       string `json:"registerIp"`
	RegisterState    string `json:"registerState"`
	RegisterRegion   string `json:"registerRegion"`
	RegisterCity     string `json:"registerCity"`
	LocationProvince string `json:"locationProvince"`
	LocationCity     string `json:"locationCity"`

	Referrer           string           `json:"referrer"`
	InviterUid         uint32           `json:"inviterUid"`
	InviteBySales      bool             `json:"inviteBySales"`
	IsActivated        bool             `json:"isActivated"`
	InternalCategory   InternalCategory `json:"internalCategory"`
	InternalDepartment int              `json:"internalDepartment"`

	CreateAt               int64     `json:"createAt"`
	CreatedAtTime          time.Time `json:"createdAtTime"`
	UpdateAt               time.Time `json:"updateAt"`
	UpgradeStdAt           time.Time `json:"upgradeStdAt"`
	UpgradeVipAt           time.Time `json:"upgradeVipAt"`
	LastPasswordModifyTime time.Time `json:"lastPasswordModifyTime"`
	LastEmailModifyTime    time.Time `json:"lastEmailModifyTime"`

	TotpStatus TotpStatus `json:"totpStatus"`
	TotpType   TotpType   `json:"totpType"`

	EmailHistory []string `json:"emailHistory"`

	SfIsEnterprise  bool   `json:"sfIsEnterprise"`
	SfLeadsId       string `json:"sfLeadsId"`
	SfAccountId     string `json:"sfAccountId"`
	SfOpportunityId string `json:"sfOpportunityId"`
	SfSalesId       string `json:"sfSalesId"`
	SfInviteCode    string `json:"sfInviteCode"`
	SfUserId        string `json:"sfUserId"`
}

type DeveloperOverview struct {
	UID      uint32 `json:"uid"`
	Email    string `json:"email"`
	Fullname string `json:"fullname"`

	IsEnterprise bool `json:"is_enterprise"`
	IsCertified  bool `json:"is_certified"`
	IsInternal   bool `json:"is_internal"`
}

type listDeveloper struct {
	Uid            uint32         `json:"uid"`
	Email          string         `json:"email"`
	Fullname       string         `json:"fullname"`
	Gender         Gender         `json:"gender"`
	PhoneNumber    string         `json:"phone_number"`
	ImNumber       string         `json:"im_number"`
	ImCategory     ImCategory     `json:"imCategory"`
	Website        string         `json:"website"`
	CompanyName    string         `json:"company_name"`
	ContactAddress string         `json:"contact_address"`
	MobileBinded   bool           `json:"mobile_binded"`
	LicenseVersion LicenseVersion `json:"license"`

	Tags []string `json:"tags"`

	RegisterIp       string `json:"reg_ip"`
	RegisterState    string `json:"reg_state"`  // 注册国家
	RegisterRegion   string `json:"reg_region"` // 注册省份
	RegisterCity     string `json:"reg_city"`   // 注册城市
	LocationProvince string `json:"location_province"`
	LocationCity     string `json:"location_city"`

	Referrer           string           `json:"referrer"`
	InviterUid         uint32           `json:"inviter_uid"`
	InviteBySales      bool             `json:"invite_by_sales"`
	IsActived          bool             `json:"is_actived"`
	InternalCategory   InternalCategory `json:"internal_category"`
	InternalDepartment int              `json:"internal_department"`

	// 各类时间
	CreatedAt              int64     `json:"created_at"`
	CreatedAtTime          time.Time `json:"created_at_time"`
	UpdatedAt              time.Time `json:"updated_at"`
	UpgradeStdAt           time.Time `json:"upgrade_std_at"`
	UpgradeVipAt           time.Time `json:"upgrade_vip_at"`
	LastPasswordModifyTime time.Time `json:"last_password_modify_time"`
	LastEmailModifyTime    time.Time `json:"last_email_modify_time"`

	// Two-factor authentication
	TotpStatus TotpStatus `json:"totp_status"`
	TotpType   TotpType   `json:"totp_type"`

	EmailHistory []string `json:"email_history"`

	SfIsEnterprise  bool   `json:"sf_is_enterprise"`
	SfLeadsId       string `json:"sf_leads_id"`
	SfAccountId     string `json:"sf_account_id"`
	SfOpportunityId string `json:"sf_opportunity_id"`
	SfSalesId       string `json:"sf_sales_id"`
	SfInviteCode    string `json:"sf_invite_code"`
	SfUserId        string `json:"sf_user_id"`
}

func (d listDeveloper) toDeveloper() Developer {
	return Developer{
		Uid:             d.Uid,
		Email:           d.Email,
		FullName:        d.Fullname,
		Gender:          d.Gender,
		PhoneNumber:     d.PhoneNumber,
		ImNumber:        d.ImNumber,
		ImCategory:      d.ImCategory,
		WebSite:         d.Website,
		CompanyName:     d.CompanyName,
		ContractAddress: d.ContactAddress,
		MobileBinded:    d.MobileBinded,
		LicenseVersion:  d.LicenseVersion,
		Tags:            d.Tags,

		RegisterIp:       d.RegisterIp,
		RegisterState:    d.RegisterState,
		RegisterRegion:   d.RegisterRegion,
		RegisterCity:     d.RegisterCity,
		LocationProvince: d.LocationProvince,
		LocationCity:     d.LocationCity,

		Referrer:           d.Referrer,
		InviterUid:         d.InviterUid,
		InviteBySales:      d.InviteBySales,
		IsActivated:        d.IsActived,
		InternalCategory:   d.InternalCategory,
		InternalDepartment: d.InternalDepartment,

		CreateAt:               d.CreatedAt,
		CreatedAtTime:          d.CreatedAtTime,
		UpdateAt:               d.UpdatedAt,
		UpgradeStdAt:           d.UpgradeStdAt,
		UpgradeVipAt:           d.UpgradeVipAt,
		LastPasswordModifyTime: d.LastPasswordModifyTime,
		LastEmailModifyTime:    d.LastEmailModifyTime,

		TotpStatus: d.TotpStatus,
		TotpType:   d.TotpType,

		EmailHistory: d.EmailHistory,

		SfIsEnterprise:  d.SfIsEnterprise,
		SfLeadsId:       d.SfLeadsId,
		SfAccountId:     d.SfAccountId,
		SfOpportunityId: d.SfOpportunityId,
		SfSalesId:       d.SfSalesId,
		SfInviteCode:    d.SfInviteCode,
		SfUserId:        d.SfUserId,
	}
}

type DeveloperGetParams struct {
	Uid   uint32
	Email string
}

func (p DeveloperGetParams) ToURLValues() url.Values {
	values := url.Values{}
	if p.Uid > 0 {
		values.Add("uid", strconv.FormatUint(uint64(p.Uid), 10))
	}

	if p.Email != "" {
		values.Add("email", p.Email)
	}

	return values
}

type SalesGetParams struct {
	CustomerEmail string
	SalesEmail    string
	Uid           uint32
}

func (p SalesGetParams) ToURLValues() url.Values {
	values := url.Values{}
	if p.Uid > 0 {
		values.Add("uid", strconv.FormatUint(uint64(p.Uid), 10))
	}

	if p.CustomerEmail != "" {
		values.Add("customer_email", p.CustomerEmail)
	}

	if p.SalesEmail != "" {
		values.Add("sales_email", p.SalesEmail)
	}

	return values
}

type Sales struct {
	Name   string `json:"name"`
	Email  string `json:"email"`
	Mobile string `json:"mobile"`
}

type DeveloperUpdateParams struct {
	FullName        string `json:"fullName,omitempty"`
	PhoneNumber     string `json:"phoneNumber,omitempty"`
	CompanyName     string `json:"CompanyName,omitempty"`
	Gender          int    `json:"gender,omitempty"`
	ImCategory      int    `json:"imCategory,omitempty"`
	ImNumber        string `json:"imNumber,omitempty"`
	WebSite         string `json:"webSite,omitempty"`
	ContractAddress string `json:"contractAddress,omitempty"`
	LicenseVersion  string `json:"licenseVersion,omitempty"`
}

type DeveloperCreateParams struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserTypeUpdateParams struct {
	IsDisable     *bool `json:"is_disable"`
	DisableParams *struct {
		DisabledType   account.DisabledType `json:"disabled_type"`
		DisabledReason string               `json:"disabled_reason"`
		Balance        int64                `json:"balance"`
		Consumption    int64                `json:"consumption"`
	} `json:"disable_params"`
	IsStd        *bool `json:"is_std"`
	IsEnterprise *bool `json:"is_enterprise"`
	IsVip        *bool `json:"is_vip"`
	IsParent     *bool `json:"is_parent"`
	ParentParams *struct {
		Domain         string `json:"domain"`
		CanGetChildKey bool   `json:"can_get_child_key"`
	} `json:"parent_params"`
	IsQcos             *bool `json:"is_qcos"`
	IsEvm              *bool `json:"is_evm"`
	IsPili             *bool `json:"is_pili"`
	IsFusion           *bool `json:"is_fusion"`
	IsPandora          *bool `json:"is_pandora"`
	IsDistribution     *bool `json:"is_distribution"`
	IsQvm              *bool `json:"is_qvm"`
	NotificationParams struct {
		NotifyUserWhenDisabled  *bool `json:"notify_user_when_disabled"`
		NotifySalesWhenDisabled *bool `json:"notify_sales_when_disabled"`
		NotifyUserWhenEnabled   *bool `json:"notify_user_when_enabled"`
		NotifySalesWhenEnabled  *bool `json:"notify_sales_when_enabled"`
	} `json:"notification_params"`
}

type DeveloperSFInfoUpdateParams struct {
	SfIsEnterprise  *bool   `json:"sf_is_enterprise"`
	SfLeadsId       *string `json:"sf_leads_id"`
	SfAccountId     *string `json:"sf_account_id"`
	SfOpportunityId *string `json:"sf_opportunity_id"`
	SfSalesId       *string `json:"sf_sales_id"`
	SfInviteCode    *string `json:"sf_invite_code"`
	SfUserId        *string `json:"sf_user_id"`
}

type DeveloperListParams struct {
	PageSize int        `param:"page_size" url:"page_size"`
	From     *time.Time `param:"from" url:"from"`     // updated_at from
	To       *time.Time `param:"to" url:"to"`         // updated_at to
	Marker   *uint32    `param:"marker" url:"marker"` // uid
}

func (p *DeveloperListParams) Values() url.Values {
	values := url.Values{}
	values.Set("page_size", strconv.Itoa(p.PageSize))

	if p.From != nil {
		values.Set("from", p.From.Format(time.RFC3339))
	}

	if p.To != nil {
		values.Set("to", p.To.Format(time.RFC3339))
	}

	if p.Marker != nil {
		values.Set("marker", strconv.FormatUint(uint64(*p.Marker), 10))
	}

	return values
}
