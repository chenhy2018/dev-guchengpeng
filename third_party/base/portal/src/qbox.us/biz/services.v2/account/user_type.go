package account

import (
	"qbox.us/admin_api/account.v2"
)

const (
	// user type
	USER_TYPE_QBOX         UserType = 0
	USER_TYPE_ADMIN                 = 0x0001
	USER_TYPE_VIP                   = 0x0002
	USER_TYPE_STDUSER               = 0x0004
	USER_TYPE_STDUSER2              = 0x0008
	USER_TYPE_EXPUSER               = 0x0010
	USER_TYPE_PARENTUSER            = 0x0020
	USER_TYPE_OP                    = 0x0040
	USER_TYPE_SUPPORT               = 0x0080
	USER_TYPE_CC                    = 0x0100
	USER_TYPE_QCOS                  = 0x0200
	USER_TYPE_FUSION                = 0x0400
	USER_TYPE_PILI                  = 0x0800
	USER_TYPE_PANDORA               = 0x1000
	USER_TYPE_DISTRIBUTION          = 0x2000
	USER_TYPE_QVM                   = 0x4000
	USER_TYPE_DISABLED              = 0x8000

	USER_TYPE_USERS            = USER_TYPE_STDUSER | USER_TYPE_STDUSER2 | USER_TYPE_EXPUSER // 一个用户必须存在的3种类型，除此之外就是无效用户类型
	USER_TYPE_SUDOERS          = USER_TYPE_ADMIN | USER_TYPE_OP | USER_TYPE_SUPPORT         // 管理员
	USER_TYPE_ENTERPRISE       = USER_TYPE_STDUSER
	USER_TYPE_ENTERPRISE_VUSER = USER_TYPE_STDUSER2
)

type UserType uint32

// admin用户
func (t UserType) IsAdmin() bool {
	return t&USER_TYPE_ADMIN != 0
}

// 无效用户
func (t UserType) IsInvalid() bool {
	return getCustomerGroup(t) == account.CUSTOMER_GROUP_INVALID
}

// 类型：体验用户
func (t UserType) IsExpUser() bool {
	return getCustomerGroup(t) == account.CUSTOMER_GROUP_EXP
}

// 类型：高级用户
func (t UserType) IsVipUser() bool {
	return getCustomerGroup(t) == account.CUSTOMER_GROUP_VIP
}

// 类型：标准用户
func (t UserType) IsNormalUser() bool {
	return getCustomerGroup(t) == account.CUSTOMER_GROUP_NORMAL
}

func (t UserType) Humanize() string {
	return getCustomerGroup(t).Humanize()
}

// 标志位：父账户
func (t UserType) IsParentUser() bool {
	return t&USER_TYPE_PARENTUSER > 0
}

// 标志位：企业账户
func (t UserType) IsEnterpriseUser() bool {
	return t&USER_TYPE_STDUSER2 > 0
}

// 标志位：CC用户
func (t UserType) IsCCUser() bool {
	return t&USER_TYPE_CC > 0
}

// 标志位：Fusion用户
func (t UserType) IsFusionUser() bool {
	return t&USER_TYPE_FUSION > 0
}

// 标志位：QCos用户
func (t UserType) IsQCosUser() bool {
	return t&USER_TYPE_QCOS > 0
}

func (t UserType) IsPiliUser() bool {
	return t&USER_TYPE_PILI > 0
}

func (t UserType) IsPandoraUser() bool {
	return t&USER_TYPE_PANDORA > 0
}

func (t UserType) IsDistributionUser() bool {
	return t&USER_TYPE_DISTRIBUTION > 0
}

func (t UserType) IsQvmUser() bool {
	return t&USER_TYPE_QVM > 0
}

func (t UserType) IsDisabled() bool {
	return t&USER_TYPE_DISABLED > 0
}

// 标志位：冻结用户
func (t UserType) SetDisabled() UserType {
	t |= USER_TYPE_DISABLED
	return t
}

// 标志位：启用用户
func (t UserType) SetEnabled() UserType {
	t &^= USER_TYPE_DISABLED
	return t
}

func (t UserType) Priority() int {
	if t.IsExpUser() {
		return 1
	}
	if t.IsNormalUser() {
		return 2
	}
	if t.IsVipUser() {
		return 3
	}
	return 0
}

// TODO 对用户类型的操作

func getCustomerGroup(uType UserType) account.CustomerGroup {
	if uType&USER_TYPE_USERS == 0 {
		return account.CUSTOMER_GROUP_INVALID
	}
	if uType&USER_TYPE_EXPUSER != 0 {
		return account.CUSTOMER_GROUP_EXP
	}
	if uType&USER_TYPE_VIP != 0 {
		return account.CUSTOMER_GROUP_VIP
	}
	return account.CUSTOMER_GROUP_NORMAL
}
