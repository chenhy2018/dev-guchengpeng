package enums

type InvoiceType uint32

const (
	// invoice type
	NormalInvoice InvoiceType = 1
	VATInvoice    InvoiceType = 2
)

func (e InvoiceType) Humanize() string {
	switch e {
	case NormalInvoice:
		return "普通发票"
	case VATInvoice:
		return "增值税专票"
	default:
		return ""
	}
}

type InvoiceSettingStatus int

const (
	InvoiceSettingAudit = iota + 1
	InvoiceSettingInvalid
	InvoiceSettingValid
)

func (s InvoiceSettingStatus) String() string {
	switch s {
	case InvoiceSettingAudit:
		return "审核中"
	case InvoiceSettingValid:
		return "审核通过"
	case InvoiceSettingInvalid:
		return "驳回"
	default:
		return "未知"
	}
}

type InvoiceStatus int

const (
	InvoiceCreated InvoiceStatus = iota + 1
	InvoiceAuditing
	InvoiceRejected
	InvoiceCompleted
	InvoiceDelivery
	InvoiceRefund
)

func (s InvoiceStatus) String() string {
	switch s {
	case InvoiceCreated:
		return "已申请"
	case InvoiceAuditing:
		return "待审核"
	case InvoiceRejected:
		return "申请已驳回"
	case InvoiceCompleted:
		return "已开票"
	case InvoiceDelivery:
		return "已配送"
	case InvoiceRefund:
		return "已退票"
	default:
		return "未知"
	}
}

type CoTaxType int

const (
	TaxPayerIdentityNum = iota + 1
	UnifiedSocialCreditCode
)

func (s CoTaxType) String() string {
	switch s {
	case TaxPayerIdentityNum:
		return "纳税人识别号"
	case UnifiedSocialCreditCode:
		return "统一社会信用代码"
	default:
		return "未知"
	}
}
