package gaeaadmin

import "fmt"

func (s *gaeaAdminService) IdentityList(params IdentityListParams) (identities []DeveloperIdentity, err error) {
	var (
		resp struct {
			apiResultBase
			Data []DeveloperIdentity `json:"data"`
		}
		api = fmt.Sprintf("%s/api/user/identity", s.host)
	)

	err = s.client.GetCallWithForm(s.reqLogger, &resp, api, params.Values())
	if err != nil || !resp.OK() {
		err = resp.Error()
		return
	}

	identities = resp.Data

	return
}

func (s *gaeaAdminService) IdentityCreate(params IdentityCreateParams) error {
	var (
		resp struct {
			apiResultBase
			Data string `json:"data"`
		}
		api = fmt.Sprintf("%s/api/user/identity", s.host)
	)

	err := s.client.CallWithJson(s.reqLogger, &resp, api, params)
	if err != nil {
		return err
	} else if !resp.OK() {
		err = resp.Error()
		if err == ErrConflict && resp.Data != "" {
			return apiError(fmt.Sprintf("%s:%s", err.Error(), resp.Data))
		}
		return err
	}

	return nil
}

func (s *gaeaAdminService) IdentityGet(uid uint32) (identity DeveloperIdentity, err error) {
	var (
		resp struct {
			apiResultBase
			Data DeveloperIdentity `json:"data"`
		}
		api = fmt.Sprintf("%s/api/user/%d/identity", s.host, uid)
	)

	err = s.client.GetCall(s.reqLogger, &resp, api)
	if err != nil {
		return
	} else if !resp.OK() {
		err = resp.Error()
		return
	}

	identity = resp.Data

	return
}

func (s *gaeaAdminService) IdentityUpdate(uid uint32, params IdentityUpdateParams) error {
	var (
		resp struct {
			apiResultBase
			Data string `json:"data"`
		}
		api = fmt.Sprintf("%s/api/user/%d/identity", s.host, uid)
	)

	err := s.client.CallWithJson(s.reqLogger, &resp, api, params)
	if err != nil {
		return err
	} else if !resp.OK() {
		err = resp.Error()
		if err == ErrConflict && resp.Data != "" {
			return apiError(fmt.Sprintf("%s:%s", err.Error(), resp.Data))
		}
		return err
	}

	return nil
}

func (s *gaeaAdminService) IdentityReview(uid uint32, params IdentityReviewParams) error {
	var (
		resp struct {
			apiResultBase
			Data string `json:"data"`
		}
		api = fmt.Sprintf("%s/api/user/%d/identity/review", s.host, uid)
	)

	err := s.client.CallWithJson(s.reqLogger, &resp, api, params)
	if err != nil {
		return err
	} else if !resp.OK() {
		err = resp.Error()
		if err == ErrConflict && resp.Data != "" {
			return apiError(fmt.Sprintf("%s:%s", err.Error(), resp.Data))
		}
		return err
	}

	return nil
}

func (s *gaeaAdminService) IdentityHistory(uid uint32, params IdentityListParams) (identities []DeveloperIdentity, err error) {
	var (
		resp struct {
			apiResultBase
			Data []DeveloperIdentity `json:"data"`
		}
		api = fmt.Sprintf("%s/api/user/%d/identity/history", s.host, uid)
	)

	err = s.client.GetCall(s.reqLogger, &resp, api)
	if err != nil {
		return
	} else if !resp.OK() {
		err = resp.Error()
		return
	}

	identities = resp.Data

	return
}

func (s *gaeaAdminService) IdentityUpToken() (token IdentityUpToken, err error) {
	var (
		resp struct {
			apiResultBase
			Data IdentityUpToken `json:"data"`
		}
		api = fmt.Sprintf("%s/api/user/identity/uptoken", s.host)
	)

	err = s.client.GetCall(s.reqLogger, &resp, api)
	if err != nil {
		return
	} else if !resp.OK() {
		err = resp.Error()
		return
	}

	token = resp.Data

	return
}

func (s *gaeaAdminService) IdentityBankVerify(uid uint32, params IdentityBankVerifyParams) (out IdentityBankVerifyOut, err error) {
	var (
		resp struct {
			apiResultBase
			Data IdentityBankVerifyOut `json:"data"`
		}
		api = fmt.Sprintf("%s/api/user/%d/identity/bank/verify", s.host, uid)
	)

	err = s.client.CallWithJson(s.reqLogger, &resp, api, params)
	if err != nil {
		return
	}

	out = resp.Data
	if !resp.OK() {
		err = resp.Error()
		return
	}

	return
}

func (s *gaeaAdminService) IdentityBankTransferList(params IdentityBankTransferListParams) (transfers []*IdentityBankTransfer, err error) {
	var (
		resp struct {
			apiResultBase
			Data []*IdentityBankTransfer `json:"data"`
		}
		api = fmt.Sprintf("%s/api/user/identity/bank", s.host)
	)

	err = s.client.GetCallWithForm(s.reqLogger, &resp, api, params.Values())
	if err != nil {
		return
	} else if !resp.OK() {
		err = resp.Error()
		return
	}

	transfers = resp.Data
	return
}

func (s *gaeaAdminService) IdentityBankTransferBatchUpdate(params IdentityBankTransferBatchUpdateParams) (err error) {
	var (
		resp apiResultBase
		api  = fmt.Sprintf("%s/api/user/identity/bank/batch", s.host)
	)

	err = s.client.PutCallWithJson(s.reqLogger, &resp, api, params)
	if err != nil {
		return
	} else if !resp.OK() {
		err = resp.Error()
		return
	}

	return
}
