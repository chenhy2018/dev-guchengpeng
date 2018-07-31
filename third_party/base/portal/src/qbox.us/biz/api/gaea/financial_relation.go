package gaea

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/qiniu/rpc.v1"

	"qbox.us/biz/api/gaea/enums"
)

type FinancialRelationService struct {
	host   string
	client rpc.Client
}

type FinancialRelationModel struct {
	Uid       uint32                      `json:"uid"`
	ParentUid uint32                      `json:"parent_uid"`
	Memo      string                      `json:"memo"`
	CreatedAt time.Time                   `json:"created_at"`
	UpdatedAt time.Time                   `json:"updated_at"`
	Type      enums.FinancialRelationType `json:"type"`
}

type ListIn struct {
	Uid    uint32 `param:"uid"`
	Offset int    `param:"offset"`
	Limit  int    `param:"limit"`
	Sort   string `param:"sort"`
}

type financialRelationResp struct {
	CommonResponse

	Data FinancialRelationModel `json:"data"`
}

type financialRelationSliceResp struct {
	CommonResponse

	Data []FinancialRelationModel `json:"data"`
}

func NewFinancialRelationService(host string, t http.RoundTripper) *FinancialRelationService {
	return &FinancialRelationService{
		host: host,
		client: rpc.Client{
			&http.Client{Transport: t},
		},
	}
}

func (s *FinancialRelationService) FindByUid(l rpc.Logger, uid uint32) (model FinancialRelationModel, err error) {
	var resp financialRelationResp

	err = s.client.GetCall(l, &resp, s.host+"/admin/financial-relation/"+strconv.FormatUint(uint64(uid), 10))
	if err != nil {
		return
	}

	if resp.Code != CodeOK {
		err = errors.New(fmt.Sprintf("%d", resp.Code))
		return
	}

	model = resp.Data

	return
}

func (s *FinancialRelationService) ListChildrenByUid(l rpc.Logger, input ListIn) (models []FinancialRelationModel, err error) {
	var resp financialRelationSliceResp

	value := url.Values{}
	value.Add("uid", strconv.Itoa(int(input.Uid)))
	value.Add("limit", strconv.Itoa(int(input.Limit)))
	value.Add("offset", strconv.Itoa(int(input.Offset)))
	value.Add("sort", input.Sort)
	err = s.client.GetCall(l, &resp, s.host+"/admin/financial-relation/children?"+value.Encode())
	if err != nil {
		return
	}

	if resp.Code != CodeOK {
		err = errors.New(fmt.Sprintf("%d", resp.Code))
		return
	}

	models = resp.Data

	return
}
