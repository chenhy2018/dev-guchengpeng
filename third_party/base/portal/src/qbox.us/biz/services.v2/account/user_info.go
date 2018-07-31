package account

import (
	"crypto/md5"
	"fmt"
	"io"
	"time"

	acc "qbox.us/servend/account"
)

type UserInfo struct {
	Uid            uint32       `json:"uid"`             // 用户数字ID唯一
	Username       string       `json:"username"`        // 用户名唯一，不可更换
	Email          string       `json:"email"`           // 电子邮箱唯一
	UserType       UserType     `json:"user_type"`       // 用户类型
	ParentUid      uint32       `json:"parent_uid"`      // 父用户Uid
	IsActivated    bool         `json:"is_activated"`    // 用户是否已经激活
	IsDisabled     bool         `json:"is_disabled"`     // 用户是否已经冻结
	DisabledType   DisabledType `json:"disabled_type"`   // 冻结类型
	DisabledReason string       `json:"disabled_reason"` // 冻结原因
	DisabledAt     time.Time    `json:"disabled_at"`     //冻结时间

	LastParentOperationAt time.Time `json:"last_parent_operation_at,omitempty"` // 父账户操作时间
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

func (cg CustomerGroup) IsInvalidUser() bool {
	return cg == CUSTOMER_GROUP_INVALID
}

func (cg CustomerGroup) IsExpUser() bool {
	return cg == CUSTOMER_GROUP_EXP
}

func (cg CustomerGroup) IsStdUser() bool {
	return cg == CUSTOMER_GROUP_NORMAL
}

func (cg CustomerGroup) IsVipUser() bool {
	return cg == CUSTOMER_GROUP_VIP
}

func (i UserInfo) GetCustomerGroup() CustomerGroup {
	if i.UserType&acc.USER_TYPE_USERS == 0 {
		return CUSTOMER_GROUP_INVALID
	}
	if i.UserType&acc.USER_TYPE_EXPUSER != 0 {
		return CUSTOMER_GROUP_EXP
	}
	if i.UserType&acc.USER_TYPE_VIP != 0 {
		return CUSTOMER_GROUP_VIP
	}
	return CUSTOMER_GROUP_NORMAL
}

func (u *UserInfo) FamilyType() FamilyType {
	if u.UserType.IsParentUser() {
		return FamilyParent
	}

	if u.ParentUid > 0 {
		return FamilyChild
	}

	return FamilyNormal
}

func (u *UserInfo) EmailMd5() string {
	if u.Email == "" {
		return ""
	}

	m := md5.New()
	io.WriteString(m, u.Email)

	return fmt.Sprintf("%x", m.Sum(nil))
}
