package account

import (
	"time"
)

type Info struct {
	UserInfo

	DisabledType     DisabledType `json:"disabled_type"`      // 用户冻结类型
	DisabledReason   string       `json:"disabled_reason"`    // 用户冻结原因
	Vendors          []*Vendor    `json:"vendors"`            //
	ChildEmailDomain string       `json:"child_email_domain"` //
	CanGetChildKey   bool         `json:"can_get_child_key"`  // 用户可以获取子账户的 AK/SK
	CreatedAt        time.Time    `json:"created_at"`         // 用户创建时间
	UpdatedAt        time.Time    `json:"updated_at"`         // 最后一次修改时间
	LastLoginAt      time.Time    `json:"last_login_at"`      // 最后一次登录时间
	DisabledAt       time.Time    `json:"disabled_at"`        // 用户冻结时间
}
