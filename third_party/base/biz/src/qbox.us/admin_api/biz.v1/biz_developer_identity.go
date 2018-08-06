package biz

import (
	"time"
)

const (
	_DeveloperIdentityStatusMin         DeveloperIdentityStatus = iota
	DeveloperIdentityPending                                    // 待审核
	DeveloperIdentityFailed                                     // 认证失败
	DeveloperIdentitySuccess                                    // 认证通过
	DeveloperIdentityExchange                                   // 更换成企业认证
	DeveloperIdentityEnterpriseExchange                         // 企业认证更换联系人信息
	_DeveloperIdentityStatusMax
)

type DeveloperIdentityStatus int

func (s DeveloperIdentityStatus) Humanize() string {
	switch s {
	case DeveloperIdentityPending:
		return "待审核"
	case DeveloperIdentityFailed:
		return "认证失败"
	case DeveloperIdentitySuccess:
		return "认证通过"
	case DeveloperIdentityExchange, DeveloperIdentityEnterpriseExchange:
		return "更换认证，等待用户重新提交认证信息"
	default:
		return "未设置"
	}
}

func (s DeveloperIdentityStatus) IsValid() bool {
	if s <= _DeveloperIdentityStatusMin || s >= _DeveloperIdentityStatusMax {
		return false
	}
	return true
}

func (s DeveloperIdentityStatus) IsPending() bool {
	return s == DeveloperIdentityPending
}

func (s DeveloperIdentityStatus) IsFailed() bool {
	return s == DeveloperIdentityFailed
}

func (s DeveloperIdentityStatus) IsSuccess() bool {
	return s == DeveloperIdentitySuccess
}

func (s DeveloperIdentityStatus) IsExchange() bool {
	return s == DeveloperIdentityExchange || s == DeveloperIdentityEnterpriseExchange
}

type AlipayUserType string

const (
	AlipayUserTypePersonal AlipayUserType = "personal"
	AlipayUserTypeCompany  AlipayUserType = "company"
)

func (t AlipayUserType) Humanize() string {
	switch t {
	case AlipayUserTypePersonal:
		return "个人"
	case AlipayUserTypeCompany:
		return "公司"
	}
	return "未知类型"
}

type DeveloperIdentity struct {
	Uid                   uint32                  `json:"uid"`
	AlipayUid             string                  `json:"alipay_uid"`               //支付宝认证的用户id
	AlipayUserType        AlipayUserType          `json:"alipay_user_type"`         //支付宝认证的用户类型（personal / company）
	EnterpriseName        string                  `json:"enterprise_name"`          // 名称
	BusinessLicenseNo     string                  `json:"business_license_no"`      // 营业执照注册号
	OrganizationNo        string                  `json:"organization_no"`          // 组织机构代码
	BusinessLicenseCopy   string                  `json:"business_license_copy"`    // 营业执照副本扫描件
	ContactName           string                  `json:"contact_name"`             // 联系人姓名
	ContactIdentityNo     string                  `json:"contact_identity_no"`      // 联系人身份证号码
	ContactIdentityPhoto  string                  `json:"contact_identity_photo"`   // 联系人身份证持证照片
	ContactIdentityPhotoB string                  `json:"contact_identity_photo_b"` // 联系人身份证持证背面照片
	ContactAddress        string                  `json:"contact_address"`          // 联系地址
	ContactProvince       string                  `json:"contact_province"`         // 所在省
	ContactCity           string                  `json:"contact_city"`             // 所在市
	ContactRegion         string                  `json:"contact_region"`           // 所在区
	Status                DeveloperIdentityStatus `json:"status"`                   // 状态
	StatusNote            string                  `json:"status_note"`              // 状态信息
	IsEnterprise          bool                    `json:"is_enterprise"`            // 是企业认证
	Memo                  string                  `json:"memo"`                     // 备忘
	CreatedAt             time.Time               `json:"created_at"`               // 创建时间
	UpdatedAt             time.Time               `json:"updated_at"`               // 更新时间
	Versions              []DeveloperIdentity     `json:"versions"`                 // 存旧数据
}
