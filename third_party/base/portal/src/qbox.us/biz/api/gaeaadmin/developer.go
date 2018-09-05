package gaeaadmin

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

func (s *gaeaAdminService) DeveloperGet(params DeveloperGetParams) (res Developer, err error) {
	var (
		resp struct {
			apiResultBase
			Data Developer `json:"data"`
		}
		api = fmt.Sprintf("%s/api/developer", s.host)
	)

	err = s.client.GetCallWithForm(s.reqLogger, &resp, api, params.ToURLValues())
	if err != nil || !resp.OK() {
		err = resp.Error()
		return
	}

	res = resp.Data

	return
}

func (s *gaeaAdminService) DeveloperOverview(uid uint32) (res DeveloperOverview, err error) {
	var (
		resp struct {
			apiResultBase
			Data DeveloperOverview `json:"data"`
		}
		api = fmt.Sprintf("%s/api/developer/%d/overview", s.host, uid)
	)

	err = s.client.GetCall(s.reqLogger, &resp, api)
	if err != nil || !resp.OK() {
		err = resp.Error()
		return
	}

	res = resp.Data
	return
}

func (s *gaeaAdminService) DeveloperOverviewWithEmail(email string) (res DeveloperOverview, err error) {
	var (
		resp struct {
			apiResultBase
			Data DeveloperOverview `json:"data"`
		}
		api = fmt.Sprintf("%s/api/developer/%s/overview", s.host, email)
	)

	err = s.client.GetCall(s.reqLogger, &resp, api)
	if err != nil || !resp.OK() {
		err = resp.Error()
		return
	}

	res = resp.Data
	return
}

// fields: developer字段, 默认全部返回
func (s *gaeaAdminService) DeveloperListByUids(uids []uint32, fields []string) (res []Developer, err error) {
	var (
		resp struct {
			apiResultBase
			Data []listDeveloper `json:"data"`
		}
		api = fmt.Sprintf("%s/api/developer/uids", s.host)
	)

	values := url.Values{}
	if len(uids) > 0 {
		values.Set("uids", strings.Join(func(uids []uint32) []string {
			uidStrs := make([]string, len(uids))
			for i, uid := range uids {
				uidStrs[i] = strconv.FormatUint(uint64(uid), 10)
			}
			return uidStrs
		}(uids), ","))
	}

	if len(fields) > 0 {
		values.Set("fields", strings.Join(fields, ","))
	}

	err = s.client.GetCallWithForm(s.reqLogger, &resp, api, values)
	if err != nil || !resp.OK() {
		err = resp.Error()
		return
	}

	for _, d := range resp.Data {
		res = append(res, d.toDeveloper())
	}

	return
}

func (s *gaeaAdminService) DeveloperList(params DeveloperListParams) (res []Developer, err error) {
	var (
		resp struct {
			apiResultBase
			Data []Developer `json:"data"`
		}
		api = fmt.Sprintf("%s/api/developer/list", s.host)
	)

	values := params.Values()
	err = s.client.GetCallWithForm(s.reqLogger, &resp, api, values)
	if err != nil || !resp.OK() {
		err = resp.Error()
		return
	}

	res = resp.Data
	return
}

func (s *gaeaAdminService) SalesGet(params SalesGetParams) (res Sales, err error) {
	var (
		resp struct {
			apiResultBase
			Data Sales `json:"data"`
		}
		api = fmt.Sprintf("%s/api/sales/get", s.host)
	)

	err = s.client.GetCallWithForm(s.reqLogger, &resp, api, params.ToURLValues())
	if err != nil || !resp.OK() {
		err = resp.Error()
		return
	}

	res = resp.Data

	return

}

func (s *gaeaAdminService) DeveloperRank(uid uint32) (rank string, err error) {
	var (
		resp struct {
			apiResultBase
			Data string `json:"data"`
		}
		api = fmt.Sprintf("%s/api/developer/%d/rank", s.host, uid)
	)

	err = s.client.GetCall(s.reqLogger, &resp, api)
	if err != nil || !resp.OK() {
		err = resp.Error()
		return
	}

	rank = resp.Data

	return
}

func (s *gaeaAdminService) DeveloperUpdate(uid uint32, params DeveloperUpdateParams) (err error) {
	var (
		resp struct {
			apiResultBase
			Data string `json:"data"`
		}
		api = fmt.Sprintf("%s/api/developer/%d", s.host, uid)
	)
	err = s.client.CallWithJson(s.reqLogger, &resp, api, params)
	if err != nil || !resp.OK() {
		err = resp.Error()
		return
	}

	return
}

func (s *gaeaAdminService) DeveloperCreate(params DeveloperCreateParams) (err error) {
	var (
		resp struct {
			apiResultBase
			Data string `json:"data"`
		}
		api = fmt.Sprintf("%s/api/developer", s.host)
	)
	err = s.client.CallWithJson(s.reqLogger, &resp, api, params)
	if err != nil || !resp.OK() {
		err = resp.Error()
		return
	}
	return
}

func (s *gaeaAdminService) UpdateUserType(uid uint32, params UserTypeUpdateParams) (err error) {
	var (
		resp struct {
			apiResultBase
			Data string `json:"data"`
		}
		api = fmt.Sprintf("%s/api/developer/%d/user-type", s.host, uid)
	)
	err = s.client.CallWithJson(s.reqLogger, &resp, api, params)
	if err != nil || !resp.OK() {
		err = resp.Error()
		return
	}
	return
}

func (s *gaeaAdminService) UpdateEmail(uid uint32, email string) (err error) {
	var (
		resp struct {
			apiResultBase
			Data string `json:"data"`
		}
		api    = fmt.Sprintf("%s/api/developer/%d/email", s.host, uid)
		params struct {
			Email string `json:"email"`
		}
	)
	params.Email = email
	err = s.client.CallWithJson(s.reqLogger, &resp, api, params)
	if err != nil || !resp.OK() {
		err = resp.Error()
		return
	}
	return
}

func (s *gaeaAdminService) DeveloperSFInfoUpdate(uid uint32, params DeveloperSFInfoUpdateParams) error {
	var (
		resp struct {
			apiResultBase
			Data string `json:"data"`
		}
		api = fmt.Sprintf("%s/api/developer/%d/sf", s.host, uid)
	)

	err := s.client.CallWithJson(s.reqLogger, &resp, api, params)
	if err != nil {
		return err
	}

	return nil
}
