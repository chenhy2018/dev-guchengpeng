// This Package is Deprecated, please use qbox.us/admin_api/v2/account .
package account

import (
	"fmt"
	"net/http"
	"strconv"
	"time"
)

import (
	"qbox.us/api/account"
	"qbox.us/oauth"
	"qbox.us/rpc"
	acc "qbox.us/servend/account"
)

type Second int64

func (s Second) Time() time.Time {
	return time.Unix(int64(s), 0)
}

type DisabledType int

const (
	DISABLED_TYPE_AUTO   DisabledType = 0 // 冻结后允许充值自动解冻
	DISABLED_TYPE_MANUAL DisabledType = 1 // 冻结后需要手动解冻
)

func (t DisabledType) Humanize() string {
	switch t {
	case DISABLED_TYPE_AUTO:
		return "欠费冻结"
	case DISABLED_TYPE_MANUAL:
		return "手动冻结"
	default:
		return fmt.Sprintf("unknown DisabledType: %d", t)
	}
}

type Info struct {
	Id             string       `json:"id"`              // 用户名(UserName)。唯一。
	Email          string       `json:"email"`           // 电子邮箱。唯一。
	CreatedAt      Second       `json:"ctime"`           // 用户创建时间。
	UpdatedAt      Second       `json:"etime"`           // 最后一次修改时间。
	LastLoginAt    Second       `json:"lgtime"`          // 最后一次登录时间。
	Uid            uint32       `json:"uid"`             // 用户数字ID。唯一。
	Utype          uint32       `json:"utype"`           // 用户类型。
	Activated      bool         `json:"activated"`       // 用户是否已经激活。
	DisabledType   DisabledType `json:"disabled_type"`   // 用户冻结类型
	DisabledReason string       `json:"disabled_reason"` // 用户冻结原因
	DisabledAt     time.Time    `json:"disabled_at"`     // 用户冻结时间
	Vendors        []Vendor     `json:"vendors"`
}

type CustomerGroup int

const (
	CUSTOMER_GROUP_EXP     CustomerGroup = 0
	CUSTOMER_GROUP_NORMAL  CustomerGroup = 1
	CUSTOMER_GROUP_VIP     CustomerGroup = 2
	CUSTOMER_GROUP_INVALID CustomerGroup = 3
)

func (cg CustomerGroup) Humanize() string {
	switch cg {
	case CUSTOMER_GROUP_EXP:
		return "体验用户"
	case CUSTOMER_GROUP_NORMAL:
		return "标准用户"
	case CUSTOMER_GROUP_VIP:
		return "高级用户"
	case CUSTOMER_GROUP_INVALID:
		return "无效用户"
	default:
		return fmt.Sprintf("未知用户类型: %d", cg)
	}
}

func (i Info) GetCustomerGroup() CustomerGroup {
	if i.Utype&acc.USER_TYPE_USERS == 0 {
		return CUSTOMER_GROUP_INVALID
	}
	if i.Utype&acc.USER_TYPE_EXPUSER != 0 {
		return CUSTOMER_GROUP_EXP
	}
	if i.Utype&acc.USER_TYPE_VIP != 0 {
		return CUSTOMER_GROUP_VIP
	}
	return CUSTOMER_GROUP_NORMAL
}

// convert to "qbox.us/api/account".UserInfo
func (i Info) UserInfo() account.UserInfo {
	return account.UserInfo{
		Uid:           i.Uid,
		UserId:        i.Id,
		Email:         i.Email,
		IsActivated:   i.Activated,
		UserType:      i.Utype,
		DeviceNum:     0,
		InvitationNum: 0,
	}
}

func (i Info) IsDisabled() bool {
	return i.Utype&acc.USER_TYPE_DISABLED != 0
}

func (i *Info) Disable() {
	i.Utype |= acc.USER_TYPE_DISABLED
}

func (i *Info) Enable() {
	i.Utype &^= acc.USER_TYPE_DISABLED
}

type Service struct {
	Host string
	Conn rpc.Client
}

type Vendor struct {
	Vendor      string    `json:"vendor"`
	VendorId    string    `json:"vendor_id"`
	VendorEmail string    `json:"vendor_email"`
	CreatedAt   time.Time `json:"created_at"`
}

const (
	VENDOR_GITHUB = "github"
	VENDOR_CSDN   = "csdn"
	VENDOR_WEIBO  = "weibo"
)

func New(host string, t http.RoundTripper) *Service {
	client := &http.Client{Transport: t}
	return &Service{host, rpc.Client{client}}
}

func NewService(host, clientId, clientSecret, username, password string) (service *Service, err error) {
	transport := &oauth.Transport{
		Config: &oauth.Config{
			ClientId:     clientId,
			ClientSecret: clientSecret,
			Scope:        "Scope",
			AuthURL:      "",
			TokenURL:     host + "/oauth2/token",
			RedirectURL:  "",
		},
		Transport: http.DefaultTransport, // it is default
	}
	_, _, err = transport.ExchangeByPassword(username, password)
	if err != nil {
		return
	}
	service = New(host, transport)
	return
}

func (r *Service) ListUsers(offset, limit int) (infos []Info, code int, err error) {
	code, err = r.Conn.CallWithForm(&infos, r.Host+"/admin/users", map[string][]string{
		"offset": {strconv.FormatInt(int64(offset), 10)},
		"limit":  {strconv.FormatInt(int64(limit), 10)},
	})
	return
}

func (r *Service) ListUsersByUtype(utype uint32, offset, limit int) (infos []Info, code int, err error) {
	code, err = r.Conn.CallWithForm(&infos, r.Host+"/admin/users", map[string][]string{
		"utype":  {strconv.FormatUint(uint64(utype), 10)},
		"offset": {strconv.FormatInt(int64(offset), 10)},
		"limit":  {strconv.FormatInt(int64(limit), 10)},
	})
	return
}

func (r *Service) ListUsersByUids(uids []uint32) (infos map[uint32]Info, code int, err error) {
	infos = map[uint32]Info{}
	infoArray := []Info{}
	uidStrs := []string{}
	for _, uid := range uids {
		uidStrs = append(uidStrs, strconv.FormatUint(uint64(uid), 10))
	}
	code, err = r.Conn.CallWithForm(&infoArray, r.Host+"/admin/users", map[string][]string{
		"uids": uidStrs,
	})
	if err != nil {
		return
	}
	for _, info := range infoArray {
		infos[info.Uid] = info
	}
	return
}

// -----------------------------------------------------------
// [DEPRECATED], please use UserInfoById
func (r *Service) Info(userName string) (info Info, code int, err error) {
	code, err = r.Conn.CallWithForm(&info, r.Host+"/admin/user_info", map[string][]string{
		"id": {userName},
	})
	return
}

func (r *Service) UserInfoById(id string) (info Info, code int, err error) {
	code, err = r.Conn.CallWithForm(&info, r.Host+"/admin/user/info", map[string][]string{
		"id": {id},
	})
	return
}

func (r *Service) UserInfoByUid(uid uint32) (info Info, code int, err error) {
	code, err = r.Conn.CallWithForm(&info, r.Host+"/admin/user/info", map[string][]string{
		"uid": {strconv.FormatUint(uint64(uid), 10)},
	})
	return
}

func (r *Service) UserInfoByVendor(vendor, vendorId string) (info Info, code int, err error) {
	code, err = r.Conn.CallWithForm(&info, r.Host+"/admin/user/info", map[string][]string{
		"vendor":    {vendor},
		"vendor_id": {vendorId},
	})
	return
}

// -----------------------------------------------------------

func (r *Service) SetUserType(userName string, utypeMask, utype uint32) (code int, err error) {
	code, err = r.Conn.CallWithForm(nil, r.Host+"/admin/set_user_type", map[string][]string{
		"id":   {userName},
		"mask": {strconv.FormatUint(uint64(utypeMask), 10)},
		"type": {strconv.FormatUint(uint64(utype), 10)},
	})
	return
}

func (r *Service) UserCreate(vendor, vendorId, vendorEmail string) (info Info, code int, err error) {
	code, err = r.Conn.CallWithForm(&info, r.Host+"/admin/user/create", map[string][]string{
		"vendor":       {vendor},
		"vendor_id":    {vendorId},
		"vendor_email": {vendorEmail},
	})
	return
}

func (r *Service) UserBindAccount(uid uint32, vendor, vendorId, vendorEmail string) (info Info, code int, err error) {
	code, err = r.Conn.CallWithForm(&info, r.Host+"/admin/user/bind_account", map[string][]string{
		"uid":          {strconv.FormatUint(uint64(uid), 10)},
		"vendor":       {vendor},
		"vendor_id":    {vendorId},
		"vendor_email": {vendorEmail},
	})
	return
}

func (r *Service) UserDisable(uid uint32, reason string, disabledType DisabledType) (info Info, code int, err error) {
	code, err = r.Conn.CallWithForm(&info, r.Host+"/admin/user/disable", map[string][]string{
		"uid":    {strconv.FormatUint(uint64(uid), 10)},
		"reason": {reason},
		"type":   {fmt.Sprintf("%d", disabledType)},
	})
	return
}

func (r *Service) UserAutoEnable(uid uint32) (info Info, code int, err error) {
	code, err = r.Conn.CallWithForm(&info, r.Host+"/admin/user/auto_enable", map[string][]string{
		"uid": {strconv.FormatUint(uint64(uid), 10)},
	})
	return
}

func (r *Service) UserForceEnable(uid uint32) (info Info, code int, err error) {
	code, err = r.Conn.CallWithForm(&info, r.Host+"/admin/user/force_enable", map[string][]string{
		"uid": {strconv.FormatUint(uint64(uid), 10)},
	})
	return
}

func (r *Service) UserUnbindAccount(uid uint32, vendor string) (info Info, code int, err error) {
	code, err = r.Conn.CallWithForm(&info, r.Host+"/admin/user/unbind_account", map[string][]string{
		"uid":    {strconv.FormatUint(uint64(uid), 10)},
		"vendor": {vendor},
	})
	return
}

func (r *Service) UserSetCustomerGroup(uid uint32, cg CustomerGroup) (info Info, code int, err error) {
	code, err = r.Conn.CallWithForm(&info, r.Host+"/admin/user/set_customer_group", map[string][]string{
		"uid":            {strconv.FormatUint(uint64(uid), 10)},
		"customer_group": {strconv.FormatInt(int64(cg), 10)},
	})
	return
}

func (r *Service) TokenCreate(uid uint32) (token oauth.Token, code int, err error) {
	code, err = r.Conn.CallWithForm(&token, r.Host+"/admin/token/create", map[string][]string{
		"uid": {strconv.FormatUint(uint64(uid), 10)},
	})
	return
}

// -----------------------------------------------------------

// depreacted by UserAutoEnable
func (r *Service) UserEnable(uid uint32) (info Info, code int, err error) {
	return r.UserAutoEnable(uid)
}
