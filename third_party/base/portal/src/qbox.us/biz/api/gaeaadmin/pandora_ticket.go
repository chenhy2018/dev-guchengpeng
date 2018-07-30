package gaeaadmin

import (
	"fmt"
	"net/url"
	"strconv"
	"time"
)

type PandoraTicketGetParams struct {
	Uid   uint32
	Email string
}

func (p PandoraTicketGetParams) ToURLValues() url.Values {
	values := url.Values{}
	if p.Uid > 0 {
		values.Add("uid", strconv.FormatUint(uint64(p.Uid), 10))
	}

	if p.Email != "" {
		values.Add("email", p.Email)
	}

	return values
}

type PandoraTicketSaveParams struct {
	Company  string   `json:"companyName"`
	Types    []string `json:"types"`
	Contacts string   `json:"contacts"`
	DataType string   `json:"dataType"`
	Remark   string   `json:"remark"`
	Uid      uint32   `json:"uid"`
}

type PandoraTicket struct {
	Uid          uint32              `json:"uid"`
	Status       PandoraTicketStatus `json:"status"`
	Company      string              `json:"companyName"`
	Types        []string            `json:"types"`
	Contacts     string              `json:"contacts"`
	DataType     string              `json:"dataType"`
	Remark       string              `json:"remark"`
	RejectReason string              `json:"rejectReason"`
	CreatedAt    time.Time           `json:"createdAt"`
	UpdatedAt    time.Time           `json:"updatedAt"`
}

type PandoraTicketStatus int

const (
	PandoraTicketStatusPending PandoraTicketStatus = iota
	PandoraTicketStatusSuccess
	PandoraTicketStatusFailed
	PandoraTicketStatusNo
)

func (p PandoraTicketStatus) String() string {
	switch p {
	case PandoraTicketStatusPending:
		return "pending"
	case PandoraTicketStatusSuccess:
		return "success"
	case PandoraTicketStatusFailed:
		return "failed"
	case PandoraTicketStatusNo:
		return "no ticket"
	default:
		return ""
	}
}

func (p PandoraTicketStatus) Humanize() string {
	switch p {
	case PandoraTicketStatusPending:
		return "审核中"
	case PandoraTicketStatusSuccess:
		return "审核通过"
	case PandoraTicketStatusFailed:
		return "审核失败"
	case PandoraTicketStatusNo:
		return "没有申请信息"
	default:
		return ""
	}
}

func (s *gaeaAdminService) PandoraTicketGet(params PandoraTicketGetParams) (ticket PandoraTicket, err error) {
	var (
		resp struct {
			apiResultBase
			Data PandoraTicket `json:"data"`
		}
		api = fmt.Sprintf("%s/api/pandora/user/ticket?%s", s.host, params.ToURLValues().Encode())
	)

	err = s.client.GetCall(s.reqLogger, &resp, api)
	if err != nil || !resp.OK() {
		err = resp.Error()
		return
	}

	ticket = resp.Data

	return
}

func (s *gaeaAdminService) PandoraTicketSave(params PandoraTicketSaveParams) error {
	var (
		resp struct {
			apiResultBase
			Data string `json:"data"`
		}
		api = fmt.Sprintf("%s/api/pandora/user/ticket", s.host)
	)

	err := s.client.CallWithJson(s.reqLogger, &resp, api, params)
	if err != nil {
		return err
	} else if !resp.OK() {
		err = resp.Error()
		return err
	}

	return nil
}
