package gaeaadmin

import (
	"fmt"
	"net/url"
	"strconv"
)

type DeveloperSearch struct {
	Total int             `json:"total"`
	Items []DeveloperItem `json:"items"`
}

type DeveloperItem struct {
	Uid          uint32 `json:"uid"`
	Email        string `json:"email"`
	FullName     string `json:"fullname"`
	PhoneNumber  string `json:"phone_number"`
	ImCategory   string `json:"im_category"`
	ImNumber     string `json:"im_number"`
	MobileBinded bool   `json:"mobile_binded"`
	IsActivated  bool   `json:"isactived"`
	CreateAt     int64  `json:"created_at"`
	SalesId      string `json:"sf_sales_id"`
}

type FeatureDeveloperGetReq struct {
	Query     string `json:"query"`
	Limit     int    `json:"limit"`
	Offset    int    `json:"offset"`
	FeatureId string `json:"featureId"`
}

func (s *gaeaAdminService) FeatureDeveloperGet(req FeatureDeveloperGetReq) (result DeveloperSearch, err error) {
	var (
		resp struct {
			apiResultBase
			Data DeveloperSearch `json:"data"`
		}
		api = fmt.Sprintf("%s/api/developer/featureConfig/dev", s.host)
	)

	values := url.Values{}
	values.Set("limit", strconv.Itoa(req.Limit))
	values.Set("offset", strconv.Itoa(req.Offset))
	values.Set("featureId", req.FeatureId)
	if req.Query != "" {
		values.Set("query", req.Query)
	}

	err = s.client.GetCall(s.reqLogger, &resp, api+"?"+values.Encode())
	if err != nil {
		return
	}

	result = resp.Data
	return
}
