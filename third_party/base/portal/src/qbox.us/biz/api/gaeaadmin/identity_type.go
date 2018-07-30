package gaeaadmin

import (
	"net/url"
	"strconv"
	"time"

	"labix.org/v2/mgo/bson"
)

const (
	_DeveloperIdentityStatusMin    DeveloperIdentityStatus = iota
	DeveloperIdentityPending                               // 待审核
	DeveloperIdentityFailed                                // 认证失败
	DeveloperIdentitySuccess                               // 认证通过
	DeveloperIdentityUpgrading                             // 认证升级中
	DeveloperIdentityUpgradeFailed                         // 认证升级失败
	_DeveloperIdentityStatusMax
)

const (
	ErrNotFound                   apiError = "not found"
	ErrInvalidArgs                apiError = "invalid args"
	ErrForbidden                  apiError = "forbidden"
	ErrConflict                   apiError = "entry exist"
	ErrContactIdConflict          apiError = "entry exist:contact_identity_no"
	ErrEnterpriseCodeConflict     apiError = "entry exist:enterprise_code"
	ErrAlipayUidConflict          apiError = "entry exist:alipay_uid"
	ErrResultError                apiError = "response result error"
	ErrDatabaseError              apiError = "database err"
	ErrAlreadyUrgedToday          apiError = "already urged today"
	ErrIdentityMaxRetry           apiError = "identity max retry"
	ErrIdentityVerifyBankMaxRetry apiError = "identity verify bank transfer max retry"
	ErrIdentityVerifyBankFail     apiError = "identity verify bank transfer fail"
)

type apiError string

func (a apiError) Error() string {
	return string(a)
}

type apiResultBase struct {
	Code int    `json:"code"`
	Msg  string `json:"message"`
}

func (a *apiResultBase) OK() bool {
	return a.Code == 200
}

func (a *apiResultBase) Error() error {
	if a.OK() {
		return nil
	}

	return apiError(a.Msg)
}

type DeveloperIdentityStatus int

func (s DeveloperIdentityStatus) Humanize() string {
	switch s {
	case DeveloperIdentityPending:
		return "待审核"
	case DeveloperIdentityFailed:
		return "认证失败"
	case DeveloperIdentitySuccess:
		return "认证通过"
	case DeveloperIdentityUpgrading:
		return "升级中"
	case DeveloperIdentityUpgradeFailed:
		return "升级失败"
	default:
		return "未知状态"
	}
}

func (s DeveloperIdentityStatus) Valid() bool {
	if s <= _DeveloperIdentityStatusMin || s >= _DeveloperIdentityStatusMax {
		return false
	}
	return true
}

// 是否通过
func (s DeveloperIdentityStatus) Passed() bool {
	return s == DeveloperIdentitySuccess
}

// 是否被拒绝
func (s DeveloperIdentityStatus) Rejected() bool {
	return s == DeveloperIdentityFailed || s == DeveloperIdentityUpgradeFailed
}

// 是否认证中（包含升级）
func (s DeveloperIdentityStatus) InProcess() bool {
	return s == DeveloperIdentityPending || s == DeveloperIdentityUpgrading
}

// 是否已认证
func (s DeveloperIdentityStatus) Identitied() bool {
	return s > DeveloperIdentityFailed
}

type IdentityType int

const (
	_IdentityTypeMin                          IdentityType = iota
	IdentityTypePersonalManual                             // 个人手动认证
	IdentityTypePersonalAlipay                             // 个人支付宝认证
	IdentityTypeEnterpriseBusiness                         // 企业营业执照号银行认证
	IdentityTypeEnterpriseOrganization                     // 企业组织机构代码银行认证
	IdentityTypeEnterpriseUnifiedSocial                    // 企业社会信用号码银行认证
	IdentityTypeEnterpriseBusinessAlipay                   // 企业营业执照号支付宝认证
	IdentityTypeEnterpriseOrganizationAlipay               // 企业组织机构代码支付宝认证
	IdentityTypeEnterpriseUnifiedSocialAlipay              // 企业社会信用号码支付宝认证
	IdentityTypeSalesGuarantee                             // 销售担保认证 = 企业认证
	_IdentityTypeMax
)

func (i IdentityType) Valid() bool {
	return i > _IdentityTypeMin && i < _IdentityTypeMax
}

func (i IdentityType) IsEnterprise() bool {
	return i == IdentityTypeEnterpriseBusiness ||
		i == IdentityTypeEnterpriseOrganization ||
		i == IdentityTypeEnterpriseUnifiedSocial ||
		i == IdentityTypeEnterpriseBusinessAlipay ||
		i == IdentityTypeEnterpriseOrganizationAlipay ||
		i == IdentityTypeEnterpriseUnifiedSocialAlipay ||
		i == IdentityTypeSalesGuarantee
}

func (i IdentityType) IsPersonal() bool {
	return i == IdentityTypePersonalManual || i == IdentityTypePersonalAlipay
}

func (i IdentityType) IsAlipayVerification() bool {
	return i == IdentityTypePersonalAlipay ||
		i == IdentityTypeEnterpriseBusinessAlipay ||
		i == IdentityTypeEnterpriseOrganizationAlipay ||
		i == IdentityTypeEnterpriseUnifiedSocialAlipay
}

func (i IdentityType) IsBankVerification() bool {
	return i == IdentityTypePersonalManual ||
		i == IdentityTypeEnterpriseBusiness ||
		i == IdentityTypeEnterpriseOrganization ||
		i == IdentityTypeEnterpriseUnifiedSocial
}

func (i IdentityType) IsSalesGuarantee() bool {
	return i == IdentityTypeSalesGuarantee
}

func (i IdentityType) Humanize() string {
	switch i {
	case IdentityTypePersonalManual:
		return "个人手动认证"
	case IdentityTypePersonalAlipay:
		return "个人支付宝认证"
	case IdentityTypeEnterpriseBusiness:
		return "企业营业执照认证"
	case IdentityTypeEnterpriseOrganization:
		return "企业组织机构代码认证"
	case IdentityTypeEnterpriseUnifiedSocial:
		return "企业社会信用号码认证"
	case IdentityTypeEnterpriseBusinessAlipay:
		return "企业营业执照支付宝认证"
	case IdentityTypeEnterpriseOrganizationAlipay:
		return "企业组织机构代码支付宝认证"
	case IdentityTypeEnterpriseUnifiedSocialAlipay:
		return "企业社会信用号码支付宝认证"
	default:
		return "无效类型"
	}
}

type IdentityStep int

const (
	IdentityStepDone                IdentityStep = iota // 已完成
	IdentityStepBasicInfoReviewing                      // 基本信息审核中
	IdentityStepBankTransferWaiting                     // 银行转账处理中
	IdentityStepBankInfoVerifying                       // 银行转账金额验证中
)

func (s IdentityStep) IsBasicInfoReviweing() bool {
	return s == IdentityStepBasicInfoReviewing
}

func (s IdentityStep) IsBankTransferWaiting() bool {
	return s == IdentityStepBankTransferWaiting
}

func (s IdentityStep) IdentityStepBankInfoVerifying() bool {
	return s == IdentityStepBankInfoVerifying
}

func (s IdentityStep) NextStep() IdentityStep {
	switch s {
	case IdentityStepBasicInfoReviewing:
		return IdentityStepBankTransferWaiting
	case IdentityStepBankTransferWaiting:
		return IdentityStepBankInfoVerifying
	case IdentityStepBankInfoVerifying:
		return IdentityStepDone
	}
	return IdentityStepDone
}

func (s IdentityStep) HasNextStep() bool {
	return s.NextStep() != s
}

type DeveloperIdentity struct {
	Id                       bson.ObjectId           `json:"id"`
	Uid                      uint32                  `json:"uid"`
	Type                     IdentityType            `json:"type"`
	EnterpriseCode           string                  `json:"enterprise_code"`              // 企业认证码
	EnterpriseCodeCopy       string                  `json:"enterprise_code_copy"`         // 企业认证码扫描件
	EnterpriseCodeCopyURL    string                  `json:"enterprise_code_copy_url"`     // 完整可访问图片地址
	AppName                  string                  `json:"app_name"`                     // 产品名称或网站URL
	EnterpriseName           string                  `json:"enterprise_name"`              // 名称
	EnterpriseIndustry       int                     `json:"enterprise_industry"`          // 公司所在行业
	ContactName              string                  `json:"contact_name"`                 // 联系人姓名
	ContactWebsite           string                  `json:"contact_website"`              // 联系人网站
	ContactIdentityNo        string                  `json:"contact_identity_no"`          // 联系人身份证号码
	ContactIdentityPhoto     string                  `json:"contact_identity_photo"`       // 联系人身份证持证照片
	ContactIdentityPhotoURL  string                  `json:"contact_identity_photo_url"`   // 完整可访问图片地址
	ContactIdentityPhotoB    string                  `json:"contact_identity_photo_b"`     // 联系人身份证持证背面照片
	ContactIdentityPhotoBURL string                  `json:"contact_identity_photo_b_url"` // 完整可访问图片地址
	ContactAddress           string                  `json:"contact_address"`              // 联系地址
	ContactProvince          string                  `json:"contact_province"`             // 所在省
	ContactCity              string                  `json:"contact_city"`                 // 所在市
	ContactRegion            string                  `json:"contact_region"`               // 所在区
	AlipayUid                string                  `json:"alipay_uid"`                   // 支付宝认证用户在支付宝的 Uid
	AlipayUserType           string                  `json:"alipay_user_type"`             // 支付宝认证用户的用户类型
	Step                     IdentityStep            `json:"step"`                         // 步骤，See: https://jira.qiniu.io/browse/BO-2266
	Status                   DeveloperIdentityStatus `json:"status"`                       // 状态
	StatusNote               string                  `json:"status_note"`                  // 状态信息
	Memo                     string                  `json:"memo"`                         // 备忘
	OperatorEmail            string                  `json:"operator_email"`               // 操作者email
	CreatorEmail             string                  `json:"creator_email"`                // 创建者
	CreatedAt                time.Time               `json:"created_at"`                   // 创建时间
	UpdatedAt                time.Time               `json:"updated_at"`                   // 更新时间

	// 新版身份认证流程相关
	// https://jira.qiniu.io/browse/BO-2257
	CompanyAccount             string `json:"company_account"`               // 对公账户
	CompanyAccountName         string `json:"company_account_name"`          // 对公账户户名
	CompanyAccountDepositBank  string `json:"company_account_deposit_bank"`  // 对公账户开户行
	PersonalAccount            string `json:"personal_account"`              // 个人账户
	PersonalAccountName        string `json:"personal_account_name"`         // 个人账户户名
	PersonalAccountDepositBank string `json:"personal_account_deposit_bank"` // 个人账户开户行
	IdentityRetryCount         int    `json:"identity_retry_count"`          // 验证重试次数
	IdentityRemains            int    `json:"identity_remains"`              // 可重新提交身份验证的剩余次数
	BankRemains                int    `json:"bank_remains"`                  // 可重新验证银行款项的剩余次数
}

func (m *DeveloperIdentity) IsEnterprise() bool {
	return m.Type.IsEnterprise()
}

func (m *DeveloperIdentity) IsPersonal() bool {
	return m.Type.IsPersonal()
}

func (m *DeveloperIdentity) IsAlipayVerification() bool {
	return m.Type.IsAlipayVerification()
}

func (m *DeveloperIdentity) IsBankVerification() bool {
	return m.Type.IsBankVerification()
}

func (m *DeveloperIdentity) EnterpriseCodeValidate() bool {
	return m.IsEnterprise() && m.EnterpriseCode != ""
}

type IdentityCreateParams struct {
	Uid                   uint32                   `json:"uid"`
	Type                  IdentityType             `json:"type"`
	EnterpriseName        string                   `json:"enterprise_name"`
	EnterpriseIndustry    int                      `json:"enterprise_industry"`
	EnterpriseCode        string                   `json:"enterprise_code"`
	EnterpriseCodeCopy    string                   `json:"enterprise_code_copy"`
	AppName               string                   `json:"app_name"`
	ContactAddress        string                   `json:"contact_address"`
	ContactName           string                   `json:"contact_name"`
	ContactWebsite        string                   `json:"contact_website"`
	ContactIdentityNo     string                   `json:"contact_identity_no"`
	ContactIdentityPhoto  string                   `json:"contact_identity_photo"`
	ContactIdentityPhotoB string                   `json:"contact_identity_photo_b"`
	ContactProvince       string                   `json:"contact_province"`
	ContactCity           string                   `json:"contact_city"`
	AlipayUid             string                   `json:"alipay_uid"`
	AlipayUserType        string                   `json:"alipay_user_type"`
	Status                *DeveloperIdentityStatus `json:"status"`
	StatusNote            *string                  `json:"status_note"`
	Memo                  *string                  `json:"memo"`
	OperatorEmail         *string                  `json:"operator_email"`

	CompanyAccount             string `json:"company_account"`               // 对公账户
	CompanyAccountName         string `json:"company_account_name"`          // 对公账户户名
	CompanyAccountDepositBank  string `json:"company_account_deposit_bank"`  // 对公账户开户行
	PersonalAccount            string `json:"personal_account"`              // 个人账户
	PersonalAccountName        string `json:"personal_account_name"`         // 个人账户户名
	PersonalAccountDepositBank string `json:"personal_account_deposit_bank"` // 个人账户开户行
}

type IdentityUpdateParams struct {
	Type                  *IdentityType            `json:"type"`
	EnterpriseName        *string                  `json:"enterprise_name"`
	EnterpriseIndustry    *int                     `json:"enterprise_industry"`
	EnterpriseCode        *string                  `json:"enterprise_code"`
	EnterpriseCodeCopy    *string                  `json:"enterprise_code_copy"`
	AppName               *string                  `json:"app_name"`
	ContactAddress        *string                  `json:"contact_address"`
	ContactName           *string                  `json:"contact_name"`
	ContactWebsite        *string                  `json:"contact_website"`
	ContactIdentityNo     *string                  `json:"contact_identity_no"`
	ContactIdentityPhoto  *string                  `json:"contact_identity_photo"`
	ContactIdentityPhotoB *string                  `json:"contact_identity_photo_b"`
	ContactProvince       *string                  `json:"contact_province"`
	ContactCity           *string                  `json:"contact_city"`
	AlipayUid             *string                  `json:"alipay_uid"`
	AlipayUserType        *string                  `json:"alipay_user_type"`
	Status                *DeveloperIdentityStatus `json:"status"`
	StatusNote            *string                  `json:"status_note"`
	Memo                  *string                  `json:"memo"`
	OperatorEmail         *string                  `json:"operator_email"`

	CompanyAccount             *string `json:"company_account"`               // 对公账户
	CompanyAccountName         *string `json:"company_account_name"`          // 对公账户户名
	CompanyAccountDepositBank  *string `json:"company_account_deposit_bank"`  // 对公账户开户行
	PersonalAccount            *string `json:"personal_account"`              // 个人账户
	PersonalAccountName        *string `json:"personal_account_name"`         // 个人账户户名
	PersonalAccountDepositBank *string `json:"personal_account_deposit_bank"` // 个人账户开户行
}

type IdentityListParams struct {
	Email             *string                  `param:"email"`
	Uid               *uint32                  `param:"uid"`
	Page              int                      `param:"page"`
	PageSize          int                      `param:"page_size"`
	Type              *IdentityType            `param:"type"`
	EnterpriseCode    *string                  `param:"enterprise_code"`
	EnterpriseName    *string                  `param:"enterprise_name"`
	ContactName       *string                  `param:"contact_name"`
	ContactIdentityNo *string                  `param:"contact_identity_no"`
	AlipayUid         *string                  `param:"alipay_uid"`
	OperatorEmail     *string                  `param:"operator_email"`
	CreatorEmail      *string                  `param:"creator_email"`
	From              *time.Time               `param:"from"`
	To                *time.Time               `param:"to"`
	Status            *DeveloperIdentityStatus `param:"status"`
}

func (p IdentityListParams) Values() url.Values {
	values := url.Values{}

	values.Set("page", strconv.Itoa(p.Page))
	values.Set("page_size", strconv.Itoa(p.PageSize))

	if p.Email != nil {
		values.Set("email", *p.Email)
	}

	if p.Uid != nil {
		values.Set("uid", strconv.FormatUint(uint64(*p.Uid), 10))
	}

	if p.Type != nil {
		values.Set("type", strconv.Itoa(int(*p.Type)))
	}

	if p.EnterpriseCode != nil {
		values.Set("enterprise_code", *p.EnterpriseCode)
	}

	if p.EnterpriseName != nil {
		values.Set("enterprise_name", *p.EnterpriseName)
	}

	if p.ContactName != nil {
		values.Set("contact_name", *p.ContactName)
	}

	if p.ContactIdentityNo != nil {
		values.Set("contact_identity_no", *p.ContactIdentityNo)
	}

	if p.AlipayUid != nil {
		values.Set("alipay_uid", *p.AlipayUid)
	}

	if p.OperatorEmail != nil {
		values.Set("operator_email", *p.OperatorEmail)
	}

	if p.CreatorEmail != nil {
		values.Set("creator_email", *p.CreatorEmail)
	}

	if p.From != nil {
		values.Set("from", p.From.Format(time.RFC3339))
	}

	if p.To != nil {
		values.Set("to", p.To.Format(time.RFC3339))
	}

	if p.Status != nil {
		values.Set("status", strconv.Itoa(int(*p.Status)))
	}

	return values
}

type IdentityReviewParams struct {
	Pass       bool   `json:"pass"`
	StatusNote string `json:"status_note"`
	Memo       string `json:"memo"`
}

type IdentityUpToken struct {
	UpHost   string `json:"up_host"`
	DownHost string `json:"down_host"`
	Token    string `json:"token"`
}

/************************************************/
/******       Identity Bank Transfer       ******/
/************************************************/

type DeveloperIdentityBankTransferStatus int

const (
	DeveloperIdentityBankTransferStatusNew     DeveloperIdentityBankTransferStatus = iota // 新建
	DeveloperIdentityBankTransferStatusPending                                            // 处理中
	DeveloperIdentityBankTransferStatusSuccess                                            // 成功
	DeveloperIdentityBankTransferStatusFailed                                             // 失败
)

func (s DeveloperIdentityBankTransferStatus) IsNew() bool {
	return s == DeveloperIdentityBankTransferStatusNew
}

func (s DeveloperIdentityBankTransferStatus) IsPending() bool {
	return s == DeveloperIdentityBankTransferStatusPending
}

func (s DeveloperIdentityBankTransferStatus) IsSuccessful() bool {
	return s == DeveloperIdentityBankTransferStatusSuccess
}

func (s DeveloperIdentityBankTransferStatus) IsFailed() bool {
	return s == DeveloperIdentityBankTransferStatusFailed
}

type IdentityBankTransfer struct {
	ID                      bson.ObjectId                       `json:"id" bson:"_id"`
	UID                     uint32                              `json:"uid" bson:"uid"`                                               // 关联的用户 uid
	IdentityID              bson.ObjectId                       `json:"identity_id" bson:"identity_id"`                               // 关联的身份认证 id
	IsEnterprise            bool                                `json:"is_enterprise" bson:"is_enterprise"`                           // 是否是企业认证
	PayerAccount            string                              `json:"payer_account" bson:"payer_account"`                           // 付款人账号
	PayeeAccount            string                              `json:"payee_account" bson:"payee_account"`                           // 收款人账号
	PayeeAccountName        string                              `json:"payee_account_name" bson:"payee_account_name"`                 // 收款人户名
	PayeeAccountDepositBank string                              `json:"payee_account_deposit_bank" bson:"payee_account_deposit_bank"` // 收款人开户行名称
	Amount                  float64                             `json:"amount" bson:"amount"`                                         // 转账金额
	Memo                    string                              `json:"memo" bson:"memo"`                                             // 转账备注
	Status                  DeveloperIdentityBankTransferStatus `json:"status" bson:"status"`                                         // 状态
	StatusNote              string                              `json:"status_note" bson:"status_note"`                               // 状态备注
	VerifyCount             int                                 `json:"verify_count" bson:"verify_count"`                             // 尝试校验次数
	ExportHistory           []string                            `json:"export_history" bson:"export_history"`                         // 历史导出批次，是过滤记录的重要条件
	CreatedAt               time.Time                           `json:"created_at" bson:"created_at"`                                 // 创建时间
	UpdatedAt               time.Time                           `json:"updated_at" bson:"updated_at"`                                 // 更新时间
	PayeeCategory           int                                 `json:"payee_category"`                                               // 收款人性质，0 - 浦发，1 - 非浦发
	PaymentPurpose          string                              `json:"payment_purpose"`                                              // 付款用途，个人一律选 5
	PaymentPath             int                                 `json:"payment_path"`                                                 // 汇路选择，0 - 同城，1 - 异地
	ShortcutFlag            string                              `json:"shortcut_flag"`                                                // 速选标志，汇路是 1 就是 1, 否则留空
}

type IdentityBankVerifyParams struct {
	Money float64 `json:"money"`
}

type IdentityBankVerifyOut struct {
	Remains int `json:"remains"`
}

type IdentityBankTransferListParams struct {
	UID              *uint32    `param:"uid"`
	PayeeAccount     *string    `param:"payee_account"`
	PayeeAccountName *string    `param:"payee_account_name"`
	CreatedAtFrom    *time.Time `param:"created_at_from"`
	CreatedAtTo      *time.Time `param:"created_at_to"`
	IsEnterprise     *bool      `param:"is_enterprise"`
	Status           *int       `param:"status"`
	Version          string     `param:"version"`
	Page             int        `param:"page"`
	PageSize         int        `param:"page_size"`
}

func (p IdentityBankTransferListParams) Values() url.Values {
	values := url.Values{}

	if p.UID != nil {
		values.Add("uid", strconv.Itoa(int(*p.UID)))
	}

	if p.PayeeAccount != nil {
		values.Add("payee_account", *p.PayeeAccount)
	}

	if p.PayeeAccountName != nil {
		values.Add("payee_account_name", *p.PayeeAccountName)
	}

	if p.CreatedAtFrom != nil {
		values.Add("created_at_from", p.CreatedAtFrom.Format(time.RFC3339))
	}

	if p.CreatedAtTo != nil {
		values.Add("created_at_to", p.CreatedAtTo.Format(time.RFC3339))
	}

	if p.IsEnterprise != nil {
		values.Add("is_enterprise", strconv.FormatBool(*p.IsEnterprise))
	}

	if p.Status != nil {
		values.Add("status", strconv.Itoa(*p.Status))
	}

	if p.Version != "" {
		values.Add("version", p.Version)
	}

	if p.Page >= 0 {
		values.Add("page", strconv.Itoa(p.Page))
	}

	if p.PageSize > 0 {
		values.Add("page_size", strconv.Itoa(p.PageSize))
	}

	return values
}

type IdentityBankTransferBatchUpdateParams struct {
	IDs           []bson.ObjectId                               `json:"ids"`
	SetStatus     *IdentityBankTransferBatchUpdateSetStatus     `json:"set_status"`
	AppendVersion *IdentityBankTransferBatchUpdateAppendVersion `json:"append_version"`
}

type IdentityBankTransferBatchUpdateSetStatus struct {
	Status DeveloperIdentityBankTransferStatus `json:"status"`
	Reason string                              `json:"reason"`
}

type IdentityBankTransferBatchUpdateAppendVersion struct {
	Version string `json:"version"`
}
