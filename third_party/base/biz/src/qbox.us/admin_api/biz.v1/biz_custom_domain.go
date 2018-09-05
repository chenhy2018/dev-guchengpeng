package biz

import (
	"net/url"
	"strconv"

	"github.com/qiniu/rpc.v1"
)

type CustomDomainData struct {
	Uid       uint32 `json:"uid,string"`
	Domain    string `json:"doamin"`
	CNAME     string `json:"cname"`
	UpdatedAt int64  `json:"updated_at"`
}

type CustomDomainData2 struct {
	Uid       uint32 `json:"uid,string"`
	Domain    string `json:"doamin"`
	CNAME     string `json:"cname"`
	UpdatedAt int64  `json:"updated_at"`
	Channel   string `json:"channel"`
}

type ListActiveCustomDomainResponse struct {
	Domains []CustomDomainData2 `json:"domains"`
	End     bool                `json:"end"`
}

func (s *BizService) CustomDomainListActive(l rpc.Logger, offset, limit int) (
	res []CustomDomainData, err error) {

	params := url.Values{
		"offset": {strconv.FormatInt(int64(offset), 10)},
		"limit":  {strconv.FormatInt(int64(limit), 10)},
	}

	err = s.rpc.CallWithForm(l, &res, s.host+"/admin/cdomain/list_active", params)
	return
}

func (s *BizService) CustomDomainListActive2(l rpc.Logger, offset, limit int) (
	res []ListActiveCustomDomainResponse, err error) {

	params := url.Values{
		"offset": {strconv.FormatInt(int64(offset), 10)},
		"limit":  {strconv.FormatInt(int64(limit), 10)},
	}

	err = s.rpc.CallWithForm(l, &res, s.host+"/admin/cdomain/list_active2", params)
	return
}
