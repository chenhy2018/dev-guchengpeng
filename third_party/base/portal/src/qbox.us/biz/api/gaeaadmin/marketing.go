package gaeaadmin

import "fmt"

type OnInviteParams struct {
	InviterUID uint32 `json:"inviter_uid"`
	INviteeUID uint32 `json:"i_nvitee_uid"`
}

func (s *gaeaAdminService) OnInvite(params OnInviteParams) (err error) {
	var (
		resp struct {
			apiResultBase
		}
		api = fmt.Sprintf("%s/api/marketing/event/invite", s.host)
	)

	err = s.client.CallWithJson(s.reqLogger, &resp, api, params)
	if err != nil || !resp.OK() {
		err = resp.Error()
		return
	}

	return
}

type OnRechargeParams struct {
	UID  uint32 `json:"uid"`
	TxID string `json:"tx_id"`
}

func (s *gaeaAdminService) OnRecharge(params OnRechargeParams) (err error) {
	var (
		resp struct {
			apiResultBase
		}
		api = fmt.Sprintf("%s/api/marketing/event/recharge", s.host)
	)

	err = s.client.CallWithJson(s.reqLogger, &resp, api, params)
	if err != nil || !resp.OK() {
		err = resp.Error()
		return
	}

	return
}

type OnIdentifyParams struct {
	UID uint32 `json:"uid"`
}

func (s *gaeaAdminService) OnIdentify(params OnIdentifyParams) (err error) {
	var (
		resp struct {
			apiResultBase
		}
		api = fmt.Sprintf("%s/api/marketing/event/identify", s.host)
	)

	err = s.client.CallWithJson(s.reqLogger, &resp, api, params)
	if err != nil || !resp.OK() {
		err = resp.Error()
		return
	}

	return
}
