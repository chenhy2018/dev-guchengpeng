package services

import (
	"fmt"

	"qbox.us/biz/utils.v2/json"
)

type CreateOpLogIn struct {
	Uid         uint32      `json:"uid"`      // customer's uid
	Email       string      `json:"email"`    // internal user's email
	Operator    string      `json:"operator"` // operator's email
	URI         string      `json:"uri"`
	Params      string      `json:"params"`
	Method      string      `json:"method"`
	Action      OplogAction `json:"action"`
	Description string      `json:"description"`
}

func (i *CreateOpLogIn) Validate() bool {
	if i.Operator == "" {
		return false
	}
	if i.Action == "" {
		return false
	}

	if i.URI == "" {
		return false
	}
	return true
}

func (s *opLogService) CreateOpLog(params CreateOpLogIn) (err error) {
	var out json.CommonResponse

	url := fmt.Sprintf("%s/api/admin/oplog", s.host)

	if err = s.client.CallWithJson(s.reqLogger, &out, url, &params); err != nil {
		return err
	}

	if out.Code/100 != 2 {
		return out.Error()
	}

	return
}
