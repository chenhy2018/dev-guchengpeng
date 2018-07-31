package biz

import (
	"net/http"
	"strconv"
	"time"
)

import (
	"github.com/qiniu/rpc.v1"
	oldRpc "qbox.us/rpc"
)

type DeveloperGender int

const (
	DEVELOPER_GENDER_MALE   DeveloperGender = 0
	DEVELOPER_GENDER_FEMALE DeveloperGender = 1
)

type CustomerType int

const (
	CUSTOMER_TYPE_STANDARD CustomerType = 0
	CUSTOMER_TYPE_DISCOUNT CustomerType = 1 // 折扣用户
	CUSTOMER_TYPE_EXP      CustomerType = 2 // 体验用户
)

type Developer struct {
	Id                     string          `json:"id"`
	Uid                    uint32          `json:"uid"`
	Email                  string          `json:"email"`
	Fullname               string          `json:"fullname"`
	PhoneNumber            string          `json:"phone_number"`
	ImCategory             int             `json:"im_category"`
	ImNumber               string          `json:"im_number"`
	CompanyCategory        int             `json:"company_category"`
	CompanySize            int             `json:"company_size"`
	CompanyName            string          `json:"company_name"`
	Website                string          `json:"website"`
	Isactived              bool            `json:"isactived"`
	InternalCategory       int             `json:"internal_category"`
	InternalDepartment     int             `json:"internal_department"`
	Ispro                  bool            `json:"ispro"`
	Gender                 DeveloperGender `json:"gender"`
	LastPasswordModifyTime time.Time       `json:"last_password_modify_time"`
	CreatedAt              int64           `json:"created_at"`
	UpdatdAt               time.Time       `json:"updated_at"`
	MobileBinded           bool            `json:"mobile_binded"`
	CrmLeadId              string          `json:"crm_lead_id"`
	CustomerType           CustomerType    `json:"customer_type"`
}

type Contact struct {
	Email       string `json:"email" json:"email"`
	FullName    string `json:"fullname" bson:"fullname"`
	PhoneNumber string `json:"phone_number" bson:"phone_number"`
	ImCategory  int    `json:"im_category" bson:"im_category"`
	ImNumber    string `json:"im_number" bson:"im_number"`
	Gender      int    `json:"gender" bson:"gender"` //0 not set; 1 Male; 2 Female
	ContactType int    `json:"contact_type" bson:"contact_type"`
}

func (s *BizService) GetDeveloper(l rpc.Logger, uid uint32) (developer Developer, err error) {
	err = s.rpc.CallWithForm(l, &developer, s.host+"/admin/developer/get", map[string][]string{
		"uid": {strconv.FormatUint(uint64(uid), 10)},
	})
	return
}

func (s *BizService) GetDeveloperIdentity(l rpc.Logger, uid uint32) (developerIdentity DeveloperIdentity, err error) {
	err = s.rpc.CallWithForm(l, &developerIdentity, s.host+"/admin/developer/identity", map[string][]string{
		"uid": {strconv.FormatUint(uint64(uid), 10)},
	})
	return
}

func (s *BizService) CountDeveloper(l rpc.Logger) (res int, err error) {
	err = s.rpc.Call(l, &res, s.host+"/admin/developer/count")
	return
}

func (s *BizService) CountDeveloperByTime(l rpc.Logger, t time.Time) (res int, err error) {
	err = s.rpc.CallWithForm(l, &res, s.host+"/admin/developer/count", map[string][]string{
		"created_before": {strconv.FormatInt(t.Unix(), 10)},
	})
	return
}

func (s *BizService) ListDevelopersByUids(l rpc.Logger, uids []uint32) (res map[uint32]Developer, err error) {
	uidStrs := []string{}
	for _, u := range uids {
		uidStrs = append(uidStrs, strconv.FormatUint(uint64(u), 10))
	}

	developers := []Developer{}

	err = s.rpc.CallWithForm(l, &developers, s.host+"/admin/developer/listbyuids", map[string][]string{
		"uids": uidStrs,
	})

	if err != nil {
		return
	}

	res = make(map[uint32]Developer, len(developers))
	for _, d := range developers {
		res[d.Uid] = d
	}

	return
}

// List接口
// offset int
// length int
// internalCategory int, 0表示非内部帐号, 1表示内部帐号, 2表示测试帐号
// detail string, 值为"true"时(大小写敏感),会返回所有字段, 否则只返回email字段
func (s *BizService) List(l rpc.Logger, offset int, length int, internalCategory int, detail string) (res []Developer, err error) {
	lengthStr := strconv.FormatInt(int64(length), 10)
	offsetStr := strconv.FormatInt(int64(offset), 10)
	internalCategoryStr := strconv.FormatInt(int64(internalCategory), 10)
	err = s.rpc.CallWithForm(l, &res, s.host+"/admin/developer/list", map[string][]string{
		"length":            []string{lengthStr},
		"offset":            []string{offsetStr},
		"internal_category": []string{internalCategoryStr},
		"detail":            []string{detail},
	})
	return
}

// ---------------------
// 废弃API，兼容保留

type Service struct {
	Host string
	Conn oldRpc.Client
}

func NewService(host string, t http.RoundTripper) *Service {
	client := &http.Client{Transport: t}
	return &Service{host, oldRpc.Client{client}}
}

func (p *Service) GetDeveloper(uid uint32) (developer Developer, code int, err error) {
	code, err = p.Conn.CallWithForm(&developer,
		p.Host+"/admin/developer/get",
		map[string][]string{
			"uid": {strconv.FormatUint(uint64(uid), 10)},
		})
	return
}
