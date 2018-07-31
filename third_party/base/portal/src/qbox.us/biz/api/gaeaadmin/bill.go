package gaeaadmin

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"qbox.us/biz/utils.v2/json"
	"qbox.us/biz/utils.v2/validator"
)

const (
	GetBillListPath            = "%s/api/user/%d/bills"
	GetBillPath                = "%s/api/user/%d/bills/%s"
	GetBillPDFPath             = "%s/api/user/%d/bills/%s/pdf"
	GetMergeAccountBillPath    = "%s/api/user/%d/merge-account-bills/%s"
	GetMergeAccountBillPDFPath = "%s/api/user/%d/merge-account-bills/%s/pdf"

	GetBillListV2Path = "%s/api/user/%d/bills/list"
	GetBillV2Path     = "%s/api/user/%d/bills/%s/detail"

	DefaultMonthLayout = "200601"

	StatusCodeStatementNotFound = 7721 //账单详情不存在
)

var (
	ErrorInvalidEmail = errors.New("invalid email")
)

type GetBillListOut struct {
	Bills   []BillBrief `json:"bills"`   // 账单列表
	Summary bool        `json:"summary"` // 是否为汇总账单
}

type BillBrief struct {
	Id     string    `json:"id"`
	Month  time.Time `json:"month"`
	Money  int64     `json:"money"` //价格要/10000
	Status int       `json:"status"`
}

type getBillListResp struct {
	json.CommonResponse

	Data GetBillListOut `json:"data"`
}

func (s *gaeaAdminService) GetBillList(uid uint32, start time.Time, end time.Time, email string) (billList GetBillListOut, err error) {
	var out getBillListResp

	_url := fmt.Sprintf(GetBillListPath, s.host, uid)

	query := url.Values{}

	if !start.IsZero() {
		query.Add("start", start.Format(DefaultMonthLayout))
	}

	if !end.IsZero() {
		query.Add("end", end.Format(DefaultMonthLayout))
	}

	if email != "" {
		if !validator.IsEmail(email) {
			err = ErrorInvalidEmail
			return
		}

		query.Add("email", email)
	}

	err = s.client.GetCall(s.reqLogger, &out, _url+"?"+query.Encode())
	if err != nil {
		return
	}

	err = out.Error()

	if err == nil {
		billList = out.Data
	}

	return
}

func (s *gaeaAdminService) GetBill(uid uint32, month time.Time) (billHtml []byte, err error) {
	_url := fmt.Sprintf(GetBillPath, s.host, uid, month.Format(DefaultMonthLayout))

	resp, err := s.client.Get(s.reqLogger, _url)
	if err != nil {
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode == StatusCodeStatementNotFound {
		billHtml = []byte("暂无账单")
		return
	}

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("get bill failed with code: %d", resp.StatusCode)
		return
	}

	billHtml, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	return
}

func (s *gaeaAdminService) GetBillPDF(uid uint32, month time.Time) (billPdf []byte, err error) {
	_url := fmt.Sprintf(GetBillPDFPath, s.host, uid, month.Format(DefaultMonthLayout))

	resp, err := s.client.Get(s.reqLogger, _url)
	if err != nil {
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode == StatusCodeStatementNotFound {
		err = errors.New("bill not found")
		return
	}

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("get bill failed with code: %d", resp.StatusCode)
		return
	}

	billPdf, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	return
}

func (s *gaeaAdminService) GetMergeAccountBill(uid uint32, month time.Time) (billHtml []byte, err error) {
	_url := fmt.Sprintf(GetMergeAccountBillPath, s.host, uid, month.Format(DefaultMonthLayout))

	resp, err := s.client.Get(s.reqLogger, _url)
	if err != nil {
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode == StatusCodeStatementNotFound {
		billHtml = []byte("暂无账单")
		return
	}

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("get bill failed with code: %d", resp.StatusCode)
		return
	}

	billHtml, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	return
}

func (s *gaeaAdminService) GetMergeAccountBillPDF(uid uint32, month time.Time) (billPdf []byte, err error) {
	_url := fmt.Sprintf(GetMergeAccountBillPDFPath, s.host, uid, month.Format(DefaultMonthLayout))

	resp, err := s.client.Get(s.reqLogger, _url)
	if err != nil {
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode == StatusCodeStatementNotFound {
		err = errors.New("bill not found")
		return
	}

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("get bill failed with code: %d", resp.StatusCode)
		return
	}

	billPdf, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	return
}

func (s *gaeaAdminService) GetBillListV2(uid uint32, start time.Time, end time.Time, merge bool) (billList GetBillListOut, err error) {
	var out getBillListResp

	_url := fmt.Sprintf(GetBillListV2Path, s.host, uid)

	query := url.Values{}

	if !start.IsZero() {
		query.Add("start", start.Format(DefaultMonthLayout))
	}

	if !end.IsZero() {
		query.Add("end", end.Format(DefaultMonthLayout))
	}

	query.Add("merge", strconv.FormatBool(merge))

	err = s.client.GetCall(s.reqLogger, &out, _url+"?"+query.Encode())
	if err != nil {
		return
	}

	err = out.Error()

	if err == nil {
		billList = out.Data
	}

	return
}

func (s *gaeaAdminService) GetBillV2(uid uint32, month time.Time, merge bool) (billHtml []byte, err error) {
	_url := fmt.Sprintf(GetBillV2Path, s.host, uid, month.Format(DefaultMonthLayout))

	query := url.Values{}
	query.Add("merge", strconv.FormatBool(merge))

	resp, err := s.client.Get(s.reqLogger, _url+"?"+query.Encode())
	if err != nil {
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode == StatusCodeStatementNotFound {
		billHtml = []byte("暂无账单")
		return
	}

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("get bill failed with code: %d", resp.StatusCode)
		return
	}

	billHtml, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	return
}
