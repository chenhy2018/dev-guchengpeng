// https://github.com/qbox/service/blob/develop/apidoc/v6/acc.md
package account

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/qiniu/log.v1"
	"github.com/qiniu/rpc.v1"
	"github.com/qiniu/rpc.v1/lb.v2.1"
	"qbox.us/api/account"
	"qbox.us/oauth"
	acc "qbox.us/servend/account"
)

type Second int64

func (s Second) Time() time.Time {
	return time.Unix(int64(s), 0)
}

type DisabledType int

const (
	DISABLED_TYPE_AUTO           DisabledType = 0 // 冻结后允许充值自动解冻
	DISABLED_TYPE_MANUAL         DisabledType = 1 // 冻结后需要手动解冻
	DISABLED_TYPE_PARENT         DisabledType = 2 // 被父账号冻结
	DISABLED_TYPE_OVERDUE        DisabledType = 3
	DISABLED_TYPE_NONSTD_OVERDUE DisabledType = 4
)

func (t DisabledType) Humanize() string {
	switch t {
	case DISABLED_TYPE_AUTO:
		return "欠费冻结"
	case DISABLED_TYPE_MANUAL:
		return "非欠费冻结"
	case DISABLED_TYPE_PARENT:
		return "被父账号冻结"
	case DISABLED_TYPE_OVERDUE:
		return "实时计费远超余额"
	case DISABLED_TYPE_NONSTD_OVERDUE:
		return "未认证用户超过免费额度"
	default:
		return fmt.Sprintf("unknown DisabledType: %d", t)
	}
}

type Info struct {
	Id               string       `bson:"id" json:"id"`                           // 用户名(UserName)。唯一。
	Email            string       `bson:"email" json:"email"`                     // 电子邮箱。唯一。
	Username         string       `bson:"username" json:"username"`               // 用户名。唯一。
	CreatedAt        Second       `bson:"ctime" json:"ctime"`                     // 用户创建时间。
	UpdatedAt        Second       `bson:"etime" json:"etime"`                     // 最后一次修改时间。
	LastLoginAt      Second       `bson:"lgtime" json:"lgtime"`                   // 最后一次登录时间。
	Uid              uint32       `bson:"uid" json:"uid"`                         // 用户数字ID。唯一。
	Utype            uint32       `bson:"utype" json:"utype"`                     // 用户类型。
	ParentUid        uint32       `bson:"parent_uid" json:"parent_uid"`           // 父用户Uid
	Activated        bool         `bson:"activated" json:"activated"`             // 用户是否已经激活。
	DisabledType     DisabledType `bson:"disabled_type" json:"disabled_type"`     // 用户冻结类型
	DisabledReason   string       `bson:"disabled_reason" json:"disabled_reason"` // 用户冻结原因
	DisabledAt       time.Time    `bson:"disabled_at" json:"disabled_at"`         // 用户冻结时间
	Vendors          []Vendor     `bson:"vendors" json:"vendors"`
	ChildEmailDomain string       `bson:"child_email_domain" json:"child_email_domain"`
	CanGetChildKey   bool         `bson:"can_get_child_key" json:"can_get_child_key"`
}

type Vendor struct {
	Vendor      string    `bson:"vendor" json:"vendor"`
	VendorId    string    `bson:"vendor_id" json:"vendor_id"`
	VendorEmail string    `bson:"vendor_email" json:"vendor_email"`
	CreatedAt   time.Time `bson:"created_at" json:"created_at"`
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
	Host string // 功能废除 兼容保留
	Conn *lb.Client
}

const (
	VENDOR_GITHUB = "github"
	VENDOR_CSDN   = "csdn"
	VENDOR_WEIBO  = "weibo"
)

func New(host string, t http.RoundTripper) *Service {
	cfg := &lb.Config{
		Hosts:    []string{host},
		TryTimes: 1,
	}
	client := lb.New(cfg, t)
	return &Service{host, client}
}

func NewWithMultiHosts(hosts []string, t http.RoundTripper) *Service {
	cfg := &lb.Config{
		Hosts:    hosts,
		TryTimes: uint32(len(hosts)),
	}
	client := lb.New(cfg, t)
	return &Service{Conn: client}
}

// NewService 内部使用 qbox.us/oauth.Transport 创建了一个 Service 对象。
// 返回的 Service 对象会继承 qbox.us/oauth.Transport 的已知限制，详见 qbox.us/oauth.Transport 的定义。
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

func (r *Service) ListUsersByUtype(utype uint32, offset, limit int, l rpc.Logger) (infos []Info, err error) {
	err = r.Conn.CallWithForm(l, &infos, "/admin/users", map[string][]string{
		"utype":  {strconv.FormatUint(uint64(utype), 10)},
		"offset": {strconv.FormatInt(int64(offset), 10)},
		"limit":  {strconv.FormatInt(int64(limit), 10)},
	})
	return
}

func (r *Service) ListUsersByUids(uids []uint32, l rpc.Logger) (infos map[uint32]Info, err error) {
	infos = map[uint32]Info{}
	infoArray := []Info{}
	uidStrs := []string{}
	for _, uid := range uids {
		uidStrs = append(uidStrs, strconv.FormatUint(uint64(uid), 10))
	}
	err = r.Conn.CallWithForm(l, &infoArray, "/admin/users", map[string][]string{
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

func (r *Service) ListUsersByLastLoginTime(from, to time.Time, offset, limit int, l rpc.Logger) (infos []Info, err error) {
	err = r.Conn.CallWithForm(l, &infos, "/admin/users", map[string][]string{
		"last_login_time_from": {strconv.FormatInt(from.Unix(), 10)},
		"last_login_time_to":   {strconv.FormatInt(to.Unix(), 10)},
		"offset":               {strconv.FormatInt(int64(offset), 10)},
		"limit":                {strconv.FormatInt(int64(limit), 10)},
	})
	return
}

func (r *Service) UserInfoById(id string, l rpc.Logger) (info Info, err error) {
	err = r.Conn.CallWithForm(l, &info, "/admin/user/info", map[string][]string{
		"id": {id},
	})
	return
}

func (r *Service) UserInfoByUid(uid uint32, l rpc.Logger) (info Info, err error) {
	err = r.Conn.CallWithForm(l, &info, "/admin/user/info", map[string][]string{
		"uid": {strconv.FormatUint(uint64(uid), 10)},
	})
	return
}

func (r *Service) UserInfoByVendor(vendor, vendorId string, l rpc.Logger) (info Info, err error) {
	err = r.Conn.CallWithForm(l, &info, "/admin/user/info", map[string][]string{
		"vendor":    {vendor},
		"vendor_id": {vendorId},
	})
	return
}

func (r *Service) UserSetUserType(userName string, utype uint32, l rpc.Logger) (info Info, err error) {
	err = r.Conn.CallWithForm(l, &info, "/admin/user/set_user_type", map[string][]string{
		"user_id":   {userName},
		"user_type": {strconv.FormatUint(uint64(utype), 10)},
	})
	return
}

func (r *Service) UserCreateByVendor(vendor, vendorId, vendorEmail string, l rpc.Logger) (info Info, err error) {
	err = r.Conn.CallWithForm(l, &info, "/admin/user/create", map[string][]string{
		"vendor":       {vendor},
		"vendor_id":    {vendorId},
		"vendor_email": {vendorEmail},
	})
	return
}

func (r *Service) UserCreateByPassword(email, password string, l rpc.Logger) (info Info, err error) {
	err = r.Conn.CallWithForm(l, &info, "/admin/user/create", map[string][]string{
		"email":    {email},
		"password": {password},
	})
	return
}

func (r *Service) UserBindAccount(uid uint32, vendor, vendorId, vendorEmail string, l rpc.Logger) (info Info, err error) {
	err = r.Conn.CallWithForm(l, &info, "/admin/user/bind_account", map[string][]string{
		"uid":          {strconv.FormatUint(uint64(uid), 10)},
		"vendor":       {vendor},
		"vendor_id":    {vendorId},
		"vendor_email": {vendorEmail},
	})
	return
}

func (r *Service) UserDisable(uid uint32, reason string, disabledType DisabledType, l rpc.Logger) (info Info, err error) {
	err = r.Conn.CallWithForm(l, &info, "/admin/user/disable", map[string][]string{
		"uid":    {strconv.FormatUint(uint64(uid), 10)},
		"reason": {reason},
		"type":   {fmt.Sprintf("%d", disabledType)},
	})
	return
}

func (r *Service) UserAutoEnable(uid uint32, l rpc.Logger) (info Info, err error) {
	err = r.Conn.CallWithForm(l, &info, "/admin/user/auto_enable", map[string][]string{
		"uid": {strconv.FormatUint(uint64(uid), 10)},
	})
	return
}

func (r *Service) UserForceEnable(uid uint32, l rpc.Logger) (info Info, err error) {
	err = r.Conn.CallWithForm(l, &info, "/admin/user/force_enable", map[string][]string{
		"uid": {strconv.FormatUint(uint64(uid), 10)},
	})
	return
}

func (r *Service) UserUnbindAccount(uid uint32, vendor string, l rpc.Logger) (info Info, err error) {
	err = r.Conn.CallWithForm(l, &info, "/admin/user/unbind_account", map[string][]string{
		"uid":    {strconv.FormatUint(uint64(uid), 10)},
		"vendor": {vendor},
	})
	return
}

func (r *Service) UserSetCustomerGroup(uid uint32, cg CustomerGroup, l rpc.Logger) (info Info, err error) {
	err = r.Conn.CallWithForm(l, &info, "/admin/user/set_customer_group", map[string][]string{
		"uid":            {strconv.FormatUint(uint64(uid), 10)},
		"customer_group": {strconv.FormatInt(int64(cg), 10)},
	})
	return
}

func (r *Service) UserSetPassword(uid uint32, password string, l rpc.Logger) (info Info, err error) {
	err = r.Conn.CallWithForm(l, &info, "/admin/user/set_password", map[string][]string{
		"uid":      {strconv.FormatUint(uint64(uid), 10)},
		"password": {password},
	})
	return
}

func (r *Service) UserSetEmail(uid uint32, email string, l rpc.Logger) (info Info, err error) {
	err = r.Conn.CallWithForm(l, &info, "/admin/user/set_email", map[string][]string{
		"uid":   {strconv.FormatUint(uint64(uid), 10)},
		"email": {email},
	})
	return
}
func (r *Service) UserUpdate(uid uint32, param url.Values, l rpc.Logger) (info Info, err error) {
	err = r.Conn.CallWithForm(l, &info, fmt.Sprintf("/admin/user/update?uid=%d", uid), param)
	return
}

func (r *Service) TokenCreate(uid uint32, l rpc.Logger) (token oauth.Token, err error) {
	err = r.Conn.CallWithForm(l, &token, "/admin/token/create", map[string][]string{
		"uid": {strconv.FormatUint(uint64(uid), 10)},
	})
	return
}

func (r *Service) ListUsersByMarker(marker uint32, limit int, l rpc.Logger) (infos []Info, err error) {
	err = r.Conn.CallWithForm(l, &infos, "/v1/admin/user", map[string][]string{
		"marker": {strconv.FormatUint(uint64(marker), 10)},
		"limit":  {strconv.FormatInt(int64(limit), 10)},
	})
	return
}

func (r *Service) ListChildrenByMarker(parent, marker uint32, limit int, l rpc.Logger) (infos []Info, err error) {
	err = r.Conn.CallWithForm(l, &infos, "/v1/admin/user", map[string][]string{
		"parent": {strconv.FormatUint(uint64(parent), 10)},
		"marker": {strconv.FormatUint(uint64(marker), 10)},
		"limit":  {strconv.FormatInt(int64(limit), 10)},
	})
	return
}

func (r *Service) CountChildren(parent uint32, l rpc.Logger) (count int, err error) {
	var retCount = struct {
		Count int `json:"count"`
	}{}

	err = r.Conn.CallWithForm(l, &retCount, "/admin/user/count_children", map[string][]string{
		"uid": {strconv.FormatUint(uint64(parent), 10)},
	})

	count = retCount.Count
	return
}

// DEPRECATED API

func (r *Service) ListUsers(offset, limit int, l rpc.Logger) (infos []Info, err error) {
	log.Warn("[DEPRECATED] please use ListUsersByMarker")
	err = r.Conn.CallWithForm(l, &infos, "/admin/users", map[string][]string{
		"offset": {strconv.FormatInt(int64(offset), 10)},
		"limit":  {strconv.FormatInt(int64(limit), 10)},
	})
	return
}

func (r *Service) UserChildren(uid uint32, offset int, limit int, l rpc.Logger) (infos []Info, err error) {
	log.Warn("[DEPRECATED] please use ListChildrenByMarker")
	err = r.Conn.Call(l, &infos,
		fmt.Sprintf("/admin/user/children?uid=%d&offset=%d&limit=%d", uid, offset, limit))
	return
}

func (r *Service) UserCreate(vendor, vendorId, vendorEmail string, l rpc.Logger) (info Info, err error) {
	log.Warn("[DEPRECATED] please use UserCreateByVendor")
	return r.UserCreateByVendor(vendor, vendorId, vendorEmail, l)
}
