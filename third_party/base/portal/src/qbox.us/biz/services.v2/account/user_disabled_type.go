package account

import (
	"fmt"
)

type DisabledType int

const (
	DISABLED_TYPE_AUTO           DisabledType = 0 // 冻结后允许充值自动解冻
	DISABLED_TYPE_MANUAL         DisabledType = 1 // 冻结后需要手动解冻
	DISABLED_TYPE_PARENT         DisabledType = 2 // 被父账号冻结
	DISABLED_TYPE_OVERDUE        DisabledType = 3 // 实时计费远超余额冻结
	DISABLED_TYPE_NONSTD_OVERDUE DisabledType = 4 // 未认证用户超过免费额度冻结
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
