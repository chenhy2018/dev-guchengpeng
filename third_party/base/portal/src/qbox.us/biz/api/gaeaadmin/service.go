package gaeaadmin

import (
	"net/http"
	"time"

	"github.com/qiniu/rpc.v1"
)

type GaeaAdminService interface {
	FreezeV2(uid uint32, in FreezeIn) (err error)
	UnfreezeV2(uid uint32, in UnfreezeIn) (err error)
	BalanceNotEnoughNotificationV2(uid uint32, in BalanceNotEnoughIn) (err error)
	GetDelayFreezeTicketV2(uid uint32) (ticket *DelayShowOut, err error)
	ListFrozenUsersV2(params ListFrozenUsersIn) (uids []uint32, err error)
	ListLogsV2(in ListLogsIn) (logs []*ListLogsOut, err error)
	GetLatestFreezeLogV2(uid uint32) (log *FreezeLog, err error)

	GetBillList(uid uint32, start time.Time, end time.Time, email string) (billList GetBillListOut, err error)
	GetBill(uid uint32, month time.Time) (billHtml []byte, err error)
	GetBillPDF(uid uint32, month time.Time) (billPdf []byte, err error)
	GetMergeAccountBill(uid uint32, month time.Time) (billHtml []byte, err error)
	GetMergeAccountBillPDF(uid uint32, month time.Time) (billPdf []byte, err error)

	GetBillListV2(uid uint32, start time.Time, end time.Time, merge bool) (billList GetBillListOut, err error)
	GetBillV2(uid uint32, month time.Time, merge bool) (billHtml []byte, err error)

	IdentityList(params IdentityListParams) (identities []DeveloperIdentity, err error)
	IdentityCreate(params IdentityCreateParams) error
	IdentityGet(uid uint32) (identity DeveloperIdentity, err error)
	IdentityUpdate(uid uint32, params IdentityUpdateParams) error
	IdentityReview(uid uint32, params IdentityReviewParams) error
	IdentityHistory(uid uint32, params IdentityListParams) (identities []DeveloperIdentity, err error)
	IdentityUpToken() (token IdentityUpToken, err error)
	IdentityBankVerify(uid uint32, params IdentityBankVerifyParams) (out IdentityBankVerifyOut, err error)
	IdentityBankTransferList(params IdentityBankTransferListParams) (transfers []*IdentityBankTransfer, err error)
	IdentityBankTransferBatchUpdate(params IdentityBankTransferBatchUpdateParams) (err error)

	GetInvoiceDelivery(uid uint32) (delivery InvoiceDelivery, err error)
	SetInvoiceDelivery(delivery InvoiceDelivery) (err error)
	GetInvoiceUploadToken() (token InvoiceUploadToken, err error)
	CreateInvoiceSetting(uid uint32, input InvoiceSettingInput) (id string, err error)
	UpdateInvoiceSetting(id string, input InvoiceSettingInput) (err error)
	GetInvoiceSetting(id string) (out InvoiceSettingOutput, err error)
	DeleteInvoiceSetting(id string) (err error)
	ListInvoiceSettingBuyUid(uid uint32, page, page_size int) (outs []InvoiceSettingOutput, err error)
	SetDefaultInvoiceSetting(uid uint32, id string) (err error)
	GetDefaultInvoiceSetting(uid uint32) (out InvoiceSettingOutput, err error)
	GetInvoiceValidTransactions(uid uint32, start, end time.Time) (out ValidTransactionsOutput, err error)
	CreateInvoice(uid uint32, input InvoiceInput) (id string, err error)
	GetInvoice(id string) (out InvoiceRecord, err error)
	ListInvoiceByUid(uid uint32, page, page_size int) (outs []InvoiceRecord, err error)
	UpdateInvoice(id string, input UpdateInvoiceInput) (err error)

	RelationCreate(FinancialRelationCreateParams) error
	RelationList(FinancialRelationListParams) (FinancialRelationList, error)
	RelationListChildren(FinancialRelationListChildrenParams) (FinancialRelationList, error)
	RelationGet(uid uint32) (FinancialRelation, error)
	RelationUpdate(uint32, FinancialRelationUpdateParams) error
	RelationRemove(uint32) error
	DeveloperListByUids(uids []uint32, fields []string) (res []Developer, err error)
	DeveloperGet(DeveloperGetParams) (Developer, error)
	DeveloperOverview(uid uint32) (res DeveloperOverview, err error)
	DeveloperOverviewWithEmail(email string) (res DeveloperOverview, err error)
	DeveloperSFInfoUpdate(uid uint32, params DeveloperSFInfoUpdateParams) error
	SalesGet(SalesGetParams) (Sales, error)
	DeveloperRank(uid uint32) (rank string, err error)

	DeveloperUpdate(uid uint32, params DeveloperUpdateParams) (err error)
	DeveloperCreate(params DeveloperCreateParams) (err error)
	UpdateUserType(uid uint32, params UserTypeUpdateParams) (err error)
	UpdateEmail(uid uint32, email string) (err error)
	UserGet(UserGetParams) (User, error)
	UserListByIds([]string) ([]User, error)
	ApplicationGet(id string) (res Application, err error)

	TransactionList(params TrListParams) (res []Transaction, err error)

	FeatureDeveloperGet(req FeatureDeveloperGetReq) (result DeveloperSearch, err error)

	PiliTicketGet(params PiliTicketGetParams) (ticket PiliTicket, err error)
	PiliTicketSave(params PiliTicketSaveParams) error
	PiliTicketUrge(uid uint32) error

	PandoraTicketGet(params PandoraTicketGetParams) (ticket PandoraTicket, err error)
	PandoraTicketSave(params PandoraTicketSaveParams) error

	SendSms(params SendSmsParams) (out SendSmsOutput, err error)
	SendMail(params SendMailParams) (out SendMailOutput, err error)

	OnInvite(params OnInviteParams) error
	OnRecharge(params OnRechargeParams) error
	OnIdentify(params OnIdentifyParams) error
}

type gaeaAdminService struct {
	host      string
	client    rpc.Client
	reqLogger rpc.Logger
}

func NewGaeaAdminService(host string, adminOAuth http.RoundTripper, reqLogger rpc.Logger) GaeaAdminService {
	return &gaeaAdminService{
		host: host,
		client: rpc.Client{
			&http.Client{Transport: adminOAuth},
		},
		reqLogger: reqLogger,
	}
}
