package gaeaadmin

import (
	"fmt"
	"time"

	"labix.org/v2/mgo/bson"
	"qbox.us/api/pay/pay"
	"qbox.us/biz/api/gaeaadmin/enums"
)

const (
	GetInvoiceDeliveryPath          = "%s/api/finance/invoice/delivery/%d"
	SetInvoiceDeliveryPath          = "%s/api/finance/invoice/delivery"
	GetInvoiceUploadTokenPath       = "%s/api/finance/invoice/upload-token"
	CreateInvoiceSettingPath        = "%s/api/finance/invoice/setting/%d"
	UpdateInvoiceSettingPath        = "%s/api/finance/invoice/setting?id=%s"
	GetInvoiceSettingPath           = "%s/api/finance/invoice/setting?id=%s"
	DeleteInvoiceSettingPath        = "%s/api/finance/invoice/setting?id=%s"
	ListInvoiceSettingByUidPath     = "%s/api/finance/invoice/setting/%d/list?page=%d&page_size=%d"
	SetDefaultInvoiceSettingPath    = "%s/api/finance/invoice/setting/%d/default?id=%s"
	GetDefaultInvoiceSettingPath    = "%s/api/finance/invoice/setting/%d/default"
	GetInvoiceValidTransactionsPath = "%s/api/finance/invoice/%d/valid-transactions?start=%d&end=%d"
	CreateInvoicePath               = "%s/api/finance/invoice/%d"
	GetInvoicePath                  = "%s/api/finance/invoice/detail/%s"
	ListInvoiceByUidPath            = "%s/api/finance/invoice/%d/list?page=%d&page_size=%d"
	UpdateInvoicePath               = "%s/api/finance/invoice/detail/%s"
)

type InvoiceDelivery struct {
	Uid      uint32    `json:"uid"`
	Address  string    `json:"address"`
	Phone    string    `json:"phone"`
	Name     string    `json:"name"`
	UpdateAt time.Time `json:"update_at"`
}

func (s *gaeaAdminService) GetInvoiceDelivery(uid uint32) (delivery InvoiceDelivery, err error) {
	var (
		resp struct {
			apiResultBase
			Data InvoiceDelivery `json:"data"`
		}
		api = fmt.Sprintf(GetInvoiceDeliveryPath, s.host, uid)
	)
	delivery.UpdateAt = time.Now()
	err = s.client.GetCall(s.reqLogger, &resp, api)
	if err != nil || !resp.OK() {
		err = resp.Error()
		return
	}

	delivery = resp.Data
	return
}

func (s *gaeaAdminService) SetInvoiceDelivery(delivery InvoiceDelivery) (err error) {
	var (
		resp struct {
			apiResultBase
			Data InvoiceDelivery `json:"data"`
		}
		api = fmt.Sprintf(SetInvoiceDeliveryPath, s.host)
	)

	err = s.client.CallWithJson(s.reqLogger, &resp, api, delivery)

	if err != nil || !resp.OK() {
		err = resp.Error()
		return
	}

	return
}

type InvoiceUploadToken struct {
	Domain string `json:"domain"`
	Token  string `json:"token"`
}

func (s *gaeaAdminService) GetInvoiceUploadToken() (token InvoiceUploadToken, err error) {
	var (
		resp struct {
			apiResultBase
			Data InvoiceUploadToken `json:"data"`
		}
		api = fmt.Sprintf(GetInvoiceUploadTokenPath, s.host)
	)

	err = s.client.GetCall(s.reqLogger, &resp, api)
	if err != nil || !resp.OK() {
		err = resp.Error()
		return
	}

	token = resp.Data

	return
}

type InvoiceSettingInput struct {
	Title          string                     `json:"title"`
	Type           enums.InvoiceType          `json:"type"`
	Status         enums.InvoiceSettingStatus `json:"status"`
	TaxPayerId     *string                    `json:"tax_payer_id"`
	BankName       *string                    `json:"bank_name"`
	BankAccount    *string                    `json:"bank_account"`
	Company        *string                    `json:"company"`
	CompanyAddress *string                    `json:"company_address"`
	CompanyTel     *string                    `json:"company_tel"`
	CertificateUrl *string                    `json:"certificate_url"`
	Reason         *string                    `json:"reason"`
	Comment        *string                    `json:"comment"`
	CoTaxType      enums.CoTaxType            `json:"co_tax_type"` // 公司纳税类型
	CoTaxId        *string                    `json:"co_tax_id"`   // 公司纳税号码
}

type InvoiceSettingOutput struct {
	Id             bson.ObjectId              `json:"id"`
	Uid            uint32                     `json:"uid"`
	Title          string                     `json:"title"`
	Type           enums.InvoiceType          `json:"type"`
	Status         enums.InvoiceSettingStatus `json:"status"`
	TaxPayerId     string                     `json:"tax_payer_id"`
	BankName       string                     `json:"bank_name"`
	BankAccount    string                     `json:"bank_account"`
	Company        string                     `json:"company"`
	CompanyAddress string                     `json:"company_address"`
	CompanyTel     string                     `json:"company_tel"`
	CertificateUrl string                     `json:"certificate_url"`
	Reason         string                     `json:"reason"`
	Comment        string                     `json:"comment"`
	IsDefault      bool                       `json:"is_default"`
	UpdatedAt      time.Time                  `json:"updated_at"`
	CoTaxType      enums.CoTaxType            `json:"co_tax_type"` // 公司纳税类型
	CoTaxId        string                     `json:"co_tax_id"`   // 公司纳税号码
}

func (s *gaeaAdminService) CreateInvoiceSetting(uid uint32, input InvoiceSettingInput) (id string, err error) {
	var (
		resp struct {
			apiResultBase
			Data string `json:"data"`
		}
		api = fmt.Sprintf(CreateInvoiceSettingPath, s.host, uid)
	)

	err = s.client.CallWithJson(s.reqLogger, &resp, api, input)

	if err != nil || !resp.OK() {
		err = resp.Error()
		return
	}

	id = resp.Data

	return
}

func (s *gaeaAdminService) UpdateInvoiceSetting(id string, input InvoiceSettingInput) (err error) {
	var (
		resp struct {
			apiResultBase
		}
		api = fmt.Sprintf(UpdateInvoiceSettingPath, s.host, id)
	)

	err = s.client.PutCallWithJson(s.reqLogger, &resp, api, input)

	if err != nil || !resp.OK() {
		err = resp.Error()
		return
	}

	return
}

func (s *gaeaAdminService) GetInvoiceSetting(id string) (out InvoiceSettingOutput, err error) {
	var (
		resp struct {
			apiResultBase
			Data InvoiceSettingOutput `json:"data"`
		}
		api = fmt.Sprintf(GetInvoiceSettingPath, s.host, id)
	)

	err = s.client.GetCall(s.reqLogger, &resp, api)

	if err != nil || !resp.OK() {
		err = resp.Error()
		return
	}

	out = resp.Data
	return
}

func (s *gaeaAdminService) DeleteInvoiceSetting(id string) (err error) {
	var (
		resp apiResultBase
		api  = fmt.Sprintf(DeleteInvoiceSettingPath, s.host, id)
	)

	err = s.client.DeleteCall(s.reqLogger, &resp, api)
	if err != nil || !resp.OK() {
		err = resp.Error()
		return
	}

	return
}

func (s *gaeaAdminService) ListInvoiceSettingBuyUid(uid uint32, page, page_size int) (outs []InvoiceSettingOutput, err error) {
	var (
		resp struct {
			apiResultBase
			Data []InvoiceSettingOutput `json:"data"`
		}
		api = fmt.Sprintf(ListInvoiceSettingByUidPath, s.host, uid, page, page_size)
	)

	err = s.client.GetCall(s.reqLogger, &resp, api)

	if err != nil || !resp.OK() {
		err = resp.Error()
		return
	}

	outs = resp.Data
	return
}

func (s *gaeaAdminService) SetDefaultInvoiceSetting(uid uint32, id string) (err error) {
	var (
		resp apiResultBase
		api  = fmt.Sprintf(SetDefaultInvoiceSettingPath, s.host, uid, id)
	)

	err = s.client.PutCallWithJson(s.reqLogger, &resp, api, nil)

	if err != nil || !resp.OK() {
		err = resp.Error()
		return
	}

	return
}

func (s *gaeaAdminService) GetDefaultInvoiceSetting(uid uint32) (out InvoiceSettingOutput, err error) {
	var (
		resp struct {
			apiResultBase
			Data InvoiceSettingOutput `json:"data"`
		}
		api = fmt.Sprintf(GetDefaultInvoiceSettingPath, s.host, uid)
	)

	err = s.client.GetCall(s.reqLogger, &resp, api)
	if err != nil || !resp.OK() {
		err = resp.Error()
		return
	}

	out = resp.Data

	return
}

type RelatedTransaction struct {
	SerialNum   string    `json:"serial_num"`
	Cash        pay.Money `json:"cash"`
	ValidAmount pay.Money `json:"valid_amount"`
	Time        time.Time `json:"time"`
	Source      string    `json:"source"`
}

type ValidTransactionsOutput struct {
	Amount              pay.Money            `json:"amount"`
	RelatedTransactions []RelatedTransaction `json:"related_transactions"`
}

func (s *gaeaAdminService) GetInvoiceValidTransactions(uid uint32, start, end time.Time) (out ValidTransactionsOutput, err error) {
	var (
		resp struct {
			apiResultBase
			Data ValidTransactionsOutput `json:"data"`
		}
		api = fmt.Sprintf(GetInvoiceValidTransactionsPath, s.host, uid, start.Unix(), end.Unix())
	)

	err = s.client.GetCall(s.reqLogger, &resp, api)

	if err != nil || !resp.OK() {
		err = resp.Error()
		return
	}

	out = resp.Data

	return
}

type Delivery struct {
	Address string `json:"address"`
	Phone   string `json:"phone"`
	Name    string `json:"name"`
}
type InvoiceInput struct {
	InvoiceSettingId    string    `json:"invoice_setting_id"`
	Amount              pay.Money `json:"amount"`
	RelatedTransactions []string  `json:"related_transactions"`
	TaxRate             int       `json:"tax_rate"`
	FeeNote             string    `json:"fee_note"`
	Delivery            Delivery  `json:"delivery"`
	CustomerComment     string    `json:"customer_comment"`
}

func (s *gaeaAdminService) CreateInvoice(uid uint32, input InvoiceInput) (id string, err error) {
	var (
		resp struct {
			apiResultBase
			Data string `json:"data"`
		}
		api = fmt.Sprintf(CreateInvoicePath, s.host, uid)
	)

	err = s.client.CallWithJson(s.reqLogger, &resp, api, input)

	if err != nil || !resp.OK() {
		err = resp.Error()
		return
	}

	id = resp.Data

	return
}

type RelateTransaction struct {
	Id         string    `json:"id"`
	Cash       pay.Money `json:"cash"`
	AmountUsed pay.Money `json:"amount_used"`
}

type InvoiceSettingRecord struct {
	Title          string            `json:"title"`
	Type           enums.InvoiceType `json:"type"`
	TaxPayerId     string            `json:"tax_payer_id"`
	BankName       string            `json:"bank_name"`
	BankAccount    string            `json:"bank_account"`
	Company        string            `json:"company"`
	CompanyAddress string            `json:"company_address"`
	CompanyTel     string            `json:"company_tel"`
	CertificateUrl string            `json:"certificate_url"`
	CoTaxType      enums.CoTaxType   `json:"co_tax_type"` // 公司纳税类型
	CoTaxId        string            `json:"co_tax_id"`   // 公司纳税号码
}

type DeliveryRecord struct {
	Address        string `json:"address"`
	Phone          string `json:"phone"`
	Name           string `json:"name"`
	ExpressCompany string `json:"express_company"`
	ExpressNumber  string `json:"express_number"`
}

type InvoiceRecord struct {
	Id                 bson.ObjectId        `json:"id"`
	Uid                uint32               `json:"uid"`
	Number             string               `json:"number"`
	Money              pay.Money            `json:"money"`
	RelateTransactions []RelateTransaction  `json:"relate_transactions"`
	InvoiceSetting     InvoiceSettingRecord `json:"invoice_setting"`
	Delivery           DeliveryRecord       `json:"delivery"`
	Status             enums.InvoiceStatus  `json:"status"`
	TaxRate            int                  `json:"tax_rate"`
	FeeNote            string               `json:"fee_note"`
	Reason             string               `json:"reason"`
	CreatedAt          time.Time            `json:"created_at"`
	UpdatedAt          time.Time            `json:"updated_at"`
}

func (s *gaeaAdminService) GetInvoice(id string) (out InvoiceRecord, err error) {
	var (
		resp struct {
			apiResultBase
			Data InvoiceRecord `json:"data"`
		}
		api = fmt.Sprintf(GetInvoicePath, s.host, id)
	)

	err = s.client.GetCall(s.reqLogger, &resp, api)

	if err != nil || !resp.OK() {
		err = resp.Error()
		return
	}

	out = resp.Data
	return
}

func (s *gaeaAdminService) ListInvoiceByUid(uid uint32, page, page_size int) (outs []InvoiceRecord, err error) {
	var (
		resp struct {
			apiResultBase
			Data []InvoiceRecord `json:"data"`
		}
		api = fmt.Sprintf(ListInvoiceByUidPath, s.host, uid, page, page_size)
	)

	err = s.client.GetCall(s.reqLogger, &resp, api)

	if err != nil || !resp.OK() {
		err = resp.Error()
		return
	}

	outs = resp.Data
	return
}

type UpdateInvoiceInput struct {
	Number                 *string              `json:"number"`
	Status                 *enums.InvoiceStatus `json:"status"`
	DeliveryExpressCompany *string              `json:"delivery_express_company"`
	DeliveryNumber         *string              `json:"delivery_number"`
	Reason                 *string              `json:"reason"`
}

func (s gaeaAdminService) UpdateInvoice(id string, input UpdateInvoiceInput) (err error) {
	var (
		resp apiResultBase
		api  = fmt.Sprintf(UpdateInvoicePath, s.host, id)
	)

	err = s.client.PutCallWithJson(s.reqLogger, &resp, api, input)

	if err != nil || !resp.OK() {
		err = resp.Error()
		return
	}

	return
}
