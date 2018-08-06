package gaeaadmin

import "fmt"

func (s *gaeaAdminService) RelationCreate(params FinancialRelationCreateParams) error {
	var (
		resp struct {
			apiResultBase
			Data string `json:"data"`
		}
		api = fmt.Sprintf("%s/api/finance/relation", s.host)
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

func (s *gaeaAdminService) RelationList(params FinancialRelationListParams) (res FinancialRelationList, err error) {
	var (
		resp struct {
			apiResultBase
			Data FinancialRelationList `json:"data"`
		}
		api = fmt.Sprintf("%s/api/finance/relation", s.host)
	)

	err = s.client.GetCallWithForm(s.reqLogger, &resp, api, params.Values())
	if err != nil {
		return
	} else if !resp.OK() {
		err = resp.Error()
		return
	}

	res = resp.Data

	return
}

func (s *gaeaAdminService) RelationListChildren(params FinancialRelationListChildrenParams) (res FinancialRelationList, err error) {
	var (
		resp struct {
			apiResultBase
			Data FinancialRelationList `json:"data"`
		}
		api = fmt.Sprintf("%s/api/finance/relation/children", s.host)
	)

	err = s.client.GetCallWithForm(s.reqLogger, &resp, api, params.Values())
	if err != nil {
		return
	} else if !resp.OK() {
		err = resp.Error()
		return
	}

	res = resp.Data

	return
}

func (s *gaeaAdminService) RelationGet(uid uint32) (res FinancialRelation, err error) {
	var (
		resp struct {
			apiResultBase
			Data FinancialRelation `json:"data"`
		}
		api = fmt.Sprintf("%s/api/finance/relation/%d", s.host, uid)
	)

	err = s.client.GetCall(s.reqLogger, &resp, api)
	if err != nil {
		return
	} else if !resp.OK() {
		err = resp.Error()
		return
	}

	res = resp.Data

	return
}

func (s *gaeaAdminService) RelationUpdate(uid uint32, params FinancialRelationUpdateParams) error {
	var (
		resp struct {
			apiResultBase
			Data string `json:"data"`
		}
		api = fmt.Sprintf("%s/api/finance/relation/%d", s.host, uid)
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

func (s *gaeaAdminService) RelationRemove(uid uint32) error {
	var (
		resp struct {
			apiResultBase
			Data string `json:"data"`
		}
		api = fmt.Sprintf("%s/api/finance/relation/%d", s.host, uid)
	)

	err := s.client.DeleteCall(s.reqLogger, &resp, api)
	if err != nil {
		return err
	} else if !resp.OK() {
		err = resp.Error()
		return err
	}

	return nil
}
