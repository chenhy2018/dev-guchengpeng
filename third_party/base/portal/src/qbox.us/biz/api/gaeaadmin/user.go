package gaeaadmin

import (
	"fmt"
	"net/url"
	"time"

	"labix.org/v2/mgo/bson"
)

func (s *gaeaAdminService) UserGet(params UserGetParams) (res User, err error) {
	var (
		resp struct {
			apiResultBase
			Data User `json:"data"`
		}
		api = fmt.Sprintf("%s/api/user", s.host)
	)

	err = s.client.GetCallWithForm(s.reqLogger, &resp, api, params.Values())
	if err != nil || !resp.OK() {
		err = resp.Error()
		return
	}

	res = resp.Data
	return
}

func (s *gaeaAdminService) UserListByIds(ids []string) (res []User, err error) {
	var (
		resp struct {
			apiResultBase
			Data []User `json:"data"`
		}
		api = fmt.Sprintf("%s/api/user/admin/list", s.host)

		params = struct {
			Ids []string `json:"ids"`
		}{
			Ids: ids,
		}
	)

	err = s.client.CallWithJson(s.reqLogger, &resp, api, params)
	if err != nil || !resp.OK() {
		err = resp.Error()
		return
	}

	res = resp.Data
	return
}

type UserGetParams struct {
	SalesId string
	Email   string
}

func (p *UserGetParams) Values() url.Values {
	values := url.Values{}
	if p.Email != "" {
		values.Add("email", p.Email)
	} else {
		values.Add("salesId", p.SalesId)
	}

	return values
}

type User struct {
	Id          bson.ObjectId `json:"id,omitempty"`
	Name        string
	CnName      string
	Email       string
	Mobile      string
	QQ          string
	GitHub      string
	WikiDot     string
	Slack       string
	Extension   string
	SfSalesId   string `json:"sf_sales_id" bson:"sf_sales_id"`
	Status      byte
	Delete      byte
	IsTotpOpen  bool
	Create_time time.Time
	Update_time time.Time
}
