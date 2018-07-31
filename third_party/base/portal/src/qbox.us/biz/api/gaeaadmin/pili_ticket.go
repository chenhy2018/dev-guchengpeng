package gaeaadmin

import (
	"fmt"
	"net/url"
	"strconv"
)

type PiliTicketGetParams struct {
	Uid   uint32
	Email string
}

func (p PiliTicketGetParams) ToURLValues() url.Values {
	values := url.Values{}
	if p.Uid > 0 {
		values.Add("uid", strconv.FormatUint(uint64(p.Uid), 10))
	}

	if p.Email != "" {
		values.Add("email", p.Email)
	}

	return values
}

type PiliTicketSaveParams struct {
	Uid              uint32   `json:"uid"`
	Types            []string `json:"types"`
	Stage            string   `json:"stage"`
	PublishEquipment []string `json:"publishEquipment"`
	PlayEquipment    []string `json:"playEquipment"`
	Domain           string   `json:"domain"`
}

type PiliTicket struct {
	Status           PiliTicketStatus `json:"status"`
	Stage            string           `json:"stage"`
	Types            []string         `json:"types"`
	PublishEquipment []string         `json:"publishEquipment"`
	PlayEquipment    []string         `json:"playEquipment"`
	RejectReason     string           `json:"rejectReason"`
	Domain           string           `json:"domain"`
}

type PiliTicketStatus string

const (
	PiliTicketStatusPending PiliTicketStatus = "pending"
	PiliTicketStatusSuccess PiliTicketStatus = "success"
	PiliTicketStatusFailed  PiliTicketStatus = "failed"
	PiliTicketStatusNo      PiliTicketStatus = "no ticket"
)

func (p PiliTicketStatus) String() string {
	return string(p)
}

func (p PiliTicketStatus) Humanize() string {
	switch p {
	case PiliTicketStatusPending:
		return "审核中"
	case PiliTicketStatusSuccess:
		return "审核通过"
	case PiliTicketStatusFailed:
		return "审核失败"
	case PiliTicketStatusNo:
		return "没有申请信息"
	default:
		return ""
	}
}

func (s *gaeaAdminService) PiliTicketGet(params PiliTicketGetParams) (ticket PiliTicket, err error) {
	var (
		resp struct {
			apiResultBase
			Data PiliTicket `json:"data"`
		}
		api = fmt.Sprintf("%s/api/pili/user/ticket?%s", s.host, params.ToURLValues().Encode())
	)

	err = s.client.GetCall(s.reqLogger, &resp, api)
	if err != nil || !resp.OK() {
		err = resp.Error()
		return
	}

	ticket = resp.Data

	return
}

func (s *gaeaAdminService) PiliTicketSave(params PiliTicketSaveParams) error {
	var (
		resp struct {
			apiResultBase
			Data string `json:"data"`
		}
		api = fmt.Sprintf("%s/api/pili/user/ticket", s.host)
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

func (s *gaeaAdminService) PiliTicketUrge(uid uint32) error {
	var (
		resp struct {
			apiResultBase
			Data string `json:"data"`
		}
		api = fmt.Sprintf("%s/api/pili/%d/ticket/urge", s.host, uid)
	)

	err := s.client.CallWithJson(s.reqLogger, &resp, api, nil)
	if err != nil {
		return err
	} else if !resp.OK() {
		return resp.Error()
	}

	return nil
}
